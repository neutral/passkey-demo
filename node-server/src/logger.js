import pino from 'pino'
import pinoHttp from 'pino-http'

export const logger = pino({ level: process.env.LOG_LEVEL || 'info' })

export const httpLogger = pinoHttp({
  logger,
  serializers: {
    req(req) {
      return { method: req.method, url: req.url, id: req.id }
    },
    res(res) {
      return { statusCode: res.statusCode }
    },
  },
})

function basePayload(context = {}) {
  const payload = {}
  if (context?.correlationId) payload.correlation_id = context.correlationId
  if (context?.rpId) payload.rp_id = context.rpId
  if (context?.origin) payload.origin = context.origin
  return payload
}

function logInfo(event, context, attrs = {}) {
  logger.info({ event, ...basePayload(context), ...attrs })
}

function logError(event, context, attrs = {}, err, message) {
  const payload = { event, ...basePayload(context), ...attrs }
  if (err) {
    logger.error({ ...payload, err }, message)
  } else if (message) {
    logger.error(payload, message)
  } else {
    logger.error(payload)
  }
}

export function logServerStart(config) {
  logInfo('server_start', { rpId: config.RP_ID, origin: config.ORIGIN }, {
    port: config.PORT,
    db_path: config.DB_PATH,
  })
}

export const logRegOptionsSuccess = (context, attrs = {}) => logInfo('reg_options', context, attrs)
export const logRegOptionsError = (context, attrs = {}, err, message = 'registration options failed') =>
  logError('reg_options_error', context, attrs, err, message)

export const logRegFinishSuccess = (context, attrs = {}) => logInfo('reg_finish', context, attrs)
export const logRegFinishError = (context, attrs = {}, err, message = 'registration finish failed') =>
  logError('reg_finish_error', context, attrs, err, message)

export const logLoginOptionsSuccess = (context, attrs = {}) => logInfo('login_options', context, attrs)
export const logLoginOptionsError = (context, attrs = {}, err, message = 'login options failed') =>
  logError('login_options_error', context, attrs, err, message)

export const logLoginFinishSuccess = (context, attrs = {}) => logInfo('login_finish', context, attrs)
export const logLoginFinishError = (context, attrs = {}, err, message = 'login finish failed') =>
  logError('login_finish_error', context, attrs, err, message)

export const logTxOptionsSuccess = (context, attrs = {}) => logInfo('tx_options', context, attrs)
export const logTxOptionsError = (context, attrs = {}, err, message = 'tx options failed') =>
  logError('tx_options_error', context, attrs, err, message)

export const logTxFinishSuccess = (context, attrs = {}) => logInfo('tx_finish', context, attrs)
export const logTxFinishError = (context, attrs = {}, err, message = 'tx finish failed') =>
  logError('tx_finish_error', context, attrs, err, message)

export const logTxListError = (context, attrs = {}, err, message = 'tx list failed') =>
  logError('tx_list_error', context, attrs, err, message)

export const logTxListSuccess = (context, attrs = {}) => logInfo('tx_list', context, attrs)

export const logMeAccountKeySuccess = (context, attrs = {}) => logInfo('me_account_key', context, attrs)
export const logMeAccountKeyError = (context, attrs = {}, err, message = 'me account key failed') =>
  logError('me_account_key_error', context, attrs, err, message)

export const logWebauthnVerifyFailure = (context, attrs = {}, err, message = 'webauthn verification failed') =>
  logError('webauthn_assert_verify', context, attrs, err, message)

export default logger
