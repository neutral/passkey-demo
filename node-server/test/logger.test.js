import test from 'node:test'
import assert from 'node:assert/strict'
import { buildServerStartEvent } from '../src/logger.js'

test('buildServerStartEvent has expected shape', () => {
  const cfg = { RP_ID: 'localhost', ORIGIN: 'http://localhost:5173', PORT: 8080, DB_PATH: 'server/demo-node.db' }
  const ev = buildServerStartEvent(cfg)
  assert.equal(ev.event, 'server_start')
  assert.equal(ev.rp_id, 'localhost')
  assert.equal(ev.origin, 'http://localhost:5173')
  assert.equal(ev.port, 8080)
  assert.equal(ev.db_path, 'server/demo-node.db')
})

