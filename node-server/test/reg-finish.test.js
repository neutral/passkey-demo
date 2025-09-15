import test from 'node:test'
import assert from 'node:assert/strict'
import express from 'express'
import Database from 'better-sqlite3'
import { Encoder } from 'cbor-x'
import { createHash } from 'node:crypto'
import { fileURLToPath } from 'node:url'

import { createRegistrationRoutes, RegistrationSessionStore } from '../src/webauthn/reg.js'
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
const encoder = new Encoder({ canonical: true, structuredClone: false, useRecords: false, mapsAsObjects: false })

function canonicalTestKey() {
  const map = new Map([
    [1, 2],
    [3, -7],
    [-1, 1],
    [-2, Uint8Array.from([0x01, 0x02, 0x03, 0x04])],
    [-3, Uint8Array.from([0x05, 0x06, 0x07, 0x08])],
  ])
  return Buffer.from(encoder.encode(map))
}

function setupApp({ store = new RegistrationSessionStore(), now, verifyRegistrationResponse, correlationId = 'test-correlation-id' } = {}) {
  const db = new Database(':memory:')
  applyMigrations(db, MIGRATIONS_PATH)
  const app = express()
  app.use(express.json())
  app.use((req, res, next) => {
    req.id = correlationId
    next()
  })
  const routes = createRegistrationRoutes({ ...CONFIG }, { store, now, verifyRegistrationResponse, db })
  app.use('/authn/passkey/registration', routes.router)
  const server = app.listen(0)
  return { server, store: routes.store, db }
}

function postFinish(port, body) {
  return fetch(`http://127.0.0.1:${port}/authn/passkey/registration/finish`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
    signal: AbortSignal.timeout(FETCH_TIMEOUT_MS),
  })
}

function closeServer(server) {
  return new Promise((resolve) => server.close(resolve))
}

function createStoreEntry(store, id, { challenge, expiresAt, rpID = CONFIG.RP_ID, origin = CONFIG.ORIGIN }) {
  store.create(id, { challenge, expiresAt, rpID, origin })
}

function computeThumb(buf) {
  return createHash('sha256').update('ACCTK1').update(buf).digest()
}

test('registration finish persists account and credential', { timeout: TEST_TIMEOUT_MS }, async () => {
  const nowSeconds = 1_700_000_000
  const store = new RegistrationSessionStore()
  const regSessionId = 'session-success'
  createStoreEntry(store, regSessionId, { challenge: 'challenge-123', expiresAt: nowSeconds + 60 })
  const coseKey = canonicalTestKey()
  const credentialId = Buffer.from('credential-id', 'utf8')
  const verifyStub = async (opts) => {
    assert.equal(opts.expectedChallenge, 'challenge-123')
    assert.equal(opts.requireUserVerification, true)
    assert.deepEqual(opts.supportedAlgorithmIDs, [-7])
    return {
      verified: true,
      registrationInfo: {
        fmt: 'none',
        counter: 7,
        aaguid: '00000000-0000-0000-0000-000000000000',
        credentialID: credentialId.toString('base64url'),
        credentialPublicKey: new Uint8Array(coseKey),
        credentialType: 'public-key',
        attestationObject: new Uint8Array(),
        userVerified: true,
        credentialDeviceType: 'singleDevice',
        credentialBackedUp: false,
        origin: CONFIG.ORIGIN,
      },
    }
  }
  const { server, store: routeStore, db } = setupApp({ store, now: () => nowSeconds, verifyRegistrationResponse: verifyStub })
  try {
    const { port } = server.address()
    const payload = {
      reg_session_id: regSessionId,
      id: credentialId.toString('base64url'),
      rawId: credentialId.toString('base64url'),
      type: 'public-key',
      response: {
        attestationObject: Buffer.from('attestation').toString('base64url'),
        clientDataJSON: Buffer.from('client-data').toString('base64url'),
      },
    }
    const res = await postFinish(port, payload)
    assert.equal(res.status, 201)
    const body = await res.json()
    const expectedThumb = computeThumb(coseKey).toString('hex')
    assert.equal(body.account_thumb_hex, expectedThumb)
    assert.equal(body.credential_id_b64, credentialId.toString('base64url'))

    const accountRow = db.prepare('SELECT acct_cbor, acct_thumb FROM accounts').get()
    assert.ok(accountRow)
    assert.equal(Buffer.compare(accountRow.acct_cbor, coseKey), 0)
    assert.equal(Buffer.compare(accountRow.acct_thumb, Buffer.from(body.account_thumb_hex, 'hex')), 0)
    const credentialRow = db
      .prepare('SELECT credential_id, acct_cbor_fk, sign_count, aaguid FROM credentials')
      .get()
    assert.ok(credentialRow)
    assert.equal(Buffer.compare(credentialRow.credential_id, credentialId), 0)
    assert.equal(Buffer.compare(credentialRow.acct_cbor_fk, coseKey), 0)
    assert.equal(credentialRow.sign_count, 7)
    assert.equal(credentialRow.aaguid.length, 16)
    assert.equal(routeStore.get(regSessionId), null)
  } finally {
    await closeServer(server)
    db.close()
  }
})

