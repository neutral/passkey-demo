import express from 'express'
import { generateAuthenticationOptions, verifyAuthenticationResponse } from '@simplewebauthn/server'
import { customAlphabet } from 'nanoid'
import { randomBytes, createHash } from 'node:crypto'
import { logger } from '../logger.js'
import {
  respondBadRequest,
  respondUnauthorized,
  respondForbidden,
  respondConflict,
  respondInternalError,
} from '../error.js'

export const LOGIN_SESSION_TTL_SECONDS = 300
export const SESSION_COOKIE_TTL_SECONDS = 3600
const SESSION_ID_ATTEMPTS = 3
const SESSION_ID_LENGTH = 24
const SESSION_ID_ALPHABET = '0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz-_' // base64url

export class LoginSessionStore {
  constructor({ ttlSeconds = LOGIN_SESSION_TTL_SECONDS } = {}) {
    this.ttlSeconds = ttlSeconds
    this.sessions = new Map()
  }

  create(sessionId, payload) {
    if (!sessionId || typeof sessionId !== 'string') {
      throw new Error('sessionId must be a non-empty string')
    }
    if (!payload || typeof payload.expiresAt !== 'number') {
      throw new Error('payload.expiresAt must be a number')
    }
    this.sessions.set(sessionId, payload)
  }

  get(sessionId) {
    return this.sessions.get(sessionId) ?? null
  }

  has(sessionId) {
    return this.sessions.has(sessionId)
  }

  delete(sessionId) {
    this.sessions.delete(sessionId)
  }

  pruneExpired(nowSeconds) {
    if (typeof nowSeconds !== 'number') return
    for (const [id, payload] of this.sessions.entries()) {
      if (!payload || typeof payload.expiresAt !== 'number') {
        this.sessions.delete(id)
        continue
      }
      if (payload.expiresAt <= nowSeconds) {
        this.sessions.delete(id)
      }
    }
  }
}

function uniqueStrings(...values) {
  return [...new Set(values.filter((v) => typeof v === 'string' && v))]
}

function hashIdentifier(buffer) {
  return createHash('sha256').update(buffer).digest('hex')
}

function computeAccountThumb(acctCbor) {
  return createHash('sha256').update('ACCTK1').update(acctCbor).digest()
}

function isSecureOrigin(origin) {
  try {
    const url = new URL(origin)
    return url.protocol === 'https:'
  } catch {
    return false
  }
}

