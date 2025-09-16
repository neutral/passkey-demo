import { test, expect } from '@playwright/test'

test('toCreationOptionsJSON maps flags and fields correctly', async ({ page }) => {
  await page.goto('/')
  const res = await page.evaluate(async () => {
    const { toCreationOptionsJSON } = await import('/src/lib/webauthn.ts')
    const sample = {
      reg_session_id: 'abc',
      expires_at: 0,
      challenge: 'AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA',
      rp: { id: 'localhost', name: 'Passkey Demo' },
      user: { id: 'dGVzdHVzZXI', name: 'demo', displayName: 'Demo' },
      pubKeyCredParams: [{ type: 'public-key', alg: -7 }],
      authenticatorSelection: {
        residentKey: 'required',
        requireResidentKey: true,
        userVerification: 'required',
      },
      attestation: 'none',
      timeout: 60000,
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
