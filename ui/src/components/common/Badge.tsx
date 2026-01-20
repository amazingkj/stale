import type { CSSProperties, ReactNode } from 'react';
import { badgeColors, type BadgeColor } from '../../constants/styles';

export type VersionDiffType = 'major' | 'minor' | 'patch' | 'unknown';

interface Props {
  children: ReactNode;
  color?: BadgeColor;
  style?: CSSProperties;
}

const baseStyle: CSSProperties = {
  padding: '3px 6px',
  borderRadius: '4px',
  fontSize: '11px',
  fontWeight: 500,
  display: 'inline-flex',
  alignItems: 'center',
  gap: '3px',
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
    maven: { color: 'purple', label: 'maven' },
    gradle: { color: 'success', label: 'gradle' },
  };

  const { color, label } = config[ecosystem] || { color: 'muted', label: ecosystem };

  return (
    <Badge color={color} style={{ fontSize: '10px', fontWeight: 600, padding: '3px 6px' }}>
      {label}
    </Badge>
  );
}

export function TypeBadge({ type }: { type: 'dependency' | 'devDependency' }) {
  return (
    <Badge
      color={type === 'dependency' ? 'warning' : 'muted'}
      style={{ fontSize: '10px' }}
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
        fontSize: '11px',
        fontWeight: 600,
        letterSpacing: '0.3px',
        padding: '4px 8px',
      }}
    >
      {version || '-'}
    </Badge>
  );
}

export function VersionDiffBadge({ diffType }: { diffType: VersionDiffType }) {
  if (diffType === 'unknown') return null;

  const config: Record<Exclude<VersionDiffType, 'unknown'>, { color: BadgeColor; label: string }> = {
    major: { color: 'danger', label: 'Major' },
    minor: { color: 'warning', label: 'Minor' },
    patch: { color: 'success', label: 'Patch' },
  };

  const { color, label } = config[diffType];

  return (
    <Badge
      color={color}
      style={{
        fontSize: '9px',
        fontWeight: 600,
        padding: '2px 5px',
        marginLeft: '4px',
      }}
    >
      {label}
    </Badge>
  );
}
