import test from 'node:test'
import assert from 'node:assert/strict'
import fs from 'node:fs'
import os from 'node:os'
import path from 'node:path'

import openDB from '../src/db.js'

test('openDB :memory: uses in-memory database without creating files', () => {
  const memoryPath = ':memory:'
  const resolved = path.resolve(process.cwd(), memoryPath)
  if (fs.existsSync(resolved)) {
    fs.rmSync(resolved)
  }

  const db = openDB(memoryPath)
  try {
    db.exec('CREATE TABLE sanity(id INTEGER PRIMARY KEY)')
    const row = db.prepare('SELECT COUNT(*) AS count FROM sqlite_master').get()
    assert.equal(typeof row.count, 'number')
  } finally {
    db.close()
  }

  assert.equal(fs.existsSync(resolved), false)
})

test('independent :memory: connections are isolated', () => {
  const db1 = openDB(':memory:')
  db1.exec('CREATE TABLE demo(id INTEGER PRIMARY KEY)')
  db1.prepare('INSERT INTO demo(id) VALUES (1)').run()
  const firstCount = db1.prepare('SELECT COUNT(*) AS count FROM demo').get().count
  assert.equal(firstCount, 1)
  db1.close()

  const db2 = openDB(':memory:')
  try {
    const row = db2.prepare("SELECT name FROM sqlite_master WHERE type='table' AND name='demo'").get()
    assert.equal(row, undefined)
  } finally {
    db2.close()
  }
})

test('openDB ensures parent directories exist for file paths', () => {
  const tmpDir = fs.mkdtempSync(path.join(os.tmpdir(), 'node-server-db-'))
  const filePath = path.join(tmpDir, 'nested', 'doc.db')
  const db = openDB(filePath)
  try {
    db.exec('CREATE TABLE sanity(id INTEGER PRIMARY KEY)')
  } finally {
    db.close()
  }

  assert.equal(fs.existsSync(path.dirname(filePath)), true)
  fs.rmSync(tmpDir, { recursive: true, force: true })
})
