import test from 'node:test'
import assert from 'node:assert/strict'
import express from 'express'
import Database from 'better-sqlite3'
import { fileURLToPath } from 'node:url'
import fs from 'node:fs'
import { createHash } from 'node:crypto'

import { createTxOptionsRoutes, TxSessionStore } from '../src/tx/options.js'
import { applyMigrations } from '../src/db.js'

const TEST_TIMEOUT_MS = 5000
const FETCH_TIMEOUT_MS = 3000

const CONFIG = {
  RP_ID: 'localhost',
  ORIGIN: 'http://localhost:5173',
}

const MIGRATIONS_PATH = fileURLToPath(new URL('../src/migrations.sql', import.meta.url))
const GOLDEN_PATH = fileURLToPath(new URL('../../specs/goldens/tx-bundle-v1.json', import.meta.url))
const GOLDEN = JSON.parse(fs.readFileSync(GOLDEN_PATH, 'utf8'))
const ACCOUNT_CBOR = Buffer.from(GOLDEN.inputs.sender_key_cbor_b64, 'base64url')
const ACCOUNT_THUMB = createHash('sha256').update('ACCTK1').update(ACCOUNT_CBOR).digest()
const CREDENTIAL_ID = Buffer.from('credential-tx-00000001', 'utf8')

function seedAccount(db, { includeCredential = true } = {}) {
  db.prepare('INSERT INTO accounts (acct_cbor, acct_thumb, created_at) VALUES (?, ?, ?)')
    .run(ACCOUNT_CBOR, ACCOUNT_THUMB, 0)
  if (includeCredential) {
    db.prepare(
      'INSERT INTO credentials (credential_id, acct_cbor_fk, sign_count, aaguid, created_at) VALUES (?, ?, ?, ?, ?)',
    ).run(CREDENTIAL_ID, ACCOUNT_CBOR, 0, Buffer.alloc(16), 0)
  }
}

function openDbWithSeed(options = {}) {
  const db = new Database(':memory:')
  applyMigrations(db, MIGRATIONS_PATH)
  seedAccount(db, options)
  return db
}

function buildApp(options = {}) {
  const app = express()
  const db = options.db ?? openDbWithSeed()
  app.use(express.json())
  app.use((req, res, next) => {
    req.id = options.correlationId || 'test-correlation-id'
    if (options.attachSession !== false) {
      req.session = { acct_cbor: ACCOUNT_CBOR }
    }
    next()
  })
  const deps = { ...(options.deps || {}), db }
  const routes = createTxOptionsRoutes(CONFIG, deps)
  app.use('/tx', routes.router)
  const server = app.listen(0)
  return { app, server, db, routes }
}

function postOptions(port, body) {
  return fetch(`http://127.0.0.1:${port}/tx/signing/options`, {
    method: 'POST',
    headers: { 'content-type': 'application/json' },
    body: JSON.stringify(body),
    signal: AbortSignal.timeout(FETCH_TIMEOUT_MS),
  })
}

function closeServer(server) {
  return new Promise((resolve) => server.close(resolve))
}

function base64url(buffer) {
  return buffer.toString('base64url')
}

test('tx signing options returns session and options', { timeout: TEST_TIMEOUT_MS }, async () => {
  const nowSeconds = 1_700_000_000
  const txSessionId = 'TXSESSIONIDABCDEFGHIJKLMNOPQRSTUV'
  const store = new TxSessionStore()
  const { server, db, routes } = buildApp({
    deps: {
      store,
      now: () => nowSeconds,
      idFactory: () => txSessionId,
    },
  })
  try {
    const { port } = server.address()
    const res = await postOptions(port, { bundle_cbor_b64: GOLDEN.bundle.bundle_cbor_b64 })
    assert.equal(res.status, 200)
    const body = await res.json()
    assert.equal(body.tx_session_id, txSessionId)
    assert.equal(body.challenge, GOLDEN.anchors.challenge_b64)
    assert.equal(body.tx_id_hex, GOLDEN.anchors.tx_id_hex)
    assert.equal(body.expires_at, nowSeconds + 300)
    assert.ok(body.options)
    assert.equal(body.options.challenge, GOLDEN.anchors.challenge_b64)
    assert.equal(body.options.rpId, CONFIG.RP_ID)
    assert.equal(body.options.origin, CONFIG.ORIGIN)
    assert.equal(body.options.userVerification, 'required')
    assert.equal(body.options.timeout, 60000)
    assert.deepEqual(body.options.allowCredentials, [{ type: 'public-key', id: base64url(CREDENTIAL_ID) }])

    const stored = routes.store.get(txSessionId)
    assert.ok(stored)
    assert.equal(stored.expiresAt, nowSeconds + 300)
    assert.equal(base64url(stored.challenge), GOLDEN.anchors.challenge_b64)
    assert.equal(base64url(stored.canonical), GOLDEN.bundle.bundle_cbor_b64)
    assert.equal(stored.txId.toString('hex'), GOLDEN.anchors.tx_id_hex)
    assert.deepEqual(stored.credentialIds.map(base64url), [base64url(CREDENTIAL_ID)])
  } finally {
    await closeServer(server)
    db.close()
  }
})

