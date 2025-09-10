import { test, expect } from '@playwright/test'

test('buildTxFinish encodes assertion fields and includes tx_session_id', async ({ page }) => {
  await page.goto('/')
  const ok = await page.evaluate(async () => {
    const { buildTxFinish } = await import('/src/lib/webauthn.ts')
    const cred = {
      id: 'cred-id',
      type: 'public-key',
      rawId: new Uint8Array([1, 2, 3, 4]).buffer,
      response: {
        authenticatorData: new Uint8Array([5, 6, 7]).buffer,
        clientDataJSON: new TextEncoder().encode('{"type":"webauthn.get"}').buffer,
        signature: new Uint8Array([8, 9]).buffer,
        userHandle: new Uint8Array([10]).buffer,
      },
    } as any
    const out = buildTxFinish(cred, 'txsess-1')
    return (
      out.tx_session_id === 'txsess-1' &&
      out.id === 'cred-id' &&
      typeof out.rawId === 'string' &&
      typeof out.response.authenticatorData === 'string' &&
      typeof out.response.clientDataJSON === 'string' &&
      typeof out.response.signature === 'string'
    )
  })
  expect(ok).toBeTruthy()
})

