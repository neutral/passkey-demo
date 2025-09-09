import { encodeCanonical } from './cbor'
import { bytesToBase64url } from './encoding'

// COSE EC2 public key (ES256 / P-256)
export type CoseEC2 = { kty: number; alg: number; crv: number; x: Uint8Array; y: Uint8Array }

// Build a logical Bundle as a CBOR map with integer keys using JS Map so keys are encoded as integers.
export function buildBundle(sender: CoseEC2, nonce: number, message: string, validUntil?: number): Map<number, any> {
  // COSE EC2 map with integer keys
  const cose = new Map<number, any>([
    [1, sender.kty],
    [3, sender.alg],
    [-1, sender.crv],
    [-2, sender.x],
    [-3, sender.y],
  ])
  const entries: [number, any][] = [
    [0, cose],
    [1, Number(nonce)],
    [2, String(message)],
  ]
  if (typeof validUntil === 'number') entries.push([3, Number(validUntil)])
  return new Map<number, any>(entries)
}

export function encodeBundleCanonical(bundle: Map<number, any>): Uint8Array {
  return encodeCanonical(bundle)
}

export function u8ToHex(u8: Uint8Array): string {
  return Array.from(u8)
    .map((b) => b.toString(16).padStart(2, '0'))
    .join('')
}

export function bundleToB64Hex(u8: Uint8Array): { b64: string; hex: string } {
  return { b64: bytesToBase64url(u8), hex: u8ToHex(u8) }
}
