import { test, expect } from '@playwright/test'

test('Register finish 400 shows toast with JSON error', async ({ page }) => {
  // Stub create()
  await page.addInitScript(() => {
    // Ensure library sees WebAuthn environment
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
      rawId: new Uint8Array([1]).buffer,
      getClientExtensionResults: () => ({}),
      response: { attestationObject: new Uint8Array([2]).buffer, clientDataJSON: new TextEncoder().encode('{"type":"webauthn.create"}').buffer },
    })
  })

  // Options ok
  await page.route('**/authn/passkey/registration/options', async (route) => {
    await route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({
      reg_session_id: 's1', challenge: 'AA', options: { rp_id: 'localhost', origin: 'http://localhost:5173', uv_required: true, attestation: 'none' }, expires_at: 0,
    }) })
  })
  // Finish 400 with JSON error
  await page.route('**/authn/passkey/registration/finish', async (route) => {
    await route.fulfill({ status: 400, contentType: 'application/json', body: JSON.stringify({ error: 'invalid attestation' }) })
  })

  await page.goto('/')
  await page.getByRole('button', { name: 'Register' }).click()
  await page.getByRole('button', { name: 'Start Registration' }).click()
  await expect(page.getByRole('alert')).toBeVisible()
  await expect(page.getByRole('alert')).toContainText('invalid attestation')
})

test('Login finish 401 shows HTTP 401 in toast', async ({ page }) => {
  // Stub get()
  await page.addInitScript(() => {
    // Ensure library sees WebAuthn environment (even though Login.tsx uses library + finish JSON)
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
      rawId: new Uint8Array([1]).buffer,
      getClientExtensionResults: () => ({}),
      response: { authenticatorData: new Uint8Array([2]).buffer, clientDataJSON: new TextEncoder().encode('{"type":"webauthn.get"}').buffer, signature: new Uint8Array([3]).buffer, userHandle: new Uint8Array([4]).buffer },
    })
  })

  // Options ok
  await page.route('**/authn/passkey/login/options', async (route) => {
    await route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({
      login_session_id: 'ls1', challenge: 'AA', options: { rp_id: 'localhost', origin: 'http://localhost:5173', uv_required: true, allow_credentials: [] }, expires_at: 0,
    }) })
  })
  // Finish 401
  await page.route('**/authn/passkey/login/finish', async (route) => {
    await route.fulfill({ status: 401, contentType: 'application/json', body: '{}' })
  })

  await page.goto('/')
  await page.getByRole('button', { name: 'Login' }).click()
  await page.getByRole('button', { name: 'Start Login' }).click()
  await expect(page.getByRole('alert')).toBeVisible()
  await expect(page.getByRole('alert')).toContainText('HTTP 401')
})

test('Dashboard account_key 401 shows toast and unauthorized prompt', async ({ page }) => {
  // Keep list authorized and empty
  await page.route('**/tx/list', async (route) => {
    await route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ items: [] }) })
  })
  // Account key 401
  await page.route('**/me/account_key', async (route) => {
    await route.fulfill({ status: 401, contentType: 'application/json', body: '{}' })
  })

  await page.goto('/')
  await page.evaluate(() => { window.location.hash = '#/dashboard' })
  await expect(page.getByRole('heading', { name: 'Dashboard' })).toBeVisible()
  // Attempt to load key
  await page.getByRole('button', { name: 'Load Key' }).click()
  await expect(page.getByRole('alert')).toBeVisible()
  await expect(page.getByText('Not logged in. Please')).toBeVisible()
})

test('Dashboard Sign: options 413 shows HTTP 413; options 429 JSON shows message and code', async ({ page }) => {
  // Stub get()
  await page.addInitScript(() => {
    // @ts-ignore
    navigator.credentials = navigator.credentials || {}
    // @ts-ignore
    navigator.credentials.get = async () => ({
      id: 'cred-id', type: 'public-key', rawId: new Uint8Array([1]).buffer,
      response: { authenticatorData: new Uint8Array([2]).buffer, clientDataJSON: new TextEncoder().encode('{"type":"webauthn.get"}').buffer, signature: new Uint8Array([3]).buffer, userHandle: new Uint8Array([4]).buffer },
    })
  })

  // Keep list authorized and empty
  await page.route('**/tx/list', async (route) => {
    await route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ items: [] }) })
  })
  // Provide sender key
  await page.route('**/me/account_key', async (route) => {
    await route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ acct_cbor_b64: 'AA', sender_key: { kty: 2, alg: -7, crv: 1, x: 'AQ', y: 'AQ' } }) })
  })

  // Case 1: 413
  await page.route('**/tx/signing/options', async (route) => {
    if (route.request().headers()['x-case'] === '429') return route.fallback()
    await route.fulfill({ status: 413, contentType: 'application/json', body: '{}' })
  })

  await page.goto('/')
  await page.evaluate(() => { window.location.hash = '#/dashboard' })
  await page.getByPlaceholder('Message').fill('m')
  await page.getByRole('button', { name: 'Load Key' }).click()
  await page.getByRole('button', { name: 'Build' }).click()
  await page.getByRole('button', { name: 'Sign' }).click()
  await expect(page.getByRole('alert')).toBeVisible()
  await expect(page.getByRole('alert')).toContainText('HTTP 413')

  // Case 2: 429 with message
  await page.unroute('**/tx/signing/options')
  await page.route('**/tx/signing/options', async (route) => {
    await route.fulfill({ status: 429, contentType: 'application/json', body: JSON.stringify({ message: 'rate limited', code: 'RATE_LIMIT' }) })
  })
  await page.getByRole('button', { name: 'Sign' }).click()
  await expect(page.getByRole('alert')).toBeVisible()
  await expect(page.getByRole('alert')).toContainText('rate limited')
  await expect(page.getByRole('alert')).toContainText('RATE_LIMIT')
})
