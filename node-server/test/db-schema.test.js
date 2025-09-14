import test from 'node:test'
import assert from 'node:assert/strict'
import fs from 'node:fs'
import os from 'node:os'
import path from 'node:path'
import openDB, { applyMigrations } from '../src/db.js'

function tempDbPath() {
  const dir = fs.mkdtempSync(path.join(os.tmpdir(), 'dbtest-'))
  return path.join(dir, 'test.db')
}

test('applyMigrations creates expected tables and enforces PRAGMAs', () => {
  const p = tempDbPath()
  const db = openDB(p)
  applyMigrations(db, path.resolve(path.dirname(new URL(import.meta.url).pathname), '../src/migrations.sql'))
  const fk = db.pragma('foreign_keys', { simple: true })
  assert.equal(fk, 1)
  const jm = db.pragma('journal_mode', { simple: true })
  assert.equal(String(jm).toLowerCase(), 'wal')
  const tables = db
    .prepare("select name from sqlite_master where type='table' order by name")
    .all()
    .map((r) => r.name)
  assert.deepEqual(tables, ['accounts', 'credentials', 'sessions', 'transactions'])
  const credFk = db.pragma("foreign_key_list('credentials')")
  assert.ok(Array.isArray(credFk) && credFk.some((r) => r.table === 'accounts' && r.from === 'acct_cbor_fk' && r.to === 'acct_cbor'))
  db.close()
})

test('applyMigrations throws for missing file', () => {
  const p = tempDbPath()
  const db = openDB(p)
  assert.throws(() => applyMigrations(db, '/no/such/file.sql'))
  db.close()
})

