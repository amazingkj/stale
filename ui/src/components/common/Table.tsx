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

export function TableRow({ children, onClick, style }: { children: ReactNode; onClick?: () => void; style?: CSSProperties }) {
  return (
    <tr
      style={{
        borderBottom: '1px solid var(--border-color)',
        cursor: onClick ? 'pointer' : 'default',
        ...style,
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
  noEllipsis?: boolean;
  colSpan?: number;
}

export function Th({ children, style, width, noEllipsis, colSpan }: TableHeaderCellProps) {
  return (
    <th
      colSpan={colSpan}
      scope="col"
      style={{
        ...tableHeaderStyle,
        width,
        ...(noEllipsis ? {} : {
          overflow: 'hidden',
          textOverflow: 'ellipsis',
          whiteSpace: 'nowrap',
        }),
        ...style,
      }}
    >
      {children}
    </th>
  );
}

interface TableCellProps {
  children: ReactNode;
  style?: CSSProperties;
  muted?: boolean;
  secondary?: boolean;
  noEllipsis?: boolean;
}

export function Td({ children, style, muted, secondary, noEllipsis }: TableCellProps) {
  return (
    <td
      style={{
        ...tableCellStyle,
        color: muted
          ? 'var(--text-muted)'
          : secondary
          ? 'var(--text-secondary)'
          : 'var(--text-primary)',
        ...(noEllipsis ? {} : {
          overflow: 'hidden',
          textOverflow: 'ellipsis',
          whiteSpace: 'nowrap',
          maxWidth: 0,
        }),
        ...style,
      }}
      title={typeof children === 'string' ? children : undefined}
    >
      {children}
    </td>
  );
}
