import test from 'node:test'
import assert from 'node:assert/strict'
import fs from 'node:fs'
import { fileURLToPath } from 'node:url'
import path from 'node:path'
import Database from 'better-sqlite3'
import { Encoder, Decoder } from 'cbor-x'

import { applyMigrations } from '../src/db.js'
import {
  validateAndAnchorBundle,
  BundleValidationError,
  ERROR_KINDS,
} from '../src/tx/bundle.js'

const canonicalEncoder = new Encoder({ canonical: true, structuredClone: false, useRecords: false, mapsAsObjects: false })
const canonicalDecoder = new Decoder({ useMaps: true, mapsAsObjects: false })

const MIGRATIONS_PATH = fileURLToPath(new URL('../src/migrations.sql', import.meta.url))
const GOLDEN_PATH = fileURLToPath(new URL('../../specs/goldens/tx-bundle-v1.json', import.meta.url))
const GOLDEN = JSON.parse(fs.readFileSync(GOLDEN_PATH, 'utf8'))

function openTestDB() {
  const db = new Database(':memory:')
  applyMigrations(db, MIGRATIONS_PATH)
  return db
}

function normalizeNumericMap(map) {
  const out = new Map()
  for (const [key, value] of map.entries()) {
    const numeric = typeof key === 'string' ? Number(key) : key
    out.set(Number.isFinite(numeric) ? numeric : key, value)
  }
  return out
}

function mutateBundle(base64, mutator) {
  const raw = Buffer.from(base64, 'base64url')
  const decoded = canonicalDecoder.decode(raw)
  const working = normalizeNumericMap(decoded)
  mutator(working)
  return canonicalEncoder.encode(working).toString('base64url')
}

test('validateAndAnchorBundle matches golden vectors', () => {
  const db = openTestDB()
  try {
    const acctCbor = Buffer.from(GOLDEN.inputs.sender_key_cbor_b64, 'base64url')
    const result = validateAndAnchorBundle(db, acctCbor, GOLDEN.bundle.bundle_cbor_b64)

    const acctDecoded = normalizeNumericMap(canonicalDecoder.decode(acctCbor))
    const acctX = Buffer.from(acctDecoded.get(-2))
    const acctY = Buffer.from(acctDecoded.get(-3))

    assert.equal(result.bundle.nonce, GOLDEN.inputs.nonce)
    assert.equal(result.bundle.message, GOLDEN.inputs.message)
    assert.deepEqual(result.bundle.validUntil, undefined)

    assert.equal(result.canonical.toString('base64url'), GOLDEN.bundle.bundle_cbor_b64)
    assert.equal(result.challenge.toString('base64url'), GOLDEN.anchors.challenge_b64)
    assert.equal(result.txId.toString('hex'), GOLDEN.anchors.tx_id_hex)
    assert.equal(result.bundle.senderKey.kty, 2)
    assert.equal(result.bundle.senderKey.alg, -7)
    assert.equal(result.bundle.senderKey.crv, 1)
    assert.equal(result.bundle.senderKey.x.toString('hex'), acctX.toString('hex'))
    assert.equal(result.bundle.senderKey.y.toString('hex'), acctY.toString('hex'))
  } finally {
    db.close()
  }
})

test('sender key mismatch yields SENDER_KEY_MISMATCH', () => {
  const db = openTestDB()
  try {
    const acct = Buffer.from(GOLDEN.inputs.sender_key_cbor_b64, 'base64url')
    const acctCorrupted = Buffer.from(acct)
    acctCorrupted[acctCorrupted.length - 1] ^= 0x01

    assert.throws(
      () => validateAndAnchorBundle(db, acctCorrupted, GOLDEN.bundle.bundle_cbor_b64),
      (err) => err instanceof BundleValidationError && err.kind === ERROR_KINDS.SENDER_KEY_MISMATCH,
    )
  } finally {
    db.close()
  }
})

