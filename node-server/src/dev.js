import 'dotenv/config'
import getPort from 'get-port'
import { start } from './server.js'

// Dev launcher: picks an available port (prefers env.PORT or 8080)
// to avoid EADDRINUSE when another instance is running.
(async () => {
  const preferred = Number(process.env.PORT || 8080)
  const candidates = [preferred, preferred + 1, preferred + 2]
  const port = await getPort({ port: candidates })
  if (port !== preferred) {
    console.warn(`[dev] Port ${preferred} busy, using ${port}`)
  }
  process.env.PORT = String(port)
  start()
})()

