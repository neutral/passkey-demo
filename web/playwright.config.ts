import { defineConfig, devices } from '@playwright/test'

export default defineConfig({
  testDir: './tests',
  fullyParallel: true,
  retries: process.env.CI ? 2 : 0,
  use: {
    baseURL: 'http://localhost:5173',
    trace: 'on-first-retry',
  },
  projects: [
    { name: 'chromium', use: { ...devices['Desktop Chrome'] } },
  ],
  webServer: [
    {
      command: 'npm run dev',
      port: 5173,
      reuseExistingServer: !process.env.CI,
      timeout: 60_000,
    },
    {
      command: 'RP_ID=localhost ORIGIN=http://localhost:5173 PORT=8080 go -C ../server run ./cmd/api',
      port: 8080,
      reuseExistingServer: !process.env.CI,
      timeout: 60_000,
    },
  ],
})
