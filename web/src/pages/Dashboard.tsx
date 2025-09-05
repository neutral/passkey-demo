type Props = {
  onBack: () => void
}

export default function Dashboard({ onBack }: Props) {
  return (
    <div style={{ padding: 24 }}>
      <h1>Dashboard</h1>
      <p>Signed messages will appear here.</p>

      <section style={{ marginTop: 16 }}>
        <h2>Sign a Message</h2>
        <div style={{ display: 'flex', flexDirection: 'column', gap: 8, maxWidth: 520 }}>
          <input placeholder="Message" disabled />
          <input placeholder="Nonce" disabled />
          <button type="button" disabled>
            Sign (coming in Steps 32–33)
          </button>
        </div>
      </section>

      <section style={{ marginTop: 24 }}>
        <h2>Transactions</h2>
        <p>(List coming in Step 31)</p>
      </section>

      <div style={{ marginTop: 24 }}>
        <button type="button" onClick={onBack}>Back</button>
      </div>
    </div>
  )
}
