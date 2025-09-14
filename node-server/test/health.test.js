import test from 'node:test'
import assert from 'node:assert/strict'
import { createApp } from '../src/server.js'
import { loadConfig } from '../src/config.js'
import openDB from '../src/db.js'

test('GET /health returns 200 and status ok', async () => {
  const cfg = loadConfig({ PORT: 0, DB_PATH: ':memory:', RP_ID: 'localhost', ORIGIN: 'http://localhost:5173' })
  const db = openDB(cfg.DB_PATH)
  const app = createApp(cfg, db)
  const server = app.listen(0)
  const { port } = server.address()
  const res = await fetch(`http://127.0.0.1:${port}/health`)
  assert.equal(res.status, 200)
  const body = await res.json()
  assert.equal(body.status, 'ok')
  server.close()
})
