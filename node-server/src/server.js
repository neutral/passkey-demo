import 'dotenv/config'
import express from 'express'
import { loadConfig } from './config.js'
import { logger, httpLogger, logServerStart } from './logger.js'
import requestId from './reqid.js'
import openDB from './db.js'

export function createApp(config, db) {
  const app = express()
  app.disable('x-powered-by')
  app.use(requestId)
  app.use(httpLogger)
  app.use(express.json({ limit: '1mb' }))

  app.get('/health', (req, res) => {
    res.setHeader('Content-Type', 'application/json')
    res.end(JSON.stringify({ status: 'ok' }))
  })

  // Placeholder: future routers mounted here
  return app
}

export function start() {
  const config = loadConfig(process.env)
  const db = openDB(config.DB_PATH)
  const app = createApp(config, db)
  const server = app.listen(config.PORT, () => {
    logServerStart(config)
    logger.info({ event: 'listening', address: server.address() })
  })
  return { app, server, db }
}

// Run only when executed directly
if (import.meta.url === `file://${process.argv[1]}`) {
  start()
}