test('returns 401 when session missing', { timeout: TEST_TIMEOUT_MS }, async () => {
  const { server, db } = buildApp({ attachSession: false })
  try {
    const { port } = server.address()
    const res = await postOptions(port, { bundle_cbor_b64: GOLDEN.bundle.bundle_cbor_b64 })
    assert.equal(res.status, 401)
    const body = await res.json()
    assert.equal(body.code, 'unauthorized')
  } finally {
    await closeServer(server)
    db.close()
  }
})

test('invalid payload yields 400', { timeout: TEST_TIMEOUT_MS }, async () => {
  const { server, db } = buildApp()
  try {
    const { port } = server.address()
    const res = await postOptions(port, { bogus: true })
    assert.equal(res.status, 400)
    const body = await res.json()
    assert.equal(body.code, 'bad_request')
  } finally {
    await closeServer(server)
    db.close()
  }
})

test('invalid bundle encoding maps to 400', { timeout: TEST_TIMEOUT_MS }, async () => {
  const { server, db } = buildApp()
  try {
    const { port } = server.address()
    const res = await postOptions(port, { bundle_cbor_b64: '@@not-base64@@' })
    assert.equal(res.status, 400)
    const body = await res.json()
    assert.equal(body.code, 'bad_request')
  } finally {
    await closeServer(server)
    db.close()
  }
})

test('nonce conflict returns 409', { timeout: TEST_TIMEOUT_MS }, async () => {
  const db = openDbWithSeed()
  db.prepare(
    'INSERT INTO transactions (tx_id, acct_cbor, nonce, message, bundle_cbor, auth_data, client_data, signature, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)',
  ).run(Buffer.from('existing'), ACCOUNT_CBOR, 1, 'x', Buffer.from('B'), Buffer.from('AD'), Buffer.from('CD'), Buffer.from('SIG'), 0)
  const { server } = buildApp({ db })
  try {
    const { port } = server.address()
    const res = await postOptions(port, { bundle_cbor_b64: GOLDEN.bundle.bundle_cbor_b64 })
    assert.equal(res.status, 409)
    const body = await res.json()
    assert.equal(body.code, 'conflict')
  } finally {
    await closeServer(server)
    db.close()
  }
})

test('missing credentials returns 409', { timeout: TEST_TIMEOUT_MS }, async () => {
  const db = openDbWithSeed({ includeCredential: false })
  const { server } = buildApp({ db })
  try {
    const { port } = server.address()
    const res = await postOptions(port, { bundle_cbor_b64: GOLDEN.bundle.bundle_cbor_b64 })
    assert.equal(res.status, 409)
    const body = await res.json()
    assert.equal(body.code, 'conflict')
  } finally {
    await closeServer(server)
    db.close()
  }
})

test('session id collision exhaustion yields 500', { timeout: TEST_TIMEOUT_MS }, async () => {
  const store = new TxSessionStore()
  const collisionId = 'collision-id-aaaaaaaaaaaaaaaaaaaa'
  store.set(collisionId, { expiresAt: Number.MAX_SAFE_INTEGER })
  const { server, db } = buildApp({
    deps: {
      store,
      idFactory: () => collisionId,
    },
  })
  try {
    const { port } = server.address()
    const res = await postOptions(port, { bundle_cbor_b64: GOLDEN.bundle.bundle_cbor_b64 })
    assert.equal(res.status, 500)
    const body = await res.json()
    assert.equal(body.code, 'internal_error')
  } finally {
    await closeServer(server)
    db.close()
  }
})

test('pruneExpired clears stale sessions before issuing new one', { timeout: TEST_TIMEOUT_MS }, async () => {
  const store = new TxSessionStore()
  store.set('expired-id', { expiresAt: 10 })
  const nowSeconds = 1000
  const txSessionId = 'fresh-session-id-abcdefghijklmnop'
  const { server, db, routes } = buildApp({
    deps: {
      store,
      now: () => nowSeconds,
      idFactory: () => txSessionId,
    },
  })
  try {
    const { port } = server.address()
    const res = await postOptions(port, { bundle_cbor_b64: GOLDEN.bundle.bundle_cbor_b64 })
    assert.equal(res.status, 200)
    assert.equal(routes.store.get('expired-id'), null)
  } finally {
    await closeServer(server)
    db.close()
  }
})

