import { test, expect } from '@playwright/test'

test('toCreationOptions maps fields and flags correctly', async ({ page }) => {
  await page.goto('/')
  const ok = await page.evaluate(async () => {
    const { toCreationOptions } = await import('/src/lib/webauthn.ts')
    const sample = {
      reg_session_id: 'abc',
      challenge: 'AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA', // valid b64url for zeros
      options: { rp_id: 'localhost', origin: 'http://localhost:5173', uv_required: true, attestation: 'none' },
      expires_at: 0,
    }
    const out = toCreationOptions(sample as any)
    const flags = out.authenticatorSelection as any
    const ch: any = out.challenge
    const isBuf = ch instanceof ArrayBuffer
    const isView = typeof ArrayBuffer !== 'undefined' && ArrayBuffer.isView(ch)
    const len = isBuf ? (ch as ArrayBuffer).byteLength : isView ? ch.byteLength : 0
    return (
      out.rp?.id === 'localhost' &&
      out.attestation === 'none' &&
      out.pubKeyCredParams?.some(p => p.alg === -7) &&
      flags?.residentKey === 'required' &&
      flags?.userVerification === 'required' &&
      len > 0
    )
  })
  expect(ok).toBeTruthy()
})

test('buildRegFinish encodes fields as base64url and includes session id', async ({ page }) => {
  await page.goto('/')
  const ok = await page.evaluate(async () => {
    const { buildRegFinish } = await import('/src/lib/webauthn.ts')
    const cred = {
      id: 'cred-id',
      type: 'public-key',
      rawId: new Uint8Array([1, 2, 3, 4]).buffer,
      response: {
        attestationObject: new Uint8Array([5, 6, 7]).buffer,
        clientDataJSON: new TextEncoder().encode('{"type":"webauthn.create"}').buffer,
      },
    } as any
    const out = buildRegFinish(cred, 'sess-1')
    return out.reg_session_id === 'sess-1' && out.id === 'cred-id' && typeof out.rawId === 'string' && typeof out.response.attestationObject === 'string' && typeof out.response.clientDataJSON === 'string'
  })
  expect(ok).toBeTruthy()
})
