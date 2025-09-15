import express from 'express'
import { generateRegistrationOptions } from '@simplewebauthn/server'
import { randomBytes } from 'node:crypto'
import { customAlphabet } from 'nanoid'
import { logger } from '../logger.js'
import writeError from '../error.js'

export const REGISTRATION_SESSION_TTL_SECONDS = 300
const SESSION_ID_ATTEMPTS = 3
const SESSION_ID_LENGTH = 24
const SESSION_ID_ALPHABET = '0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz-_' // base64url
const DEFAULT_USER_NAME = 'passkey-user'
const DEFAULT_USER_DISPLAY_NAME = 'Passkey User'
const DEFAULT_RP_NAME = 'Passkey Demo'

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
          userID: randomBytes(32).toString('base64url'),
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

  return { router, store }
}

export default createRegistrationRoutes
