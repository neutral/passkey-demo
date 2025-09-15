import express from 'express'
import { Decoder } from 'cbor-x'
import { createHash } from 'node:crypto'
import { logger } from './logger.js'
import writeError from './error.js'

const coseDecoder = new Decoder({ useMaps: true })
const ACCOUNT_THUMB_PREFIX = Buffer.from('ACCTK1')

function toBuffer(value) {
  if (value instanceof Buffer) return value
  if (value instanceof Uint8Array) return Buffer.from(value)
  if (Array.isArray(value)) return Buffer.from(value)
  if (value && typeof value === 'object' && 'buffer' in value) {
    try {
      return Buffer.from(value.buffer)
    } catch {}
  }
  return Buffer.alloc(0)
}

function mapFromDecoded(value) {
  if (value instanceof Map) return value
  if (value && typeof value === 'object') {
    const m = new Map()
    for (const [key, v] of Object.entries(value)) {
      const numKey = Number(key)
      if (!Number.isFinite(numKey)) continue
      m.set(numKey, v)
    }
    return m
  }
  return new Map()
}

function computeAccountThumb(acctCbor) {
  return createHash('sha256').update(ACCOUNT_THUMB_PREFIX).update(acctCbor).digest()
}

function decodeAccountKey(acctCbor, decode) {
  const decoded = decode(acctCbor)
  const map = mapFromDecoded(decoded)
  const kty = Number(map.get(1))
  const alg = Number(map.get(3))
  const crv = Number(map.get(-1))
  const xBuf = toBuffer(map.get(-2))
  const yBuf = toBuffer(map.get(-3))
  return {
    kty: Number.isFinite(kty) ? kty : 0,
    alg: Number.isFinite(alg) ? alg : 0,
    crv: Number.isFinite(crv) ? crv : 0,
    x: xBuf,
    y: yBuf,
  }
}

export function createMeRoutes(config, deps = {}) {
  if (!config) throw new Error('createMeRoutes requires config')
  const db = deps.db
  if (!db) throw new Error('createMeRoutes requires db')
  const decode = deps.decode ?? ((buf) => coseDecoder.decode(buf))

  const router = express.Router()

  const selectAccount = db.prepare(
    'SELECT acct_cbor, acct_thumb, created_at FROM accounts WHERE acct_cbor = ?',
  )

  router.get('/account_key', (req, res) => {
    const correlationId = req.id
    const session = req.session
    const sessionAcct = session ? toBuffer(session.acct_cbor) : Buffer.alloc(0)
    if (!session || sessionAcct.length === 0) {
      return writeError(res, 401, 'unauthorized', 'Unauthorized', correlationId)
    }

    let accountRow
    try {
      accountRow = selectAccount.get(sessionAcct)
    } catch (err) {
      logger.error({ event: 'me_account_key_error', correlation_id: correlationId, err }, 'failed to read account')
      return writeError(res, 500, 'internal_error', 'Internal server error', correlationId)
    }

    if (!accountRow) {
      return writeError(res, 401, 'unauthorized', 'Unauthorized', correlationId)
    }

    const acctCbor = toBuffer(accountRow.acct_cbor)
    if (acctCbor.length === 0) {
      logger.error({ event: 'me_account_key_error', correlation_id: correlationId }, 'account missing acct_cbor')
      return writeError(res, 500, 'internal_error', 'Internal server error', correlationId)
    }

    const thumbBufRaw = toBuffer(accountRow.acct_thumb)
    const thumbBuf = thumbBufRaw.length > 0 ? thumbBufRaw : computeAccountThumb(acctCbor)
    const accountThumbHex = thumbBuf.toString('hex')

    let decoded
    try {
      decoded = decodeAccountKey(acctCbor, decode)
    } catch (err) {
      logger.error({ event: 'me_account_key_error', correlation_id: correlationId, err }, 'failed to decode account key')
      return writeError(res, 500, 'internal_error', 'Internal server error', correlationId)
    }

    if (decoded.x.length === 0 || decoded.y.length === 0) {
      logger.error({ event: 'me_account_key_error', correlation_id: correlationId }, 'account key missing x or y coordinate')
      return writeError(res, 500, 'internal_error', 'Internal server error', correlationId)
    }

    const senderKey = {
      kty: decoded.kty,
      alg: decoded.alg,
      crv: decoded.crv,
      x: decoded.x.toString('base64url'),
      y: decoded.y.toString('base64url'),
    }

    const responseBody = {
      acct_cbor_b64: acctCbor.toString('base64url'),
      account_thumb_hex: accountThumbHex,
      pubkey_x_hex: decoded.x.toString('hex'),
      pubkey_y_hex: decoded.y.toString('hex'),
      created_at: typeof accountRow.created_at === 'number' ? accountRow.created_at : Number(accountRow.created_at) || 0,
      sender_key: senderKey,
    }

    logger.info({
      event: 'me_account_key',
      correlation_id: correlationId,
      account_thumb_hex: accountThumbHex,
      rp_id: config.RP_ID,
      origin: config.ORIGIN,
    })

    return res.status(200).json(responseBody)
  })

  return { router }
}

export default createMeRoutes
