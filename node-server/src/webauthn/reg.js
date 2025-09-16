import express from 'express'
import { generateRegistrationOptions, verifyRegistrationResponse } from '@simplewebauthn/server'
import { randomBytes, createHash } from 'node:crypto'
import { customAlphabet } from 'nanoid'
import {
  logRegOptionsSuccess,
  logRegOptionsError,
  logRegFinishSuccess,
  logRegFinishError,
} from '../logger.js'
import {
  respondBadRequest,
  respondUnauthorized,
  respondForbidden,
  respondConflict,
  respondInternalError,
} from '../error.js'
import { Encoder, Decoder } from 'cbor-x'

const canonicalEncoder = new Encoder({ canonical: true, structuredClone: false, useRecords: false, mapsAsObjects: false })
const canonicalDecoder = new Decoder({ useMaps: true })

export const REGISTRATION_SESSION_TTL_SECONDS = 300
const SESSION_ID_ATTEMPTS = 3
const SESSION_ID_LENGTH = 24
const SESSION_ID_ALPHABET = '0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz-_' // base64url
const DEFAULT_USER_NAME = 'passkey-user'
const DEFAULT_USER_DISPLAY_NAME = 'Passkey User'
const DEFAULT_RP_NAME = 'Passkey Demo'
const ACCOUNT_THUMB_PREFIX = Buffer.from('ACCTK1')

function hashIdentifier(buffer) {
  return createHash('sha256').update(buffer).digest('hex')
}

function computeAccountThumb(acctCbor) {
  return createHash('sha256').update(ACCOUNT_THUMB_PREFIX).update(acctCbor).digest()
}

function canonicalizeCoseKey(coseBytes) {
  const decoded = canonicalDecoder.decode(coseBytes)
  const canonical = canonicalEncoder.encode(decoded)
  return Buffer.from(canonical)
}

function parseAAGUID(aaguid) {
  if (typeof aaguid !== 'string' || aaguid.length === 0) {
    return Buffer.alloc(0)
  }
  if (aaguid.includes('-')) {
    const hex = aaguid.replace(/-/g, '')
    return Buffer.from(hex, 'hex')
  }
  try {
    return Buffer.from(aaguid, 'base64url')
  } catch {
    return Buffer.alloc(0)
  }
}

function uniqueStrings(...values) {
  return [...new Set(values.filter((v) => typeof v === 'string' && v))]
}

