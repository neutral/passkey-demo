import Database from 'better-sqlite3'
import fs from 'node:fs'
import path from 'node:path'

export function openDB(path) {
  const db = new Database(path)
  // Apply recommended PRAGMAs similar to Go server
  db.pragma('journal_mode = WAL')
  db.pragma('synchronous = NORMAL')
  db.pragma('foreign_keys = ON')
  return db
}

export function applyMigrations(db, sqlPath) {
  const p = path.resolve(sqlPath)
  if (!fs.existsSync(p)) throw new Error(`migrations file not found: ${p}`)
  const sql = fs.readFileSync(p, 'utf8')
  db.exec('BEGIN')
  try {
    db.exec(sql)
    db.exec('COMMIT')
  } catch (e) {
    try { db.exec('ROLLBACK') } catch {}
    throw e
  }
}

export default openDB
