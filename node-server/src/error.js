export const ERROR_CODES = {
  BAD_REQUEST: 'ERR_BAD_REQUEST',
  UNAUTHORIZED: 'ERR_UNAUTHORIZED',
  FORBIDDEN: 'ERR_FORBIDDEN',
  CONFLICT: 'ERR_CONFLICT',
  PAYLOAD_TOO_LARGE: 'ERR_PAYLOAD_TOO_LARGE',
  RATE_LIMIT: 'ERR_RATE_LIMIT',
  INTERNAL: 'ERR_INTERNAL',
}

const DEFAULT_MESSAGES = {
  [ERROR_CODES.BAD_REQUEST]: 'Bad request',
  [ERROR_CODES.UNAUTHORIZED]: 'Unauthorized',
  [ERROR_CODES.FORBIDDEN]: 'Forbidden',
  [ERROR_CODES.CONFLICT]: 'Conflict',
  [ERROR_CODES.PAYLOAD_TOO_LARGE]: 'Payload too large',
  [ERROR_CODES.RATE_LIMIT]: 'Too many requests',
  [ERROR_CODES.INTERNAL]: 'Internal server error',
}

export function respondError(res, status, code, message, correlationId, details) {
  const responseCode = code || ERROR_CODES.INTERNAL
  const body = {
    code: responseCode,
    error: message || DEFAULT_MESSAGES[responseCode] || DEFAULT_MESSAGES[ERROR_CODES.INTERNAL],
  }
  if (correlationId) body.correlation_id = correlationId
  if (details && typeof details === 'object' && Object.keys(details).length > 0) {
    body.details = details
  }
  res.status(status).json(body)
  return res
}

export const respondBadRequest = (res, correlationId, message, details) =>
  respondError(res, 400, ERROR_CODES.BAD_REQUEST, message, correlationId, details)

export const respondUnauthorized = (res, correlationId, message) =>
  respondError(res, 401, ERROR_CODES.UNAUTHORIZED, message, correlationId)

export const respondForbidden = (res, correlationId, message) =>
  respondError(res, 403, ERROR_CODES.FORBIDDEN, message, correlationId)

export const respondConflict = (res, correlationId, message) =>
  respondError(res, 409, ERROR_CODES.CONFLICT, message, correlationId)

export const respondPayloadTooLarge = (res, correlationId, details) =>
  respondError(res, 413, ERROR_CODES.PAYLOAD_TOO_LARGE, DEFAULT_MESSAGES[ERROR_CODES.PAYLOAD_TOO_LARGE], correlationId, details)

export const respondRateLimit = (res, correlationId, retryAfterSeconds) => {
  if (typeof retryAfterSeconds === 'number' && Number.isFinite(retryAfterSeconds) && retryAfterSeconds > 0) {
    res.setHeader('Retry-After', String(Math.ceil(retryAfterSeconds)))
  }
  return respondError(res, 429, ERROR_CODES.RATE_LIMIT, DEFAULT_MESSAGES[ERROR_CODES.RATE_LIMIT], correlationId)
}

export const respondInternalError = (res, correlationId) =>
  respondError(res, 500, ERROR_CODES.INTERNAL, DEFAULT_MESSAGES[ERROR_CODES.INTERNAL], correlationId)

export function buildExpressErrorHandler() {
  return function errorHandler(err, req, res, next) {
    if (res.headersSent) return next(err)
    const correlationId = req?.id
    if (err && err.type === 'entity.too.large') {
      return respondPayloadTooLarge(res, correlationId, { max_bytes: err.limit })
    }
    if (err && err.type === 'entity.parse.failed') {
      return respondBadRequest(res, correlationId)
    }
    if (err && err.status === 413) {
      return respondPayloadTooLarge(res, correlationId)
    }
    if (err && err.status === 400) {
      return respondBadRequest(res, correlationId)
    }
    return respondInternalError(res, correlationId)
  }
}

export default respondError
