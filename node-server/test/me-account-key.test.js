import test from 'node:test'
import assert from 'node:assert/strict'
import express from 'express'
import Database from 'better-sqlite3'
import { Encoder } from 'cbor-x'
import { createHash } from 'node:crypto'
import { fileURLToPath } from 'node:url'

import { createMeRoutes } from '../src/me.js'
import { applyMigrations } from '../src/db.js'

const TEST_TIMEOUT_MS = 3000
const FETCH_TIMEOUT_MS = 2500
const CONFIG = {
  RP_ID: 'localhost',
  ORIGIN: 'http://localhost:5173',
}
const MIGRATIONS_PATH = fileURLToPath(new URL('../src/migrations.sql', import.meta.url))
const encoder = new Encoder({ canonical: true, structuredClone: false, useRecords: false, mapsAsObjects: false })

function canonicalTestKey() {
  const map = new Map([
    [1, 2],
    [3, -7],
    [-1, 1],
    [-2, Uint8Array.from([0x0a, 0x0b, 0x0c, 0x0d])],
    [-3, Uint8Array.from([0x0e, 0x0f, 0x10, 0x11])],
  ])
  return Buffer.from(encoder.encode(map))
}

function computeThumb(buf) {
  return createHash('sha256').update('ACCTK1').update(buf).digest()
}

function startServer({ sessionFactory } = {}) {
  const db = new Database(':memory:')
  applyMigrations(db, MIGRATIONS_PATH)
  const app = express()
  app.use(express.json())
  app.use((req, res, next) => {
    req.id = 'test-correlation-id'
    if (typeof sessionFactory === 'function') {
      const session = sessionFactory()
      if (session !== undefined) req.session = session
    }
    next()
  })
  const routes = createMeRoutes({ ...CONFIG }, { db })
  app.use('/me', routes.router)
  const server = app.listen(0)
  return { server, db }
}

function closeServer(server) {
  return new Promise((resolve) => server.close(resolve))
}

test('GET /me/account_key returns account metadata for authenticated session', { timeout: TEST_TIMEOUT_MS }, async () => {
  const acctCbor = canonicalTestKey()
  const acctThumb = computeThumb(acctCbor)
  const createdAt = 1_700_001_234

  const { server, db } = startServer({
    sessionFactory: () => ({ sid: 'sid-123', acct_cbor: acctCbor, expires_at: createdAt + 300 }),
  })
  try {
    db.prepare('INSERT INTO accounts (acct_cbor, acct_thumb, created_at) VALUES (?, ?, ?)').run(
      acctCbor,
      acctThumb,
      createdAt,
    )

    const { port } = server.address()
    const res = await fetch(`http://127.0.0.1:${port}/me/account_key`, {
      method: 'GET',
      signal: AbortSignal.timeout(FETCH_TIMEOUT_MS),
    })
    assert.equal(res.status, 200)
    const body = await res.json()
    assert.equal(body.acct_cbor_b64, acctCbor.toString('base64url'))
    assert.equal(body.account_thumb_hex, acctThumb.toString('hex'))
    assert.equal(body.created_at, createdAt)
    assert.equal(body.pubkey_x_hex, Buffer.from([0x0a, 0x0b, 0x0c, 0x0d]).toString('hex'))
    assert.equal(body.pubkey_y_hex, Buffer.from([0x0e, 0x0f, 0x10, 0x11]).toString('hex'))
    assert.ok(body.sender_key)
    assert.equal(body.sender_key.kty, 2)
    assert.equal(body.sender_key.alg, -7)
    assert.equal(body.sender_key.crv, 1)
    assert.equal(body.sender_key.x, Buffer.from([0x0a, 0x0b, 0x0c, 0x0d]).toString('base64url'))
    assert.equal(body.sender_key.y, Buffer.from([0x0e, 0x0f, 0x10, 0x11]).toString('base64url'))
  } finally {
    await closeServer(server)
    db.close()
  }
})

test('GET /me/account_key returns 401 when session missing', { timeout: TEST_TIMEOUT_MS }, async () => {
  const { server, db } = startServer({ sessionFactory: () => undefined })
  try {
    const { port } = server.address()
    const res = await fetch(`http://127.0.0.1:${port}/me/account_key`, {
      method: 'GET',
      signal: AbortSignal.timeout(FETCH_TIMEOUT_MS),
    })
    assert.equal(res.status, 401)
    const body = await res.json()
    assert.equal(body.code, 'ERR_UNAUTHORIZED')
  } finally {
    await closeServer(server)
    db.close()
  }
})

test('GET /me/account_key returns 401 when account row missing', { timeout: TEST_TIMEOUT_MS }, async () => {
  const acctCbor = canonicalTestKey()
  const { server, db } = startServer({
    sessionFactory: () => ({ sid: 'sid-404', acct_cbor: acctCbor, expires_at: Date.now() / 1000 + 60 }),
  })
  try {
    const { port } = server.address()
    const res = await fetch(`http://127.0.0.1:${port}/me/account_key`, {
      method: 'GET',
      signal: AbortSignal.timeout(FETCH_TIMEOUT_MS),
    })
    assert.equal(res.status, 401)
    const body = await res.json()
    assert.equal(body.code, 'ERR_UNAUTHORIZED')
  } finally {
    await closeServer(server)
    db.close()
  }
})
