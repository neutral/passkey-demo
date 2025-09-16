import test from 'node:test'
import assert from 'node:assert/strict'
import express from 'express'
import Database from 'better-sqlite3'
import { fileURLToPath } from 'node:url'

import { createTxListRoutes } from '../src/tx/list.js'
import { applyMigrations } from '../src/db.js'

const TEST_TIMEOUT_MS = 5000
const FETCH_TIMEOUT_MS = 3000

const CONFIG = {
  RP_ID: 'localhost',
  ORIGIN: 'http://localhost:5173',
}

const MIGRATIONS_PATH = fileURLToPath(new URL('../src/migrations.sql', import.meta.url))

const ACCT_CBOR = Buffer.from('account-cbor-0001')

function seedAccount(db) {
  db.prepare('INSERT INTO accounts (acct_cbor, acct_thumb, created_at) VALUES (?, ?, ?)')
    .run(ACCT_CBOR, Buffer.alloc(32), 0)
  db.prepare(
    'INSERT INTO sessions (session_id, acct_cbor, expires_at, created_at) VALUES (?, ?, ?, ?)',
  ).run('sid-123', ACCT_CBOR, 1_700_000_500, 1_700_000_000)
}

function seedTransactions(db, rows = []) {
  const stmt = db.prepare(
    'INSERT INTO transactions (tx_id, acct_cbor, nonce, message, bundle_cbor, auth_data, client_data, signature, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)',
  )
  for (const row of rows) {
    stmt.run(
      row.txId,
      ACCT_CBOR,
      row.nonce,
      row.message,
      row.bundle ?? Buffer.alloc(0),
      row.authData ?? Buffer.alloc(0),
      row.clientData ?? Buffer.alloc(0),
      row.signature ?? Buffer.alloc(0),
      row.createdAt,
    )
  }
}

function buildApp(options = {}) {
  const app = express()
  const db = options.db ?? (() => {
    const database = new Database(':memory:')
    applyMigrations(database, MIGRATIONS_PATH)
    seedAccount(database)
    if (options.seedTransactions) seedTransactions(database, options.seedTransactions)
    return database
  })()

  app.use(express.json())
  app.use((req, res, next) => {
    req.id = options.correlationId || 'test-correlation-id'
    if (options.attachSession !== false) {
      const acctValue = options.sessionValue ?? ACCT_CBOR
      req.session = { acct_cbor: acctValue, sid: 'sid-123' }
    }
    next()
  })

  const deps = {
    db,
    selectTransactions: options.selectTransactions,
  }

  const routes = createTxListRoutes(CONFIG, deps)
  app.use('/tx', routes.router)
  const server = app.listen(0)
  return { server, db }
}

function closeServer(server) {
  return new Promise((resolve) => server.close(resolve))
}

function fetchList(port) {
  return fetch(`http://127.0.0.1:${port}/tx/list`, {
    method: 'GET',
    headers: { 'content-type': 'application/json' },
    signal: AbortSignal.timeout(FETCH_TIMEOUT_MS),
  })
}

test('returns transactions sorted by created_at desc', { timeout: TEST_TIMEOUT_MS }, async () => {
  const txRows = [
    { txId: Buffer.from('0000000000000001', 'hex'), nonce: 1, message: 'older', createdAt: 1_700_000_100 },
    { txId: Buffer.from('0000000000000002', 'hex'), nonce: 2, message: 'newer', createdAt: 1_700_000_200 },
  ]
  const { server, db } = buildApp({ seedTransactions: txRows })
  try {
    const { port } = server.address()
    const res = await fetchList(port)
    assert.equal(res.status, 200)
    const body = await res.json()
    assert.deepEqual(body.items, [
      {
        tx_id_hex: '0000000000000002',
        nonce: 2,
        message: 'newer',
        created_at: 1_700_000_200,
      },
      {
        tx_id_hex: '0000000000000001',
        nonce: 1,
        message: 'older',
        created_at: 1_700_000_100,
      },
    ])
  } finally {
    await closeServer(server)
    db.close()
  }
})

test('returns empty list when no transactions exist', { timeout: TEST_TIMEOUT_MS }, async () => {
  const { server, db } = buildApp()
  try {
    const { port } = server.address()
    const res = await fetchList(port)
    assert.equal(res.status, 200)
    const body = await res.json()
    assert.deepEqual(body.items, [])
  } finally {
    await closeServer(server)
    db.close()
  }
})

test('handles session acct_cbor as Uint8Array', { timeout: TEST_TIMEOUT_MS }, async () => {
  const txRows = [{ txId: Buffer.from('01', 'hex'), nonce: 9, message: 'value', createdAt: 1_700_000_300 }]
  const { server, db } = buildApp({ seedTransactions: txRows, sessionValue: new Uint8Array(ACCT_CBOR) })
  try {
    const { port } = server.address()
    const res = await fetchList(port)
    assert.equal(res.status, 200)
    const body = await res.json()
    assert.equal(body.items.length, 1)
    assert.equal(body.items[0].nonce, 9)
  } finally {
    await closeServer(server)
    db.close()
  }
})

test('returns 401 when session missing', { timeout: TEST_TIMEOUT_MS }, async () => {
  const { server, db } = buildApp({ attachSession: false })
  try {
    const { port } = server.address()
    const res = await fetchList(port)
    assert.equal(res.status, 401)
    const body = await res.json()
    assert.equal(body.code, 'unauthorized')
  } finally {
    await closeServer(server)
    db.close()
  }
})

test('database failure returns 500', { timeout: TEST_TIMEOUT_MS }, async () => {
  const failingSelect = { all() { throw new Error('boom') } }
  const { server, db } = buildApp({ selectTransactions: failingSelect })
  try {
    const { port } = server.address()
    const res = await fetchList(port)
    assert.equal(res.status, 500)
    const body = await res.json()
    assert.equal(body.code, 'internal_error')
  } finally {
    await closeServer(server)
    db.close()
  }
})
