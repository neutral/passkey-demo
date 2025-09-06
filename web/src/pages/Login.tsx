import { useState } from 'react'
import { apiUrl } from '../config'
import { toRequestOptions, buildLoginFinish, type LoginOptionsResponse } from '../lib/webauthn'

type Props = { onBack: () => void }

export default function Login({ onBack }: Props) {
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [thumb, setThumb] = useState<string | null>(null)

  async function startLogin() {
    setError(null)
    setThumb(null)
    setLoading(true)
    try {
      const r = await fetch(apiUrl('/authn/passkey/login/options'), {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        mode: 'cors',
      })
      if (!r.ok) throw new Error(`options: HTTP ${r.status}`)
      const data = (await r.json()) as LoginOptionsResponse
      const publicKey = toRequestOptions(data)
      const cred = (await navigator.credentials.get({ publicKey })) as PublicKeyCredential
      if (!cred) throw new Error('get() returned null')
      const payload = buildLoginFinish(cred, data.login_session_id)
      const r2 = await fetch(apiUrl('/authn/passkey/login/finish'), {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        mode: 'cors',
        credentials: 'include',
        body: JSON.stringify(payload),
      })
      if (!r2.ok) throw new Error(`finish: HTTP ${r2.status}`)
      const out = (await r2.json()) as { account_thumb_hex: string; credential_id_b64: string }
      setThumb(out.account_thumb_hex)
      window.location.hash = '#/dashboard'
    } catch (e: any) {
      setError(e?.message || String(e))
    } finally {
      setLoading(false)
    }
  }

  return (
    <div style={{ padding: 24 }}>
      <h1>Login</h1>
      <p>Authenticate with your passkey to start a session.</p>
      <div style={{ display: 'flex', gap: 12, marginTop: 12 }}>
        <button type="button" onClick={startLogin} disabled={loading}>
          {loading ? 'Working…' : 'Start Login'}
        </button>
        <button type="button" onClick={onBack} disabled={loading}>
          Back
        </button>
      </div>
      {error && <p style={{ color: 'crimson', marginTop: 12 }}>Error: {error}</p>}
      {thumb && (
        <div style={{ marginTop: 16 }}>
          <div>
            <strong>Account thumb:</strong> <code>{thumb}</code>
          </div>
        </div>
      )}
    </div>
  )
}
