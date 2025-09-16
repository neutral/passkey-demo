import { test, expect } from '@playwright/test'

test('toRequestOptionsJSON sets rpId/uv and omits allowCredentials when empty', async ({ page }) => {
  await page.goto('/')
  const res = await page.evaluate(async () => {
    const { toRequestOptionsJSON } = await import('/src/lib/webauthn.ts')
    const sample = {
      login_session_id: 'abc',
      expires_at: 0,
      challenge: 'AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA',
      rpId: 'localhost',
      userVerification: 'required',
      allowCredentials: [],
      timeout: 60000,
    }
    const out = toRequestOptionsJSON(sample as any)
    const hasAC = Object.prototype.hasOwnProperty.call(out as any, 'allowCredentials')
    return { rpId: (out as any).rpId, uv: (out as any).userVerification, hasAC }
  })
  expect(res.rpId).toBe('localhost')
  expect(res.uv).toBe('required')
  expect(res.hasAC).toBeFalsy()
})

test('toRequestOptionsJSON includes allowCredentials when provided', async ({ page }) => {
  await page.goto('/')
  const has = await page.evaluate(async () => {
    const { toRequestOptionsJSON } = await import('/src/lib/webauthn.ts')
    const credId = 'AQIDBA' // base64url for [1,2,3,4]
    const sample = {
      login_session_id: 'abc',
      expires_at: 0,
      challenge: 'AA',
      rpId: 'localhost',
      userVerification: 'required',
      allowCredentials: [{ type: 'public-key', id: credId }],
      timeout: 60000,
    }
    const out = toRequestOptionsJSON(sample as any)
    return Array.isArray((out as any).allowCredentials) && (out as any).allowCredentials.length === 1
  })
  expect(has).toBeTruthy()
})