test('nonce policy enforces strict monotonicity', () => {
  const db = openTestDB()
  try {
    const acct = Buffer.from(GOLDEN.inputs.sender_key_cbor_b64, 'base64url')
    db.prepare('INSERT INTO accounts (acct_cbor, acct_thumb, created_at) VALUES (?, ?, ?)').run(
      acct,
      Buffer.alloc(32),
      0,
    )
    const insertTx = db.prepare(
      'INSERT INTO transactions (tx_id, acct_cbor, nonce, message, bundle_cbor, auth_data, client_data, signature, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)',
    )
    insertTx.run(Buffer.from('prior'), acct, 10, 'hello', Buffer.from('B'), Buffer.from('AD'), Buffer.from('CD'), Buffer.from('SIG'), 0)

    const lowerNonce = mutateBundle(GOLDEN.bundle.bundle_cbor_b64, (map) => {
      map.set(1, 9)
    })
    assert.throws(
      () => validateAndAnchorBundle(db, acct, lowerNonce),
      (err) => err instanceof BundleValidationError && err.kind === ERROR_KINDS.NONCE_NOT_MONOTONIC,
    )

    const equalNonce = mutateBundle(GOLDEN.bundle.bundle_cbor_b64, (map) => {
      map.set(1, 10)
    })
    assert.throws(
      () => validateAndAnchorBundle(db, acct, equalNonce),
      (err) => err instanceof BundleValidationError && err.kind === ERROR_KINDS.NONCE_NOT_MONOTONIC,
    )

    const higherNonce = mutateBundle(GOLDEN.bundle.bundle_cbor_b64, (map) => {
      map.set(1, 11)
    })
    const res = validateAndAnchorBundle(db, acct, higherNonce)
    assert.equal(res.bundle.nonce, 11)
  } finally {
    db.close()
  }
})

test('message length and nonce range limits', () => {
  const db = openTestDB()
  try {
    const acct = Buffer.from(GOLDEN.inputs.sender_key_cbor_b64, 'base64url')

    const longMessage = mutateBundle(GOLDEN.bundle.bundle_cbor_b64, (map) => {
      map.set(2, 'x'.repeat(1025))
    })
    assert.throws(
      () => validateAndAnchorBundle(db, acct, longMessage),
      (err) => err instanceof BundleValidationError && err.kind === ERROR_KINDS.MESSAGE_TOO_LONG,
    )

    const hugeNonce = mutateBundle(GOLDEN.bundle.bundle_cbor_b64, (map) => {
      map.set(1, BigInt(Number.MAX_SAFE_INTEGER) + 1n)
    })
    assert.throws(
      () => validateAndAnchorBundle(db, acct, hugeNonce),
      (err) => err instanceof BundleValidationError && err.kind === ERROR_KINDS.NONCE_OUT_OF_RANGE,
    )
  } finally {
    db.close()
  }
})

test('base64 and CBOR errors surface typed BundleValidationError', () => {
  const db = openTestDB()
  try {
    const acct = Buffer.from(GOLDEN.inputs.sender_key_cbor_b64, 'base64url')

    assert.throws(
      () => validateAndAnchorBundle(db, acct, '@@not-base64@@'),
      (err) => err instanceof BundleValidationError && err.kind === ERROR_KINDS.BUNDLE_BASE64,
    )

    const invalidCbor = Buffer.from('deadbeef', 'hex').toString('base64url')
    assert.throws(
      () => validateAndAnchorBundle(db, acct, invalidCbor),
      (err) => err instanceof BundleValidationError && err.kind === ERROR_KINDS.BUNDLE_CBOR,
    )
  } finally {
    db.close()
  }
})

test('decoder fallback tolerates alternative sender key shapes', () => {
  const db = openTestDB()
  try {
    const acct = Buffer.from(GOLDEN.inputs.sender_key_cbor_b64, 'base64url')
    const senderFields = normalizeNumericMap(canonicalDecoder.decode(acct))
    const mutated = mutateBundle(GOLDEN.bundle.bundle_cbor_b64, (map) => {
      const objectSender = Object.create(null)
      objectSender.kty = senderFields.get(1)
      objectSender.alg = senderFields.get(3)
      objectSender.crv = senderFields.get(-1)
      objectSender.x = Buffer.from(senderFields.get(-2))
      objectSender.y = Buffer.from(senderFields.get(-3))
      map.set(0, objectSender)
    })
    const res = validateAndAnchorBundle(db, acct, mutated)
    assert.equal(res.bundle.nonce, GOLDEN.inputs.nonce)
  } finally {
    db.close()
  }
})