test('expired registration session returns 401 and removes entry', { timeout: TEST_TIMEOUT_MS }, async () => {
  const nowSeconds = 2
  const store = new RegistrationSessionStore()
  const regSessionId = 'expired-session'
  createStoreEntry(store, regSessionId, { challenge: 'whatever', expiresAt: nowSeconds - 1 })
  let verifyCalled = false
  const verifyStub = async () => {
    verifyCalled = true
    return { verified: true }
  }
  const { server, store: routeStore, db } = setupApp({ store, now: () => nowSeconds, verifyRegistrationResponse: verifyStub })
  try {
    const { port } = server.address()
    const res = await postFinish(port, {
      reg_session_id: regSessionId,
      id: 'abc',
      rawId: 'abc',
      type: 'public-key',
      response: { attestationObject: 'AA', clientDataJSON: 'AA' },
    })
    assert.equal(res.status, 401)
    const body = await res.json()
    assert.equal(body.code, 'unauthorized')
    assert.equal(routeStore.get(regSessionId), null)
    assert.equal(verifyCalled, false)
  } finally {
    await closeServer(server)
    db.close()
  }
})

test('verification failure yields unauthorized', { timeout: TEST_TIMEOUT_MS }, async () => {
  const nowSeconds = 10
  const store = new RegistrationSessionStore()
  const regSessionId = 'verify-fail'
  createStoreEntry(store, regSessionId, { challenge: 'challenge', expiresAt: nowSeconds + 10 })
  const verifyStub = async () => ({ verified: false })
  const { server, db } = setupApp({ store, now: () => nowSeconds, verifyRegistrationResponse: verifyStub })
  try {
    const { port } = server.address()
    const res = await postFinish(port, {
      reg_session_id: regSessionId,
      id: 'abc',
      rawId: 'abc',
      type: 'public-key',
      response: { attestationObject: 'AA', clientDataJSON: 'AA' },
    })
    assert.equal(res.status, 401)
    const body = await res.json()
    assert.equal(body.code, 'unauthorized')
    assert.equal(store.get(regSessionId), null)
  } finally {
    await closeServer(server)
    db.close()
  }
})

