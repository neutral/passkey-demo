import test from 'node:test'
import assert from 'node:assert/strict'
import express from 'express'
import Database from 'better-sqlite3'
import { Encoder } from 'cbor-x'
import { createHash } from 'node:crypto'
import { fileURLToPath } from 'node:url'

import { createLoginRoutes, LoginSessionStore, SESSION_COOKIE_TTL_SECONDS } from '../src/webauthn/login.js'
import { applyMigrations } from '../src/db.js'
import { logger } from '../src/logger.js'

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

function computeThumb(buf) {
  return createHash('sha256').update('ACCTK1').update(buf).digest()
}

function hashIdentifier(buf) {
  return createHash('sha256').update(buf).digest('hex')
}

function captureLogs(method, run) {
  const original = logger[method]
  const calls = []
  logger[method] = (...args) => {
    calls.push(args)
  }
  return run(calls).finally(() => {
    logger[method] = original
  })
}

function setupApp({
  store = new LoginSessionStore(),
  now,
  verifyAuthenticationResponse,
  sessionIdFactory,
  correlationId = 'test-correlation-id',
} = {}) {
  const db = new Database(':memory:')
  applyMigrations(db, MIGRATIONS_PATH)
  const app = express()
  app.use(express.json())
  app.use((req, res, next) => {
    req.id = correlationId
    next()
  })
  const routes = createLoginRoutes({ ...CONFIG }, { store, now, verifyAuthenticationResponse, sessionIdFactory, db })
  app.use('/authn/passkey/login', routes.router)
  const server = app.listen(0)
  return { server, store: routes.store, db }
}

