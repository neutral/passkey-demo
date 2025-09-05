type Props = {
  onBack: () => void
}

export default function Register({ onBack }: Props) {
  return (
    <div style={{ padding: 24 }}>
      <h1>Register</h1>
      <p>Use your platform authenticator to create a passkey.</p>
      <div style={{ display: 'flex', gap: 12, marginTop: 12 }}>
        <button type="button" disabled>
          Start Registration (coming in Step 29)
        </button>
        <button type="button" onClick={onBack}>
          Back
        </button>
      </div>
    </div>
  )
}
