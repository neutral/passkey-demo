import { encode, decode } from 'cbor-x'

// Encode a JS value to CBOR bytes. We rely on using Map with numeric keys
// for integer-key maps (e.g., COSE and Bundle), which cbor-x encodes as CBOR maps
// with integer keys. Encoding is deterministic for our usage.
export function encodeCanonical(value: any): Uint8Array {
  return encode(value)
}

export function decodeCBOR<T = any>(data: Uint8Array): T {
  return decode(data) as T
}

