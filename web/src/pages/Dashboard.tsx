import { useEffect, useState } from 'react'
import { apiUrl } from '../config'
import { base64urlToBytes } from '../lib/encoding'
import { buildBundle, bundleToB64Hex, encodeBundleCanonical, type CoseEC2 } from '../lib/bundle'
import { decodeCBOR } from '../lib/cbor'
import { toRequestOptions, buildTxFinish } from '../lib/webauthn'
import ErrorToast from '../components/ErrorToast'
import { parseHttpError, normalizeError } from '../lib/http'
import { postJson, ApiError } from '../lib/api'

type Props = { onBack: () => void }

type TxItem = { tx_id_hex: string; nonce: number; message: string; created_at: number }

export default function Dashboard({ onBack }: Props) {
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [unauthorized, setUnauthorized] = useState(false)
  const [items, setItems] = useState<TxItem[]>([])
  const [senderKey, setSenderKey] = useState<CoseEC2 | null>(null)
  const [msg, setMsg] = useState('')
  const [nonce, setNonce] = useState('')
  const [bundleB64, setBundleB64] = useState('')
  const [bundleHex, setBundleHex] = useState('')
  const [signing, setSigning] = useState(false)
  const [senderCoseObj, setSenderCoseObj] = useState<any>(null)

  function nextNonceFrom(items: TxItem[]): string {
    if (!Array.isArray(items) || items.length === 0) return '1'
    const max = items.reduce((m, it) => (it.nonce > m ? it.nonce : m), 0)
    return String(max + 1)
  }

  function toUint8(v: any): Uint8Array {
    if (v instanceof Uint8Array) return v
    if (Array.isArray(v)) return new Uint8Array(v)
    if (v && typeof v === 'object' && 'buffer' in v) {
      try { return new Uint8Array(v as ArrayBufferLike) } catch {}
    }
    return new Uint8Array(0)
  }

  function toCoseMap(obj: any): Map<number, any> | null {
    if (obj instanceof Map) return obj as Map<number, any>
    if (obj && typeof obj === 'object') {
      const m = new Map<number, any>()
      for (const k of Object.keys(obj)) {
        const ik = parseInt(k, 10)
        if (!Number.isFinite(ik)) continue
        const v = (obj as any)[k]
        if (ik === -2 || ik === -3) {
          m.set(ik, toUint8(v))
        } else if (ik === 1 || ik === 3 || ik === -1) {
          m.set(ik, Number(v))
        }
      }
      if (m.has(1) && (m.has(-2) || m.has(-3))) return m
    }
    return null
  }

  async function loadList() {
    setLoading(true)
    setError(null)
    setUnauthorized(false)
    try {
      const r = await fetch(apiUrl('/tx/list'), {
        method: 'GET',
        mode: 'cors',
        credentials: 'include',
      })
      if (r.status === 401) {
        setUnauthorized(true)
        setItems([])
        return
      }
      if (!r.ok) {
        const pe = await parseHttpError(r)
        setError(`${pe.title}${pe.detail ? ` — ${pe.detail}` : ''}`)
        return
      }
      const data = (await r.json()) as { items?: TxItem[] }
      const arr = Array.isArray(data.items) ? data.items : []
      setItems(arr)
      setNonce(nextNonceFrom(arr))
    } catch (e: any) {
      const ne = normalizeError(e)
      setError(ne.detail)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    void loadList()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  async function ensureSenderKey(force = false): Promise<CoseEC2 | null> {
    if (!force && (senderKey || unauthorized)) return senderKey
    try {
      const r = await fetch(apiUrl('/me/account_key'), { method: 'GET', mode: 'cors', credentials: 'include' })
      if (r.status === 401) {
        setUnauthorized(true)
        setError('HTTP 401')
        return null
      }
      if (!r.ok) {
        const pe = await parseHttpError(r)
        setError(`${pe.title}${pe.detail ? ` — ${pe.detail}` : ''}`)
        return null
      }
      const data = await r.json() as { acct_cbor_b64?: string, sender_key: { kty: number, alg: number, crv: number, x: string, y: string } }
      const sk: CoseEC2 = {
        kty: data.sender_key.kty,
        alg: data.sender_key.alg,
        crv: data.sender_key.crv,
        x: base64urlToBytes(data.sender_key.x),
        y: base64urlToBytes(data.sender_key.y),
      }
      setSenderKey(sk)
      if (data.acct_cbor_b64) {
        try {
          const buf = base64urlToBytes(data.acct_cbor_b64)
          const obj = decodeCBOR(buf)
          const mm = toCoseMap(obj)
          setSenderCoseObj(mm || obj)
        } catch { setSenderCoseObj(null) }
      } else { setSenderCoseObj(null) }
      return sk
    } catch (e: any) {
      const ne = normalizeError(e)
      setError(ne.detail)
      return null
    }
  }

  async function buildBundlePreview() {
    setError(null)
    setBundleB64('')
    setBundleHex('')
    // Always force-refresh the account key to avoid stale state across re-logins/browsers
    const sk = await ensureSenderKey(true)
    if (unauthorized) return
    if (!sk) return
    const n = Number(nonce)
    if (!Number.isFinite(n) || n <= 0) {
      setError('Enter a positive integer nonce')
      return
    }
    if (!msg.trim()) {
      setError('Enter a message')
      return
    }
    try {
      let B: Uint8Array
      if (senderCoseObj) {
        const b = new Map<number, any>([[0, senderCoseObj], [1, n], [2, msg.trim()]])
        B = encodeBundleCanonical(b)
      } else {
        const b = buildBundle(sk, n, msg.trim())
        B = encodeBundleCanonical(b)
      }
      const { b64, hex } = bundleToB64Hex(B)
      setBundleB64(b64)
      setBundleHex(hex)
    } catch (e: any) {
      const ne = normalizeError(e)
      setError(ne.detail)
    }
  }

  async function signTransaction() {
    if (!bundleB64) { setError('Build the bundle first'); return }
    setSigning(true)
    setError(null)
    try {
      // 1) Options
      const data = await postJson<any>(apiUrl('/tx/signing/options'), { bundle_cbor_b64: bundleB64 })
      const publicKey = toRequestOptions(data)
      const cred = (await navigator.credentials.get({ publicKey })) as PublicKeyCredential
      if (!cred) throw new Error('get() returned null')
      // 2) Finish
      const payload = buildTxFinish(cred, data.tx_session_id)
      const rf = await fetch(apiUrl('/tx/signing/finish'), {
        method: 'POST',
        mode: 'cors',
        credentials: 'include',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload),
      })
      if (!rf.ok) throw new Error(`finish: HTTP ${rf.status}`)
      // Success: clear message and bundle; refresh list (which auto-advances nonce)
      setMsg('')
      setBundleB64('')
      setBundleHex('')
      await loadList()
    } catch (e: any) {
      if (e instanceof ApiError && e.status === 401) {
        // Distinguish missing session vs sender_key mismatch by probing account_key
        const probe = await fetch(apiUrl('/me/account_key'), { method: 'GET', mode: 'cors', credentials: 'include' })
        if (probe.status === 200) {
          setError('Account key mismatch — click Build to refresh and try again')
          return
        }
        setUnauthorized(true)
        return
      }
      if (e instanceof ApiError) {
        setError(`HTTP ${e.status} — ${e.message}${e.code ? ` (${e.code})` : ''}`)
        return
      }
      setError(e?.message || String(e))
    } finally {
      setSigning(false)
    }
  }

  return (
    <div style={{ padding: 24 }}>
      <h1>Dashboard</h1>

      <section style={{ marginTop: 16 }}>
        <h2>Transactions</h2>
        <div style={{ display: 'flex', gap: 8, alignItems: 'center' }}>
          <button type="button" onClick={loadList} disabled={loading}>
            {loading ? 'Refreshing…' : 'Refresh'}
          </button>
          <button type="button" onClick={onBack} disabled={loading}>Back</button>
        </div>

        {error && (
          <ErrorToast title="Error" detail={error} onClose={() => setError(null)} />
        )}
        {unauthorized && (
          <p style={{ marginTop: 12 }}>
            Not logged in. Please <a href="#/login">Login</a>.
          </p>
        )}
        {!unauthorized && !error && !loading && items.length === 0 && (
          <p style={{ marginTop: 12 }}>No transactions yet.</p>
        )}
        {!unauthorized && !error && items.length > 0 && (
          <table style={{ width: '100%', marginTop: 12, borderCollapse: 'collapse' }}>
            <thead>
              <tr>
                <th style={{ textAlign: 'left', borderBottom: '1px solid #444' }}>Time</th>
                <th style={{ textAlign: 'right', borderBottom: '1px solid #444' }}>Nonce</th>
                <th style={{ textAlign: 'left', borderBottom: '1px solid #444' }}>Message</th>
                <th style={{ textAlign: 'left', borderBottom: '1px solid #444' }}>Tx ID</th>
              </tr>
            </thead>
            <tbody>
              {items.map((it) => (
                <tr key={it.tx_id_hex}>
                  <td>{new Date(it.created_at * 1000).toLocaleString()}</td>
                  <td style={{ textAlign: 'right' }}>{it.nonce}</td>
                  <td>{it.message}</td>
                  <td>
                    <code>
                      {it.tx_id_hex.length > 16
                        ? `${it.tx_id_hex.slice(0, 16)}…`
                        : it.tx_id_hex}
                    </code>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </section>

      <section style={{ marginTop: 24 }}>
        <h2>Sign a Message</h2>
        <div style={{ display: 'flex', flexDirection: 'column', gap: 8, maxWidth: 520 }}>
          <input placeholder="Message" value={msg} onChange={(e) => setMsg(e.target.value)} />
          <div style={{ display: 'flex', gap: 8, alignItems: 'center' }}>
            <input placeholder="Nonce" value={nonce} readOnly />
            <small style={{ opacity: 0.8 }}>(auto-filled)</small>
          </div>
          <div style={{ display: 'flex', gap: 8 }}>
            <button type="button" onClick={buildBundlePreview} disabled={unauthorized || loading}>Build</button>
            {!senderKey && !unauthorized && (
              <button type="button" onClick={ensureSenderKey} disabled={loading}>Load Key</button>
            )}
            <button type="button" onClick={signTransaction} disabled={unauthorized || loading || signing || !bundleB64}>Sign</button>
          </div>
          {bundleB64 && (
            <div style={{ marginTop: 8 }}>
              <div><strong>bundle_cbor_b64</strong></div>
              <textarea readOnly value={bundleB64} style={{ width: '100%', height: 60 }} />
              <div><strong>hex(B)</strong></div>
              <textarea readOnly value={bundleHex} style={{ width: '100%', height: 60 }} />
            </div>
          )}
        </div>
      </section>
    </div>
  )
}
