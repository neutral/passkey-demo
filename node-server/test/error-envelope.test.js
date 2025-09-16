import test from 'node:test'
import assert from 'node:assert/strict'
import express from 'express'

import {
  respondBadRequest,
  respondInternalError,
  buildExpressErrorHandler,
  ERROR_CODES,
} from '../src/error.js'

const TEST_TIMEOUT_MS = 3000
const FETCH_TIMEOUT_MS = 2500

test('respondBadRequest emits standardized envelope with correlation id', { timeout: TEST_TIMEOUT_MS }, async () => {
  const app = express()
  app.get('/bad', (req, res) => respondBadRequest(res, 'corr-1', 'Invalid input'))
  const server = app.listen(0)
  try {
    const { port } = server.address()
    const res = await fetch(`http://127.0.0.1:${port}/bad`, {
      signal: AbortSignal.timeout(FETCH_TIMEOUT_MS),
    })
    assert.equal(res.status, 400)
    const body = await res.json()
    assert.equal(body.code, ERROR_CODES.BAD_REQUEST)
    assert.equal(body.error, 'Invalid input')
    assert.equal(body.correlation_id, 'corr-1')
  } finally {
    await new Promise((resolve) => server.close(resolve))
  }
})

test('buildExpressErrorHandler maps JSON parse errors to ERR_BAD_REQUEST', { timeout: TEST_TIMEOUT_MS }, async () => {
  const app = express()
  app.use((req, res, next) => {
    req.id = 'corr-json'
    next()
  })
  app.use(express.json())
  app.post('/json', (req, res) => respondInternalError(res, req.id))
  app.use(buildExpressErrorHandler())
  const server = app.listen(0)
  try {
    const { port } = server.address()
    const res = await fetch(`http://127.0.0.1:${port}/json`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: '{"invalid"',
      signal: AbortSignal.timeout(FETCH_TIMEOUT_MS),
    })
    assert.equal(res.status, 400)
    const body = await res.json()
    assert.equal(body.code, ERROR_CODES.BAD_REQUEST)
    assert.equal(body.correlation_id, 'corr-json')
  } finally {
    await new Promise((resolve) => server.close(resolve))
  }
})
