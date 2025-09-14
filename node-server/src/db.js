import Database from 'better-sqlite3'

export function openDB(path) {
  const db = new Database(path)
  // Apply recommended PRAGMAs similar to Go server
  db.pragma('journal_mode = WAL')
  db.pragma('synchronous = NORMAL')
  db.pragma('foreign_keys = ON')
  return db
}

export default openDB

