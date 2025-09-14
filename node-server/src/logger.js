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

export default logger

