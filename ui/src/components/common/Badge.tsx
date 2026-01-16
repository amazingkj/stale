import type { CSSProperties, ReactNode } from 'react';
import { badgeColors, type BadgeColor } from '../../constants/styles';

interface Props {
  children: ReactNode;
  color?: BadgeColor;
  style?: CSSProperties;
}

const baseStyle: CSSProperties = {
  padding: '4px 8px',
  borderRadius: '6px',
  fontSize: '12px',
  fontWeight: 500,
  display: 'inline-flex',
  alignItems: 'center',
  gap: '4px',
};

export function Badge({ children, color = 'muted', style }: Props) {
  return (
    <span
      style={{
        ...baseStyle,
        ...badgeColors[color],
        ...style,
      }}
    >
      {children}
    </span>
  );
}

// Specialized badges for common use cases
export function EcosystemBadge({ ecosystem }: { ecosystem: string }) {
  const config: Record<string, { color: BadgeColor; label: string }> = {
    npm: { color: 'warning', label: 'npm' },
    go: { color: 'accent', label: 'go' },
    maven: { color: 'danger', label: 'maven' },
    gradle: { color: 'success', label: 'gradle' },
  };

  const { color, label } = config[ecosystem] || { color: 'muted', label: ecosystem };

  return (
    <Badge color={color} style={{ fontSize: '12px', fontWeight: 600, padding: '4px 8px' }}>
      {label}
    </Badge>
  );
}

export function TypeBadge({ type }: { type: 'dependency' | 'devDependency' }) {
  return (
    <Badge
      color={type === 'dependency' ? 'warning' : 'muted'}
      style={{ fontSize: '12px' }}
    >
      {type === 'dependency' ? 'prod' : 'dev'}
    </Badge>
  );
}

export function VersionBadge({ version, isOutdated }: { version: string; isOutdated?: boolean }) {
  // If isOutdated is undefined (Latest column), use neutral styling
  const color = isOutdated === undefined ? 'muted' : isOutdated ? 'danger' : 'success';

  return (
    <Badge
      color={color}
      style={{
        fontFamily: "'JetBrains Mono', 'Fira Code', monospace",
        fontSize: '13px',
        fontWeight: 600,
        letterSpacing: '0.3px',
        padding: '5px 10px',
      }}
    >
      {version || '-'}
    </Badge>
  );
}
