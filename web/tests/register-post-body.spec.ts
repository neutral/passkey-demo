import { test, expect } from '@playwright/test'

test('Register posts RegistrationResponseJSON with reg_session_id', async ({ page }) => {
  page.on('console', (msg) => console.log('[console]', msg.type(), msg.text()))
  // Stub create() so the library gets a predictable credential
  await page.addInitScript(() => {
    // Minimal WebAuthn globals for @simplewebauthn/browser
    // @ts-ignore
    window.PublicKeyCredential = (function () {}) as any
    // @ts-ignore
    window.PublicKeyCredential.isUserVerifyingPlatformAuthenticatorAvailable = async () => true
    // @ts-ignore
    navigator.credentials = navigator.credentials || {}
    // @ts-ignore
    navigator.credentials.create = async () => ({
      id: 'cred-id',
      type: 'public-key',
      rawId: new Uint8Array([1,2,3,4]).buffer,
      getClientExtensionResults: () => ({}),
      response: {
        attestationObject: new Uint8Array([5,6,7]).buffer,
        clientDataJSON: new TextEncoder().encode('{"type":"webauthn.create"}').buffer,
      },
    })
  })

  // Options (Go shape)
  await page.route('**/authn/passkey/registration/options', async (route) => {
    await route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({
      reg_session_id: 's1', challenge: 'AA', options: { rp_id: 'localhost', origin: 'http://localhost:5173', uv_required: true, attestation: 'none' }, expires_at: 0,
    }) })
  })

  await page.route('**/authn/passkey/registration/finish', async (route) => {
    await route.fulfill({ status: 201, contentType: 'application/json', body: JSON.stringify({ account_thumb_hex: '00', credential_id_b64: 'AA' }) })
  })

  await page.goto('/')
  await page.getByRole('button', { name: 'Register' }).click()
  const reqP = page.waitForRequest('**/authn/passkey/registration/finish')
  await page.getByRole('button', { name: 'Start Registration' }).click()
  const req = await reqP

  // Validate payload shape
  const posted = await req.postDataJSON()
  expect(posted).toBeTruthy()
  expect((posted as any).reg_session_id).toBe('s1')
  expect(typeof (posted as any).id).toBe('string')
  expect(typeof (posted as any).rawId).toBe('string')
  expect((posted as any).type).toBe('public-key')
  expect(typeof (posted as any).response?.attestationObject).toBe('string')
  expect(typeof (posted as any).response?.clientDataJSON).toBe('string')
})
