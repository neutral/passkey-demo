import test from 'node:test'
import assert from 'node:assert/strict'
import { createApp } from '../src/server.js'
import { loadConfig } from '../src/config.js'
import openDB from '../src/db.js'

test('preflight allowed returns 204 with headers', async () => {
  const cfg = loadConfig({ ORIGIN: 'http://localhost:5173', ORIGIN_ALLOWLIST: '' })
  const db = openDB(':memory:')
  const app = createApp(cfg, db)
  const server = app.listen(0)
  const { port } = server.address()
  const res = await fetch(`http://127.0.0.1:${port}/authn/passkey/login/options`, {
    method: 'OPTIONS',
    headers: {
      Origin: 'http://localhost:5173',
      'Access-Control-Request-Method': 'POST',
      'Access-Control-Request-Headers': 'Content-Type',
    },
  })
  assert.equal(res.status, 204)
  assert.equal(res.headers.get('access-control-allow-origin'), 'http://localhost:5173')
  assert.equal(res.headers.get('access-control-allow-credentials'), 'true')
  assert.match(res.headers.get('vary') || '', /Origin/i)
  server.close()
})

test('preflight disallowed returns 403 without CORS headers', async () => {
  const cfg = loadConfig({ ORIGIN: 'http://localhost:5173', ORIGIN_ALLOWLIST: '' })
  const db = openDB(':memory:')
  const app = createApp(cfg, db)
  const server = app.listen(0)
  const { port } = server.address()
  const res = await fetch(`http://127.0.0.1:${port}/authn/passkey/login/options`, {
    method: 'OPTIONS',
    headers: {
      Origin: 'https://evil.example',
      'Access-Control-Request-Method': 'POST',
      'Access-Control-Request-Headers': 'Content-Type',
    },
  })
  assert.equal(res.status, 403)
  assert.equal(res.headers.get('access-control-allow-origin'), null)
  server.close()
})

test('actual allowed includes ACAO and ACAC', async () => {
  const cfg = loadConfig({ ORIGIN: 'http://localhost:5173', ORIGIN_ALLOWLIST: '' })
  const db = openDB(':memory:')
  const app = createApp(cfg, db)
  const server = app.listen(0)
  const { port } = server.address()
  const res = await fetch(`http://127.0.0.1:${port}/health`, {
    method: 'GET',
    headers: { Origin: 'http://localhost:5173' },
  })
  assert.equal(res.status, 200)
  assert.equal(res.headers.get('access-control-allow-origin'), 'http://localhost:5173')
  assert.equal(res.headers.get('access-control-allow-credentials'), 'true')
  server.close()
})

