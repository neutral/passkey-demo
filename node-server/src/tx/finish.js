import express from 'express'
import { verifyAuthenticationResponse } from '@simplewebauthn/server'
import { createHash } from 'node:crypto'
import { Decoder } from 'cbor-x'

import writeError from '../error.js'
import {
  logTxFinishSuccess,
  logTxFinishError,
} from '../logger.js'
import { TxSessionStore } from './options.js'

const canonicalDecoder = new Decoder({ useMaps: true, mapsAsObjects: false })
const ACCOUNT_THUMB_PREFIX = Buffer.from('ACCTK1')

function base64urlToBuffer(value) {
  if (typeof value !== 'string') throw new Error('expected base64url string')
  return Buffer.from(value, 'base64url')
}

function uniqueStrings(...values) {
  return [...new Set(values.filter((v) => typeof v === 'string' && v))]
}

function computeAccountThumbHex(acctCbor) {
  return createHash('sha256').update(ACCOUNT_THUMB_PREFIX).update(acctCbor).digest('hex')
}

function hashIdentifier(buffer) {
  return createHash('sha256').update(buffer).digest('hex')
}

function readMapValue(map, key) {
  if (!map) return undefined
  if (map instanceof Map) {
    if (map.has(key)) return map.get(key)
    const strKey = String(key)
    if (map.has(strKey)) return map.get(strKey)
    return undefined
  }
  if (typeof map === 'object' && map !== null) {
    if (Object.prototype.hasOwnProperty.call(map, key)) return map[key]
    const strKey = String(key)
    if (Object.prototype.hasOwnProperty.call(map, strKey)) return map[strKey]
  }
  return undefined
}

function decodeBundleMeta(canonical) {
  let decoded
  try {
    decoded = canonicalDecoder.decode(canonical)
  } catch (err) {
    throw new Error('bundle_decode_failed')
  }
  const nonceVal = readMapValue(decoded, 1)
  const messageVal = readMapValue(decoded, 2)
  let nonce
  if (typeof nonceVal === 'number' && Number.isFinite(nonceVal)) {
    nonce = nonceVal
  } else if (typeof nonceVal === 'bigint') {
    nonce = Number(nonceVal)
  } else {
    throw new Error('invalid_nonce')
  }
  if (!Number.isSafeInteger(nonce) || nonce < 0) {
    throw new Error('invalid_nonce')
  }
  if (typeof messageVal !== 'string') {
    throw new Error('invalid_message')
  }
  return { nonce, message: messageVal }
}

