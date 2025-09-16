import { test, expect } from '@playwright/test'

test('Login posts AuthenticationResponseJSON with login_session_id', async ({ page }) => {
  page.on('console', (msg) => console.log('[console]', msg.type(), msg.text()))
  // Stub get() so the library gets a predictable assertion
  await page.addInitScript(() => {
    // Minimal WebAuthn globals for @simplewebauthn/browser
    // @ts-ignore
    window.PublicKeyCredential = (function () {}) as any
    // @ts-ignore
    window.PublicKeyCredential.isUserVerifyingPlatformAuthenticatorAvailable = async () => true
    // @ts-ignore
    navigator.credentials = navigator.credentials || {}
    // @ts-ignore
    navigator.credentials.get = async () => ({
      id: 'cred-id',
      type: 'public-key',
      rawId: new Uint8Array([1,2,3,4]).buffer,
      getClientExtensionResults: () => ({}),
      response: {
        authenticatorData: new Uint8Array([5,6,7]).buffer,
        clientDataJSON: new TextEncoder().encode('{"type":"webauthn.get"}').buffer,
        signature: new Uint8Array([8,9]).buffer,
        userHandle: new Uint8Array([10]).buffer,
      },
    })
  })

  // Options (Go shape)
  await page.route('**/authn/passkey/login/options', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        login_session_id: 'ls1',
        expires_at: 0,
        challenge: 'AA',
        rpId: 'localhost',
        userVerification: 'required',
        allowCredentials: [],
        timeout: 60000,
      }),
    })
  })

  await page.route('**/authn/passkey/login/finish', async (route) => {
    await route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ account_thumb_hex: '00', credential_id_b64: 'AA' }) })
  })

  await page.goto('/')
  await page.getByRole('button', { name: 'Login' }).click()
  const reqP = page.waitForRequest('**/authn/passkey/login/finish')
  await page.getByRole('button', { name: 'Start Login' }).click()
  const req = await reqP

  // Validate payload shape
  const posted = await req.postDataJSON()
  expect(posted).toBeTruthy()
  expect((posted as any).login_session_id).toBe('ls1')
  expect(typeof (posted as any).id).toBe('string')
  expect(typeof (posted as any).rawId).toBe('string')
  expect((posted as any).type).toBe('public-key')
  expect(typeof (posted as any).response?.authenticatorData).toBe('string')
  expect(typeof (posted as any).response?.clientDataJSON).toBe('string')
  expect(typeof (posted as any).response?.signature).toBe('string')
})
