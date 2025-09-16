import { useState } from 'react'
import { apiUrl } from '../config'
import {
  toCreationOptionsJSON,
  type RegistrationOptionsResponse,
  mapDomException,
} from '../lib/webauthn'
import { startRegistration as swuStartRegistration } from '@simplewebauthn/browser'
import ErrorToast from '../components/ErrorToast'
import { normalizeError } from '../lib/http'
import { postJson, ApiError, formatApiError } from '../lib/api'

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
      const data = await postJson<RegistrationOptionsResponse>(
        apiUrl('/authn/passkey/registration/options'),
        {},
      )

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
      const out = await postJson<{ account_thumb_hex: string; credential_id_b64: string }>(
        apiUrl('/authn/passkey/registration/finish'),
        payload,
      )
      setThumb(out.account_thumb_hex)
    } catch (e: any) {
      console.debug('[Register] error', e)
      if (e instanceof ApiError) {
        setError(formatApiError(e))
        return
      }
      const mapped = mapDomException(e)
      const ne = normalizeError(mapped)
      setError(ne.detail || mapped.detail || mapped.title)
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
