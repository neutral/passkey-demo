import { createHash } from 'node:crypto'
import Database from 'better-sqlite3'
import { Encoder, Decoder, Tag } from 'cbor-x'

const CHALLENGE_PREFIX = Buffer.from('CHALv1', 'utf8')
const TX_ID_PREFIX = Buffer.from('TXIDv1', 'utf8')
const MAX_MESSAGE_BYTES = 1024
const MAX_SAFE_NONCE = Number.MAX_SAFE_INTEGER

export const ERROR_KINDS = {
  BUNDLE_BASE64: 'bundle_base64',
  BUNDLE_CBOR: 'bundle_cbor',
  SENDER_KEY_MISMATCH: 'sender_key_mismatch',
  NONCE_NOT_MONOTONIC: 'nonce_not_monotonic',
  MESSAGE_TOO_LONG: 'message_too_long',
  NONCE_OUT_OF_RANGE: 'nonce_out_of_range',
}

const canonicalEncoder = new Encoder({ canonical: true, structuredClone: false, useRecords: false, mapsAsObjects: false })
const canonicalDecoder = new Decoder({ useMaps: true, mapsAsObjects: false })
const looseDecoder = new Decoder({ useMaps: true })

const selectMaxNonceStmtByDB = new WeakMap()

export class BundleValidationError extends Error {
  constructor(kind, message) {
    super(message)
    this.name = 'BundleValidationError'
    this.kind = kind
  }
}

const REQUIRED_COSE_KEYS = [1, 3, -1, -2, -3]

function toBuffer(value) {
  if (Buffer.isBuffer(value)) return value
  if (value instanceof Uint8Array) return Buffer.from(value)
  return Buffer.from(value)
}

function readMapLike(container, key) {
  if (!container) return undefined
  if (container instanceof Map) {
    if (container.has(key)) return container.get(key)
    if (typeof key === 'number') {
      const asString = String(key)
      if (container.has(asString)) return container.get(asString)
    }
    return undefined
  }
  if (typeof container === 'object' && container !== null) {
    if (Object.prototype.hasOwnProperty.call(container, key)) return container[key]
    if (typeof key === 'number') {
      const asString = String(key)
      if (Object.prototype.hasOwnProperty.call(container, asString)) return container[asString]
    }
  }
  return undefined
}

function decodeBase64Url(input) {
  try {
    if (typeof input !== 'string' || input.length === 0) {
      throw new Error('invalid')
    }
    const buffer = Buffer.from(input, 'base64url')
    const normalized = buffer.toString('base64url')
    const trimmed = input.replace(/=+$/, '')
    if (normalized !== trimmed) {
      throw new Error('mismatch')
    }
    return buffer
  } catch {
    throw new BundleValidationError(ERROR_KINDS.BUNDLE_BASE64, 'invalid bundle base64url')
  }
}

function decodeBundle(rawBytes) {
  try {
    const decoded = canonicalDecoder.decode(rawBytes)
    if (!(decoded instanceof Map)) {
      throw new Error('bundle must be CBOR map')
    }
    return decoded
  } catch {
    throw new BundleValidationError(ERROR_KINDS.BUNDLE_CBOR, 'invalid bundle CBOR')
  }
}

