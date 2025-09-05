type Props = {
  onBack: () => void
}

export default function Login({ onBack }: Props) {
  return (
    <div style={{ padding: 24 }}>
      <h1>Login</h1>
      <p>Authenticate with your passkey to start a session.</p>
      <div style={{ display: 'flex', gap: 12, marginTop: 12 }}>
        <button type="button" disabled>
          Start Login (coming in Step 30)
        </button>
        <button type="button" onClick={onBack}>
          Back
        </button>
      </div>
    </div>
  )
}