function postFinish(port, body) {
  return fetch(`http://127.0.0.1:${port}/authn/passkey/login/finish`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    credentials: 'include',
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

test('login finish verifies assertion, updates counters, and sets session cookie', { timeout: TEST_TIMEOUT_MS }, async () => {
  const nowSeconds = 1_700_000_000
  const store = new LoginSessionStore()
  const loginSessionId = 'session-success'
  createStoreEntry(store, loginSessionId, { challenge: 'challenge-login-123', expiresAt: nowSeconds + 60 })
  const credentialId = Buffer.from('credential-login', 'utf8')
  const accountCbor = canonicalTestKey()
  const accountThumb = computeThumb(accountCbor)
  const sessionId = 'cookie-session-id'

  const verifyStub = async (opts) => {
    assert.equal(opts.expectedChallenge, 'challenge-login-123')
    assert.equal(opts.requireUserVerification, true)
    assert(opts.authenticator)
    assert.equal(opts.authenticator.credentialID.toString(), credentialId.toString())
    return {
      verified: true,
      authenticationInfo: {
        credentialID: credentialId.toString('base64url'),
        newCounter: 10,
        userVerified: true,
        credentialDeviceType: 'singleDevice',
        credentialBackedUp: false,
        origin: CONFIG.ORIGIN,
        rpID: CONFIG.RP_ID,
      },
    }
  }

  await captureLogs('info', async (infoCalls) => {
    const { server, store: routeStore, db } = setupApp({
      store,
      now: () => nowSeconds,
      verifyAuthenticationResponse: verifyStub,
      sessionIdFactory: () => sessionId,
    })
    try {
      db.prepare('INSERT INTO accounts (acct_cbor, acct_thumb, created_at) VALUES (?, ?, ?)').run(
        accountCbor,
        accountThumb,
        nowSeconds,
      )
      db.prepare('INSERT INTO credentials (credential_id, acct_cbor_fk, sign_count, aaguid, created_at) VALUES (?, ?, ?, ?, ?)').run(
        credentialId,
        accountCbor,
        5,
        Buffer.alloc(16, 0),
        nowSeconds,
      )

      const { port } = server.address()
      const payload = {
        login_session_id: loginSessionId,
        id: credentialId.toString('base64url'),
        rawId: credentialId.toString('base64url'),
        type: 'public-key',
        response: {
          authenticatorData: Buffer.from('authdata').toString('base64url'),
          clientDataJSON: Buffer.from('clientdata').toString('base64url'),
          signature: Buffer.from('signature').toString('base64url'),
          userHandle: '',
        },
      }
      const res = await postFinish(port, payload)
      assert.equal(res.status, 200)
      const body = await res.json()
      assert.equal(body.account_thumb_hex, accountThumb.toString('hex'))
      assert.equal(body.credential_id_b64, credentialId.toString('base64url'))

      const setCookie = res.headers.get('set-cookie')
      assert.ok(setCookie && setCookie.includes('sid=cookie-session-id'))
      assert.ok(setCookie.includes('HttpOnly'))
      assert.ok(setCookie.includes('SameSite=Lax'))
      assert.ok(setCookie.includes('Path=/'))
      assert.ok(setCookie.includes('Max-Age='))
      assert.ok(!setCookie.includes('Secure'))

      const credentialRow = db
        .prepare('SELECT sign_count FROM credentials WHERE credential_id = ?')
        .get(credentialId)
      assert.equal(credentialRow.sign_count, 10)

      const sessionRow = db.prepare('SELECT acct_cbor, expires_at FROM sessions WHERE session_id = ?').get(sessionId)
      assert.ok(sessionRow)
      assert.equal(Buffer.compare(sessionRow.acct_cbor, accountCbor), 0)
      assert.equal(sessionRow.expires_at, nowSeconds + SESSION_COOKIE_TTL_SECONDS)
      assert.equal(routeStore.get(loginSessionId), null)

      const loginFinishLog = infoCalls.find(([p]) => p?.event === 'login_finish')
      assert.ok(loginFinishLog)
      const [logPayload] = loginFinishLog
      assert.equal(logPayload.account_thumb_hex, accountThumb.toString('hex'))
      assert.equal(logPayload.credential_id_hash, hashIdentifier(credentialId))
    } finally {
      await closeServer(server)
      db.close()
    }
  })
})

test('expired login session returns 401 and removes entry', { timeout: TEST_TIMEOUT_MS }, async () => {
  const nowSeconds = 10
  const store = new LoginSessionStore()
  const loginSessionId = 'expired-login'
  createStoreEntry(store, loginSessionId, { challenge: 'challenge', expiresAt: nowSeconds - 1 })
  const { server, db } = setupApp({
    store,
    now: () => nowSeconds,
    verifyAuthenticationResponse: async () => ({ verified: true, authenticationInfo: { newCounter: 1, credentialID: 'abc', userVerified: true, credentialDeviceType: 'singleDevice', credentialBackedUp: false, origin: CONFIG.ORIGIN, rpID: CONFIG.RP_ID } }),
  })
  try {
    const { port } = server.address()
    const res = await postFinish(port, {
      login_session_id: loginSessionId,
      id: 'abc',
      rawId: 'abc',
      type: 'public-key',
      response: { authenticatorData: 'AA', clientDataJSON: 'AA', signature: 'AA', userHandle: '' },
    })
    assert.equal(res.status, 401)
    const body = await res.json()
    assert.equal(body.code, 'ERR_UNAUTHORIZED')
    assert.equal(store.get(loginSessionId), null)
  } finally {
    await closeServer(server)
    db.close()
  }
})

test('credential not found returns unauthorized', { timeout: TEST_TIMEOUT_MS }, async () => {
  const nowSeconds = 20
  const store = new LoginSessionStore()
  const loginSessionId = 'missing-cred'
  createStoreEntry(store, loginSessionId, { challenge: 'challenge', expiresAt: nowSeconds + 10 })
  const { server, db } = setupApp({
    store,
    now: () => nowSeconds,
    verifyAuthenticationResponse: async () => ({ verified: true, authenticationInfo: { newCounter: 1, credentialID: 'abc', userVerified: true, credentialDeviceType: 'singleDevice', credentialBackedUp: false, origin: CONFIG.ORIGIN, rpID: CONFIG.RP_ID } }),
  })
  try {
    const { port } = server.address()
    const res = await postFinish(port, {
      login_session_id: loginSessionId,
      id: 'abc',
      rawId: 'abc',
      type: 'public-key',
      response: { authenticatorData: 'AA', clientDataJSON: 'AA', signature: 'AA', userHandle: '' },
    })
    assert.equal(res.status, 401)
    const body = await res.json()
    assert.equal(body.code, 'ERR_UNAUTHORIZED')
    assert.equal(store.get(loginSessionId), null)
  } finally {
    await closeServer(server)
    db.close()
  }
})

test('missing user verification returns 403', { timeout: TEST_TIMEOUT_MS }, async () => {
  const nowSeconds = 30
  const store = new LoginSessionStore()
  const loginSessionId = 'no-uv'
  createStoreEntry(store, loginSessionId, { challenge: 'challenge', expiresAt: nowSeconds + 10 })
  const accountCbor = canonicalTestKey()
  const accountThumb = computeThumb(accountCbor)
  const credentialId = Buffer.from('cred', 'utf8')
  const { server, db } = setupApp({
    store,
    now: () => nowSeconds,
    verifyAuthenticationResponse: async () => ({
      verified: true,
      authenticationInfo: {
        credentialID: credentialId.toString('base64url'),
        newCounter: 1,
        userVerified: false,
        credentialDeviceType: 'singleDevice',
        credentialBackedUp: false,
        origin: CONFIG.ORIGIN,
        rpID: CONFIG.RP_ID,
      },
    }),
  })
  try {
    db.prepare('INSERT INTO accounts (acct_cbor, acct_thumb, created_at) VALUES (?, ?, ?)').run(
      accountCbor,
      accountThumb,
      nowSeconds,
    )
    db.prepare('INSERT INTO credentials (credential_id, acct_cbor_fk, sign_count, aaguid, created_at) VALUES (?, ?, ?, ?, ?)').run(
      credentialId,
      accountCbor,
      1,
      Buffer.alloc(16, 0),
      nowSeconds,
    )
    const { port } = server.address()
    const res = await postFinish(port, {
      login_session_id: loginSessionId,
      id: credentialId.toString('base64url'),
      rawId: credentialId.toString('base64url'),
      type: 'public-key',
      response: { authenticatorData: 'AA', clientDataJSON: 'AA', signature: 'AA', userHandle: '' },
    })
    assert.equal(res.status, 403)
    const body = await res.json()
    assert.equal(body.code, 'ERR_FORBIDDEN')
    assert.equal(store.get(loginSessionId), null)
  } finally {
    await closeServer(server)
    db.close()
  }
})

test('sign count regression returns conflict', { timeout: TEST_TIMEOUT_MS }, async () => {
  const nowSeconds = 40
  const store = new LoginSessionStore()
  const loginSessionId = 'signcount'
  createStoreEntry(store, loginSessionId, { challenge: 'challenge', expiresAt: nowSeconds + 10 })
  const accountCbor = canonicalTestKey()
  const accountThumb = computeThumb(accountCbor)
  const credentialId = Buffer.from('cred', 'utf8')
  const { server, db } = setupApp({
    store,
    now: () => nowSeconds,
    verifyAuthenticationResponse: async () => ({
      verified: true,
      authenticationInfo: {
        credentialID: credentialId.toString('base64url'),
        newCounter: 5,
        userVerified: true,
        credentialDeviceType: 'singleDevice',
        credentialBackedUp: false,
        origin: CONFIG.ORIGIN,
        rpID: CONFIG.RP_ID,
      },
    }),
  })
  try {
    db.prepare('INSERT INTO accounts (acct_cbor, acct_thumb, created_at) VALUES (?, ?, ?)').run(
      accountCbor,
      accountThumb,
      nowSeconds,
    )
    db.prepare('INSERT INTO credentials (credential_id, acct_cbor_fk, sign_count, aaguid, created_at) VALUES (?, ?, ?, ?, ?)').run(
      credentialId,
      accountCbor,
      10,
      Buffer.alloc(16, 0),
      nowSeconds,
    )
    const { port } = server.address()
    const res = await postFinish(port, {
      login_session_id: loginSessionId,
      id: credentialId.toString('base64url'),
      rawId: credentialId.toString('base64url'),
      type: 'public-key',
      response: { authenticatorData: 'AA', clientDataJSON: 'AA', signature: 'AA', userHandle: '' },
    })
    assert.equal(res.status, 409)
    const body = await res.json()
    assert.equal(body.code, 'ERR_CONFLICT')
    assert.equal(store.get(loginSessionId), null)
    const credentialRow = db
      .prepare('SELECT sign_count FROM credentials WHERE credential_id = ?')
      .get(credentialId)
    assert.equal(credentialRow.sign_count, 10)
  } finally {
    await closeServer(server)
    db.close()
  }
})

test('invalid payload returns 400 without invoking verifier', { timeout: TEST_TIMEOUT_MS }, async () => {
  const nowSeconds = 50
  const store = new LoginSessionStore()
  const loginSessionId = 'bad-body'
  createStoreEntry(store, loginSessionId, { challenge: 'challenge', expiresAt: nowSeconds + 10 })
  let verifyCalled = false
  const { server, db } = setupApp({
    store,
    now: () => nowSeconds,
    verifyAuthenticationResponse: async () => {
      verifyCalled = true
      return { verified: true, authenticationInfo: { credentialID: 'abc', newCounter: 1, userVerified: true, credentialDeviceType: 'singleDevice', credentialBackedUp: false, origin: CONFIG.ORIGIN, rpID: CONFIG.RP_ID } }
    },
  })
  try {
    const { port } = server.address()
    const res = await postFinish(port, {
      login_session_id: loginSessionId,
      response: null,
    })
    assert.equal(res.status, 400)
    const body = await res.json()
    assert.equal(body.code, 'ERR_BAD_REQUEST')
    assert.equal(store.get(loginSessionId), null)
    assert.equal(verifyCalled, false)
  } finally {
    await closeServer(server)
    db.close()
  }
})
