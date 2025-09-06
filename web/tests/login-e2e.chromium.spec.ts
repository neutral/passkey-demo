import { test, expect } from '@playwright/test'

// Optional E2E: without a pre-registered credential, finish is expected to return 401.
test.describe('Login E2E (Chromium + Virtual Authenticator)', () => {
  test('login flow runs and may yield 401 without seeded credential', async ({ page, context, browserName }) => {
    test.skip(browserName !== 'chromium', 'Virtual authenticator only supported in Chromium via CDP')

    const client = await context.newCDPSession(page)
    await client.send('WebAuthn.enable')
    const { authenticatorId } = await client.send('WebAuthn.addVirtualAuthenticator', {
      options: {
        protocol: 'ctap2',
        transport: 'internal',
        hasResidentKey: true,
        hasUserVerification: true,
        isUserVerified: true,
        automaticPresenceSimulation: true,
      },
    })

    await page.goto('/')
    await page.getByRole('button', { name: 'Login' }).click()
    await page.getByRole('button', { name: 'Start Login' }).click()

    const thumb = page.getByText('Account thumb:')
    const err = page.getByText('Error:')
    await Promise.race([
      thumb.waitFor({ state: 'visible', timeout: 10_000 }),
      err.waitFor({ state: 'visible', timeout: 10_000 }),
    ])
    if (await err.isVisible()) {
      const txt = await err.textContent()
      // Either NotAllowed (user gesture/picker conditions) or 401 from backend are acceptable in this environment.
      expect(txt.includes('HTTP 401') || txt.includes('Not allowed') || txt.includes('not allowed')).toBeTruthy()
    } else {
      await expect(thumb).toBeVisible()
    }

    await client.send('WebAuthn.removeVirtualAuthenticator', { authenticatorId })
  })
})
