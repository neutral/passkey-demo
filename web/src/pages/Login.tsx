import { useState } from 'react'
import { apiUrl } from '../config'
import {
  toRequestOptionsJSON,
  type LoginOptionsResponse,
  mapDomException,
} from '../lib/webauthn'
import { startAuthentication } from '@simplewebauthn/browser'
import ErrorToast from '../components/ErrorToast'
import { parseHttpError, normalizeError } from '../lib/http'

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
      console.debug('[Login] fetch options')
      const r = await fetch(apiUrl('/authn/passkey/login/options'), {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        mode: 'cors',
      })
      if (!r.ok) {
        const pe = await parseHttpError(r)
        setError(pe.title) // e.g., "HTTP 401"
        return
      }
      const data = (await r.json()) as LoginOptionsResponse
      const optionsJSON = toRequestOptionsJSON(data)
      console.debug('[Login] startAuthentication')
      const asg = await startAuthentication(optionsJSON)
      console.debug('[Login] assertion result received')
      const payload = {
        login_session_id: data.login_session_id,
        id: asg.id,
        rawId: asg.rawId,
        type: asg.type,
        response: {
          authenticatorData: asg.response.authenticatorData,
          clientDataJSON: asg.response.clientDataJSON,
          signature: asg.response.signature,
          userHandle: (asg.response as any).userHandle || '',
        },
      }
      console.debug('[Login] POST finish')
      const r2 = await fetch(apiUrl('/authn/passkey/login/finish'), {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        mode: 'cors',
        credentials: 'include',
        body: JSON.stringify(payload),
      })
      if (!r2.ok) {
        const pe2 = await parseHttpError(r2)
        setError(pe2.title) // keep string that includes HTTP status for tests
        return
      }
      const out = (await r2.json()) as { account_thumb_hex: string; credential_id_b64: string }
      setThumb(out.account_thumb_hex)
      window.location.hash = '#/dashboard'
    } catch (e: any) {
      console.debug('[Login] error', e)
      const ne = normalizeError(mapDomException(e))
      setError(ne.detail)
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
      {error && <ErrorToast title="Error" detail={error} onClose={() => setError(null)} />}
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
