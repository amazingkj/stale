import type { CSSProperties } from 'react';
import { errorStyle } from '../../constants/styles';

interface Props {
  message: string;
  onDismiss?: () => void;
  style?: CSSProperties;
}

export function ErrorMessage({ message, onDismiss, style }: Props) {
  return (
    <div
      style={{
        ...errorStyle,
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'space-between',
        ...style,
      }}
      role="alert"
    >
      <span>{message}</span>
      {onDismiss && (
        <button
          onClick={onDismiss}
          style={{
            background: 'none',
            border: 'none',
            color: 'inherit',
            cursor: 'pointer',
            padding: '4px',
            marginLeft: '8px',
            opacity: 0.7,
          }}
          aria-label="Dismiss error"
        >
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
            <path d="M18 6L6 18M6 6l12 12" />
          </svg>
        </button>
      )}
    </div>
  );
}
