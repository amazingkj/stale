import type { CSSProperties } from 'react';

/**
 * Table header cell styles
 */
export const tableHeaderStyle: CSSProperties = {
  padding: '10px 16px',
  textAlign: 'left',
  fontSize: '11px',
  fontWeight: 600,
  color: 'var(--text-muted)',
  textTransform: 'uppercase',
  letterSpacing: '0.5px',
};

/**
 * Table body cell styles
 */
export const tableCellStyle: CSSProperties = {
  padding: '12px 16px',
  fontSize: '13px',
  color: 'var(--text-primary)',
};

/**
 * Input field styles
 */
export const inputStyle: CSSProperties = {
  width: '100%',
  padding: '8px 10px',
  borderRadius: '6px',
  border: '1px solid var(--border-color)',
  backgroundColor: 'var(--bg-card)',
  color: 'var(--text-primary)',
  fontSize: '13px',
  outline: 'none',
  boxSizing: 'border-box',
};

/**
 * Select dropdown styles
 */
export const selectStyle: CSSProperties = {
  padding: '8px 10px',
  borderRadius: '6px',
  border: '1px solid var(--border-color)',
  backgroundColor: 'var(--bg-card)',
  color: 'var(--text-primary)',
  fontSize: '13px',
  outline: 'none',
  cursor: 'pointer',
};

/**
 * Primary button styles
 */
export const primaryButtonStyle: CSSProperties = {
  padding: '8px 16px',
  borderRadius: '6px',
  border: 'none',
  backgroundColor: 'var(--accent)',
  color: 'white',
  fontSize: '13px',
  fontWeight: 500,
  cursor: 'pointer',
};

/**
 * Secondary button styles
 */
export const secondaryButtonStyle: CSSProperties = {
  padding: '8px 14px',
  borderRadius: '6px',
  border: '1px solid var(--border-color)',
  backgroundColor: 'var(--bg-card)',
  color: 'var(--text-primary)',
  fontSize: '13px',
  cursor: 'pointer',
};

/**
 * Card container styles
 */
export const cardStyle: CSSProperties = {
  backgroundColor: 'var(--bg-card)',
  borderRadius: '10px',
  border: '1px solid var(--border-color)',
  overflow: 'hidden',
};

/**
 * Error message styles
 */
export const errorStyle: CSSProperties = {
  padding: '10px 14px',
  borderRadius: '6px',
  backgroundColor: 'var(--danger-bg)',
  color: 'var(--danger-text)',
  fontSize: '13px',
};

/**
 * Badge color configurations
 */
export const badgeColors = {
  success: {
    backgroundColor: 'var(--success-bg)',
    color: 'var(--success-text)',
  },
  warning: {
    backgroundColor: 'var(--warning-bg)',
    color: 'var(--warning-text)',
  },
  danger: {
    backgroundColor: 'var(--danger-bg)',
    color: 'var(--danger-text)',
  },
  accent: {
    backgroundColor: 'var(--accent)',
    color: 'white',
  },
  muted: {
    backgroundColor: 'var(--bg-hover)',
    color: 'var(--text-muted)',
  },
  purple: {
    backgroundColor: 'var(--purple-bg)',
    color: 'var(--purple-text)',
  },
} as const;

/**
 * Stat card color configurations
 */
export const statCardColors = {
  accent: { bg: 'var(--bg-hover)', text: 'var(--accent)' },
  success: { bg: 'var(--success-bg)', text: 'var(--success-text)' },
  warning: { bg: 'var(--warning-bg)', text: 'var(--warning-text)' },
  danger: { bg: 'var(--danger-bg)', text: 'var(--danger-text)' },
} as const;

export type StatCardColor = keyof typeof statCardColors;
export type BadgeColor = keyof typeof badgeColors;
