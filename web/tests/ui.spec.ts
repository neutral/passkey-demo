import { test, expect } from '@playwright/test'

test('home shows exactly two buttons and navigation works', async ({ page }) => {
  await page.goto('/')
  // Home header
  await expect(page.getByRole('heading', { name: 'Passkey Demo' })).toBeVisible()

  // Two primary buttons
  const buttons = page.getByRole('button')
  await expect(buttons).toHaveCount(2)
  await expect(page.getByRole('button', { name: 'Register' })).toBeVisible()
  await expect(page.getByRole('button', { name: 'Login' })).toBeVisible()

  // Navigate to Register
  await page.getByRole('button', { name: 'Register' }).click()
  await expect(page.getByRole('heading', { name: 'Register' })).toBeVisible()
  await page.getByRole('button', { name: 'Back' }).click()
  await expect(page.getByRole('heading', { name: 'Passkey Demo' })).toBeVisible()

  // Navigate to Login
  await page.getByRole('button', { name: 'Login' }).click()
  await expect(page.getByRole('heading', { name: 'Login' })).toBeVisible()
  await page.getByRole('button', { name: 'Back' }).click()
  await expect(page.getByRole('heading', { name: 'Passkey Demo' })).toBeVisible()
})

