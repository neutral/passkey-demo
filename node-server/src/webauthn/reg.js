import express from 'express'
import { generateRegistrationOptions, verifyRegistrationResponse } from '@simplewebauthn/server'
import { randomBytes, createHash } from 'node:crypto'
import { customAlphabet } from 'nanoid'
import { logger } from '../logger.js'
import writeError from '../error.js'
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

      logger.info({
        event: 'reg_options',
        correlation_id: req.id,
        rp_id: config.RP_ID,
        origin: config.ORIGIN,
        expires_at: expiresAt,
        session_id_len: sessionId.length,
      })

      res.status(200).json(responseBody)
    } catch (err) {
      logger.error({ event: 'reg_options_error', correlation_id: req.id, err }, 'registration options failed')
      writeError(res, 500, 'internal_error', 'Internal server error', req.id)
    }
  })

  router.post('/finish', async (req, res) => {
    const correlationId = req.id
    try {
      const body = req.body
      if (!body || typeof body !== 'object') {
        return writeError(res, 400, 'bad_request', 'Bad request', correlationId)
      }
      const { reg_session_id: regSessionId, ...responsePayload } = body
      if (typeof regSessionId !== 'string' || regSessionId.trim().length === 0) {
        return writeError(res, 400, 'bad_request', 'Bad request', correlationId)
      }
      if (!responsePayload || typeof responsePayload !== 'object' || typeof responsePayload.response !== 'object') {
        return writeError(res, 400, 'bad_request', 'Bad request', correlationId)
      }
      if (typeof responsePayload.id !== 'string' || typeof responsePayload.rawId !== 'string') {
        return writeError(res, 400, 'bad_request', 'Bad request', correlationId)
      }
      const session = store.get(regSessionId)
      const nowSeconds = now()
      if (!session || typeof session.expiresAt !== 'number' || session.expiresAt <= nowSeconds) {
        if (session) store.delete(regSessionId)
        return writeError(res, 401, 'unauthorized', 'Unauthorized', correlationId)
      }

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
        return writeError(res, 401, 'unauthorized', 'Unauthorized', correlationId)
      }
      if (info.fmt && info.fmt !== 'none') {
        store.delete(regSessionId)
        return writeError(res, 403, 'policy_violation', 'Policy violation', correlationId)
      }
      if (!info.userVerified) {
        store.delete(regSessionId)
        return writeError(res, 403, 'policy_violation', 'Policy violation', correlationId)
      }

      const credentialId = Buffer.from(info.credentialID, 'base64url')
      const coseKey = Buffer.from(info.credentialPublicKey)
      let canonicalCose
      try {
        canonicalCose = canonicalizeCoseKey(coseKey)
      } catch (err) {
        logger.error({ event: 'reg_finish_error', correlation_id: correlationId, err }, 'failed to canonicalize COSE key')
        store.delete(regSessionId)
        return writeError(res, 400, 'bad_request', 'Bad request', correlationId)
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
          return writeError(res, 409, 'conflict', 'Conflict', correlationId)
        }
        logger.error({ event: 'reg_finish_error', correlation_id: correlationId, err }, 'registration persistence failed')
        return writeError(res, 500, 'internal_error', 'Internal server error', correlationId)
      }

      store.delete(regSessionId)

      const accountThumbHex = accountThumb.toString('hex')
      const credentialIdHash = hashIdentifier(credentialId)
      logger.info({
        event: 'reg_finish',
        correlation_id: correlationId,
        rp_id: config.RP_ID,
        origin: session.origin,
        account_thumb_hex: accountThumbHex,
        credential_id_hash: credentialIdHash,
        sign_count: signCount,
      })

      res.status(201).json({
        account_thumb_hex: accountThumbHex,
        credential_id_b64: credentialId.toString('base64url'),
      })
    } catch (err) {
      logger.error({ event: 'reg_finish_error', correlation_id: req.id, err }, 'registration finish failed')
      writeError(res, 500, 'internal_error', 'Internal server error', req.id)
    }
  })

  return { router, store }
}

export default createRegistrationRoutes
