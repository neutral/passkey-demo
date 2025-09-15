import test from 'node:test'
import assert from 'node:assert/strict'
import express from 'express'
import Database from 'better-sqlite3'
import { fileURLToPath } from 'node:url'

import { createLoginRoutes, LoginSessionStore, LOGIN_SESSION_TTL_SECONDS } from '../src/webauthn/login.js'
import { applyMigrations } from '../src/db.js'

const TEST_TIMEOUT_MS = 3000
const FETCH_TIMEOUT_MS = 2500

const CONFIG = {
  RP_ID: 'localhost',
  ORIGIN: 'http://localhost:5173',
  RP_ID_ALLOWLIST: [],
  ORIGIN_ALLOWLIST: [],
}
const MIGRATIONS_PATH = fileURLToPath(new URL('../src/migrations.sql', import.meta.url))

function buildApp(options = {}) {
  const app = express()
  const db = new Database(':memory:')
  applyMigrations(db, MIGRATIONS_PATH)
  app.use(express.json())
  app.use((req, res, next) => {
    req.id = options.correlationId || 'test-correlation-id'
    next()
  })
  const deps = options.deps || {}
  const routes = createLoginRoutes(CONFIG, { ...deps, db })
  app.use('/authn/passkey/login', routes.router)
  const server = app.listen(0)
  return { server, routes, db }
}

function postOptions(port) {
  return fetch(`http://127.0.0.1:${port}/authn/passkey/login/options`, {
    method: 'POST',
    signal: AbortSignal.timeout(FETCH_TIMEOUT_MS),
  })
}

function closeServer(server) {
  return new Promise((resolve) => server.close(resolve))
}

test('login options returns policy-compliant JSON and stores session', { timeout: TEST_TIMEOUT_MS }, async () => {
  const nowSeconds = 1_700_000_000
  const sessionId = 'ABCDEFGHIJKLMNOPQRSTUVWX'
  const deps = {
    store: new LoginSessionStore(),
    now: () => nowSeconds,
    idFactory: () => sessionId,
    generateAuthenticationOptions: () => ({
      challenge: 'mock-challenge',
      rpId: CONFIG.RP_ID,
      timeout: 60000,
      userVerification: 'required',
      allowCredentials: [],
    }),
  }
  const { server, routes, db } = buildApp({ deps })
  try {
    const { port } = server.address()

    const res = await postOptions(port)
    assert.equal(res.status, 200)
    assert.match(res.headers.get('content-type') || '', /^application\/json/)
    const body = await res.json()
    assert.equal(body.login_session_id, sessionId)
    assert.equal(body.expires_at, nowSeconds + LOGIN_SESSION_TTL_SECONDS)
    assert.equal(body.challenge, 'mock-challenge')
    assert.equal(body.userVerification, 'required')
    assert.equal(body.rpId, CONFIG.RP_ID)

    const stored = routes.store.get(sessionId)
    assert.deepEqual(stored, {
      challenge: 'mock-challenge',
      rpID: CONFIG.RP_ID,
      origin: CONFIG.ORIGIN,
      expiresAt: nowSeconds + LOGIN_SESSION_TTL_SECONDS,
    })
  } finally {
    await closeServer(server)
    routes.store?.pruneExpired(Number.MAX_SAFE_INTEGER)
    db.close()
  }
})

test('pruneExpired removes stale sessions', () => {
  const store = new LoginSessionStore()
  store.create('old', { challenge: 'c', rpID: CONFIG.RP_ID, origin: CONFIG.ORIGIN, expiresAt: 10 })
  store.create('fresh', { challenge: 'c2', rpID: CONFIG.RP_ID, origin: CONFIG.ORIGIN, expiresAt: 20 })

  store.pruneExpired(10)

  assert.equal(store.get('old'), null)
  assert.deepEqual(store.get('fresh'), {
    challenge: 'c2',
    rpID: CONFIG.RP_ID,
    origin: CONFIG.ORIGIN,
    expiresAt: 20,
  })
})

test('generator failure yields internal_error envelope', { timeout: TEST_TIMEOUT_MS }, async () => {
  const deps = {
    store: new LoginSessionStore(),
    now: () => 5,
    idFactory: () => 'ABCDEFGHIJKLMNOPQRSTUVWX',
    generateAuthenticationOptions: () => {
      throw new Error('boom')
    },
  }
  const { server, db } = buildApp({ deps, correlationId: 'corr-id' })
  try {
    const { port } = server.address()
    const res = await postOptions(port)
    assert.equal(res.status, 500)
    const body = await res.json()
    assert.equal(body.code, 'internal_error')
    assert.equal(body.correlation_id, 'corr-id')
  } finally {
    await closeServer(server)
    db.close()
  }
})

test('handler retries session id on collision', { timeout: TEST_TIMEOUT_MS }, async () => {
  const ids = ['duplicate-id', 'duplicate-id', 'unique-id-1234567890']
  let idx = 0
  const store = new LoginSessionStore()
  store.create('duplicate-id', { challenge: 'old', rpID: CONFIG.RP_ID, origin: CONFIG.ORIGIN, expiresAt: 999 })
  const deps = {
    store,
    now: () => 100,
    idFactory: () => ids[idx++],
    generateAuthenticationOptions: () => ({
      challenge: 'new-challenge',
      rpId: CONFIG.RP_ID,
      timeout: 60000,
      userVerification: 'required',
      allowCredentials: [],
    }),
  }
  const { server, routes, db } = buildApp({ deps })
  try {
    const { port } = server.address()
    const res = await postOptions(port)
    assert.equal(res.status, 200)
    const body = await res.json()
    assert.equal(body.login_session_id, 'unique-id-1234567890')
    const stored = routes.store.get('unique-id-1234567890')
    assert.equal(stored.challenge, 'new-challenge')
  } finally {
    await closeServer(server)
    db.close()
  }
})

test('session id exhaustion returns 500', { timeout: TEST_TIMEOUT_MS }, async () => {
  const store = new LoginSessionStore()
  store.create('duplicate', { challenge: 'c', rpID: CONFIG.RP_ID, origin: CONFIG.ORIGIN, expiresAt: 999 })
  const deps = {
    store,
    now: () => 1,
    idFactory: () => 'duplicate',
    generateAuthenticationOptions: () => ({
      challenge: 'new',
      rpId: CONFIG.RP_ID,
      timeout: 60000,
      userVerification: 'required',
      allowCredentials: [],
    }),
  }
  const { server, db } = buildApp({ deps })
  try {
    const { port } = server.address()
    const res = await postOptions(port)
    assert.equal(res.status, 500)
    const body = await res.json()
    assert.equal(body.code, 'internal_error')
  } finally {
    await closeServer(server)
    db.close()
  }
})
