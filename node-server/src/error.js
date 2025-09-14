// JSON error envelope helper

export function writeError(res, status, code, message, correlationId) {
  res.status(status)
  res.setHeader('Content-Type', 'application/json')
  const body = { code, error: message }
  if (correlationId) body.correlation_id = correlationId
  res.end(JSON.stringify(body))
}

export default writeError