export function createTxFinishRoutes(config, deps = {}) {
  const router = express.Router()
  const db = deps.db
  if (!db) throw new Error('createTxFinishRoutes requires db')
  const store = deps.store ?? new TxSessionStore()
  const now = deps.now ?? (() => Math.floor(Date.now() / 1000))
  const verifier = deps.verifyAuthenticationResponse ?? verifyAuthenticationResponse

  const selectCredential = deps.selectCredential ?? db.prepare(
    'SELECT credential_id, sign_count FROM credentials WHERE credential_id = ? AND acct_cbor_fk = ?',
  )
  const updateSignCount = deps.updateSignCount ?? db.prepare(
    'UPDATE credentials SET sign_count = ? WHERE credential_id = ?',
  )
  const insertTransaction = deps.insertTransaction ?? db.prepare(
    'INSERT INTO transactions (tx_id, acct_cbor, nonce, message, bundle_cbor, auth_data, client_data, signature, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)',
  )

  router.post('/signing/finish', async (req, res) => {
    const correlationId = req.id
    const finishError = (status, code, message, reason, txSessionId, err) => {
      if (txSessionId) store.delete(txSessionId)
      if (reason || err) {
        logTxFinishError({ correlation_id: correlationId, reason }, err)
      }
      return writeError(res, status, code, message, correlationId)
    }

    try {
      if (!req.session || !req.session.acct_cbor) {
        return writeError(res, 401, 'unauthorized', 'Unauthorized', correlationId)
      }
      const body = req.body
      if (!body || typeof body !== 'object') {
        return writeError(res, 400, 'bad_request', 'Bad request', correlationId)
      }

      const { tx_session_id: txSessionId, ...responsePayload } = body
      if (typeof txSessionId !== 'string' || txSessionId.length === 0) {
        return writeError(res, 400, 'bad_request', 'Bad request', correlationId)
      }

      const txSession = store.get(txSessionId)
      const nowSeconds = now()
      if (!txSession || typeof txSession.expiresAt !== 'number' || txSession.expiresAt <= nowSeconds) {
        store.delete(txSessionId)
        return writeError(res, 401, 'unauthorized', 'Unauthorized', correlationId)
      }

      const accountCborBuffer = Buffer.isBuffer(req.session.acct_cbor)
        ? Buffer.from(req.session.acct_cbor)
        : Buffer.from(req.session.acct_cbor)

      if (!txSession.acctCbor || !Buffer.from(txSession.acctCbor).equals(accountCborBuffer)) {
        return finishError(401, 'unauthorized', 'Unauthorized', 'acct_mismatch', txSessionId)
      }

      const rawIdBase64 = typeof responsePayload.rawId === 'string' && responsePayload.rawId.length > 0
        ? responsePayload.rawId
        : typeof responsePayload.id === 'string'
          ? responsePayload.id
          : null
      if (!rawIdBase64) {
        return finishError(400, 'bad_request', 'Bad request', 'missing_raw_id', txSessionId)
      }

      let credentialId
      let authenticatorData
      let clientDataJSON
      let signature
      try {
        credentialId = base64urlToBuffer(rawIdBase64)
        if (!responsePayload.response || typeof responsePayload.response !== 'object') {
          throw new Error('missing_response')
        }
        const resp = responsePayload.response
        authenticatorData = base64urlToBuffer(resp.authenticatorData)
        clientDataJSON = base64urlToBuffer(resp.clientDataJSON)
        signature = base64urlToBuffer(resp.signature)
      } catch (err) {
        return finishError(400, 'bad_request', 'Bad request', 'payload_decode', txSessionId, err)
      }

      const allowed = Array.isArray(txSession.credentialIds)
        ? txSession.credentialIds.some((id) => Buffer.from(id).equals(credentialId))
        : false
      if (!allowed) {
        return finishError(401, 'unauthorized', 'Unauthorized', 'credential_not_allowed', txSessionId)
      }

      const credentialRow = selectCredential.get(credentialId, accountCborBuffer)
      if (!credentialRow) {
        return finishError(401, 'unauthorized', 'Unauthorized', 'credential_not_found', txSessionId)
      }

      const expectedOrigins = uniqueStrings(config.ORIGIN, ...(config.ORIGIN_ALLOWLIST || []))
      const expectedRPIDs = uniqueStrings(config.RP_ID, ...(config.RP_ID_ALLOWLIST || []))

      let verification
      try {
        verification = await verifier({
          response: responsePayload,
          expectedChallenge: txSession.challenge.toString('base64url'),
          expectedOrigin: expectedOrigins.length === 1 ? expectedOrigins[0] : expectedOrigins,
          expectedRPID: expectedRPIDs.length === 1 ? expectedRPIDs[0] : expectedRPIDs,
          requireUserVerification: true,
          authenticator: {
            credentialID: credentialId,
            credentialPublicKey: Buffer.from(txSession.acctCbor),
            counter: credentialRow.sign_count || 0,
          },
        })
      } catch (err) {
        return finishError(401, 'unauthorized', 'Unauthorized', 'verification_exception', txSessionId, err)
      }

      const info = verification?.authenticationInfo
      if (!verification?.verified || !info) {
        return finishError(401, 'unauthorized', 'Unauthorized', 'verification_failed', txSessionId)
      }
      if (!info.userVerified) {
        return finishError(403, 'policy_violation', 'Policy violation', 'uv_required', txSessionId)
      }

      const newCounter = typeof info.newCounter === 'number' ? info.newCounter : 0
      const storedCounter = typeof credentialRow.sign_count === 'number' ? credentialRow.sign_count : 0
      if (newCounter > 0 && newCounter <= storedCounter) {
        return finishError(409, 'conflict', 'Conflict', 'sign_count_regression', txSessionId)
      }

      let bundleMeta
      try {
        bundleMeta = decodeBundleMeta(Buffer.from(txSession.canonical))
      } catch (err) {
        return finishError(500, 'internal_error', 'Internal server error', 'bundle_decode', txSessionId, err)
      }

      const txIdBuffer = Buffer.from(txSession.txId)
      const canonicalBuffer = Buffer.from(txSession.canonical)
      const authDataBuffer = Buffer.from(authenticatorData)
      const clientDataBuffer = Buffer.from(clientDataJSON)
      const signatureBuffer = Buffer.from(signature)

      const writeTransaction = db.transaction((counterToPersist) => {
        if (counterToPersist > storedCounter) {
          updateSignCount.run(counterToPersist, credentialId)
        }
        insertTransaction.run(
          txIdBuffer,
          accountCborBuffer,
          bundleMeta.nonce,
          bundleMeta.message,
          canonicalBuffer,
          authDataBuffer,
          clientDataBuffer,
          signatureBuffer,
          nowSeconds,
        )
      })

      try {
        writeTransaction(newCounter > 0 ? newCounter : storedCounter)
      } catch (err) {
        const message = String(err?.message || '')
        const isConstraint = message.includes('UNIQUE') || message.includes('constraint')
        if (isConstraint) {
          return finishError(409, 'conflict', 'Conflict', 'transaction_conflict', txSessionId, err)
        }
        return finishError(500, 'internal_error', 'Internal server error', 'db_error', txSessionId, err)
      }

      store.delete(txSessionId)

      const accountThumbHex = computeAccountThumbHex(accountCborBuffer)
      const credentialIdHash = hashIdentifier(credentialId)
      const finalCounter = newCounter > 0 ? newCounter : storedCounter

      logTxFinishSuccess({
        correlation_id: correlationId,
        tx_id_hex: txIdBuffer.toString('hex'),
        account_thumb_hex: accountThumbHex,
        credential_id_hash: credentialIdHash,
        nonce: bundleMeta.nonce,
        sign_count: finalCounter,
      })

      res.status(201).json({ tx_id_hex: txIdBuffer.toString('hex') })
    } catch (err) {
      logTxFinishError({ correlation_id: req.id }, err)
      writeError(res, 500, 'internal_error', 'Internal server error', req.id)
    }
  })

  return { router, store }
}

export default createTxFinishRoutes
