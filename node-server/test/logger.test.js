import test from 'node:test'
import assert from 'node:assert/strict'

import { logServerStart, logger } from '../src/logger.js'

test('logServerStart emits server_start payload', async () => {
  const cfg = { RP_ID: 'localhost', ORIGIN: 'http://localhost:5173', PORT: 8080, DB_PATH: 'server/demo-node.db' }
  const original = logger.info
  const calls = []
  logger.info = (...args) => {
    calls.push(args)
  }
  try {
    logServerStart(cfg)
    assert.equal(calls.length, 1)
    const [payload] = calls[0]
    assert.equal(payload.event, 'server_start')
    assert.equal(payload.rp_id, 'localhost')
    assert.equal(payload.origin, 'http://localhost:5173')
    assert.equal(payload.port, 8080)
    assert.equal(payload.db_path, 'server/demo-node.db')
  } finally {
    logger.info = original
  }
})
