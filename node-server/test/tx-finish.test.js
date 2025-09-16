import test from 'node:test'
import assert from 'node:assert/strict'
import express from 'express'
import Database from 'better-sqlite3'
import { fileURLToPath } from 'node:url'
import fs from 'node:fs'

import { createTxFinishRoutes } from '../src/tx/finish.js'
import { createTxSessionStore } from '../src/tx/options.js'
import { applyMigrations } from '../src/db.js'

const TEST_TIMEOUT_MS = 5000
const FETCH_TIMEOUT_MS = 3000

const CONFIG = {
  RP_ID: 'localhost',
  ORIGIN: 'http://localhost:5173',
  RP_ID_ALLOWLIST: [],
  ORIGIN_ALLOWLIST: [],
}

const MIGRATIONS_PATH = fileURLToPath(new URL('../src/migrations.sql', import.meta.url))
const GOLDEN_PATH = fileURLToPath(new URL('../../specs/goldens/tx-bundle-v1.json', import.meta.url))
const GOLDEN = JSON.parse(fs.readFileSync(GOLDEN_PATH, 'utf8'))
const ACCOUNT_CBOR = Buffer.from(GOLDEN.inputs.sender_key_cbor_b64, 'base64url')
const CREDENTIAL_ID = Buffer.from('credential-tx-00000001', 'utf8')
const AUTH_DATA = Buffer.from('auth-data-sample')
const CLIENT_DATA = Buffer.from('client-data-sample')
const SIGNATURE = Buffer.from('signature-sample')

function seedAccountAndCredential(db, { signCount = 10 } = {}) {
  db.prepare('INSERT INTO accounts (acct_cbor, acct_thumb, created_at) VALUES (?, ?, ?)')
    .run(ACCOUNT_CBOR, Buffer.alloc(32), 0)
  db.prepare(
    'INSERT INTO credentials (credential_id, acct_cbor_fk, sign_count, aaguid, created_at) VALUES (?, ?, ?, ?, ?)',
  ).run(CREDENTIAL_ID, ACCOUNT_CBOR, signCount, Buffer.alloc(16), 0)
}

function makeSessionStore(overrides = {}) {
  const store = createTxSessionStore()
  const sessionId = overrides.id || 'tx-session-id'
  store.set(sessionId, {
    canonical: Buffer.from(GOLDEN.bundle.bundle_cbor_b64, 'base64url'),
    challenge: Buffer.from(GOLDEN.anchors.challenge_b64, 'base64url'),
    txId: Buffer.from(GOLDEN.anchors.tx_id_hex, 'hex'),
    acctCbor: ACCOUNT_CBOR,
    credentialIds: [CREDENTIAL_ID],
    expiresAt: overrides.expiresAt ?? 1_700_000_300,
  })
  return { store, sessionId }
}

function buildApp(options = {}) {
  const app = express()
  const db = options.db ?? (() => {
    const database = new Database(':memory:')
    applyMigrations(database, MIGRATIONS_PATH)
    seedAccountAndCredential(database, options.seedOptions)
    return database
  })()
  const { store, sessionId } = options.storeSetup ?? makeSessionStore()
  app.use(express.json())
  app.use((req, res, next) => {
    req.id = options.correlationId || 'test-correlation-id'
    if (options.attachSession !== false) {
      req.session = { acct_cbor: ACCOUNT_CBOR }
    }
    next()
  })
  const deps = {
    db,
    store,
    verifyAuthenticationResponse: options.verifyStub,
    now: options.now ?? (() => 1_700_000_000),
  }
  const routes = createTxFinishRoutes(CONFIG, deps)
  app.use('/tx', routes.router)
  const server = app.listen(0)
  return { server, routes, db, txSessionId: sessionId }
}

function finishPayload(sessionId) {
  return {
    tx_session_id: sessionId,
    id: CREDENTIAL_ID.toString('base64url'),
    rawId: CREDENTIAL_ID.toString('base64url'),
    type: 'public-key',
    response: {
      authenticatorData: AUTH_DATA.toString('base64url'),
      clientDataJSON: CLIENT_DATA.toString('base64url'),
      signature: SIGNATURE.toString('base64url'),
      userHandle: '',
    },
  }
}

function postFinish(port, body) {
  return fetch(`http://127.0.0.1:${port}/tx/signing/finish`, {
    method: 'POST',
    headers: { 'content-type': 'application/json' },
    body: JSON.stringify(body),
    signal: AbortSignal.timeout(FETCH_TIMEOUT_MS),
  })
}

function closeServer(server) {
  return new Promise((resolve) => server.close(resolve))
}

const VERIFIED_RESPONSE = {
  verified: true,
  authenticationInfo: {
    newCounter: 11,
    userVerified: true,
  },
}

