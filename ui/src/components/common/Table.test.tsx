import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
import { Table, TableHead, TableBody, TableRow, Th, Td } from './Table';

describe('Table', () => {
  it('renders children correctly', () => {
    render(
      <Table>
        <tbody>
          <tr>
            <td>Cell content</td>
          </tr>
        </tbody>
      </Table>
    );
    expect(screen.getByText('Cell content')).toBeInTheDocument();
  });

  it('applies custom style', () => {
    render(
      <Table style={{ margin: '10px' }}>
        <tbody>
          <tr>
            <td>Content</td>
          </tr>
        </tbody>
      </Table>
    );
    expect(screen.getByRole('table')).toHaveStyle({ margin: '10px' });
  });

  it('applies fixed table layout when fixed prop is true', () => {
    render(
      <Table fixed>
        <tbody>
          <tr>
            <td>Content</td>
          </tr>
        </tbody>
      </Table>
    );
    expect(screen.getByRole('table')).toHaveStyle({ tableLayout: 'fixed' });
  });

  it('applies minHeight to wrapper', () => {
    render(
      <Table minHeight={300}>
        <tbody>
          <tr>
            <td>Content</td>
          </tr>
        </tbody>
      </Table>
    );
    const wrapper = screen.getByRole('table').parentElement;
    expect(wrapper).toHaveStyle({ minHeight: '300px' });
  });
});

describe('TableHead', () => {
  it('renders header row', () => {
    render(
      <table>
        <TableHead>
          <th>Header</th>
        </TableHead>
      </table>
    );
    expect(screen.getByText('Header')).toBeInTheDocument();
    expect(screen.getByRole('rowgroup')).toBeInTheDocument();
  });
});

describe('TableBody', () => {
  it('renders body rows', () => {
    render(
      <table>
        <TableBody>
          <tr>
            <td>Row content</td>
          </tr>
        </TableBody>
      </table>
    );
    expect(screen.getByText('Row content')).toBeInTheDocument();
  });
});

describe('TableRow', () => {
  it('renders children correctly', () => {
    render(
      <table>
        <tbody>
          <TableRow>
            <td>Cell</td>
          </TableRow>
        </tbody>
      </table>
    );
    expect(screen.getByText('Cell')).toBeInTheDocument();
  });

  it('handles click events', () => {
    const handleClick = vi.fn();
    render(
      <table>
        <tbody>
          <TableRow onClick={handleClick}>
            <td>Clickable row</td>
          </TableRow>
        </tbody>
      </table>
    );

    fireEvent.click(screen.getByRole('row'));
    expect(handleClick).toHaveBeenCalledTimes(1);
  });

  it('shows pointer cursor when clickable', () => {
    const handleClick = vi.fn();
    render(
      <table>
        <tbody>
          <TableRow onClick={handleClick}>
            <td>Clickable</td>
          </TableRow>
        </tbody>
      </table>
    );
    expect(screen.getByRole('row')).toHaveStyle({ cursor: 'pointer' });
  });

  it('shows default cursor when not clickable', () => {
    render(
      <table>
        <tbody>
          <TableRow>
            <td>Non-clickable</td>
          </TableRow>
        </tbody>
      </table>
    );
    expect(screen.getByRole('row')).toHaveStyle({ cursor: 'default' });
  });

  it('applies custom style', () => {
    render(
      <table>
        <tbody>
          <TableRow style={{ backgroundColor: 'red' }}>
            <td>Styled row</td>
          </TableRow>
        </tbody>
      </table>
    );
    expect(screen.getByRole('row').style.backgroundColor).toBe('red');
  });
});