function coerceCoseFields(value) {
  if (!value) return null
  if (value instanceof Tag) return coerceCoseFields(value.content)
  if (value instanceof Uint8Array || Buffer.isBuffer(value)) {
    try {
      return coerceCoseFields(canonicalDecoder.decode(value))
    } catch {
      return null
    }
  }
  const fields = {}
  if (value instanceof Map) {
    const shortcutKeys = ['kty', 'alg', 'crv', 'x', 'y']
    const hasShortcut = shortcutKeys.every((key) => value.has(key))
    if (hasShortcut) {
      return {
        kty: Number(value.get('kty')),
        alg: Number(value.get('alg')),
        crv: Number(value.get('crv')),
        x: toBuffer(value.get('x')),
        y: toBuffer(value.get('y')),
      }
    }
    for (const [rawKey, rawValue] of value.entries()) {
      const key = typeof rawKey === 'string' ? Number(rawKey) : rawKey
      if (!Number.isFinite(key)) continue
      fields[key] = rawValue
    }
  } else if (typeof value === 'object' && value !== null) {
    const shortcutKeys = ['kty', 'alg', 'crv', 'x', 'y']
    const hasShortcutShape = shortcutKeys.every((key) => key in value)
    if (hasShortcutShape) {
      return {
        kty: Number(value.kty),
        alg: Number(value.alg),
        crv: Number(value.crv),
        x: toBuffer(value.x),
        y: toBuffer(value.y),
      }
    }
    for (const rawKey of Object.keys(value)) {
      const key = Number(rawKey)
      if (!Number.isFinite(key)) continue
      fields[key] = value[rawKey]
    }
  } else {
    return null
  }
  for (const required of REQUIRED_COSE_KEYS) {
    if (!(required in fields)) return null
  }
  return {
    kty: Number(fields[1]),
    alg: Number(fields[3]),
    crv: Number(fields[-1]),
    x: toBuffer(fields[-2]),
    y: toBuffer(fields[-3]),
  }
}

function fallbackSender(rawBytes) {
  try {
    const loose = looseDecoder.decode(rawBytes)
    const candidate = readMapLike(loose, 0)
    return coerceCoseFields(candidate)
  } catch {
    return null
  }
}

function ensureSenderFields(bundleMap, rawBytes) {
  const direct = readMapLike(bundleMap, 0)
  let fields = coerceCoseFields(direct)
  if (!fields) {
    fields = fallbackSender(rawBytes)
  }
  if (!fields) {
    throw new BundleValidationError(ERROR_KINDS.BUNDLE_CBOR, 'invalid sender key')
  }
  return fields
}

function decodeUnsigned(value) {
  if (typeof value === 'number') {
    if (!Number.isFinite(value) || !Number.isInteger(value) || value < 0) {
      throw new BundleValidationError(ERROR_KINDS.BUNDLE_CBOR, 'invalid unsigned integer')
    }
    if (value > MAX_SAFE_NONCE) {
      throw new BundleValidationError(ERROR_KINDS.NONCE_OUT_OF_RANGE, 'nonce out of range')
    }
    return value
  }
  if (typeof value === 'bigint') {
    if (value < 0n) {
      throw new BundleValidationError(ERROR_KINDS.BUNDLE_CBOR, 'invalid unsigned integer')
    }
    if (value > BigInt(MAX_SAFE_NONCE)) {
      throw new BundleValidationError(ERROR_KINDS.NONCE_OUT_OF_RANGE, 'nonce out of range')
    }
    return Number(value)
  }
  throw new BundleValidationError(ERROR_KINDS.BUNDLE_CBOR, 'invalid unsigned integer')
}

function decodeMessage(value) {
  if (typeof value !== 'string') {
    throw new BundleValidationError(ERROR_KINDS.BUNDLE_CBOR, 'invalid message')
  }
  return value
}

function decodeOptionalUnsigned(value) {
  if (value === undefined || value === null) return undefined
  return decodeUnsigned(value)
}

function enforceMessage(message) {
  if (Buffer.byteLength(message, 'utf8') > MAX_MESSAGE_BYTES) {
    throw new BundleValidationError(ERROR_KINDS.MESSAGE_TOO_LONG, 'message too long')
  }
}

function enforceNonceBounds(nonce) {
  if (nonce <= 0) {
    throw new BundleValidationError(ERROR_KINDS.BUNDLE_CBOR, 'invalid nonce')
  }
}

function decodeAccountCose(acctCbor) {
  try {
    const decoded = canonicalDecoder.decode(acctCbor)
    const fields = coerceCoseFields(decoded)
    if (!fields) {
      throw new Error('missing cose fields')
    }
    return fields
  } catch {
    throw new BundleValidationError(ERROR_KINDS.BUNDLE_CBOR, 'invalid account CBOR')
  }
}

