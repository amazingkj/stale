import type { CSSProperties } from 'react';

/**
 * Table header cell styles
 */
export const tableHeaderStyle: CSSProperties = {
  padding: '14px 18px',
  textAlign: 'left',
  fontSize: '11px',
  fontWeight: 700,
  color: 'var(--text-muted)',
  textTransform: 'uppercase',
  letterSpacing: '0.6px',
};

/**
 * Table body cell styles
 */
export const tableCellStyle: CSSProperties = {
  padding: '14px 18px',
  fontSize: '13px',
  color: 'var(--text-primary)',
};

/**
 * Input field styles
 */
export const inputStyle: CSSProperties = {
  width: '100%',
  padding: '10px 14px',
  borderRadius: 'var(--radius-md)',
  border: '1.5px solid var(--border-color)',
  backgroundColor: 'var(--bg-card)',
  color: 'var(--text-primary)',
  fontSize: '13px',
  outline: 'none',
  boxSizing: 'border-box',
  transition: 'all 0.2s ease',
};

/**
 * Select dropdown styles
 */
export const selectStyle: CSSProperties = {
  padding: '10px 14px',
  borderRadius: 'var(--radius-md)',
  border: '1.5px solid var(--border-color)',
  backgroundColor: 'var(--bg-card)',
  color: 'var(--text-primary)',
  fontSize: '13px',
  outline: 'none',
  cursor: 'pointer',
  transition: 'all 0.2s ease',
};

/**
 * Primary button styles
 */
export const primaryButtonStyle: CSSProperties = {
  padding: '10px 20px',
  borderRadius: 'var(--radius-full)',
  border: 'none',
  background: 'var(--accent-gradient)',
  color: 'white',
  fontSize: '13px',
  fontWeight: 600,
  cursor: 'pointer',
  boxShadow: '0 2px 8px -2px rgba(124, 181, 149, 0.4)',
};

/**
 * Secondary button styles
 */
export const secondaryButtonStyle: CSSProperties = {
  padding: '10px 18px',
  borderRadius: 'var(--radius-full)',
  border: '1.5px solid var(--border-color)',
  backgroundColor: 'var(--bg-card)',
  color: 'var(--text-primary)',
  fontSize: '13px',
  fontWeight: 500,
  cursor: 'pointer',
};

/**
 * Card container styles
 */
export const cardStyle: CSSProperties = {
  backgroundColor: 'var(--bg-card)',
  borderRadius: 'var(--radius-xl)',
  border: 'none',
  boxShadow: 'var(--shadow-md)',
  overflow: 'hidden',
};

/**
 * Error message styles
 */
export const errorStyle: CSSProperties = {
  padding: '12px 16px',
  borderRadius: 'var(--radius-md)',
  backgroundColor: 'var(--danger-bg)',
  color: 'var(--danger-text)',
  fontSize: '13px',
  fontWeight: 500,
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
    backgroundColor: 'var(--accent-light)',
    color: 'var(--accent)',
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
  accent: { bg: 'var(--accent-light)', text: 'var(--accent)' },
  success: { bg: 'var(--success-bg)', text: 'var(--success-text)' },
  warning: { bg: 'var(--warning-bg)', text: 'var(--warning-text)' },
  danger: { bg: 'var(--danger-bg)', text: 'var(--danger-text)' },
} as const;

export type StatCardColor = keyof typeof statCardColors;
export type BadgeColor = keyof typeof badgeColors;
