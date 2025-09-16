import express from 'express'
import { randomBytes, createHash } from 'node:crypto'

import { logTxOptionsSuccess, logTxOptionsError } from '../logger.js'
import writeError from '../error.js'
import validateAndAnchorBundle, { ERROR_KINDS, BundleValidationError } from './bundle.js'

const TX_SESSION_TTL_SECONDS = 300
const SESSION_ID_ATTEMPTS = 3
const ACCOUNT_THUMB_PREFIX = Buffer.from('ACCTK1')

export class TxSessionStore {
  constructor() {
    this.sessions = new Map()
  }

  pruneExpired(nowSeconds) {
    for (const [id, payload] of this.sessions.entries()) {
      if (!payload || typeof payload.expiresAt !== 'number' || payload.expiresAt <= nowSeconds) {
        this.sessions.delete(id)
      }
    }
  }

  has(id) {
    return this.sessions.has(id)
  }

  set(id, payload) {
    this.sessions.set(id, payload)
  }

  get(id) {
    return this.sessions.get(id) ?? null
  }

  delete(id) {
    this.sessions.delete(id)
  }
}

function base64url(buffer) {
  return buffer.toString('base64url')
}

function generateSessionId() {
  return randomBytes(24).toString('base64url')
}

function computeAccountThumbHex(acctCbor) {
  return createHash('sha256').update(ACCOUNT_THUMB_PREFIX).update(acctCbor).digest('hex')
}

export function createTxOptionsRoutes(config, deps = {}) {
  const router = express.Router()
  const db = deps.db
  if (!db) throw new Error('createTxOptionsRoutes requires db')
  const store = deps.store ?? new TxSessionStore()
  const now = deps.now ?? (() => Math.floor(Date.now() / 1000))
  const idFactory = deps.idFactory ?? generateSessionId
  const credentialStmt = deps.credentialStmt ?? db.prepare('SELECT credential_id FROM credentials WHERE acct_cbor_fk = ?')

  router.post('/signing/options', (req, res) => {
    const correlationId = req.id
    try {
      if (!req.session || !req.session.acct_cbor) {
        return writeError(res, 401, 'unauthorized', 'Unauthorized', correlationId)
      }
      const body = req.body
      if (!body || typeof body !== 'object' || typeof body.bundle_cbor_b64 !== 'string') {
        return writeError(res, 400, 'bad_request', 'Bad request', correlationId)
      }

      const sessionAcct = Buffer.isBuffer(req.session.acct_cbor)
        ? Buffer.from(req.session.acct_cbor)
        : Buffer.from(req.session.acct_cbor)

      let anchored
      try {
        anchored = validateAndAnchorBundle(db, sessionAcct, body.bundle_cbor_b64)
      } catch (err) {
        if (err instanceof BundleValidationError) {
          switch (err.kind) {
            case ERROR_KINDS.BUNDLE_BASE64:
            case ERROR_KINDS.BUNDLE_CBOR:
            case ERROR_KINDS.MESSAGE_TOO_LONG:
            case ERROR_KINDS.NONCE_OUT_OF_RANGE:
              return writeError(res, 400, 'bad_request', 'Invalid bundle', correlationId)
            case ERROR_KINDS.SENDER_KEY_MISMATCH:
              return writeError(res, 401, 'unauthorized', 'Unauthorized', correlationId)
            case ERROR_KINDS.NONCE_NOT_MONOTONIC:
              return writeError(res, 409, 'conflict', 'Nonce conflict', correlationId)
            default:
              return writeError(res, 400, 'bad_request', 'Invalid bundle', correlationId)
          }
        }
        logTxOptionsError({ correlation_id: correlationId }, err, 'bundle validation failed')
        return writeError(res, 500, 'internal_error', 'Internal server error', correlationId)
      }

      let credentialIds
      try {
        const rows = credentialStmt.all(sessionAcct)
        credentialIds = rows.map((row) => Buffer.from(row.credential_id))
      } catch (err) {
        logTxOptionsError({ correlation_id: correlationId }, err, 'credential lookup failed')
        return writeError(res, 500, 'internal_error', 'Internal server error', correlationId)
      }
      if (credentialIds.length === 0) {
        return writeError(res, 409, 'conflict', 'No credentials for account', correlationId)
      }

      const nowSeconds = now()
      store.pruneExpired(nowSeconds)

      let txSessionId = ''
      for (let attempt = 0; attempt < SESSION_ID_ATTEMPTS; attempt += 1) {
        txSessionId = idFactory()
        if (!store.has(txSessionId)) break
        txSessionId = ''
      }
      if (!txSessionId) {
        logTxOptionsError({ correlation_id: correlationId }, null, 'session id collision exhaustion')
        return writeError(res, 500, 'internal_error', 'Internal server error', correlationId)
      }

      const expiresAt = nowSeconds + TX_SESSION_TTL_SECONDS
      const canonical = Buffer.from(anchored.canonical)
      const challenge = Buffer.from(anchored.challenge)
      const txId = Buffer.from(anchored.txId)
      store.set(txSessionId, {
        canonical,
        challenge,
        txId,
        acctCbor: Buffer.from(sessionAcct),
        credentialIds: credentialIds.map((id) => Buffer.from(id)),
        expiresAt,
      })

      const challengeB64 = base64url(challenge)
      const txIdHex = txId.toString('hex')
      const accountThumbHex = computeAccountThumbHex(sessionAcct)
      const allowCredentials = credentialIds.map((id) => ({
        type: 'public-key',
        id: base64url(id),
      }))

      logTxOptionsSuccess({
        correlation_id: correlationId,
        account_thumb_hex: accountThumbHex,
        tx_id_hex: txIdHex,
        allow_credentials_count: allowCredentials.length,
        expires_at: expiresAt,
      })

      res.status(200).json({
        tx_session_id: txSessionId,
        challenge: challengeB64,
        options: {
          rpId: config.RP_ID,
          origin: config.ORIGIN,
          timeout: 60000,
          userVerification: 'required',
          challenge: challengeB64,
          allowCredentials,
        },
        tx_id_hex: txIdHex,
        expires_at: expiresAt,
      })
    } catch (err) {
      logTxOptionsError({ correlation_id: correlationId }, err, 'tx options failed')
      writeError(res, 500, 'internal_error', 'Internal server error', correlationId)
    }
  })

  return { router, store }
}

export default createTxOptionsRoutes
