import type { CSSProperties } from 'react';

interface Props {
  size?: 'sm' | 'md' | 'lg';
  text?: string;
  fullPage?: boolean;
}

const sizes = {
  sm: 16,
  md: 24,
  lg: 32,
};

export function LoadingSpinner({ size = 'md', text, fullPage }: Props) {
  const spinnerSize = sizes[size];

  const containerStyle: CSSProperties = fullPage
    ? {
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        justifyContent: 'center',
        padding: '48px',
        color: 'var(--text-muted)',
        gap: '12px',
      }
    : {
        display: 'inline-flex',
        alignItems: 'center',
        gap: '8px',
        color: 'var(--text-muted)',
      };

  return (
    <div style={containerStyle}>
      <svg
        width={spinnerSize}
        height={spinnerSize}
        viewBox="0 0 24 24"
        style={{ animation: 'spin 1s linear infinite' }}
      >
        <style>{`@keyframes spin { from { transform: rotate(0deg); } to { transform: rotate(360deg); } }`}</style>
        <circle
          cx="12"
          cy="12"
          r="10"
          stroke="currentColor"
          strokeWidth="3"
          fill="none"
          opacity="0.25"
        />
        <path
          d="M12 2a10 10 0 0 1 10 10"
          stroke="currentColor"
          strokeWidth="3"
          fill="none"
          strokeLinecap="round"
        />
      </svg>
      {text && <span style={{ fontSize: '14px' }}>{text}</span>}
    </div>
  );
}
