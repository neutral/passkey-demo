import Database from 'better-sqlite3'
import fs from 'node:fs'
import path from 'node:path'

function isInMemory(dbPath) {
  if (!dbPath) return false
  if (dbPath === ':memory:') return true
  if (!dbPath.startsWith('file:')) return false
  // Handle URIs like file::memory:?cache=shared or file:memdb1?mode=memory
  const uri = dbPath.slice('file:'.length)
  if (uri.startsWith(':memory:')) return true
  return /mode=memory/i.test(dbPath)
}

export function openDB(dbPath) {
  const useMemory = isInMemory(dbPath)
  const targetPath = useMemory
    ? dbPath
    : path.isAbsolute(dbPath)
      ? dbPath
      : path.resolve(process.cwd(), dbPath)

  if (!useMemory) {
    const dir = path.dirname(targetPath)
    if (dir && dir !== '.' && !fs.existsSync(dir)) {
      fs.mkdirSync(dir, { recursive: true })
    }
  }

  const db = new Database(targetPath)
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
