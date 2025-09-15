import test from 'node:test'
import assert from 'node:assert/strict'
import express from 'express'
import Database from 'better-sqlite3'

import { createRegistrationRoutes, RegistrationSessionStore, REGISTRATION_SESSION_TTL_SECONDS } from '../src/webauthn/reg.js'
import { applyMigrations } from '../src/db.js'
import { fileURLToPath } from 'node:url'

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
  const routes = createRegistrationRoutes(CONFIG, { ...deps, db })
  app.use('/authn/passkey/registration', routes.router)
  const server = app.listen(0)
  return { server, routes, db }
}

function postOptions(port) {
  return fetch(`http://127.0.0.1:${port}/authn/passkey/registration/options`, {
    method: 'POST',
    signal: AbortSignal.timeout(FETCH_TIMEOUT_MS),
  })
}

function closeServer(server) {
  return new Promise((resolve) => server.close(resolve))
}

test('registration options returns policy-compliant JSON and stores session', { timeout: TEST_TIMEOUT_MS }, async () => {
  const fixedNow = 1_700_000_000
  const idValue = 'ABCDEFGHIJKLMNOPQRSTUVWX'
  const deps = {
    store: new RegistrationSessionStore(),
    now: () => fixedNow,
    idFactory: () => idValue,
    generateRegistrationOptions: () => ({
      challenge: 'mock-challenge',
      rp: { id: CONFIG.RP_ID, name: 'ignored' },
      user: { id: 'user', name: 'n', displayName: 'd' },
      pubKeyCredParams: [{ type: 'public-key', alg: -7 }],
      attestation: 'none',
      authenticatorSelection: { residentKey: 'required', requireResidentKey: true, userVerification: 'required' },
      timeout: 60000,
    }),
  }
  const { server, routes, db } = buildApp({ deps })
  try {
    const { port } = server.address()

    const res = await postOptions(port)
    assert.equal(res.status, 200)
    assert.match(res.headers.get('content-type') || '', /^application\/json/)
    const body = await res.json()
    assert.equal(body.reg_session_id, idValue)
    assert.equal(body.expires_at, fixedNow + REGISTRATION_SESSION_TTL_SECONDS)
    assert.equal(body.attestation, 'none')
    assert.equal(body.authenticatorSelection.residentKey, 'required')
    assert.equal(body.authenticatorSelection.userVerification, 'required')
    assert.equal(body.pubKeyCredParams[0].alg, -7)
    assert.ok(body.challenge)

    const stored = routes.store.get(idValue)
    assert.deepEqual(stored, {
      challenge: 'mock-challenge',
      rpID: CONFIG.RP_ID,
      origin: CONFIG.ORIGIN,
      expiresAt: fixedNow + REGISTRATION_SESSION_TTL_SECONDS,
    })
  } finally {
    await closeServer(server)
    routes.store?.pruneExpired(Number.MAX_SAFE_INTEGER)
    db.close()
  }
})

test('pruneExpired removes sessions at or before now', () => {
  const store = new RegistrationSessionStore()
  store.create('old', { challenge: 'c', rpID: CONFIG.RP_ID, origin: CONFIG.ORIGIN, expiresAt: 10 })
  store.create('fresh', { challenge: 'c2', rpID: CONFIG.RP_ID, origin: CONFIG.ORIGIN, expiresAt: 20 })

  store.pruneExpired(10)

  assert.equal(store.get('old'), null)
  assert.deepEqual(store.get('fresh'), { challenge: 'c2', rpID: CONFIG.RP_ID, origin: CONFIG.ORIGIN, expiresAt: 20 })
})

test('handler returns internal_error envelope when option generation fails', { timeout: TEST_TIMEOUT_MS }, async () => {
  const deps = {
    store: new RegistrationSessionStore(),
    now: () => 5,
    idFactory: () => 'ABCDEFGHIJKLMNOPQRSTUVWX',
    generateRegistrationOptions: () => {
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

test('handler regenerates session id when collision occurs', { timeout: TEST_TIMEOUT_MS }, async () => {
  const ids = ['DUPLICATESESSION0000000000', 'DUPLICATESESSION0000000000', 'UNIQUESESSION000000000000']
  let idx = 0
  const store = new RegistrationSessionStore()
  store.create('DUPLICATESESSION0000000000', { challenge: 'old', rpID: CONFIG.RP_ID, origin: CONFIG.ORIGIN, expiresAt: 999 })
  const deps = {
    store,
    now: () => 100,
    idFactory: () => ids[idx++],
    generateRegistrationOptions: () => ({
      challenge: 'new-challenge',
      rp: { id: CONFIG.RP_ID, name: 'ignored' },
      user: { id: 'user', name: 'n', displayName: 'd' },
      pubKeyCredParams: [{ type: 'public-key', alg: -7 }],
      attestation: 'none',
      authenticatorSelection: { residentKey: 'required', requireResidentKey: true, userVerification: 'required' },
      timeout: 60000,
    }),
  }
  const { server, routes, db } = buildApp({ deps })
  try {
    const { port } = server.address()

    const res = await postOptions(port)
    assert.equal(res.status, 200)
    const body = await res.json()
    assert.equal(body.reg_session_id, 'UNIQUESESSION000000000000')

    const stored = routes.store.get('UNIQUESESSION000000000000')
    assert.equal(stored.challenge, 'new-challenge')
  } finally {
    await closeServer(server)
    db.close()
  }
})

test('handler fails after exhausting session id attempts', { timeout: TEST_TIMEOUT_MS }, async () => {
  const store = {
    has: () => true,
    pruneExpired: () => {},
    create: () => {
      throw new Error('should not create when collisions persist')
    },
  }
  const deps = {
    store,
    now: () => 1,
    idFactory: () => 'duplicate',
    generateRegistrationOptions: () => ({
      challenge: 'new',
      rp: { id: CONFIG.RP_ID, name: 'ignored' },
      user: { id: 'user', name: 'n', displayName: 'd' },
      pubKeyCredParams: [{ type: 'public-key', alg: -7 }],
      attestation: 'none',
      authenticatorSelection: { residentKey: 'required', requireResidentKey: true, userVerification: 'required' },
      timeout: 60000,
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
