import express from 'express'

import writeError from '../error.js'
import { logger } from '../logger.js'

function toBuffer(value) {
  if (Buffer.isBuffer(value)) return Buffer.from(value)
  if (value instanceof Uint8Array) return Buffer.from(value)
  if (Array.isArray(value)) return Buffer.from(value)
  if (value && typeof value === 'object' && 'buffer' in value) {
    try {
      return Buffer.from(value.buffer)
    } catch {}
  }
  return Buffer.alloc(0)
}

export function createTxListRoutes(config, deps = {}) {
  if (!config) throw new Error('createTxListRoutes requires config')
  const db = deps.db
  if (!db) throw new Error('createTxListRoutes requires db')
  const selectTransactions =
    deps.selectTransactions ??
    db.prepare(
      'SELECT tx_id, nonce, message, created_at FROM transactions WHERE acct_cbor = ? ORDER BY created_at DESC',
    )

  const router = express.Router()

  router.get('/list', (req, res) => {
    const correlationId = req.id
    const session = req.session
    const acctBuffer = session ? toBuffer(session.acct_cbor) : Buffer.alloc(0)

    if (!session || acctBuffer.length === 0) {
      return writeError(res, 401, 'unauthorized', 'Unauthorized', correlationId)
    }

    let rows
    try {
      rows = selectTransactions.all(acctBuffer)
    } catch (err) {
      logger.error({ event: 'tx_list_error', correlation_id: correlationId, err }, 'failed to read transactions')
      return writeError(res, 500, 'internal_error', 'Internal server error', correlationId)
    }

    const items = rows.map((row) => {
      const txId = row.tx_id ? Buffer.from(row.tx_id) : Buffer.alloc(0)
      const createdAtRaw = row.created_at
      const nonceRaw = row.nonce
      return {
        tx_id_hex: txId.toString('hex'),
        nonce: typeof nonceRaw === 'number' ? nonceRaw : Number(nonceRaw) || 0,
        message: typeof row.message === 'string' ? row.message : String(row.message ?? ''),
        created_at:
          typeof createdAtRaw === 'number'
            ? createdAtRaw
            : Number(createdAtRaw) || 0,
      }
    })

    return res.status(200).json({ items })
  })

  return { router }
}

export default createTxListRoutes
