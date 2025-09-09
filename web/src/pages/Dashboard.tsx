import { useEffect, useState } from 'react'
import { apiUrl } from '../config'
import { base64urlToBytes } from '../lib/encoding'
import { buildBundle, bundleToB64Hex, encodeBundleCanonical, type CoseEC2 } from '../lib/bundle'
import { toRequestOptions, buildTxFinish } from '../lib/webauthn'

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
      if (!r.ok) throw new Error(`list: HTTP ${r.status}`)
      const data = (await r.json()) as { items?: TxItem[] }
      setItems(Array.isArray(data.items) ? data.items : [])
    } catch (e: any) {
      setError(e?.message || String(e))
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    void loadList()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  async function ensureSenderKey() {
    if (senderKey || unauthorized) return
    try {
      const r = await fetch(apiUrl('/me/account_key'), { method: 'GET', mode: 'cors', credentials: 'include' })
      if (r.status === 401) {
        setUnauthorized(true)
        return
      }
      if (!r.ok) throw new Error(`me/account_key: HTTP ${r.status}`)
      const data = await r.json() as { sender_key: { kty: number, alg: number, crv: number, x: string, y: string } }
      const sk: CoseEC2 = {
        kty: data.sender_key.kty,
        alg: data.sender_key.alg,
        crv: data.sender_key.crv,
        x: base64urlToBytes(data.sender_key.x),
        y: base64urlToBytes(data.sender_key.y),
      }
      setSenderKey(sk)
    } catch (e: any) {
      setError(e?.message || String(e))
    }
  }

  async function buildBundlePreview() {
    setError(null)
    setBundleB64('')
    setBundleHex('')
    if (!senderKey) await ensureSenderKey()
    if (unauthorized) return
    if (!senderKey) return
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
      const b = buildBundle(senderKey, n, msg.trim())
      const B = encodeBundleCanonical(b)
      const { b64, hex } = bundleToB64Hex(B)
      setBundleB64(b64)
      setBundleHex(hex)
    } catch (e: any) {
      setError(e?.message || String(e))
    }
  }

  async function signTransaction() {
    if (!bundleB64) { setError('Build the bundle first'); return }
    setSigning(true)
    setError(null)
    try {
      // 1) Options
      const ro = await fetch(apiUrl('/tx/signing/options'), {
        method: 'POST',
        mode: 'cors',
        credentials: 'include',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ bundle_cbor_b64: bundleB64 }),
      })
      if (ro.status === 401) { setUnauthorized(true); return }
      if (ro.status === 409) { throw new Error('conflict (nonce or credentials)') }
      if (ro.status === 400) { throw new Error('invalid bundle') }
      if (!ro.ok) throw new Error(`options: HTTP ${ro.status}`)
      const data = await ro.json()
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
      // Success: clear inputs and refresh list
      setMsg('')
      setNonce('')
      setBundleB64('')
      setBundleHex('')
      await loadList()
    } catch (e: any) {
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
          <p style={{ color: 'crimson', marginTop: 12 }}>Error: {error}</p>
        )}
        {unauthorized && !error && (
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
          <input placeholder="Nonce" value={nonce} onChange={(e) => setNonce(e.target.value)} />
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
