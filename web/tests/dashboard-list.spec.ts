import { test, expect } from '@playwright/test'

test('dashboard shows unauthorized prompt when not logged in', async ({ page }) => {
  await page.goto('/')
  await page.evaluate(() => { window.location.hash = '#/dashboard' })
  await expect(page.getByRole('heading', { name: 'Dashboard' })).toBeVisible()
  // Either an unauthorized prompt appears or, if a session is already present from prior manual tests,
  // we allow an empty state. Prefer asserting the unauthorized text by default.
  const unauthorized = page.getByText('Not logged in. Please Login.')
  await expect(unauthorized).toBeVisible()
})
