import test from 'node:test'
import assert from 'node:assert/strict'
import express from 'express'

import { buildBodyLimit, buildRateLimiter } from '../src/limits.js'
import { buildExpressErrorHandler, ERROR_CODES } from '../src/error.js'

const TEST_TIMEOUT_MS = 4000
const FETCH_TIMEOUT_MS = 2500

test('body limit middleware returns 413 with standardized envelope', { timeout: TEST_TIMEOUT_MS }, async () => {
  const app = express()
  app.use((req, res, next) => {
    req.id = 'body-limit-test'
    next()
  })
  app.post('/limited', buildBodyLimit({ bytes: 64 }), (req, res) => {
    res.json({ ok: true })
  })
  app.use(buildExpressErrorHandler())
  const server = app.listen(0)
  try {
    const { port } = server.address()
    const payload = { data: 'x'.repeat(1024) }
    const res = await fetch(`http://127.0.0.1:${port}/limited`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload),
      signal: AbortSignal.timeout(FETCH_TIMEOUT_MS),
    })
    assert.equal(res.status, 413)
    const body = await res.json()
    assert.equal(body.code, ERROR_CODES.PAYLOAD_TOO_LARGE)
    assert.equal(body.correlation_id, 'body-limit-test')
    assert.equal(body.details?.max_bytes, 64)
  } finally {
    await new Promise((resolve) => server.close(resolve))
  }
})

test('rate limiter enforces burst and sets Retry-After', { timeout: TEST_TIMEOUT_MS }, async () => {
  let nowMs = 0
  const app = express()
  app.use((req, res, next) => {
    req.id = 'rate-limit-test'
    next()
  })
  app.use(
    buildRateLimiter({
      burst: 2,
      refillPerSecond: 1,
      now: () => nowMs,
      getKey: (req) => req.headers['x-test-key'] || req.ip,
    }),
  )
  app.get('/ping', (req, res) => {
    res.json({ ok: true })
  })
  app.use(buildExpressErrorHandler())
  const server = app.listen(0)
  try {
    const { port } = server.address()
    const url = `http://127.0.0.1:${port}/ping`
    const doFetch = () =>
      fetch(url, {
        headers: { 'X-Test-Key': 'client-1' },
        signal: AbortSignal.timeout(FETCH_TIMEOUT_MS),
      })

    nowMs = 0
    let res = await doFetch()
    assert.equal(res.status, 200)

    nowMs = 1
    res = await doFetch()
    assert.equal(res.status, 200)

    nowMs = 2
    res = await doFetch()
    assert.equal(res.status, 429)
    assert.equal(res.headers.get('Retry-After'), '1')
    const body = await res.json()
    assert.equal(body.code, ERROR_CODES.RATE_LIMIT)
    assert.equal(body.correlation_id, 'rate-limit-test')

    nowMs = 3000
    res = await doFetch()
    assert.equal(res.status, 200)
  } finally {
    await new Promise((resolve) => server.close(resolve))
  }
})