test('missing user verification returns 403', { timeout: TEST_TIMEOUT_MS }, async () => {
  const nowSeconds = 10
  const store = new RegistrationSessionStore()
  const regSessionId = 'no-uv'
  createStoreEntry(store, regSessionId, { challenge: 'challenge', expiresAt: nowSeconds + 10 })
  const verifyStub = async () => ({
    verified: true,
    registrationInfo: {
      fmt: 'none',
      counter: 0,
      aaguid: '00000000-0000-0000-0000-000000000000',
      credentialID: 'abc',
      credentialPublicKey: new Uint8Array(canonicalTestKey()),
      credentialType: 'public-key',
      attestationObject: new Uint8Array(),
      userVerified: false,
      credentialDeviceType: 'singleDevice',
      credentialBackedUp: false,
      origin: CONFIG.ORIGIN,
    },
  })
  const { server, db } = setupApp({ store, now: () => nowSeconds, verifyRegistrationResponse: verifyStub })
  try {
    const { port } = server.address()
    const res = await postFinish(port, {
      reg_session_id: regSessionId,
      id: 'abc',
      rawId: 'abc',
      type: 'public-key',
      response: { attestationObject: 'AA', clientDataJSON: 'AA' },
    })
    assert.equal(res.status, 403)
    const body = await res.json()
    assert.equal(body.code, 'policy_violation')
    assert.equal(store.get(regSessionId), null)
  } finally {
    await closeServer(server)
    db.close()
  }
})

test('duplicate credential returns conflict and consumes session', { timeout: TEST_TIMEOUT_MS }, async () => {
  const nowSeconds = 20
  const store = new RegistrationSessionStore()
  const regSessionId = 'dup'
  createStoreEntry(store, regSessionId, { challenge: 'challenge', expiresAt: nowSeconds + 10 })
  const coseKey = canonicalTestKey()
  const credentialId = Buffer.from('credential-id', 'utf8')
  const verifyStub = async () => ({
    verified: true,
    registrationInfo: {
      fmt: 'none',
      counter: 1,
      aaguid: '00000000-0000-0000-0000-000000000000',
      credentialID: credentialId.toString('base64url'),
      credentialPublicKey: new Uint8Array(coseKey),
      credentialType: 'public-key',
      attestationObject: new Uint8Array(),
      userVerified: true,
      credentialDeviceType: 'singleDevice',
      credentialBackedUp: false,
      origin: CONFIG.ORIGIN,
    },
  })
  const { server, store: routeStore, db } = setupApp({ store, now: () => nowSeconds, verifyRegistrationResponse: verifyStub })
  try {
    const thumb = computeThumb(coseKey)
    db.prepare('INSERT OR IGNORE INTO accounts (acct_cbor, acct_thumb, created_at) VALUES (?, ?, ?)').run(
      coseKey,
      thumb,
      nowSeconds,
    )
    db.prepare('INSERT OR IGNORE INTO credentials (credential_id, acct_cbor_fk, sign_count, aaguid, created_at) VALUES (?, ?, ?, ?, ?)').run(
      credentialId,
      coseKey,
      1,
      Buffer.alloc(16, 0),
      nowSeconds,
    )

    const { port } = server.address()
    const res = await postFinish(port, {
      reg_session_id: regSessionId,
      id: credentialId.toString('base64url'),
      rawId: credentialId.toString('base64url'),
      type: 'public-key',
      response: { attestationObject: 'AA', clientDataJSON: 'AA' },
    })
    assert.equal(res.status, 409)
    const body = await res.json()
    assert.equal(body.code, 'conflict')
    assert.equal(routeStore.get(regSessionId), null)
  } finally {
    await closeServer(server)
    db.close()
  }
})

test('invalid request body returns 400', { timeout: TEST_TIMEOUT_MS }, async () => {
  const nowSeconds = 30
  const store = new RegistrationSessionStore()
  const regSessionId = 'bad-body'
  createStoreEntry(store, regSessionId, { challenge: 'challenge', expiresAt: nowSeconds + 10 })
  let verifyCalled = false
  const { server, db } = setupApp({ store, now: () => nowSeconds, verifyRegistrationResponse: async () => {
    verifyCalled = true
    return { verified: true }
  } })
  try {
    const { port } = server.address()
    const res = await postFinish(port, {
      reg_session_id: regSessionId,
      response: null,
    })
    assert.equal(res.status, 400)
    const body = await res.json()
    assert.equal(body.code, 'bad_request')
    assert.equal(verifyCalled, false)
  } finally {
    await closeServer(server)
    db.close()
  }
})
