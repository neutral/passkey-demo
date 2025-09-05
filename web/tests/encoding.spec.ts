import { test, expect } from '@playwright/test'

async function importEncoding(page: import('@playwright/test').Page) {
  await page.goto('/')
  // Dynamically import from Vite dev server
  const mod = await page.evaluate(async () => {
    const m = await import('/src/lib/encoding.ts')
    return {
      has: {
        bytesToBase64url: typeof m.bytesToBase64url === 'function',
        base64urlToBytes: typeof m.base64urlToBytes === 'function',
        utf8ToBytes: typeof m.utf8ToBytes === 'function',
        bytesToUtf8: typeof m.bytesToUtf8 === 'function',
      },
    }
  })
  expect(mod.has).toEqual({
    bytesToBase64url: true,
    base64urlToBytes: true,
    utf8ToBytes: true,
    bytesToUtf8: true,
  })
}

test('bytes ⇄ base64url round-trip and character set', async ({ page }) => {
  await page.goto('/')
  const result = await page.evaluate(async () => {
    const { bytesToBase64url, base64urlToBytes } = await import('/src/lib/encoding.ts')
    const cases: number[][] = [
      [], [0], [0, 255], [1, 2, 3, 4, 5]
    ]
    const alpha = /^[A-Za-z0-9_-]*$/
    const ok = cases.every(arr => {
      const u8 = new Uint8Array(arr)
      const s = bytesToBase64url(u8)
      if (!alpha.test(s)) return false
      const back = base64urlToBytes(s)
      if (back.length !== u8.length) return false
      for (let i = 0; i < back.length; i++) if (back[i] !== u8[i]) return false
      return true
    })
    return ok
  })
  expect(result).toBeTruthy()
})

test('UTF-8 ⇄ bytes round-trip including multibyte', async ({ page }) => {
  await page.goto('/')
  const ok = await page.evaluate(async () => {
    const { utf8ToBytes, bytesToUtf8 } = await import('/src/lib/encoding.ts')
    const samples = ['', 'foo', 'Hello, world!', 'π🙂']
    return samples.every(s => bytesToUtf8(utf8ToBytes(s)) === s)
  })
  expect(ok).toBeTruthy()
})

test('RFC 4648 known vectors (no padding)', async ({ page }) => {
  await page.goto('/')
  const outputs = await page.evaluate(async () => {
    const { utf8ToBytes, bytesToBase64url } = await import('/src/lib/encoding.ts')
    const inputs = ['f', 'fo', 'foo', 'foob', 'fooba', 'foobar']
    return inputs.map(s => bytesToBase64url(utf8ToBytes(s)))
  })
  expect(outputs).toEqual(['Zg', 'Zm8', 'Zm9v', 'Zm9vYg', 'Zm9vYmE', 'Zm9vYmFy'])
})

test('decoder rejects invalid characters and accepts padding', async ({ page }) => {
  await page.goto('/')
  const res = await page.evaluate(async () => {
    const { base64urlToBytes, bytesToUtf8 } = await import('/src/lib/encoding.ts')
    let invalidRejected = false
    try {
      base64urlToBytes('Z*g') // '*' invalid
    } catch (e) {
      invalidRejected = e instanceof TypeError
    }
    const padded = bytesToUtf8(base64urlToBytes('Zg==')) // should decode to 'f'
    const unpadded = bytesToUtf8(base64urlToBytes('Zg'))
    return { invalidRejected, padded, unpadded }
  })
  expect(res.invalidRejected).toBeTruthy()
  expect(res.padded).toBe('f')
  expect(res.unpadded).toBe('f')
})

test('large random input round-trip (4 KiB)', async ({ page }) => {
  await page.goto('/')
  const ok = await page.evaluate(async () => {
    const { bytesToBase64url, base64urlToBytes } = await import('/src/lib/encoding.ts')
    const size = 4096
    const buf = new Uint8Array(size)
    crypto.getRandomValues(buf)
    const s = bytesToBase64url(buf)
    const back = base64urlToBytes(s)
    if (back.length !== buf.length) return false
    for (let i = 0; i < size; i++) if (back[i] !== buf[i]) return false
    return true
  })
  expect(ok).toBeTruthy()
})

