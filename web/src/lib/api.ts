import { parseHttpError } from './http'

export class ApiError extends Error {
  status: number
  code?: string
  correlationId?: string
  constructor(status: number, message: string, code?: string, correlationId?: string) {
    super(message)
    this.status = status
    this.code = code
    this.correlationId = correlationId
  }
}

export async function apiFetch<T>(url: string, init: RequestInit = {}): Promise<T> {
  const opts: RequestInit = {
    ...init,
    mode: 'cors',
    credentials: 'include',
    headers: { 'Content-Type': 'application/json', ...(init.headers || {}) },
  }
  const r = await fetch(url, opts)
  if (!r.ok) {
    let code: string | undefined
    let corr: string | undefined
    try {
      const data = await r.clone().json()
      if (data && typeof data === 'object') {
        code = (data as any).code
        corr = (data as any).correlation_id
      }
    } catch {}
    const parsed = await parseHttpError(r)
    throw new ApiError(r.status, parsed.detail || parsed.title, code || parsed.code, corr)
  }
  return (await r.json()) as T
}

export async function postJson<T>(url: string, body: any): Promise<T> {
  return apiFetch<T>(url, { method: 'POST', body: JSON.stringify(body) })
}

export function formatApiError(err: ApiError): string {
  const prefix = `HTTP ${err.status}`
  const detail = typeof err.message === 'string' ? err.message.trim() : ''
  const code = err.code ? ` (${err.code})` : ''
  if (detail.length > 0 && !detail.startsWith('HTTP ')) {
    return `${prefix} — ${detail}${code}`
  }
  return `${detail.length > 0 ? detail : prefix}${code}`
}
