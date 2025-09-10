import { test, expect } from '@playwright/test'

test('signing flow (mocked): options → get → finish and refresh', async ({ page }) => {
  page.on('console', (m) => console.log('[console]', m.type(), m.text()))
  // Stub navigator.credentials.get before any script runs
  await page.addInitScript(() => {
    // @ts-ignore
    const fakeGet = async () => {
      return {
        id: 'cred-id',
        type: 'public-key',
        rawId: new Uint8Array([1, 2, 3, 4]).buffer,
        response: {
          authenticatorData: new Uint8Array([5]).buffer,
          clientDataJSON: new TextEncoder().encode('{"type":"webauthn.get"}').buffer,
          signature: new Uint8Array([6]).buffer,
          userHandle: new Uint8Array([7]).buffer,
        },
      } as any
    }
    if (navigator.credentials) {
      // @ts-ignore
      navigator.credentials.get = fakeGet
    } else {
      // @ts-ignore
      navigator.credentials = { get: fakeGet }
    }
  })

  // Mock /me/account_key to return a simple COSE EC2 key (small x/y accepted by client)
  await page.route('**/me/account_key', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        acct_cbor_b64: 'AA',
        sender_key: { kty: 2, alg: -7, crv: 1, x: 'AQ', y: 'AQ' },
      }),
    })
  })

  // Mock /tx/list to avoid unauthorized state and keep Build/Sign enabled
  await page.route('**/tx/list', async (route) => {
    await route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ items: [] }) })
  })

  // Capture options and finish requests
  let optionsBody: any | null = null
  await page.route('**/tx/signing/options', async (route) => {
    const postData = route.request().postData()
    optionsBody = postData ? JSON.parse(postData) : null
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        tx_session_id: 'txsess-1',
        challenge: 'AA',
        tx_id_hex: '00',
        options: { rp_id: 'localhost', origin: 'http://localhost:5173', uv_required: true, allow_credentials: ['AQID'] },
        expires_at: Math.floor(Date.now() / 1000) + 300,
      }),
    })
  })

  let finishBody: any | null = null
  await page.route('**/tx/signing/finish', async (route) => {
    const postData = route.request().postData()
    finishBody = postData ? JSON.parse(postData) : null
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ tx_id_hex: 'abcd', stored: true }),
    })
  })

  await page.goto('/')
  // Navigate to dashboard
  await page.evaluate(() => { window.location.hash = '#/dashboard' })
  await expect(page.getByRole('heading', { name: 'Dashboard' })).toBeVisible()

  // Enter inputs, load key, build bundle, then sign
  await page.getByPlaceholder('Message').fill('hello world')
  await page.getByPlaceholder('Nonce').fill('123')
  await page.getByRole('button', { name: 'Load Key' }).click()
  await page.getByRole('button', { name: 'Build' }).click()
  // Ensure preview appears
  await expect(page.getByText('bundle_cbor_b64')).toBeVisible({ timeout: 5000 })
  await page.getByRole('button', { name: 'Sign' }).click()

  // Assert options was called with bundle
  expect(optionsBody).toBeTruthy()
  expect(typeof optionsBody.bundle_cbor_b64).toBe('string')
  expect(optionsBody.bundle_cbor_b64.length).toBeGreaterThan(0)

  // Assert finish payload contains expected fields
  expect(finishBody).toBeTruthy()
  expect(finishBody.tx_session_id).toBe('txsess-1')
  expect(typeof finishBody.rawId).toBe('string')
  expect(typeof finishBody.response?.clientDataJSON).toBe('string')

  // After success, preview should clear (no bundle textareas present)
  await expect(page.getByText('bundle_cbor_b64')).toHaveCount(0)
})
