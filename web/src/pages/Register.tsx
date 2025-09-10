import { useState } from 'react'
import { apiUrl } from '../config'
import { toCreationOptions, buildRegFinish, type RegistrationOptionsResponse } from '../lib/webauthn'
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
      // 1) Fetch options from backend
      const r = await fetch(apiUrl('/authn/passkey/registration/options'), {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        mode: 'cors',
      })
      if (!r.ok) {
        const pe = await parseHttpError(r)
        setError(pe.title)
        return
      }
      const data = (await r.json()) as RegistrationOptionsResponse

      // 2) Build WebAuthn creation options
      const publicKey = toCreationOptions(data)

      // 3) Invoke WebAuthn
      const cred = (await navigator.credentials.create({ publicKey })) as PublicKeyCredential
      if (!cred) throw new Error('create() returned null')

      // 4) Build finish payload and POST
      const payload = buildRegFinish(cred, data.reg_session_id)
      const r2 = await fetch(apiUrl('/authn/passkey/registration/finish'), {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        mode: 'cors',
        body: JSON.stringify(payload),
      })
      if (!r2.ok) {
        const pe2 = await parseHttpError(r2)
        setError(pe2.title)
        return
      }
      const out = (await r2.json()) as { account_thumb_hex: string; credential_id_b64: string }
      setThumb(out.account_thumb_hex)
    } catch (e: any) {
      const ne = normalizeError(e)
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
