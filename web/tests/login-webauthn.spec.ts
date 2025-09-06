import { test, expect } from '@playwright/test'

test('toRequestOptions maps fields and omits allowCredentials when empty', async ({ page }) => {
  await page.goto('/')
  const checks = await page.evaluate(async () => {
    const { toRequestOptions } = await import('/src/lib/webauthn.ts')
    const sample = {
      login_session_id: 'abc',
      challenge: 'AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA',
      options: { rp_id: 'localhost', origin: 'http://localhost:5173', uv_required: true, allow_credentials: [] },
      expires_at: 0,
    }
    const out = toRequestOptions(sample as any)
    const hasAC = Object.prototype.hasOwnProperty.call(out as any, 'allowCredentials')
    const ch: any = out.challenge
    const isBuf = ch instanceof ArrayBuffer
    const isView = typeof ArrayBuffer !== 'undefined' && ArrayBuffer.isView(ch)
    const len = isBuf ? (ch as ArrayBuffer).byteLength : isView ? ch.byteLength : 0
    return { rpId: (out as any).rpId, uv: (out as any).userVerification, hasAC, lenOK: len > 0 }
  })
  expect(checks.rpId).toBe('localhost')
  expect(checks.uv).toBe('required')
  expect(checks.hasAC).toBeFalsy()
  expect(checks.lenOK).toBeTruthy()
})

test('toRequestOptions includes allowCredentials when non-empty', async ({ page }) => {
  await page.goto('/')
  const has = await page.evaluate(async () => {
    const { toRequestOptions } = await import('/src/lib/webauthn.ts')
    const credId = 'AQIDBA' // base64url for [1,2,3,4]
    const sample = {
      login_session_id: 'abc',
      challenge: 'AA',
      options: { rp_id: 'localhost', origin: 'http://localhost:5173', uv_required: true, allow_credentials: [credId] },
      expires_at: 0,
    }
    const out = toRequestOptions(sample as any)
    return Array.isArray((out as any).allowCredentials) && (out as any).allowCredentials.length === 1
  })
  expect(has).toBeTruthy()
})

test('buildLoginFinish encodes assertion fields properly', async ({ page }) => {
  await page.goto('/')
  const ok = await page.evaluate(async () => {
    const { buildLoginFinish } = await import('/src/lib/webauthn.ts')
    const cred = {
      id: 'cred-id',
      type: 'public-key',
      rawId: new Uint8Array([1, 2, 3, 4]).buffer,
      response: {
        authenticatorData: new Uint8Array([5, 6, 7]).buffer,
        clientDataJSON: new TextEncoder().encode('{"type":"webauthn.get"}').buffer,
        signature: new Uint8Array([8, 9]).buffer,
        userHandle: new Uint8Array([10]).buffer,
      },
    } as any
    const out = buildLoginFinish(cred, 'sess-1')
    return out.login_session_id === 'sess-1' && typeof out.rawId === 'string' && typeof out.response.authenticatorData === 'string' && typeof out.response.clientDataJSON === 'string' && typeof out.response.signature === 'string'
  })
  expect(ok).toBeTruthy()
})

