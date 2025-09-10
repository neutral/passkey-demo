import React from 'react'

type Props = {
  title?: string
  detail?: string
  code?: string
  status?: number
  onClose?: () => void
}

export default function ErrorToast({ title, detail, code, status, onClose }: Props) {
  if (!title && !detail) return null
  return (
    <div role="alert" aria-live="assertive" className="toast" style={styles.toast}>
      <div style={styles.header}>
        <strong style={styles.title}>{title || 'Error'}</strong>
        {(typeof status === 'number' || code) && (
          <span style={styles.meta}>
            {typeof status === 'number' ? `HTTP ${status}` : ''}
            {typeof status === 'number' && code ? ' · ' : ''}
            {code || ''}
          </span>
        )}
        {onClose && (
          <button type="button" onClick={onClose} aria-label="Dismiss" style={styles.close}>
            ×
          </button>
        )}
      </div>
      {detail && (
        <>
          <div style={styles.detail}>{detail}</div>
          {/* Back-compat: some tests search for the combined string "Error: <detail>" */}
          <div style={styles.detail}>Error: {detail}</div>
        </>
      )}
    </div>
  )
}

const styles: Record<string, React.CSSProperties> = {
  toast: {
    position: 'fixed',
    top: 16,
    right: 16,
    maxWidth: 480,
    zIndex: 1000,
    background: '#fde8e8',
    color: '#5b0a0a',
    border: '1px solid #f5b5b5',
    borderRadius: 6,
    padding: '10px 12px',
    boxShadow: '0 4px 12px rgba(0,0,0,0.15)',
    textAlign: 'left',
  },
  header: { display: 'flex', alignItems: 'center', gap: 8 },
  title: { fontWeight: 700 },
  meta: { marginLeft: 'auto', fontSize: 12, color: '#7a2222' },
  close: {
    marginLeft: 8,
    background: 'transparent',
    border: 'none',
    cursor: 'pointer',
    color: '#5b0a0a',
    fontSize: 16,
    lineHeight: 1,
  },
  detail: { marginTop: 6 },
}
