import type { CSSProperties } from 'react';

/**
 * Table header cell styles
 */
export const tableHeaderStyle: CSSProperties = {
  padding: '12px 20px',
  textAlign: 'left',
  fontSize: '12px',
  fontWeight: 600,
  color: 'var(--text-muted)',
  textTransform: 'uppercase',
  letterSpacing: '0.5px',
};

/**
 * Table body cell styles
 */
export const tableCellStyle: CSSProperties = {
  padding: '14px 20px',
  fontSize: '14px',
  color: 'var(--text-primary)',
};

/**
 * Input field styles
 */
export const inputStyle: CSSProperties = {
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

/**
 * Select dropdown styles
 */
export const selectStyle: CSSProperties = {
  padding: '10px 12px',
  borderRadius: '8px',
  border: '1px solid var(--border-color)',
  backgroundColor: 'var(--bg-card)',
  color: 'var(--text-primary)',
  fontSize: '14px',
  outline: 'none',
  cursor: 'pointer',
};

/**
 * Primary button styles
 */
export const primaryButtonStyle: CSSProperties = {
  padding: '10px 20px',
  borderRadius: '8px',
  border: 'none',
  backgroundColor: 'var(--accent)',
  color: 'white',
  fontSize: '14px',
  fontWeight: 500,
  cursor: 'pointer',
};

/**
 * Secondary button styles
 */
export const secondaryButtonStyle: CSSProperties = {
  padding: '10px 16px',
  borderRadius: '8px',
  border: '1px solid var(--border-color)',
  backgroundColor: 'var(--bg-card)',
  color: 'var(--text-primary)',
  fontSize: '14px',
  cursor: 'pointer',
};

/**
 * Card container styles
 */
export const cardStyle: CSSProperties = {
  backgroundColor: 'var(--bg-card)',
  borderRadius: '12px',
  border: '1px solid var(--border-color)',
  overflow: 'hidden',
};

/**
 * Error message styles
 */
export const errorStyle: CSSProperties = {
  padding: '12px 16px',
  borderRadius: '8px',
  backgroundColor: 'var(--danger-bg)',
  color: 'var(--danger-text)',
  fontSize: '14px',
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
