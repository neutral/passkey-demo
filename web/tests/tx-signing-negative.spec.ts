import { test, expect } from '@playwright/test'

// Reusable helper to prepare page: stub navigator.credentials.get and basic routes
async function preparePage(page: import('@playwright/test').Page) {
  // Stub navigator.credentials.get to avoid real platform prompts where flow reaches get()
  await page.addInitScript(() => {
    // @ts-ignore
    const fakeGet = async () => ({
      id: 'cred-id',
      type: 'public-key',
      rawId: new Uint8Array([1, 2, 3, 4]).buffer,
      response: {
        authenticatorData: new Uint8Array([5]).buffer,
        clientDataJSON: new TextEncoder().encode('{"type":"webauthn.get"}').buffer,
        signature: new Uint8Array([6]).buffer,
        userHandle: new Uint8Array([7]).buffer,
      },
    })
    if (navigator.credentials) {
      // @ts-ignore
      navigator.credentials.get = fakeGet
    } else {
      // @ts-ignore
      navigator.credentials = { get: fakeGet }
    }
  })

  // Keep dashboard list non-blocking (authorized empty list)
  await page.route('**/tx/list', async (route) => {
    await route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ items: [] }) })
  })

  // Provide a minimal sender key for Build
  await page.route('**/me/account_key', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ acct_cbor_b64: 'AA', sender_key: { kty: 2, alg: -7, crv: 1, x: 'AQ', y: 'AQ' } }),
    })
  })
}

test('options 401 → unauthorized prompt shown', async ({ page }) => {
  await preparePage(page)

  await page.route('**/tx/signing/options', async (route) => {
    await route.fulfill({ status: 401, contentType: 'application/json', body: '{}' })
  })
  await page.unroute('**/me/account_key')
  let keyCalls = 0
  await page.route('**/me/account_key', async (route) => {
    keyCalls += 1
    if (keyCalls <= 2) {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ acct_cbor_b64: 'AA', sender_key: { kty: 2, alg: -7, crv: 1, x: 'AQ', y: 'AQ' } }),
      })
      return
    }
    await route.fulfill({ status: 401, contentType: 'application/json', body: '{}' })
  })

  await page.goto('/')
  await page.evaluate(() => { window.location.hash = '#/dashboard' })
  await expect(page.getByRole('heading', { name: 'Dashboard' })).toBeVisible()

  await page.getByPlaceholder('Message').fill('hello')
  await page.getByRole('button', { name: 'Load Key' }).click()
  await page.getByRole('button', { name: 'Build' }).click()
  await expect(page.getByText('bundle_cbor_b64')).toBeVisible({ timeout: 5000 })

  await page.getByRole('button', { name: 'Sign' }).click()
  await expect(page.getByText('Not logged in. Please')).toBeVisible()
})

test('options 409 → conflict error surfaced', async ({ page }) => {
  await preparePage(page)

  await page.route('**/tx/signing/options', async (route) => {
    await route.fulfill({
      status: 409,
      contentType: 'application/json',
      body: JSON.stringify({ message: 'conflict (nonce or credentials)' }),
    })
  })

  await page.goto('/')
  await page.evaluate(() => { window.location.hash = '#/dashboard' })
  await expect(page.getByRole('heading', { name: 'Dashboard' })).toBeVisible()
  await page.getByPlaceholder('Message').fill('hello')
  await page.getByRole('button', { name: 'Load Key' }).click()
  await page.getByRole('button', { name: 'Build' }).click()
  await page.getByRole('button', { name: 'Sign' }).click()

  await expect(page.getByText('Error: HTTP 409 — conflict (nonce or credentials)')).toBeVisible()
})

test('options 400 → invalid bundle error surfaced', async ({ page }) => {
  await preparePage(page)

  await page.route('**/tx/signing/options', async (route) => {
    await route.fulfill({
      status: 400,
      contentType: 'application/json',
      body: JSON.stringify({ message: 'invalid bundle' }),
    })
  })

  await page.goto('/')
  await page.evaluate(() => { window.location.hash = '#/dashboard' })
  await page.getByPlaceholder('Message').fill('hello')
  await page.getByRole('button', { name: 'Load Key' }).click()
  await page.getByRole('button', { name: 'Build' }).click()
  await page.getByRole('button', { name: 'Sign' }).click()

  await expect(page.getByText('Error: HTTP 400 — invalid bundle')).toBeVisible()
})

test('finish 409 → finish HTTP error surfaced', async ({ page }) => {
  await preparePage(page)

  await page.route('**/tx/signing/options', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        tx_session_id: 'txsess-2',
        challenge: 'AA',
        tx_id_hex: '00',
        expires_at: Math.floor(Date.now() / 1000) + 300,
        options: {
          rpId: 'localhost',
          origin: 'http://localhost:5173',
          timeout: 60000,
          userVerification: 'required',
          challenge: 'AA',
          allowCredentials: [{ type: 'public-key', id: 'AQID' }],
        },
      }),
    })
  })

  await page.route('**/tx/signing/finish', async (route) => {
    await route.fulfill({
      status: 409,
      contentType: 'application/json',
      body: JSON.stringify({ message: 'transaction conflict' }),
    })
  })

  await page.goto('/')
  await page.evaluate(() => { window.location.hash = '#/dashboard' })
  await page.getByPlaceholder('Message').fill('hello')
  await page.getByRole('button', { name: 'Load Key' }).click()
  await page.getByRole('button', { name: 'Build' }).click()
  await page.getByRole('button', { name: 'Sign' }).click()

  await expect(page.getByText('Error: HTTP 409 — transaction conflict')).toBeVisible()
})

test('finish 400 → finish HTTP error surfaced', async ({ page }) => {
  await preparePage(page)

  await page.route('**/tx/signing/options', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        tx_session_id: 'txsess-3',
        challenge: 'AA',
        tx_id_hex: '00',
        expires_at: Math.floor(Date.now() / 1000) + 300,
        options: {
          rpId: 'localhost',
          origin: 'http://localhost:5173',
          timeout: 60000,
          userVerification: 'required',
          challenge: 'AA',
          allowCredentials: [{ type: 'public-key', id: 'AQID' }],
        },
      }),
    })
  })

  await page.route('**/tx/signing/finish', async (route) => {
    await route.fulfill({
      status: 400,
      contentType: 'application/json',
      body: JSON.stringify({ message: 'verification failed' }),
    })
  })

  await page.goto('/')
  await page.evaluate(() => { window.location.hash = '#/dashboard' })
  await page.getByPlaceholder('Message').fill('hello')
  await page.getByRole('button', { name: 'Load Key' }).click()
  await page.getByRole('button', { name: 'Build' }).click()
  await page.getByRole('button', { name: 'Sign' }).click()

  await expect(page.getByText('Error: HTTP 400 — verification failed')).toBeVisible()
})
