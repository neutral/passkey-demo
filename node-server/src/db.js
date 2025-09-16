import Database from 'better-sqlite3'
import fs from 'node:fs'
import path from 'node:path'

export function openDB(dbPath) {
  // Resolve to absolute path and ensure parent directory exists
  const absPath = path.isAbsolute(dbPath) ? dbPath : path.resolve(process.cwd(), dbPath)
  const dir = path.dirname(absPath)
  if (dir && dir !== '.' && !fs.existsSync(dir)) {
    fs.mkdirSync(dir, { recursive: true })
  }
  const db = new Database(absPath)
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
