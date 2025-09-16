import { base64urlToBytes, bytesToBase64url } from './encoding'
import type {
  PublicKeyCredentialCreationOptionsJSON,
  PublicKeyCredentialRequestOptionsJSON,
  PublicKeyCredentialDescriptorJSON,
} from '@simplewebauthn/types'

type NativeDescriptor = PublicKeyCredentialDescriptor
type AuthenticatorTransport = NativeDescriptor['transports'] extends Array<infer X> ? X : never

type WithMeta<T> = T & { reg_session_id?: string; login_session_id?: string; tx_session_id?: string; expires_at?: number }

type AnyJson = Record<string, unknown>

export type RegistrationOptionsResponse = WithMeta<PublicKeyCredentialCreationOptionsJSON | AnyJson>

export type LoginOptionsResponse = WithMeta<PublicKeyCredentialRequestOptionsJSON | AnyJson>

export type TxOptionsResponse = WithMeta<AnyJson> & {
  tx_session_id: string
  tx_id_hex?: string
  challenge?: string
  options?: AnyJson
}

// Adapters: Convert Node server option shapes to @simplewebauthn/browser JSON shapes
export function toCreationOptionsJSON(
  resp: RegistrationOptionsResponse,
): PublicKeyCredentialCreationOptionsJSON {
  const flattened = flattenOptions(resp)
  const {
    reg_session_id: _ignoreSession,
    login_session_id: _ignoreLogin,
    tx_session_id: _ignoreTx,
    expires_at: _ignoreExpires,
    ...rest
  } = flattened
  return normalizeCreationOptions(rest)
}

export function toRequestOptionsJSON(
  resp: LoginOptionsResponse | TxOptionsResponse | PublicKeyCredentialRequestOptionsJSON,
): PublicKeyCredentialRequestOptionsJSON {
  const flattened = flattenOptions(resp)
  const {
    reg_session_id: _ignoreReg,
    login_session_id: _ignoreLogin,
    tx_session_id: _ignoreTx,
    expires_at: _ignoreExpires,
    ...rest
  } = flattened
  return normalizeRequestOptions(rest)
}

// Back-compat helper: convert JSON request options into native WebAuthn request options (ArrayBuffers, etc.)
export function toRequestOptions(
  resp: LoginOptionsResponse | TxOptionsResponse | PublicKeyCredentialRequestOptionsJSON,
): PublicKeyCredentialRequestOptions {
  const json = toRequestOptionsJSON(resp)
  const challengeBytes = base64urlToBytes(json.challenge)
  const out: PublicKeyCredentialRequestOptions = {
    challenge: challengeBytes,
    userVerification: json.userVerification || 'required',
  }
  if (json.rpId) out.rpId = json.rpId
  if (typeof json.timeout === 'number') out.timeout = json.timeout
  if (Array.isArray(json.allowCredentials) && json.allowCredentials.length > 0) {
    const descriptors = json.allowCredentials
      .map((cred) => {
        if (!cred || typeof cred.id !== 'string' || cred.id.length === 0) return null
        const descriptor: PublicKeyCredentialDescriptor = {
          type: 'public-key',
          id: base64urlToBytes(cred.id),
        }
        if (Array.isArray(cred.transports) && cred.transports.length > 0) {
          descriptor.transports = cred.transports.filter((t): t is AuthenticatorTransport => typeof t === 'string')
        }
        return descriptor
      })
      .filter((descriptor): descriptor is PublicKeyCredentialDescriptor => descriptor !== null)
    if (descriptors.length > 0) out.allowCredentials = descriptors
  }
  if (json.extensions && typeof json.extensions === 'object') {
    ;(out as any).extensions = json.extensions
  }
  return out
}

// Error mapping helper: present DOMException/NotAllowed consistently via normalizeError caller
export function mapDomException(e: unknown): { title: string; detail: string } {
  if (e && typeof e === 'object' && (e as any).name && (e as any).message) {
    const name = String((e as any).name)
    const msg = String((e as any).message)
    return { title: name, detail: msg }
  }
  return { title: 'Error', detail: String(e) }
}