describe('Th', () => {
  it('renders header cell content', () => {
    render(
      <table>
        <thead>
          <tr>
            <Th>Header</Th>
          </tr>
        </thead>
      </table>
    );
    expect(screen.getByRole('columnheader')).toHaveTextContent('Header');
  });

  it('applies width when provided', () => {
    render(
      <table>
        <thead>
          <tr>
            <Th width="200px">Header</Th>
          </tr>
        </thead>
      </table>
    );
    expect(screen.getByRole('columnheader')).toHaveStyle({ width: '200px' });
  });

  it('applies colSpan when provided', () => {
    render(
      <table>
        <thead>
          <tr>
            <Th colSpan={2}>Spanning header</Th>
          </tr>
        </thead>
      </table>
    );
    expect(screen.getByRole('columnheader')).toHaveAttribute('colspan', '2');
  });

  it('has scope="col" for accessibility', () => {
    render(
      <table>
        <thead>
          <tr>
            <Th>Header</Th>
          </tr>
        </thead>
      </table>
    );
    expect(screen.getByRole('columnheader')).toHaveAttribute('scope', 'col');
  });

  it('applies ellipsis by default', () => {
    render(
      <table>
        <thead>
          <tr>
            <Th>Header</Th>
          </tr>
        </thead>
      </table>
    );
    const th = screen.getByRole('columnheader');
    expect(th).toHaveStyle({ overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' });
  });

  it('disables ellipsis when noEllipsis is true', () => {
    render(
      <table>
        <thead>
          <tr>
            <Th noEllipsis>Header</Th>
          </tr>
        </thead>
      </table>
    );
    const th = screen.getByRole('columnheader');
    expect(th.style.overflow).toBeFalsy();
  });
});

describe('Td', () => {
  it('renders cell content', () => {
    render(
      <table>
        <tbody>
          <tr>
            <Td>Cell data</Td>
          </tr>
        </tbody>
      </table>
    );
    expect(screen.getByRole('cell')).toHaveTextContent('Cell data');
  });

  it('applies muted style', () => {
    render(
      <table>
        <tbody>
          <tr>
            <Td muted>Muted text</Td>
          </tr>
        </tbody>
      </table>
    );
    expect(screen.getByRole('cell')).toHaveStyle({ color: 'var(--text-muted)' });
  });

  it('applies secondary style', () => {
    render(
      <table>
        <tbody>
          <tr>
            <Td secondary>Secondary text</Td>
          </tr>
        </tbody>
      </table>
    );
    expect(screen.getByRole('cell')).toHaveStyle({ color: 'var(--text-secondary)' });
  });

  it('applies primary color by default', () => {
    render(
      <table>
        <tbody>
          <tr>
            <Td>Primary text</Td>
          </tr>
        </tbody>
      </table>
    );
    expect(screen.getByRole('cell')).toHaveStyle({ color: 'var(--text-primary)' });
  });

  it('adds title attribute for string children', () => {
    render(
      <table>
        <tbody>
          <tr>
            <Td>Cell content</Td>
          </tr>
        </tbody>
      </table>
    );
    expect(screen.getByRole('cell')).toHaveAttribute('title', 'Cell content');
  });

  it('does not add title for non-string children', () => {
    render(
      <table>
        <tbody>
          <tr>
            <Td><span>Complex content</span></Td>
          </tr>
        </tbody>
      </table>
    );
    expect(screen.getByRole('cell')).not.toHaveAttribute('title');
  });

  it('applies custom style', () => {
    render(
      <table>
        <tbody>
          <tr>
            <Td style={{ fontWeight: 'bold' }}>Bold text</Td>
          </tr>
        </tbody>
      </table>
    );
    expect(screen.getByRole('cell')).toHaveStyle({ fontWeight: 'bold' });
  });
});

describe('Table composition', () => {
  it('renders complete table structure', () => {
    render(
      <Table>
        <TableHead>
          <Th>Name</Th>
          <Th>Value</Th>
        </TableHead>
        <TableBody>
          <TableRow>
            <Td>Item 1</Td>
            <Td>100</Td>
          </TableRow>
          <TableRow>
            <Td>Item 2</Td>
            <Td>200</Td>
          </TableRow>
        </TableBody>
      </Table>
    );

    expect(screen.getAllByRole('columnheader')).toHaveLength(2);
    expect(screen.getAllByRole('row')).toHaveLength(3); // 1 header + 2 body rows
    expect(screen.getAllByRole('cell')).toHaveLength(4);
  });
});
