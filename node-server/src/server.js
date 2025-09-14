import 'dotenv/config'
import express from 'express'
import { loadConfig } from './config.js'
import { logger, httpLogger, logServerStart } from './logger.js'
import requestId from './reqid.js'
import openDB, { applyMigrations } from './db.js'
import path from 'node:path'

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
  // Apply schema migrations once at startup
  const sqlPath = path.resolve(path.dirname(new URL(import.meta.url).pathname), './migrations.sql')
  applyMigrations(db, sqlPath)
  try {
    const fk = db.pragma('foreign_keys', { simple: true })
    const jm = db.pragma('journal_mode', { simple: true })
    logger.info({ event: 'db_init', foreign_keys: fk, journal_mode: jm })
  } catch {}
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