test('tx signing finish stores transaction and updates counter', { timeout: TEST_TIMEOUT_MS }, async () => {
  const verifyStub = () => VERIFIED_RESPONSE
  const { server, routes, db, txSessionId } = buildApp({ verifyStub })
  try {
    const { port } = server.address()
    const res = await postFinish(port, finishPayload(txSessionId))
    assert.equal(res.status, 201)
    const body = await res.json()
    assert.equal(body.tx_id_hex, GOLDEN.anchors.tx_id_hex)

    // session deleted
    assert.equal(routes.store.get(txSessionId), null)

    const credRow = db
      .prepare('SELECT sign_count FROM credentials WHERE credential_id = ?')
      .get(CREDENTIAL_ID)
    assert.equal(credRow.sign_count, 11)

    const txRow = db
      .prepare('SELECT nonce, message, bundle_cbor, auth_data, client_data, signature FROM transactions WHERE tx_id = ?')
      .get(Buffer.from(GOLDEN.anchors.tx_id_hex, 'hex'))
    assert.equal(txRow.nonce, GOLDEN.inputs.nonce)
    assert.equal(txRow.message, GOLDEN.inputs.message)
    assert.equal(Buffer.from(txRow.bundle_cbor).toString('base64url'), GOLDEN.bundle.bundle_cbor_b64)
    assert.equal(Buffer.from(txRow.auth_data).toString(), AUTH_DATA.toString())
    assert.equal(Buffer.from(txRow.client_data).toString(), CLIENT_DATA.toString())
    assert.equal(Buffer.from(txRow.signature).toString(), SIGNATURE.toString())
  } finally {
    await closeServer(server)
    db.close()
  }
})

test('missing login session returns 401', { timeout: TEST_TIMEOUT_MS }, async () => {
  const { server, db, txSessionId } = buildApp({ attachSession: false, verifyStub: () => VERIFIED_RESPONSE })
  try {
    const { port } = server.address()
    const res = await postFinish(port, finishPayload(txSessionId))
    assert.equal(res.status, 401)
  } finally {
    await closeServer(server)
    db.close()
  }
})

test('missing tx session returns 401', { timeout: TEST_TIMEOUT_MS }, async () => {
  const storeSetup = { store: createTxSessionStore(), sessionId: 'unknown' }
  const { server, db } = buildApp({ verifyStub: () => VERIFIED_RESPONSE, storeSetup })
  try {
    const { port } = server.address()
    const res = await postFinish(port, finishPayload('unknown'))
    assert.equal(res.status, 401)
  } finally {
    await closeServer(server)
    db.close()
  }
})

test('credential not in session allowlist returns 401', { timeout: TEST_TIMEOUT_MS }, async () => {
  const storeSetup = makeSessionStore({})
  storeSetup.store.set(storeSetup.sessionId, { ...storeSetup.store.get(storeSetup.sessionId), credentialIds: [Buffer.from('other')] })
  const { server, db, txSessionId } = buildApp({ verifyStub: () => VERIFIED_RESPONSE, storeSetup })
  try {
    const { port } = server.address()
    const res = await postFinish(port, finishPayload(txSessionId))
    assert.equal(res.status, 401)
  } finally {
    await closeServer(server)
    db.close()
  }
})

test('verification throws -> 401 and session removed', { timeout: TEST_TIMEOUT_MS }, async () => {
  const verifyStub = () => {
    throw new Error('mismatch')
  }
  const { server, db, routes, txSessionId } = buildApp({ verifyStub })
  try {
    const { port } = server.address()
    const res = await postFinish(port, finishPayload(txSessionId))
    assert.equal(res.status, 401)
    assert.equal(routes.store.get(txSessionId), null)
  } finally {
    await closeServer(server)
    db.close()
  }
})

test('user verification missing -> 403', { timeout: TEST_TIMEOUT_MS }, async () => {
  const verifyStub = () => ({ verified: true, authenticationInfo: { userVerified: false, newCounter: 12 } })
  const { server, db, txSessionId } = buildApp({ verifyStub })
  try {
    const { port } = server.address()
    const res = await postFinish(port, finishPayload(txSessionId))
    assert.equal(res.status, 403)
  } finally {
    await closeServer(server)
    db.close()
  }
})

test('sign count regression -> 409', { timeout: TEST_TIMEOUT_MS }, async () => {
  const verifyStub = () => ({ verified: true, authenticationInfo: { userVerified: true, newCounter: 5 } })
  const { server, db, txSessionId } = buildApp({ verifyStub, seedOptions: { signCount: 10 } })
  try {
    const { port } = server.address()
    const res = await postFinish(port, finishPayload(txSessionId))
    assert.equal(res.status, 409)
  } finally {
    await closeServer(server)
    db.close()
  }
})

test('duplicate transaction inserts return 409', { timeout: TEST_TIMEOUT_MS }, async () => {
  const db = new Database(':memory:')
  applyMigrations(db, MIGRATIONS_PATH)
  seedAccountAndCredential(db, { signCount: 10 })
  db.prepare(
    'INSERT INTO transactions (tx_id, acct_cbor, nonce, message, bundle_cbor, auth_data, client_data, signature, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)',
  ).run(
    Buffer.from(GOLDEN.anchors.tx_id_hex, 'hex'),
    ACCOUNT_CBOR,
    GOLDEN.inputs.nonce,
    GOLDEN.inputs.message,
    Buffer.from(GOLDEN.bundle.bundle_cbor_b64, 'base64url'),
    AUTH_DATA,
    CLIENT_DATA,
    SIGNATURE,
    0,
  )
  const verifyStub = () => VERIFIED_RESPONSE
  const storeSetup = makeSessionStore()
  const { server } = buildApp({ verifyStub, storeSetup, db })
  try {
    const { port } = server.address()
    const res = await postFinish(port, finishPayload(storeSetup.sessionId))
    assert.equal(res.status, 409)
  } finally {
    await closeServer(server)
    db.close()
  }
})

test('invalid payload yields 400', { timeout: TEST_TIMEOUT_MS }, async () => {
  const { server, db } = buildApp({ verifyStub: () => VERIFIED_RESPONSE })
  try {
    const { port } = server.address()
    const res = await postFinish(port, { bogus: true })
    assert.equal(res.status, 400)
  } finally {
    await closeServer(server)
    db.close()
  }
})
