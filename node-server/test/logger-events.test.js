import test from 'node:test'
import assert from 'node:assert/strict'

import {
  logger,
  logRegOptionsSuccess,
  logRegOptionsError,
  logWebauthnVerifyFailure,
} from '../src/logger.js'

const TEST_TIMEOUT_MS = 3000

function capture(method, fn) {
  const original = logger[method]
  const calls = []
  logger[method] = (...args) => {
    calls.push(args)
  }
  return async () => {
    try {
      await fn(calls)
    } finally {
      logger[method] = original
    }
  }
}

test('logRegOptionsSuccess attaches context and attributes', { timeout: TEST_TIMEOUT_MS }, async (t) => {
  await capture('info', async (calls) => {
    logRegOptionsSuccess(
      { correlationId: 'req-1', rpId: 'example.com', origin: 'https://example.com' },
      { expires_at: 123, session_id_len: 24 },
    )
    assert.equal(calls.length, 1)
    const [payload] = calls[0]
    assert.equal(payload.event, 'reg_options')
    assert.equal(payload.correlation_id, 'req-1')
    assert.equal(payload.rp_id, 'example.com')
    assert.equal(payload.origin, 'https://example.com')
    assert.equal(payload.expires_at, 123)
    assert.equal(payload.session_id_len, 24)
  })()
})

test('logRegOptionsError emits error event with message', { timeout: TEST_TIMEOUT_MS }, async () => {
  await capture('error', async (calls) => {
    const err = new Error('boom')
    logRegOptionsError(
      { correlationId: 'req-2', rpId: 'example.com', origin: 'https://example.com' },
      { reason: 'invalid_payload' },
      err,
      'registration options failed',
    )
    assert.equal(calls.length, 1)
    const [payload, message] = calls[0]
    assert.equal(payload.event, 'reg_options_error')
    assert.equal(payload.reason, 'invalid_payload')
    assert.equal(payload.err, err)
    assert.equal(message, 'registration options failed')
  })()
})

test('logWebauthnVerifyFailure emits verification event', { timeout: TEST_TIMEOUT_MS }, async () => {
  await capture('error', async (calls) => {
    logWebauthnVerifyFailure(
      { correlationId: 'req-3', rpId: 'example.com', origin: 'https://example.com' },
      {
        error_kind: 'uv_required',
        account_thumb_hex: 'aa',
        credential_id_hash: 'bb',
        tx_id_hex: 'cc',
      },
    )
    assert.equal(calls.length, 1)
    const [payload] = calls[0]
    assert.equal(payload.event, 'webauthn_assert_verify')
    assert.equal(payload.error_kind, 'uv_required')
    assert.equal(payload.account_thumb_hex, 'aa')
    assert.equal(payload.credential_id_hash, 'bb')
    assert.equal(payload.tx_id_hex, 'cc')
    assert.equal(payload.correlation_id, 'req-3')
  })()
})
