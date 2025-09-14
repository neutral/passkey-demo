import { test, expect } from '@playwright/test'

test('toCreationOptionsJSON maps flags and fields correctly', async ({ page }) => {
  await page.goto('/')
  const res = await page.evaluate(async () => {
    const { toCreationOptionsJSON } = await import('/src/lib/webauthn.ts')
    const sample = {
      reg_session_id: 'abc',
      challenge: 'AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA', // base64url zeros
      options: { rp_id: 'localhost', origin: 'http://localhost:5173', uv_required: true, attestation: 'none' },
      expires_at: 0,
    }
    const out = toCreationOptionsJSON(sample as any)
    const hasES256 = Array.isArray(out.pubKeyCredParams) && out.pubKeyCredParams.some(p => (p as any).alg === -7)
    const authSel = out.authenticatorSelection as any
    const userId = (out.user as any).id
    return {
      rpId: out.rp?.id,
      att: out.attestation,
      rk: authSel?.residentKey,
      uv: authSel?.userVerification,
      challenge: out.challenge,
      userIdType: typeof userId,
      hasES256,
    }
  })
  expect(res.rpId).toBe('localhost')
  expect(res.att).toBe('none')
  expect(res.rk).toBe('required')
  expect(res.uv).toBe('required')
  expect(res.challenge).toBe('AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA')
  expect(res.userIdType).toBe('string')
  expect(res.hasES256).toBeTruthy()
})

