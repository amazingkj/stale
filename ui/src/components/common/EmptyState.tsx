import type { CSSProperties, ReactNode } from 'react';

interface Props {
  icon?: string;
  title?: string;
  description: string;
  action?: ReactNode;
}

const containerStyle: CSSProperties = {
  padding: '60px 24px',
  textAlign: 'center',
  display: 'flex',
  flexDirection: 'column',
  alignItems: 'center',
};

const iconContainerStyle: CSSProperties = {
  width: '80px',
  height: '80px',
  borderRadius: 'var(--radius-full)',
  backgroundColor: 'var(--bg-secondary)',
  display: 'flex',
  alignItems: 'center',
  justifyContent: 'center',
  marginBottom: '20px',
};

const iconStyle: CSSProperties = {
  fontSize: '36px',
};

const titleStyle: CSSProperties = {
  fontSize: '18px',
  fontWeight: 600,
  color: 'var(--text-primary)',
  marginBottom: '8px',
};

const descriptionStyle: CSSProperties = {
  color: 'var(--text-muted)',
  fontSize: '14px',
  maxWidth: '400px',
  lineHeight: 1.6,
  whiteSpace: 'pre-line',
};

export function EmptyState({ icon, title, description, action }: Props) {
  return (
    <div style={containerStyle}>
      {icon && (
        <div style={iconContainerStyle}>
          <span style={iconStyle}>{icon}</span>
        </div>
      )}
      {title && <div style={titleStyle}>{title}</div>}
      <p style={descriptionStyle}>{description}</p>
      {action && <div style={{ marginTop: '20px' }}>{action}</div>}
    </div>
  );
}
