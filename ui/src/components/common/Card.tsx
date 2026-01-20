import type { ReactNode, CSSProperties } from 'react';

interface Props {
  children: ReactNode;
  style?: CSSProperties;
  hoverable?: boolean;
}

export function Card({ children, style, hoverable }: Props) {
  return (
    <div
      style={{
        backgroundColor: 'var(--bg-card)',
        borderRadius: 'var(--radius-xl)',
        border: 'none',
        boxShadow: 'var(--shadow-md)',
        overflow: 'hidden',
        transition: hoverable ? 'all 0.3s cubic-bezier(0.16, 1, 0.3, 1)' : undefined,
        ...style,
      }}
    >
      {children}
    </div>
  );
}

export function CardHeader({ children, style }: { children: ReactNode; style?: CSSProperties }) {
  return (
    <div
      style={{
        padding: '18px 24px',
        borderBottom: '1px solid var(--border-light)',
        backgroundColor: 'var(--bg-card)',
        ...style,
      }}
    >
      {children}
    </div>
  );
}

export function CardBody({ children, style }: { children: ReactNode; style?: CSSProperties }) {
  return (
    <div style={{ padding: '24px', ...style }}>
      {children}
    </div>
  );
}