export function buildTxFinish(cred: PublicKeyCredential, txSessionId: string) {
  const asr = cred.response as AuthenticatorAssertionResponse
  const rawId = bytesToBase64url(new Uint8Array(cred.rawId as ArrayBuffer))
  const authenticatorData = bytesToBase64url(new Uint8Array(asr.authenticatorData))
  const clientDataJSON = bytesToBase64url(new Uint8Array(asr.clientDataJSON))
  const signature = bytesToBase64url(new Uint8Array(asr.signature))
  const userHandle = asr.userHandle ? bytesToBase64url(new Uint8Array(asr.userHandle)) : ''

  return {
    tx_session_id: txSessionId,
    id: cred.id,
    rawId,
    type: cred.type,
    response: {
      authenticatorData,
      clientDataJSON,
      signature,
      userHandle,
    },
  }
}

function flattenOptions(input: unknown): AnyJson {
  if (!input || typeof input !== 'object') return {}
  const clone: AnyJson = Array.isArray(input) ? {} : { ...(input as AnyJson) }
  if ('options' in clone && clone.options && typeof clone.options === 'object') {
    const nested = flattenOptions(clone.options)
    delete clone.options
    return { ...nested, ...clone }
  }
  return clone
}

function normalizeCreationOptions(raw: AnyJson): PublicKeyCredentialCreationOptionsJSON {
  let challenge = ensureString(raw.challenge)
  if (!challenge) {
    const random = new Uint8Array(32)
    crypto.getRandomValues(random)
    challenge = bytesToBase64url(random)
  }
  const rp = normalizeRp(raw)
  const user = normalizeUser(raw.user)
  const pubKeyCredParams = normalizePubKeyCredParams(raw.pubKeyCredParams)
  const authenticatorSelection = normalizeAuthenticatorSelection(raw.authenticatorSelection)
  const attestation = normalizeAttestation(raw.attestation)

  const out: PublicKeyCredentialCreationOptionsJSON = {
    challenge,
    rp,
    user,
    pubKeyCredParams,
    authenticatorSelection,
    attestation,
  }

  const timeout = Number(raw.timeout)
  if (Number.isFinite(timeout) && timeout > 0) out.timeout = timeout

  const exclude = raw.excludeCredentials
  const normalizedExclude = normalizeDescriptorList(exclude)
  if (normalizedExclude.length > 0) out.excludeCredentials = normalizedExclude

  if (raw.extensions && typeof raw.extensions === 'object') {
    out.extensions = raw.extensions as Record<string, unknown>
  }

  return out
}

function normalizeAttestation(value: unknown): 'none' | 'direct' | 'indirect' | 'enterprise' {
  if (typeof value === 'string') {
    const trimmed = value.trim()
    if (trimmed === 'direct' || trimmed === 'indirect' || trimmed === 'enterprise') {
      return trimmed
    }
  }
  return 'none'
}

function normalizeRequestOptions(raw: AnyJson): PublicKeyCredentialRequestOptionsJSON {
  let challenge = ensureString(raw.challenge)
  if (!challenge) {
    const random = new Uint8Array(32)
    crypto.getRandomValues(random)
    challenge = bytesToBase64url(random)
  }
  const userVerification = ensureUserVerification(raw)
  const rpId = ensureString(raw.rpId, true)
  const allowCredentials = normalizeDescriptorList(raw.allowCredentials)

  const out: PublicKeyCredentialRequestOptionsJSON = {
    challenge,
    userVerification,
  }
  if (rpId) out.rpId = rpId
  if (allowCredentials.length > 0) out.allowCredentials = allowCredentials

  const timeout = Number(raw.timeout)
  if (Number.isFinite(timeout) && timeout > 0) out.timeout = timeout

  if (raw.extensions && typeof raw.extensions === 'object') {
    out.extensions = raw.extensions as Record<string, unknown>
  }

  return out
}

function normalizeUser(user: unknown): PublicKeyCredentialCreationOptionsJSON['user'] {
  if (user && typeof user === 'object') {
    const obj = user as AnyJson
    const id = typeof obj.id === 'string' && obj.id.length > 0 ? obj.id : null
    const name = ensureString(obj.name) || 'demo'
    const displayName = ensureString(obj.displayName) || 'Demo'
    if (id) {
      return { id, name, displayName }
    }
  }
  const random = new Uint8Array(32)
  crypto.getRandomValues(random)
  return {
    id: bytesToBase64url(random),
    name: 'demo',
    displayName: 'Demo',
  }
}

