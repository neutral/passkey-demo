export type ParsedHttpError = {
  title: string
  detail: string
  code?: string
  status: number
}

export async function parseHttpError(r: Response): Promise<ParsedHttpError> {
  const base: ParsedHttpError = {
    title: `HTTP ${r.status}`,
    detail: r.statusText || 'Request failed',
    status: r.status,
  }
  try {
    // Clone to avoid consuming the original if caller still needs it
    const data = await r.clone().json()
    if (data && typeof data === 'object') {
      const d = data as Record<string, any>
      const msg = d.error || d.message || d.detail || d.title
      if (typeof msg === 'string' && msg.trim()) base.detail = msg
      if (typeof d.code === 'string') base.code = d.code
    }
  } catch {
    // Non-JSON body; keep base title/detail
  }
  return base
}

export function normalizeError(e: unknown): { title: string; detail: string } {
  if (e && typeof e === 'object' && 'message' in (e as any)) {
    const msg = String((e as any).message || '')
    return { title: 'Error', detail: msg || 'Request failed' }
  }
  try {
    return { title: 'Error', detail: JSON.stringify(e) }
  } catch {
    return { title: 'Error', detail: String(e) }
  }
}