export class RegistrationSessionStore {
  constructor({ ttlSeconds = REGISTRATION_SESSION_TTL_SECONDS } = {}) {
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

export function createRegistrationRoutes(config, deps = {}) {
  const router = express.Router()
  const store = deps.store ?? new RegistrationSessionStore()
  const now = deps.now ?? (() => Math.floor(Date.now() / 1000))
  const idFactory = deps.idFactory ?? customAlphabet(SESSION_ID_ALPHABET, SESSION_ID_LENGTH)
  const optionsFactory = deps.generateRegistrationOptions ?? generateRegistrationOptions
  const verifier = deps.verifyRegistrationResponse ?? verifyRegistrationResponse
  const db = deps.db

  if (!db) {
    throw new Error('createRegistrationRoutes requires db')
  }

  const insertAccount = db.prepare(
    'INSERT OR IGNORE INTO accounts (acct_cbor, acct_thumb, created_at) VALUES (?, ?, ?)',
  )
  const insertCredential = db.prepare(
    'INSERT INTO credentials (credential_id, acct_cbor_fk, sign_count, aaguid, created_at) VALUES (?, ?, ?, ?, ?)',
  )
  const persistCredential = db.transaction((acctCbor, thumb, credentialId, signCount, aaguidBuf, createdAt) => {
    insertAccount.run(acctCbor, thumb, createdAt)
    insertCredential.run(credentialId, acctCbor, signCount, aaguidBuf, createdAt)
  })

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

      const expiresAt = nowSeconds + REGISTRATION_SESSION_TTL_SECONDS
      const options = await Promise.resolve(
        optionsFactory({
          rpName: DEFAULT_RP_NAME,
          rpID: config.RP_ID,
          userID: randomBytes(32),
          userName: DEFAULT_USER_NAME,
          userDisplayName: DEFAULT_USER_DISPLAY_NAME,
          attestationType: 'none',
          authenticatorSelection: {
            residentKey: 'required',
            requireResidentKey: true,
            userVerification: 'required',
          },
          supportedAlgorithmIDs: [-7],
          timeout: 60000,
        }),
      )

      if (!options || typeof options.challenge !== 'string' || options.challenge.length === 0) {
        throw new Error('invalid_challenge')
      }
      const challenge = options.challenge
      store.create(sessionId, {
        challenge,
        rpID: config.RP_ID,
        origin: config.ORIGIN,
        expiresAt,
      })

      const responseBody = {
        ...options,
        reg_session_id: sessionId,
        expires_at: expiresAt,
      }

      logRegOptionsSuccess(
        { correlationId: req.id, rpId: config.RP_ID, origin: config.ORIGIN },
        { expires_at: expiresAt, session_id_len: sessionId.length },
      )

      res.status(200).json(responseBody)
    } catch (err) {
      logRegOptionsError({ correlationId: req.id, rpId: config.RP_ID, origin: config.ORIGIN }, {}, err)
      respondInternalError(res, req.id)
    }
  })

  router.post('/finish', async (req, res) => {
    const correlationId = req.id
    const baseContext = { correlationId, rpId: config.RP_ID, origin: config.ORIGIN }
    const logFailure = (reason, attrs = {}, err) => {
      logRegFinishError(baseContext, { reason, ...attrs }, err)
    }
    try {
      const body = req.body
      if (!body || typeof body !== 'object') {
        logFailure('invalid_payload')
        return respondBadRequest(res, correlationId)
      }
      const { reg_session_id: regSessionId, ...responsePayload } = body
      if (typeof regSessionId !== 'string' || regSessionId.trim().length === 0) {
        logFailure('missing_session_id')
        return respondBadRequest(res, correlationId)
      }
      if (!responsePayload || typeof responsePayload !== 'object' || typeof responsePayload.response !== 'object') {
        logFailure('invalid_response')
        return respondBadRequest(res, correlationId)
      }
      if (typeof responsePayload.id !== 'string' || typeof responsePayload.rawId !== 'string') {
        return respondBadRequest(res, correlationId)
      }
      const session = store.get(regSessionId)
      const nowSeconds = now()
      if (!session || typeof session.expiresAt !== 'number' || session.expiresAt <= nowSeconds) {
        if (session) store.delete(regSessionId)
        logRegFinishError({ ...baseContext, origin: session?.origin || config.ORIGIN }, { reason: 'session_expired' })
        return respondUnauthorized(res, correlationId)
      }
      const sessionContext = { correlationId, rpId: config.RP_ID, origin: session.origin }

      const expectedOrigins = uniqueStrings(config.ORIGIN, ...(config.ORIGIN_ALLOWLIST || []))
      const expectedRPIDs = uniqueStrings(config.RP_ID, ...(config.RP_ID_ALLOWLIST || []))

      const verification = await verifier({
        response: responsePayload,
        expectedChallenge: session.challenge,
        expectedOrigin: expectedOrigins.length === 1 ? expectedOrigins[0] : expectedOrigins,
        expectedRPID: expectedRPIDs.length === 1 ? expectedRPIDs[0] : expectedRPIDs,
        requireUserVerification: true,
        supportedAlgorithmIDs: [-7],
      })

      const info = verification.registrationInfo
      if (!verification.verified || !info) {
        store.delete(regSessionId)
        logRegFinishError(sessionContext, { reason: 'verification_failed' })
        return respondUnauthorized(res, correlationId)
      }
      if (info.fmt && info.fmt !== 'none') {
        store.delete(regSessionId)
        logRegFinishError(sessionContext, { reason: 'attestation_forbidden' })
        return respondForbidden(res, correlationId, 'Policy violation')
      }
      if (!info.userVerified) {
        store.delete(regSessionId)
        logRegFinishError(sessionContext, { reason: 'uv_required' })
        return respondForbidden(res, correlationId, 'Policy violation')
      }

      const credentialId = Buffer.from(info.credentialID, 'base64url')
      const coseKey = Buffer.from(info.credentialPublicKey)
      let canonicalCose
      try {
        canonicalCose = canonicalizeCoseKey(coseKey)
      } catch (err) {
        logRegFinishError(sessionContext, { reason: 'canonicalize_cose_failed' }, err)
        store.delete(regSessionId)
        return respondBadRequest(res, correlationId)
      }
      const accountThumb = computeAccountThumb(canonicalCose)
      const aaguidBuf = parseAAGUID(info.aaguid)
      const signCount = Number.isFinite(info.counter) && info.counter >= 0 ? Math.floor(info.counter) : 0

      const createdAt = nowSeconds
      try {
        persistCredential(canonicalCose, accountThumb, credentialId, signCount, aaguidBuf, createdAt)
      } catch (err) {
        const message = String(err?.message || '')
        if (message.toLowerCase().includes('unique')) {
          store.delete(regSessionId)
          logRegFinishError(sessionContext, { reason: 'credential_conflict' }, err)
          return respondConflict(res, correlationId)
        }
        logRegFinishError(sessionContext, { reason: 'persistence_failed' }, err)
        return respondInternalError(res, correlationId)
      }

      store.delete(regSessionId)

      const accountThumbHex = accountThumb.toString('hex')
      const credentialIdHash = hashIdentifier(credentialId)
      logRegFinishSuccess(sessionContext, {
        account_thumb_hex: accountThumbHex,
        credential_id_hash: credentialIdHash,
        sign_count: signCount,
      })

      res.status(201).json({
        account_thumb_hex: accountThumbHex,
        credential_id_b64: credentialId.toString('base64url'),
      })
    } catch (err) {
      logRegFinishError(baseContext, { reason: 'exception' }, err)
      respondInternalError(res, req.id)
    }
  })

  return { router, store }
}

export default createRegistrationRoutes