function normalizeRp(raw: AnyJson): PublicKeyCredentialCreationOptionsJSON['rp'] {
  if (raw.rp && typeof raw.rp === 'object') {
    const rpObj = raw.rp as AnyJson
    const id = ensureString(rpObj.id)
    const name = ensureString(rpObj.name) || 'Passkey Demo'
    return { id: id || 'localhost', name }
  }
  return { id: 'localhost', name: 'Passkey Demo' }
}

function normalizePubKeyCredParams(input: unknown): PublicKeyCredentialCreationOptionsJSON['pubKeyCredParams'] {
  const normalized: PublicKeyCredentialCreationOptionsJSON['pubKeyCredParams'] = []
  if (Array.isArray(input)) {
    for (const entry of input) {
      if (!entry || typeof entry !== 'object') continue
      const alg = (entry as AnyJson).alg
      if (typeof alg !== 'number') continue
      normalized.push({ type: 'public-key', alg })
    }
  }
  if (!normalized.some((entry) => entry.alg === -7)) {
    normalized.push({ type: 'public-key', alg: -7 })
  }
  return normalized
}

function normalizeAuthenticatorSelection(selection: unknown): PublicKeyCredentialCreationOptionsJSON['authenticatorSelection'] {
  const base = selection && typeof selection === 'object' ? { ...(selection as AnyJson) } : {}
  const uv = ensureUserVerification(base)
  const residentKey = ensureResidentKey(base)
  const requireResidentKey = typeof base.requireResidentKey === 'boolean' ? base.requireResidentKey : residentKey === 'required'
  const attachmentValue = typeof base.authenticatorAttachment === 'string'
    ? base.authenticatorAttachment.trim()
    : ''
  const authenticatorSelection: PublicKeyCredentialCreationOptionsJSON['authenticatorSelection'] = {
    residentKey,
    userVerification: uv,
    requireResidentKey,
  }
  if (attachmentValue === 'platform' || attachmentValue === 'cross-platform') {
    authenticatorSelection.authenticatorAttachment = attachmentValue
  }
  return authenticatorSelection
}

function ensureUserVerification(input: AnyJson): PublicKeyCredentialRequestOptionsJSON['userVerification'] {
  const value = typeof input.userVerification === 'string' ? input.userVerification.trim() : ''
  if (value === 'required' || value === 'preferred' || value === 'discouraged') {
    return value
  }
  return 'required'
}

function ensureResidentKey(input: AnyJson): 'required' | 'preferred' | 'discouraged' {
  if (typeof input.residentKey === 'string') {
    const value = input.residentKey.trim()
    if (value === 'preferred' || value === 'discouraged') return value
  }
  return 'required'
}

function normalizeDescriptorList(input: unknown): PublicKeyCredentialDescriptorJSON[] {
  if (!Array.isArray(input)) return []
  const result: PublicKeyCredentialDescriptorJSON[] = []
  for (const entry of input) {
    if (typeof entry === 'string') {
      result.push({ type: 'public-key', id: entry })
      continue
    }
    if (!entry || typeof entry !== 'object') continue
    const obj = entry as AnyJson
    const idSrc = typeof obj.id === 'string' && obj.id.length > 0 ? obj.id : ''
    if (!idSrc) continue
    let transports: PublicKeyCredentialDescriptorJSON['transports']
    if (Array.isArray(obj.transports)) {
      const filtered = obj.transports.filter((t) => typeof t === 'string')
      if (filtered.length > 0) {
        transports = filtered as PublicKeyCredentialDescriptorJSON['transports']
      }
    }
    result.push({
      type: 'public-key',
      id: idSrc,
      ...(transports ? { transports } : {}),
    })
  }
  return result
}

function ensureString(value: unknown, allowEmpty = false): string {
  if (typeof value === 'string') return value
  return allowEmpty ? '' : ''
}
