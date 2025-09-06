import { base64urlToBytes, bytesToBase64url } from './encoding'

export type RegistrationOptionsResponse = {
  reg_session_id: string
  challenge: string
  options: {
    rp_id: string
    origin: string
    uv_required: boolean
    attestation: 'none'
  }
  expires_at: number
}

export function toCreationOptions(resp: RegistrationOptionsResponse): PublicKeyCredentialCreationOptions {
  const challenge = base64urlToBytes(resp.challenge)
  const userId = new Uint8Array(32)
  crypto.getRandomValues(userId)

  return {
    rp: {
      id: resp.options.rp_id,
      name: 'Passkey Demo',
    },
    user: {
      id: userId,
      name: 'demo',
      displayName: 'Demo',
    },
    challenge,
    pubKeyCredParams: [
      { type: 'public-key', alg: -7 }, // ES256
    ],
    authenticatorSelection: {
      residentKey: 'required',
      userVerification: 'required',
    },
    attestation: resp.options.attestation,
  }
}

export function buildRegFinish(cred: PublicKeyCredential, regSessionId: string) {
  const att = cred.response as AuthenticatorAttestationResponse
  const rawId = bytesToBase64url(new Uint8Array(cred.rawId as ArrayBuffer))
  const attObj = bytesToBase64url(new Uint8Array(att.attestationObject))
  const cdj = bytesToBase64url(new Uint8Array(att.clientDataJSON))

  return {
    reg_session_id: regSessionId,
    id: cred.id,
    rawId,
    type: cred.type,
    response: {
      attestationObject: attObj,
      clientDataJSON: cdj,
    },
  }
}

