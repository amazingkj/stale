import type { ButtonHTMLAttributes, ReactNode, CSSProperties } from 'react';

interface Props extends ButtonHTMLAttributes<HTMLButtonElement> {
  children: ReactNode;
  variant?: 'primary' | 'secondary' | 'danger' | 'ghost';
  size?: 'sm' | 'md' | 'lg';
  loading?: boolean;
}

const baseStyle: CSSProperties = {
  display: 'inline-flex',
  alignItems: 'center',
  justifyContent: 'center',
  fontWeight: 600,
  borderRadius: 'var(--radius-full)',
  transition: 'all 0.25s cubic-bezier(0.16, 1, 0.3, 1)',
  cursor: 'pointer',
  border: 'none',
  outline: 'none',
  gap: '8px',
  letterSpacing: '-0.01em',
};

const variantStyles: Record<NonNullable<Props['variant']>, CSSProperties> = {
  primary: {
    background: 'var(--accent-gradient)',
    color: 'white',
    boxShadow: '0 2px 8px -2px rgba(124, 181, 149, 0.4)',
  },
  secondary: {
    backgroundColor: 'var(--bg-card)',
    color: 'var(--text-primary)',
    border: '1.5px solid var(--border-color)',
    boxShadow: 'var(--shadow-sm)',
  },
  danger: {
    backgroundColor: 'var(--danger)',
    color: 'white',
    boxShadow: '0 2px 8px -2px rgba(212, 132, 122, 0.4)',
  },
  ghost: {
    backgroundColor: 'transparent',
    color: 'var(--text-secondary)',
  },
};

const sizeStyles: Record<NonNullable<Props['size']>, CSSProperties> = {
  sm: { padding: '8px 16px', fontSize: '13px' },
  md: { padding: '12px 22px', fontSize: '14px' },
  lg: { padding: '14px 28px', fontSize: '15px' },
};

export function Button({
  children,
  variant = 'primary',
  size = 'md',
  loading = false,
  disabled,
  style,
  onMouseEnter,
  onMouseLeave,
  ...props
}: Props) {
  const isDisabled = disabled || loading;

  const handleMouseEnter = (e: React.MouseEvent<HTMLButtonElement>) => {
    if (!isDisabled) {
      if (variant === 'primary') {
        e.currentTarget.style.transform = 'translateY(-2px) scale(1.02)';
        e.currentTarget.style.boxShadow = '0 8px 20px -4px rgba(124, 181, 149, 0.5)';
      } else if (variant === 'secondary') {
        e.currentTarget.style.backgroundColor = 'var(--bg-hover)';
        e.currentTarget.style.borderColor = 'var(--accent)';
        e.currentTarget.style.transform = 'translateY(-1px)';
      } else if (variant === 'danger') {
        e.currentTarget.style.transform = 'translateY(-2px) scale(1.02)';
        e.currentTarget.style.boxShadow = '0 8px 20px -4px rgba(212, 132, 122, 0.5)';
      } else if (variant === 'ghost') {
        e.currentTarget.style.backgroundColor = 'var(--bg-hover)';
        e.currentTarget.style.color = 'var(--text-primary)';
      }
    }
    onMouseEnter?.(e);
  };

  const handleMouseLeave = (e: React.MouseEvent<HTMLButtonElement>) => {
    if (!isDisabled) {
      if (variant === 'primary') {
        e.currentTarget.style.transform = 'translateY(0) scale(1)';
        e.currentTarget.style.boxShadow = '0 2px 8px -2px rgba(124, 181, 149, 0.4)';
      } else if (variant === 'secondary') {
        e.currentTarget.style.backgroundColor = 'var(--bg-card)';
        e.currentTarget.style.borderColor = 'var(--border-color)';
        e.currentTarget.style.transform = 'translateY(0)';
      } else if (variant === 'danger') {
        e.currentTarget.style.transform = 'translateY(0) scale(1)';
        e.currentTarget.style.boxShadow = '0 2px 8px -2px rgba(212, 132, 122, 0.4)';
      } else if (variant === 'ghost') {
        e.currentTarget.style.backgroundColor = 'transparent';
        e.currentTarget.style.color = 'var(--text-secondary)';
      }
    }
    onMouseLeave?.(e);
  };

  return (
    <button
      style={{
        ...baseStyle,
        ...variantStyles[variant],
        ...sizeStyles[size],
        opacity: isDisabled ? 0.6 : 1,
        cursor: isDisabled ? 'not-allowed' : 'pointer',
        ...style,
      }}
      disabled={isDisabled}
      aria-busy={loading}
      aria-disabled={isDisabled}
      onMouseEnter={handleMouseEnter}
      onMouseLeave={handleMouseLeave}
      {...props}
    >
      {loading && <Spinner />}
      {children}
    </button>
  );
}

function Spinner() {
  return (
    <svg
      width="16"
      height="16"
      viewBox="0 0 24 24"
      style={{ animation: 'spin 1s linear infinite' }}
      aria-hidden="true"
    >
      <circle
        cx="12"
        cy="12"
        r="10"
        stroke="currentColor"
        strokeWidth="3"
        fill="none"
        opacity="0.25"
      />
      <path
        d="M12 2a10 10 0 0 1 10 10"
        stroke="currentColor"
        strokeWidth="3"
        fill="none"
        strokeLinecap="round"
      />
    </svg>
  );
}
