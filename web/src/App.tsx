import { useEffect, useState } from 'react'
import './App.css'
import Register from './pages/Register'
import Login from './pages/Login'
import Dashboard from './pages/Dashboard'

type Route = 'home' | 'register' | 'login' | 'dashboard'

function App() {
  const [route, setRoute] = useState<Route>('home')

  // Optional: sync initial state with hash (e.g., #/register)
  useEffect(() => {
    const parseHash = (): Route => {
      const h = window.location.hash.replace(/^#\//, '')
      if (h === 'register' || h === 'login' || h === 'dashboard') return h
      return 'home'
    }
    setRoute(parseHash())
    const onHash = () => setRoute(parseHash())
    window.addEventListener('hashchange', onHash)
    return () => window.removeEventListener('hashchange', onHash)
  }, [])

  // Keep hash in sync with route (no navigation library)
  useEffect(() => {
    const expected = route === 'home' ? '' : `#/${route}`
    if (window.location.hash !== expected) {
      window.location.hash = expected
    }
  }, [route])

  if (route === 'register') {
    return <Register onBack={() => setRoute('home')} />
  }
  if (route === 'login') {
    return <Login onBack={() => setRoute('home')} />
  }
  if (route === 'dashboard') {
    return <Dashboard onBack={() => setRoute('home')} />
  }

  // Home view with exactly two primary actions
  return (
    <div style={{ padding: 24 }}>
      <h1>Passkey Demo</h1>
      <p>Choose an action to continue:</p>
      <div style={{ display: 'flex', gap: 12, marginTop: 12 }}>
        <button type="button" onClick={() => setRoute('register')}>Register</button>
        <button type="button" onClick={() => setRoute('login')}>Login</button>
      </div>
    </div>
  )
}

export default App
