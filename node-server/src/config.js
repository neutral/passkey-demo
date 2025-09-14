// Environment configuration loader for Node server (ESM)

export function loadConfig(env = process.env) {
  const cfg = {
    RP_ID: String(env.RP_ID || '').trim(),
    ORIGIN: String(env.ORIGIN || '').trim(),
    PORT: Number(env.PORT || 8080),
    DB_PATH: String(env.DB_PATH || 'server/demo-node.db'),
    RP_ID_ALLOWLIST: parseList(env.RP_ID_ALLOWLIST),
    ORIGIN_ALLOWLIST: parseList(env.ORIGIN_ALLOWLIST),
  }
  if (!Number.isFinite(cfg.PORT) || cfg.PORT <= 0) {
    throw new Error('Invalid PORT')
  }
  // RP_ID and ORIGIN are required for real flows; allow missing in early scaffold but warn.
  if (!cfg.RP_ID) console.warn('[config] RP_ID not set (using scaffold).')
  if (!cfg.ORIGIN) console.warn('[config] ORIGIN not set (using scaffold).')
  return Object.freeze(cfg)
}

function parseList(v) {
  if (!v) return []
  const s = String(v)
  return s
    .split(',')
    .map((x) => x.trim())
    .filter(Boolean)
}

export default loadConfig