export function createLoginRoutes(config, deps = {}) {
  const router = express.Router()
  const store = deps.store ?? new LoginSessionStore()
  const now = deps.now ?? (() => Math.floor(Date.now() / 1000))
  const idFactory = deps.idFactory ?? customAlphabet(SESSION_ID_ALPHABET, SESSION_ID_LENGTH)
  const optionsFactory = deps.generateAuthenticationOptions ?? generateAuthenticationOptions
  const verifier = deps.verifyAuthenticationResponse ?? verifyAuthenticationResponse
  const sessionIdFactory = deps.sessionIdFactory ?? (() => randomBytes(24).toString('base64url'))
  const db = deps.db

  if (!db) {
    throw new Error('createLoginRoutes requires db')
  }

  const selectCredential = db.prepare(
    'SELECT acct_cbor_fk AS acct_cbor, sign_count FROM credentials WHERE credential_id = ?',
  )
  const selectAccount = db.prepare('SELECT acct_thumb, acct_cbor FROM accounts WHERE acct_cbor = ?')
  const updateSignCount = db.prepare('UPDATE credentials SET sign_count = ? WHERE credential_id = ?')
  const insertSession = db.prepare(
    'INSERT INTO sessions (session_id, acct_cbor, expires_at, created_at) VALUES (?, ?, ?, ?)',
  )

  router.post('/options', async (req, res) => {
    try {
      const nowSeconds = now()
      store.pruneExpired(nowSeconds)

      let sessionId = ''
      for (let attempt = 0; attempt < SESSION_ID_ATTEMPTS; attempt += 1) {
        sessionId = idFactory()
        if (!store.has(sessionId)) break
        sessionId = ''
      }
      if (!sessionId) {
        throw new Error('session_id_collision')
      }

      const expiresAt = nowSeconds + LOGIN_SESSION_TTL_SECONDS
      const options = await Promise.resolve(
        optionsFactory({
          rpID: config.RP_ID,
          userVerification: 'required',
          timeout: 60000,
          allowCredentials: [],
        }),
      )

      if (!options || typeof options.challenge !== 'string' || options.challenge.length === 0) {
        throw new Error('invalid_challenge')
      }

      store.create(sessionId, {
        challenge: options.challenge,
        rpID: config.RP_ID,
        origin: config.ORIGIN,
        expiresAt,
      })

      const responseBody = {
        ...options,
        login_session_id: sessionId,
        expires_at: expiresAt,
      }

      logger.info({
        event: 'login_options',
        correlation_id: req.id,
        rp_id: config.RP_ID,
        origin: config.ORIGIN,
        expires_at: expiresAt,
        session_id_len: sessionId.length,
        allow_credentials_count: Array.isArray(options.allowCredentials) ? options.allowCredentials.length : 0,
      })

      res.status(200).json(responseBody)
    } catch (err) {
      logger.error({ event: 'login_options_error', correlation_id: req.id, err }, 'login options failed')
      respondInternalError(res, req.id)
    }
  })

  router.post('/finish', async (req, res) => {
    const correlationId = req.id
    try {
      const body = req.body
      if (!body || typeof body !== 'object') {
        return respondBadRequest(res, correlationId)
      }
      const { login_session_id: loginSessionId, ...responsePayload } = body
      if (typeof loginSessionId !== 'string' || loginSessionId.trim().length === 0) {
        return respondBadRequest(res, correlationId)
      }
      if (!responsePayload || typeof responsePayload !== 'object' || typeof responsePayload.response !== 'object') {
        return respondBadRequest(res, correlationId)
      }

      const nowSeconds = now()
      const session = store.get(loginSessionId)
      if (!session || typeof session.expiresAt !== 'number' || session.expiresAt <= nowSeconds) {
        if (session) store.delete(loginSessionId)
        return respondUnauthorized(res, correlationId)
      }

      let credentialIdBase64 = ''
      if (typeof responsePayload.rawId === 'string' && responsePayload.rawId.length > 0) {
        credentialIdBase64 = responsePayload.rawId
      } else if (typeof responsePayload.id === 'string' && responsePayload.id.length > 0) {
        credentialIdBase64 = responsePayload.id
      }
      let credentialId
      try {
        credentialId = Buffer.from(credentialIdBase64, 'base64url')
      } catch {
        credentialId = Buffer.alloc(0)
      }
      if (!credentialId || credentialId.length === 0) {
        store.delete(loginSessionId)
        return respondBadRequest(res, correlationId)
      }

      const credentialRow = selectCredential.get(credentialId)
      if (!credentialRow) {
        store.delete(loginSessionId)
        return respondUnauthorized(res, correlationId)
      }
      const accountRow = selectAccount.get(credentialRow.acct_cbor)
      if (!accountRow) {
        store.delete(loginSessionId)
        return respondUnauthorized(res, correlationId)
      }

      const expectedOrigins = uniqueStrings(config.ORIGIN, session.origin, ...(config.ORIGIN_ALLOWLIST || []))
      const expectedRPIDs = uniqueStrings(config.RP_ID, session.rpID, ...(config.RP_ID_ALLOWLIST || []))

      let verification
      try {
        verification = await verifier({
          response: responsePayload,
          expectedChallenge: session.challenge,
          expectedOrigin: expectedOrigins.length === 1 ? expectedOrigins[0] : expectedOrigins,
          expectedRPID: expectedRPIDs.length === 1 ? expectedRPIDs[0] : expectedRPIDs,
          requireUserVerification: true,
          authenticator: {
            credentialID: credentialId,
            credentialPublicKey: accountRow.acct_cbor,
            counter: credentialRow.sign_count,
          },
        })
      } catch (err) {
        logger.error({ event: 'login_finish_error', correlation_id: correlationId, err }, 'login verification failed')
        store.delete(loginSessionId)
        return respondUnauthorized(res, correlationId)
      }

      const info = verification?.authenticationInfo
      if (!verification?.verified || !info) {
        store.delete(loginSessionId)
        return respondUnauthorized(res, correlationId)
      }
      if (!info.userVerified) {
        store.delete(loginSessionId)
        return respondForbidden(res, correlationId, 'Policy violation')
      }

      const newCounter = typeof info.newCounter === 'number' ? info.newCounter : 0
      const storedCount = typeof credentialRow.sign_count === 'number' ? credentialRow.sign_count : 0
      if (newCounter > 0 && newCounter <= storedCount) {
        store.delete(loginSessionId)
        return respondConflict(res, correlationId)
      }

      try {
        if (newCounter > 0) {
          updateSignCount.run(newCounter, credentialId)
        }
      } catch (err) {
        logger.error({ event: 'login_finish_error', correlation_id: correlationId, err }, 'failed to update sign_count')
        store.delete(loginSessionId)
        return respondInternalError(res, correlationId)
      }

      const sessionId = sessionIdFactory()
      const createdAt = nowSeconds
      const expiresAt = nowSeconds + SESSION_COOKIE_TTL_SECONDS
      try {
        insertSession.run(sessionId, credentialRow.acct_cbor, expiresAt, createdAt)
      } catch (err) {
        logger.error({ event: 'login_finish_error', correlation_id: correlationId, err }, 'failed to persist session')
        store.delete(loginSessionId)
        return respondInternalError(res, correlationId)
      }

      const secure = isSecureOrigin(config.ORIGIN)
      const cookieParts = [
        `sid=${sessionId}`,
        'HttpOnly',
        'Path=/',
        'SameSite=Lax',
        `Max-Age=${SESSION_COOKIE_TTL_SECONDS}`,
      ]
      if (secure) cookieParts.push('Secure')
      res.setHeader('Set-Cookie', cookieParts.join('; '))

      store.delete(loginSessionId)

      const thumbBuffer = accountRow.acct_thumb instanceof Buffer ? accountRow.acct_thumb : Buffer.from(accountRow.acct_thumb)
      const accountThumbHex = thumbBuffer.byteLength > 0 ? thumbBuffer.toString('hex') : computeAccountThumb(accountRow.acct_cbor).toString('hex')
      const credentialIdHash = hashIdentifier(credentialId)

      logger.info({
        event: 'login_finish',
        correlation_id: correlationId,
        rp_id: config.RP_ID,
        origin: session.origin,
        account_thumb_hex: accountThumbHex,
        credential_id_hash: credentialIdHash,
        sign_count: newCounter > 0 ? newCounter : storedCount,
      })

      res.status(200).json({
        account_thumb_hex: accountThumbHex,
        credential_id_b64: credentialId.toString('base64url'),
      })
    } catch (err) {
      logger.error({ event: 'login_finish_error', correlation_id: req.id, err }, 'login finish failed')
      respondInternalError(res, req.id)
    }
  })

  return { router, store }
}

export default createLoginRoutes
