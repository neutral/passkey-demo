// Credentialed CORS middleware with exact-origin allowlist

export function corsMiddleware(config) {
  const allowedOrigins = new Set([config.ORIGIN, ...config.ORIGIN_ALLOWLIST].filter(Boolean))
  const allowMethods = ['GET', 'POST', 'OPTIONS']
  const allowHeaders = ['content-type'] // case-insensitive check

  return function cors(req, res, next) {
    const origin = req.headers.origin
    const hasOrigin = typeof origin === 'string' && origin.length > 0
    const isAllowed = hasOrigin && allowedOrigins.has(origin)

    // Always set Vary: Origin when Origin header present
    if (hasOrigin) res.setHeader('Vary', 'Origin')

    if (req.method === 'OPTIONS') {
      if (!isAllowed) {
        res.statusCode = 403
        return res.end()
      }
      // Validate requested method and headers
      const reqMethod = (req.headers['access-control-request-method'] || '').toString().toUpperCase()
      const reqHeaders = (req.headers['access-control-request-headers'] || '')
        .toString()
        .split(',')
        .map((h) => h.trim().toLowerCase())
        .filter(Boolean)
      const methodsOk = allowMethods.includes(reqMethod)
      const headersOk = reqHeaders.every((h) => allowHeaders.includes(h))
      if (!methodsOk || !headersOk) {
        res.statusCode = 403
        return res.end()
      }
      res.statusCode = 204
      res.setHeader('Access-Control-Allow-Origin', origin)
      res.setHeader('Access-Control-Allow-Methods', allowMethods.join(', '))
      res.setHeader('Access-Control-Allow-Headers', 'Content-Type')
      res.setHeader('Access-Control-Allow-Credentials', 'true')
      res.setHeader('Access-Control-Max-Age', '600')
      return res.end()
    }

    if (hasOrigin && isAllowed) {
      res.setHeader('Access-Control-Allow-Origin', origin)
      res.setHeader('Access-Control-Allow-Credentials', 'true')
    }
    return next()
  }
}

export default corsMiddleware

