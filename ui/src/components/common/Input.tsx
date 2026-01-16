import type { InputHTMLAttributes, CSSProperties } from 'react';

interface Props extends InputHTMLAttributes<HTMLInputElement> {
  label?: string;
  error?: string;
  wrapperStyle?: CSSProperties;
}

const inputBaseStyle: CSSProperties = {
  width: '100%',
  padding: '10px 12px',
  borderRadius: '8px',
  border: '1px solid var(--border-color)',
  backgroundColor: 'var(--bg-card)',
  color: 'var(--text-primary)',
  fontSize: '14px',
  outline: 'none',
  boxSizing: 'border-box',
};

export function Input({ label, error, id, wrapperStyle, style, ...props }: Props) {
  const inputId = id || (label ? label.toLowerCase().replace(/\s+/g, '-') : undefined);

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: '6px', ...wrapperStyle }}>
      {label && (
        <label
          htmlFor={inputId}
          style={{
            fontSize: '14px',
            fontWeight: 500,
            color: 'var(--text-primary)',
          }}
        >
          {label}
        </label>
      )}
      <input
        id={inputId}
        style={{
          ...inputBaseStyle,
          borderColor: error ? 'var(--danger)' : 'var(--border-color)',
          ...style,
        }}
        aria-invalid={error ? 'true' : undefined}
        aria-describedby={error ? `${inputId}-error` : undefined}
        {...props}
      />
      {error && (
        <span
          id={`${inputId}-error`}
          style={{
            fontSize: '13px',
            color: 'var(--danger-text)',
          }}
          role="alert"
        >
          {error}
        </span>
      )}
    </div>
  );
}
