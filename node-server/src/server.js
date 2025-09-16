import 'dotenv/config'
import express from 'express'
import { loadConfig } from './config.js'
import { logger, httpLogger, logServerStart } from './logger.js'
import requestId from './reqid.js'
import openDB, { applyMigrations } from './db.js'
import corsMiddleware from './cors.js'
import sessionMiddleware from './session.js'
import path from 'node:path'
import { createRegistrationRoutes } from './webauthn/reg.js'
import { createLoginRoutes } from './webauthn/login.js'
import { createMeRoutes } from './me.js'
import { createTxOptionsRoutes } from './tx/options.js'
import { createTxFinishRoutes } from './tx/finish.js'
import { createTxListRoutes } from './tx/list.js'
import { buildBodyLimit, buildRateLimiter, DEFAULT_LIMITS } from './limits.js'
import { buildExpressErrorHandler } from './error.js'

export function createApp(config, db) {
  const app = express()
  app.disable('x-powered-by')
  app.use(requestId)
  app.use(httpLogger)
  app.use(corsMiddleware(config))
  app.use(sessionMiddleware(db))

  const registration = createRegistrationRoutes(config, { db })
  const login = createLoginRoutes(config, { db })
  const txOptions = createTxOptionsRoutes(config, { db })
  const txFinish = createTxFinishRoutes(config, { db, store: txOptions.store })
  const txList = createTxListRoutes(config, { db })

  const authnLimiter = buildRateLimiter(DEFAULT_LIMITS.AUTHN_RATE)
  const authnBody = buildBodyLimit({ bytes: DEFAULT_LIMITS.AUTHN_BODY_LIMIT_BYTES })
  const authnGroup = express.Router()
  authnGroup.use(authnLimiter)
  authnGroup.use(authnBody)
  authnGroup.use('/passkey/registration', registration.router)
  authnGroup.use('/passkey/login', login.router)
  app.locals.registration = registration
  app.locals.login = login
  app.use('/authn', authnGroup)

  const txLimiter = buildRateLimiter(DEFAULT_LIMITS.TX_RATE)
  const txBody = buildBodyLimit({ bytes: DEFAULT_LIMITS.TX_BODY_LIMIT_BYTES })
  const txGroup = express.Router()
  txGroup.use(txLimiter)
  txGroup.use(txBody)
  txGroup.use(txOptions.router)
  txGroup.use(txFinish.router)
  txGroup.use(txList.router)
  app.locals.txOptions = txOptions
  app.locals.txFinish = txFinish
  app.locals.txList = txList
  app.locals.tx = { options: txOptions, finish: txFinish, list: txList }
  app.use('/tx', txGroup)

  const me = createMeRoutes(config, { db })
  app.locals.me = me
  app.use('/me', me.router)

  app.get('/health', (req, res) => {
    res.setHeader('Content-Type', 'application/json')
    res.end(JSON.stringify({ status: 'ok' }))
  })

  // Placeholder: future routers mounted here
  app.use(buildExpressErrorHandler())

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
