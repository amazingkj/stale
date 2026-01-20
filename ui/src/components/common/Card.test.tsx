import { render, screen } from '@testing-library/react';
import { describe, it, expect } from 'vitest';
import { Card, CardHeader, CardBody } from './Card';

describe('Card', () => {
  it('renders children correctly', () => {
    render(<Card>Card content</Card>);
    expect(screen.getByText('Card content')).toBeInTheDocument();
  });

  it('applies custom style', () => {
    render(<Card style={{ width: '300px' }}><span>Content</span></Card>);
    const card = screen.getByText('Content').parentElement as HTMLElement;
    expect(card.style.width).toBe('300px');
  });

  it('applies minHeight when provided', () => {
    render(<Card minHeight={400}><span>Content</span></Card>);
    const card = screen.getByText('Content').parentElement as HTMLElement;
    expect(card.style.minHeight).toBe('400px');
  });

  it('applies transition when hoverable is true', () => {
    render(<Card hoverable><span>Hoverable card</span></Card>);
    const card = screen.getByText('Hoverable card').parentElement as HTMLElement;
    expect(card.style.transition).toBe('all 0.3s cubic-bezier(0.16, 1, 0.3, 1)');
  });

  it('does not apply transition when hoverable is false', () => {
    render(<Card hoverable={false}><span>Non-hoverable card</span></Card>);
    const card = screen.getByText('Non-hoverable card').parentElement as HTMLElement;
    expect(card.style.transition).toBeFalsy();
  });
});

describe('CardHeader', () => {
  it('renders children correctly', () => {
    render(<CardHeader>Header content</CardHeader>);
    expect(screen.getByText('Header content')).toBeInTheDocument();
  });

  it('applies custom style', () => {
    render(<CardHeader style={{ fontWeight: 'bold' }}>Header</CardHeader>);
    expect(screen.getByText('Header')).toHaveStyle({ fontWeight: 'bold' });
  });

  it('has bottom border', () => {
    render(<CardHeader>Header</CardHeader>);
    const header = screen.getByText('Header');
    expect(header.style.borderBottom).toBe('1px solid var(--border-light)');
  });
});

describe('CardBody', () => {
  it('renders children correctly', () => {
    render(<CardBody>Body content</CardBody>);
    expect(screen.getByText('Body content')).toBeInTheDocument();
  });

  it('applies default padding', () => {
    render(<CardBody>Body</CardBody>);
    expect(screen.getByText('Body')).toHaveStyle({ padding: '24px' });
  });

  it('applies custom style', () => {
    render(<CardBody style={{ padding: '16px' }}>Custom body</CardBody>);
    expect(screen.getByText('Custom body')).toHaveStyle({ padding: '16px' });
  });
});

describe('Card composition', () => {
  it('renders complete card with header and body', () => {
    render(
      <Card>
        <CardHeader>Title</CardHeader>
        <CardBody>Content here</CardBody>
      </Card>
    );

    expect(screen.getByText('Title')).toBeInTheDocument();
    expect(screen.getByText('Content here')).toBeInTheDocument();
  });
});
