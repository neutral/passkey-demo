import { base64urlToBytes, bytesToBase64url } from './encoding'
import type {
  PublicKeyCredentialCreationOptionsJSON,
  PublicKeyCredentialRequestOptionsJSON,
} from '@simplewebauthn/types'

export type RegistrationOptionsResponse = PublicKeyCredentialCreationOptionsJSON & {
  reg_session_id: string
  expires_at: number
}

export type LoginOptionsResponse = {
  login_session_id: string
  challenge: string
  options: {
    rp_id: string
    origin: string
    uv_required: boolean
    allow_credentials: string[]
  }
  expires_at: number
}

// Adapters: Convert current Go server option shapes to @simplewebauthn/browser JSON shapes
export function toCreationOptionsJSON(
  resp: RegistrationOptionsResponse,
): PublicKeyCredentialCreationOptionsJSON {
  const { reg_session_id: _ignoreSession, expires_at: _ignoreExpiry, user, ...rest } = resp
  let resolvedUser = user
  if (!resolvedUser || typeof resolvedUser.id !== 'string' || resolvedUser.id.length === 0) {
    const u8 = new Uint8Array(32)
    crypto.getRandomValues(u8)
    resolvedUser = {
      id: bytesToBase64url(u8),
      name: user?.name || 'demo',
      displayName: user?.displayName || 'Demo',
    }
  }
  return {
    ...rest,
    user: resolvedUser,
  }
}

export function toRequestOptionsJSON(
  resp: LoginOptionsResponse,
): PublicKeyCredentialRequestOptionsJSON {
  const ids = (resp.options.allow_credentials || []).filter((s) => !!s && s.length > 0)
  const out: PublicKeyCredentialRequestOptionsJSON = {
    challenge: resp.challenge,
    userVerification: 'required',
  }
  if (resp.options.rp_id) (out as any).rpId = resp.options.rp_id
  if (ids.length > 0) (out as any).allowCredentials = ids.map((id) => ({ type: 'public-key', id }))
  return out
}

// Back-compat for Dashboard signing flow which still uses native WebAuthn in this step
export function toRequestOptions(resp: LoginOptionsResponse): PublicKeyCredentialRequestOptions {
  const challenge = base64urlToBytes(resp.challenge)
  const ids = (resp.options.allow_credentials || [])
    .map((b64) => ({ type: 'public-key', id: base64urlToBytes(b64) }))
    .filter((d) => d.id.byteLength > 0)
  const out: PublicKeyCredentialRequestOptions = {
    challenge,
    userVerification: 'required',
  }
  if (resp.options.rp_id) (out as any).rpId = resp.options.rp_id
  if (ids.length > 0) (out as any).allowCredentials = ids
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
