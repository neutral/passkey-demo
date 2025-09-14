import { useState } from 'react'
import { apiUrl } from '../config'
import {
  toCreationOptionsJSON,
  type RegistrationOptionsResponse,
  mapDomException,
} from '../lib/webauthn'
import { startRegistration as swuStartRegistration } from '@simplewebauthn/browser'
import ErrorToast from '../components/ErrorToast'
import { parseHttpError, normalizeError } from '../lib/http'

type Props = { onBack: () => void }

export default function Register({ onBack }: Props) {
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [thumb, setThumb] = useState<string | null>(null)

  async function startRegistration() {
    setError(null)
    setThumb(null)
    setLoading(true)
    try {
      console.debug('[Register] fetch options')
      // 1) Fetch options from backend
      const r = await fetch(apiUrl('/authn/passkey/registration/options'), {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        mode: 'cors',
      })
      if (!r.ok) {
        const pe = await parseHttpError(r)
        setError(pe.detail || pe.title)
        return
      }
      const data = (await r.json()) as RegistrationOptionsResponse

      // 2) Build options JSON for @simplewebauthn/browser
      const optionsJSON = toCreationOptionsJSON(data)

      // 3) Invoke library to perform navigator.credentials.create and return JSON response
      console.debug('[Register] startRegistration')
      const att = await swuStartRegistration(optionsJSON)
      console.debug('[Register] attestation result received')

      // 4) Compose finish payload (server expects reg_session_id)
      const payload = {
        reg_session_id: data.reg_session_id,
        id: att.id,
        rawId: att.rawId,
        type: att.type,
        response: {
          attestationObject: att.response.attestationObject,
          clientDataJSON: att.response.clientDataJSON,
        },
      }
      console.debug('[Register] POST finish')
      const r2 = await fetch(apiUrl('/authn/passkey/registration/finish'), {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        mode: 'cors',
        body: JSON.stringify(payload),
      })
      if (!r2.ok) {
        const pe2 = await parseHttpError(r2)
        setError(pe2.detail || pe2.title)
        return
      }
      const out = (await r2.json()) as { account_thumb_hex: string; credential_id_b64: string }
      setThumb(out.account_thumb_hex)
    } catch (e: any) {
      console.debug('[Register] error', e)
      const ne = normalizeError(mapDomException(e))
      setError(ne.detail)
    } finally {
      setLoading(false)
    }
  }

  return (
    <div style={{ padding: 24 }}>
      <h1>Register</h1>
      <p>Use your platform authenticator to create a passkey.</p>
      <div style={{ display: 'flex', gap: 12, marginTop: 12 }}>
        <button type="button" onClick={startRegistration} disabled={loading}>
          {loading ? 'Working…' : 'Start Registration'}
        </button>
        <button type="button" onClick={onBack} disabled={loading}>
          Back
        </button>
      </div>
      {error && (
        <ErrorToast title="Error" detail={error} onClose={() => setError(null)} />
      )}
      {thumb && (
        <div style={{ marginTop: 16 }}>
          <div>
            <strong>Account thumb:</strong> <code>{thumb}</code>
          </div>
          <button
            type="button"
            style={{ marginTop: 12 }}
            onClick={() => {
              // Route to Login via hash to avoid changing App props
              window.location.hash = '#/login'
            }}
          >
            Go to Login
          </button>
        </div>
      )}
    </div>
  )
}
