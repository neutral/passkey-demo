import { customAlphabet } from 'nanoid'

const nano = customAlphabet('0123456789abcdefghijklmnopqrstuvwxyz', 16)

export function requestIdMiddleware(req, res, next) {
  const incoming = req.headers['x-request-id']
  const id = typeof incoming === 'string' && incoming.trim() ? incoming.trim() : nano()
  req.id = id
  res.setHeader('X-Request-ID', id)
  next()
}

export default requestIdMiddleware