function enforceAccountBinding(accountFields, senderFields) {
  if (
    accountFields.kty !== senderFields.kty ||
    accountFields.alg !== senderFields.alg ||
    accountFields.crv !== senderFields.crv ||
    !accountFields.x.equals(senderFields.x) ||
    !accountFields.y.equals(senderFields.y)
  ) {
    throw new BundleValidationError(ERROR_KINDS.SENDER_KEY_MISMATCH, 'sender key mismatch')
  }
}

function getSelectMaxNonceStmt(db) {
  if (!(db instanceof Database)) {
    throw new TypeError('db must be a better-sqlite3 Database instance')
  }
  let stmt = selectMaxNonceStmtByDB.get(db)
  if (!stmt) {
    stmt = db.prepare('SELECT MAX(nonce) AS max_nonce FROM transactions WHERE acct_cbor = ?')
    selectMaxNonceStmtByDB.set(db, stmt)
  }
  return stmt
}

function enforceNonceMonotonic(db, acctCbor, nonce) {
  const stmt = getSelectMaxNonceStmt(db)
  const row = stmt.get(acctCbor)
  if (!row) return
  const { max_nonce: maxNonce } = row
  if (maxNonce === null || maxNonce === undefined) return
  const last = Number(maxNonce)
  if (Number.isFinite(last) && nonce <= last) {
    throw new BundleValidationError(ERROR_KINDS.NONCE_NOT_MONOTONIC, 'nonce must increase')
  }
}

function buildCanonicalBundle(senderFields, nonce, message, validUntil) {
  const senderMap = new Map()
  senderMap.set('1', senderFields.kty)
  senderMap.set('3', senderFields.alg)
  senderMap.set('-1', senderFields.crv)
  senderMap.set('-2', senderFields.x)
  senderMap.set('-3', senderFields.y)

  const bundleMap = new Map()
  bundleMap.set('0', senderMap)
  bundleMap.set('1', nonce)
  bundleMap.set('2', message)
  if (validUntil !== undefined) {
    bundleMap.set('3', validUntil)
  }

  return canonicalEncoder.encode(bundleMap)
}

function deriveAnchor(prefix, canonicalBundle) {
  return createHash('sha256').update(prefix).update(canonicalBundle).digest()
}

export function validateAndAnchorBundle(db, acctCborBytes, bundleBase64Url) {
  if (!db || typeof db.prepare !== 'function') {
    throw new TypeError('db must be a better-sqlite3 Database instance')
  }
  const acctCbor = Buffer.isBuffer(acctCborBytes) ? acctCborBytes : Buffer.from(acctCborBytes)
  const bundleRaw = decodeBase64Url(bundleBase64Url)
  const bundleMap = decodeBundle(bundleRaw)

  const senderFields = ensureSenderFields(bundleMap, bundleRaw)
  const nonce = decodeUnsigned(readMapLike(bundleMap, 1))
  const message = decodeMessage(readMapLike(bundleMap, 2))
  const validUntil = decodeOptionalUnsigned(readMapLike(bundleMap, 3))

  enforceMessage(message)
  enforceNonceBounds(nonce)

  const canonicalBundle = buildCanonicalBundle(senderFields, nonce, message, validUntil)

  const accountFields = decodeAccountCose(acctCbor)
  enforceAccountBinding(accountFields, senderFields)
  enforceNonceMonotonic(db, acctCbor, nonce)

  const challenge = deriveAnchor(CHALLENGE_PREFIX, canonicalBundle)
  const txId = deriveAnchor(TX_ID_PREFIX, canonicalBundle)

  return {
    bundle: {
      senderKey: senderFields,
      nonce,
      message,
      validUntil,
    },
    canonical: canonicalBundle,
    challenge,
    txId,
  }
}

export default validateAndAnchorBundle
