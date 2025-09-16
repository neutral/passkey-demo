import express from 'express'

import {
  respondPayloadTooLarge,
  respondRateLimit,
  ERROR_CODES,
} from './error.js'

function formatJsonLimit(bytes) {
  if (!bytes || bytes <= 0) return undefined
  if (bytes % (1 << 20) === 0) return `${bytes / (1 << 20)}mb`
  if (bytes % 1024 === 0) return `${bytes / 1024}kb`
  return `${bytes}b`
}

export function buildBodyLimit({ bytes, json = true } = {}) {
  const maxBytes = typeof bytes === 'number' && bytes > 0 ? bytes : 0
  const parser = json
    ? express.json({ limit: formatJsonLimit(maxBytes) || '1mb', strict: true })
    : express.urlencoded({ limit: formatJsonLimit(maxBytes) || '1mb', extended: false })

  return function bodyLimitMiddleware(req, res, next) {
    if (maxBytes > 0) {
      const header = req.headers['content-length']
      if (header) {
        const requested = Number(header)
        if (Number.isFinite(requested) && requested > maxBytes) {
          respondPayloadTooLarge(res, req.id, { max_bytes: maxBytes })
          return
        }
      }
    }
    parser(req, res, (err) => {
      if (err) {
        next(err)
        return
      }
      next()
    })
  }
}

export function buildRateLimiter({
  burst = 20,
  refillPerSecond = 10,
  getKey,
  now = () => Date.now(),
} = {}) {
  if (!burst || burst <= 0 || !refillPerSecond || refillPerSecond <= 0) {
    return function noopRateLimiter(req, res, next) {
      next()
    }
  }

  const buckets = new Map()

  return function rateLimitMiddleware(req, res, next) {
    const key = (typeof getKey === 'function' && getKey(req, res)) || req.ip || req.headers['x-forwarded-for']
    if (!key) return next()

    const nowMs = now()
    const bucket = buckets.get(key) || { tokens: burst, updatedAt: nowMs }
    const elapsedSeconds = Math.max(0, (nowMs - bucket.updatedAt) / 1000)
    bucket.tokens = Math.min(burst, bucket.tokens + elapsedSeconds * refillPerSecond)
    bucket.updatedAt = nowMs

    if (bucket.tokens < 1) {
      const deficit = 1 - bucket.tokens
      const retryAfter = deficit / refillPerSecond
      buckets.set(key, bucket)
      respondRateLimit(res, req.id, retryAfter)
      return
    }

    bucket.tokens -= 1
    buckets.set(key, bucket)

    const remaining = Math.max(0, Math.floor(bucket.tokens))
    res.setHeader('X-RateLimit-Limit', String(burst))
    res.setHeader('X-RateLimit-Remaining', String(remaining))

    next()
  }
}

export const DEFAULT_LIMITS = {
  AUTHN_BODY_LIMIT_BYTES: 1 << 20, // 1 MiB
  TX_BODY_LIMIT_BYTES: 1 << 20, // 1 MiB
  AUTHN_RATE: { burst: 20, refillPerSecond: 10 },
  TX_RATE: { burst: 20, refillPerSecond: 10 },
}

export { ERROR_CODES }

export default {
  buildBodyLimit,
  buildRateLimiter,
  DEFAULT_LIMITS,
}
