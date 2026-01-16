import type { ReactNode, CSSProperties } from 'react';
import { tableHeaderStyle, tableCellStyle } from '../../constants/styles';

interface TableProps {
  children: ReactNode;
  style?: CSSProperties;
  fixed?: boolean;
}

export function Table({ children, style, fixed }: TableProps) {
  return (
    <div style={{ overflowX: 'auto' }}>
      <table style={{
        width: '100%',
        borderCollapse: 'collapse',
        tableLayout: fixed ? 'fixed' : 'auto',
        ...style,
      }}>
        {children}
      </table>
    </div>
  );
}

export function TableHead({ children }: { children: ReactNode }) {
  return (
    <thead>
      <tr style={{ backgroundColor: 'var(--bg-secondary)' }}>{children}</tr>
    </thead>
  );
}

export function TableBody({ children }: { children: ReactNode }) {
  return <tbody>{children}</tbody>;
}

export function TableRow({ children, onClick }: { children: ReactNode; onClick?: () => void }) {
  return (
    <tr
      style={{
        borderBottom: '1px solid var(--border-color)',
        cursor: onClick ? 'pointer' : 'default',
      }}
      onClick={onClick}
    >
      {children}
    </tr>
  );
}

interface TableHeaderCellProps {
  children?: ReactNode;
  style?: CSSProperties;
  width?: string;
}

export function Th({ children, style, width }: TableHeaderCellProps) {
  return (
    <th style={{
      ...tableHeaderStyle,
      width,
      overflow: 'hidden',
      textOverflow: 'ellipsis',
      whiteSpace: 'nowrap',
      ...style,
    }}>
      {children}
    </th>
  );
}

interface TableCellProps {
  children: ReactNode;
  style?: CSSProperties;
  muted?: boolean;
  secondary?: boolean;
}

export function Td({ children, style, muted, secondary }: TableCellProps) {
  return (
    <td
      style={{
        ...tableCellStyle,
        color: muted
          ? 'var(--text-muted)'
          : secondary
          ? 'var(--text-secondary)'
          : 'var(--text-primary)',
        overflow: 'hidden',
        textOverflow: 'ellipsis',
        whiteSpace: 'nowrap',
        maxWidth: 0,
        ...style,
      }}
      title={typeof children === 'string' ? children : undefined}
    >
      {children}
    </td>
  );
}
