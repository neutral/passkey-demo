import test from 'node:test'
import assert from 'node:assert/strict'
import { loadConfig } from '../src/config.js'

test('loadConfig applies defaults and parses lists', () => {
  const cfg = loadConfig({ PORT: undefined, DB_PATH: undefined, RP_ID_ALLOWLIST: 'example.com, sub.example', ORIGIN_ALLOWLIST: '' })
  assert.equal(cfg.PORT, 8080)
  assert.equal(cfg.DB_PATH, 'server/demo-node.db')
  assert.deepEqual(cfg.RP_ID_ALLOWLIST, ['example.com', 'sub.example'])
  assert.deepEqual(cfg.ORIGIN_ALLOWLIST, [])
})

test('loadConfig rejects invalid PORT', () => {
  assert.throws(() => loadConfig({ PORT: 'abc' }))
})

