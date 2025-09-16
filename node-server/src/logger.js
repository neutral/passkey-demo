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

export function buildServerStartEvent(config) {
  return {
    event: 'server_start',
    rp_id: config.RP_ID,
    origin: config.ORIGIN,
    port: config.PORT,
    db_path: config.DB_PATH,
  }
}

export function logServerStart(config) {
  logger.info(buildServerStartEvent(config))
}

export function logTxOptionsSuccess(payload) {
  logger.info({ event: 'tx_options', ...payload })
}

export function logTxOptionsError(payload, err, message = 'tx options failed') {
  if (err) {
    logger.error({ event: 'tx_options_error', ...payload, err }, message)
  } else {
    logger.error({ event: 'tx_options_error', ...payload }, message)
  }
}

export function logTxFinishSuccess(payload) {
  logger.info({ event: 'tx_finish', ...payload })
}

export function logTxFinishError(payload, err, message = 'tx finish failed') {
  if (err) {
    logger.error({ event: 'tx_finish_error', ...payload, err }, message)
  } else {
    logger.error({ event: 'tx_finish_error', ...payload }, message)
  }
}

export default logger
