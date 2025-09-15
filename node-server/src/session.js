import cookie from 'cookie'

export function sessionMiddleware(db, opts = {}) {
  const rollingTTLSeconds = opts.rollingTTLSeconds || 0 // 0 = disabled
  return function session(req, res, next) {
    const header = req.headers['cookie']
    if (typeof header !== 'string' || header.length === 0) return next()
    let sid = ''
    try {
      const parsed = cookie.parse(header)
      sid = parsed.sid || ''
    } catch {
      // ignore malformed cookies
    }
    if (!sid) return next()
    try {
      const row = db.prepare('SELECT acct_cbor, expires_at FROM sessions WHERE session_id = ?').get(sid)
      const now = Math.floor(Date.now() / 1000)
      if (!row || typeof row.expires_at !== 'number' || row.expires_at <= now) return next()
      const ctx = { sid, acct_cbor: row.acct_cbor, expires_at: row.expires_at }
      if (opts.preloadCredentialIds) {
        const ids = db.prepare('SELECT credential_id FROM credentials WHERE acct_cbor_fk = ?').all(row.acct_cbor)
        ctx.credential_ids = ids.map((r) => r.credential_id)
      }
      req.session = ctx
      // Optional rolling refresh when within window
      if (rollingTTLSeconds > 0 && row.expires_at - now < rollingTTLSeconds) {
        const newExp = now + rollingTTLSeconds
        db.prepare('UPDATE sessions SET expires_at = ? WHERE session_id = ?').run(newExp, sid)
        req.session.expires_at = newExp
      }
    } catch {
      // swallow DB errors here; protected routes will enforce auth and map errors
    }
    return next()
  }
}

export default sessionMiddleware

