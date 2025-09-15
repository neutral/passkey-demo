import test from 'node:test'
import assert from 'node:assert/strict'
import path from 'node:path'
import openDB, { applyMigrations } from '../src/db.js'
import { createApp } from '../src/server.js'
import { loadConfig } from '../src/config.js'

function withApp(fn) {
  const cfg = loadConfig({ ORIGIN: 'http://localhost:5173' })
  const db = openDB(':memory:')
  applyMigrations(db, path.resolve(path.dirname(new URL(import.meta.url).pathname), '../src/migrations.sql'))
  const app = createApp(cfg, db)
  app.get('/whoami', (req, res) => {
    res.setHeader('Content-Type', 'application/json')
    res.end(JSON.stringify({ hasSession: !!req.session, sid: req.session?.sid || null }))
  })
  const server = app.listen(0)
  const { port } = server.address()
  return fn({ db, server, port }).finally(() => server.close())
}

test('session present when cookie matches unexpired row', async () => {
  await withApp(async ({ db, port }) => {
    const now = Math.floor(Date.now() / 1000)
    // Insert required account row to satisfy FK
    db.prepare('INSERT INTO accounts(acct_cbor, acct_thumb, created_at) VALUES (?, ?, ?)')
      .run(Buffer.from([1]), Buffer.from([2]), now)
    db.prepare('INSERT INTO sessions(session_id, acct_cbor, expires_at, created_at) VALUES (?, ?, ?, ?)')
      .run('abc', Buffer.from([1]), now + 3600, now)
    const res = await fetch(`http://127.0.0.1:${port}/whoami`, { headers: { Cookie: 'sid=abc' } })
    const body = await res.json()
    assert.equal(body.hasSession, true)
    assert.equal(body.sid, 'abc')
  })
})

test('no session when cookie missing or expired/unknown', async () => {
  await withApp(async ({ db, port }) => {
    const now = Math.floor(Date.now() / 1000)
    db.prepare('INSERT INTO accounts(acct_cbor, acct_thumb, created_at) VALUES (?, ?, ?)')
      .run(Buffer.from([1]), Buffer.from([2]), now)
    db.prepare('INSERT INTO sessions(session_id, acct_cbor, expires_at, created_at) VALUES (?, ?, ?, ?)')
      .run('expired', Buffer.from([1]), now - 10, now - 100)
    let res = await fetch(`http://127.0.0.1:${port}/whoami`)
    let body = await res.json()
    assert.equal(body.hasSession, false)
    res = await fetch(`http://127.0.0.1:${port}/whoami`, { headers: { Cookie: 'sid=unknown' } })
    body = await res.json()
    assert.equal(body.hasSession, false)
    res = await fetch(`http://127.0.0.1:${port}/whoami`, { headers: { Cookie: 'sid=expired' } })
    body = await res.json()
    assert.equal(body.hasSession, false)
  })
})
