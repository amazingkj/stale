import type { SelectHTMLAttributes, CSSProperties } from 'react';

interface Props extends SelectHTMLAttributes<HTMLSelectElement> {
  label?: string;
  wrapperStyle?: CSSProperties;
}

const selectBaseStyle: CSSProperties = {
  padding: '10px 12px',
  borderRadius: '8px',
  border: '1px solid var(--border-color)',
  backgroundColor: 'var(--bg-card)',
  color: 'var(--text-primary)',
  fontSize: '14px',
  outline: 'none',
  cursor: 'pointer',
  boxSizing: 'border-box',
};

export function Select({ label, id, children, wrapperStyle, style, ...props }: Props) {
  const selectId = id || (label ? label.toLowerCase().replace(/\s+/g, '-') : undefined);

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: '6px', ...wrapperStyle }}>
      {label && (
        <label
          htmlFor={selectId}
          style={{
            fontSize: '14px',
            fontWeight: 500,
            color: 'var(--text-primary)',
          }}
        >
          {label}
        </label>
      )}
      <select
        id={selectId}
        style={{
          ...selectBaseStyle,
          ...style,
        }}
        {...props}
      >
        {children}
      </select>
    </div>
  );
}
