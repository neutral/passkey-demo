// Base64url and UTF-8 helpers for browser/Node (Playwright) environments.
// - URL-safe per RFC 4648: '+' -> '-', '/' -> '_', no padding on encode
// - Decoder accepts optional '=' padding; rejects invalid characters

function toUint8(input: ArrayBuffer | Uint8Array): Uint8Array {
  return input instanceof Uint8Array ? input : new Uint8Array(input)
}

function encodeBase64(bytes: Uint8Array): string {
  // Prefer browser btoa for performance; fallback to Node Buffer in tests
  if (typeof (globalThis as any).btoa === 'function') {
    let binary = ''
    // Chunk to avoid call stack / argument length limits
    const chunkSize = 0x8000 // 32 KiB
    for (let i = 0; i < bytes.length; i += chunkSize) {
      const sub = bytes.subarray(i, i + chunkSize)
      binary += String.fromCharCode(...sub)
    }
    return (globalThis as any).btoa(binary)
  }
  const BufferAny = (globalThis as any).Buffer
  if (BufferAny && typeof BufferAny.from === 'function') {
    return BufferAny.from(bytes).toString('base64')
  }
  throw new Error('No base64 encoder available')
}

function decodeBase64(b64: string): Uint8Array {
  if (typeof (globalThis as any).atob === 'function') {
    const binary = (globalThis as any).atob(b64)
    const out = new Uint8Array(binary.length)
    for (let i = 0; i < binary.length; i++) out[i] = binary.charCodeAt(i)
    return out
  }
  const BufferAny = (globalThis as any).Buffer
  if (BufferAny && typeof BufferAny.from === 'function') {
    const buf = BufferAny.from(b64, 'base64')
    return new Uint8Array(buf.buffer, buf.byteOffset, buf.byteLength)
  }
  throw new Error('No base64 decoder available')
}

export function bytesToBase64url(input: ArrayBuffer | Uint8Array): string {
  const bytes = toUint8(input)
  const b64 = encodeBase64(bytes)
  return b64.replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/g, '')
}

export function base64urlToBytes(s: string): Uint8Array {
  if (typeof s !== 'string') throw new TypeError('Expected string')
  if (s.length === 0) return new Uint8Array(0)
  // Reject whitespace and invalid characters; allow up to two '=' at end (optional padding)
  if (!/^[A-Za-z0-9\-_]+={0,2}$/.test(s)) {
    throw new TypeError('Invalid base64url characters')
  }
  // Convert to base64 and normalize padding
  let b64 = s.replace(/-/g, '+').replace(/_/g, '/')
  const remainder = b64.length % 4
  if (remainder === 2) b64 += '=='
  else if (remainder === 3) b64 += '='
  else if (remainder === 1) throw new TypeError('Invalid base64url length')
  // remainder 0: leave as-is (may already include padding)
  return decodeBase64(b64)
}

export function utf8ToBytes(s: string): Uint8Array {
  return new TextEncoder().encode(s)
}

export function bytesToUtf8(b: ArrayBuffer | Uint8Array): string {
  const bytes = toUint8(b)
  return new TextDecoder('utf-8', { fatal: true }).decode(bytes)
}

