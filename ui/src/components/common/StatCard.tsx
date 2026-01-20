import type { CSSProperties, ReactNode } from 'react';
import { statCardColors, type StatCardColor } from '../../constants/styles';

interface Props {
  label: string;
  value: number;
  subtitle?: string;
  color: StatCardColor;
  icon?: ReactNode;
  active?: boolean;
  onClick?: () => void;
}

const icons: Record<string, ReactNode> = {
  total: (
    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
      <path d="M16.5 9.4l-9-5.19M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z" />
      <polyline points="3.27 6.96 12 12.01 20.73 6.96" />
      <line x1="12" y1="22.08" x2="12" y2="12" />
    </svg>
  ),
  upgradable: (
    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
      <circle cx="12" cy="12" r="10" />
      <polyline points="16 12 12 8 8 12" />
      <line x1="12" y1="16" x2="12" y2="8" />
    </svg>
  ),
  uptodate: (
    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
      <path d="M22 11.08V12a10 10 0 1 1-5.93-9.14" />
      <polyline points="22 4 12 14.01 9 11.01" />
    </svg>
  ),
  prod: (
    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
      <path d="M12 2L2 7l10 5 10-5-10-5z" />
      <path d="M2 17l10 5 10-5" />
      <path d="M2 12l10 5 10-5" />
    </svg>
  ),
  dev: (
    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
      <polyline points="16 18 22 12 16 6" />
      <polyline points="8 6 2 12 8 18" />
    </svg>
  ),
};

export function StatCard({ label, value, subtitle, color, icon, active, onClick }: Props) {
  const colors = statCardColors[color];
  const labelLower = label.toLowerCase().replace(/\s+/g, '');
  const displayIcon = icon || icons[labelLower];

  const cardStyle: CSSProperties = {
    padding: '16px',
    borderRadius: 'var(--radius-md)',
    backgroundColor: active ? colors.bg : 'var(--bg-card)',
    cursor: onClick ? 'pointer' : 'default',
    border: `2px solid ${active ? colors.text : 'var(--border-color)'}`,
    boxShadow: active ? 'var(--shadow-md)' : 'var(--shadow-sm)',
    transition: 'all 0.2s ease',
    transform: active ? 'translateY(-2px)' : 'translateY(0)',
  };

  return (
    <div
      onClick={onClick}
      style={cardStyle}
      role={onClick ? 'button' : undefined}
      tabIndex={onClick ? 0 : undefined}
      onKeyDown={onClick ? (e) => e.key === 'Enter' && onClick() : undefined}
      aria-pressed={onClick ? active : undefined}
      onMouseEnter={(e) => {
        if (onClick && !active) {
          e.currentTarget.style.borderColor = 'var(--border-color)';
          e.currentTarget.style.boxShadow = 'var(--shadow-md)';
          e.currentTarget.style.transform = 'translateY(-2px)';
        }
      }}
      onMouseLeave={(e) => {
        if (onClick && !active) {
          e.currentTarget.style.borderColor = 'var(--border-color)';
          e.currentTarget.style.boxShadow = 'var(--shadow-sm)';
          e.currentTarget.style.transform = 'translateY(0)';
        }
      }}
    >
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', marginBottom: '10px' }}>
        <div style={{
          fontSize: '11px',
          fontWeight: 500,
          color: active ? colors.text : 'var(--text-secondary)',
          textTransform: 'uppercase',
          letterSpacing: '0.5px',
        }}>
          {label}
        </div>
        {displayIcon && (
          <div style={{
            color: active ? colors.text : 'var(--text-muted)',
            opacity: active ? 1 : 0.6,
          }}>
            {displayIcon}
          </div>
        )}
      </div>
      <div style={{
        fontSize: '30px',
        fontWeight: 700,
        color: active ? colors.text : 'var(--text-primary)',
        lineHeight: 1,
        letterSpacing: '-1px',
      }}>
        {value.toLocaleString()}
      </div>
      {subtitle && (
        <div style={{
          fontSize: '12px',
          color: active ? colors.text : 'var(--text-muted)',
          marginTop: '6px',
          opacity: active ? 0.8 : 1,
        }}>
          {subtitle}
        </div>
      )}
    </div>
  );
}
