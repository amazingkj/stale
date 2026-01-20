import type { CSSProperties } from 'react';

interface PaginationProps {
  currentPage: number;
  totalPages: number;
  totalItems?: number;
  pageSize?: number;
  onPageChange: (page: number) => void;
  showPageNumbers?: boolean;
  maxPageButtons?: number;
}

export function Pagination({
  currentPage,
  totalPages,
  totalItems,
  pageSize,
  onPageChange,
  showPageNumbers = true,
  maxPageButtons = 5,
}: PaginationProps) {
  if (totalPages <= 1) return null;

  const getPageNumbers = (): (number | 'ellipsis')[] => {
    const pages: (number | 'ellipsis')[] = [];

    if (totalPages <= maxPageButtons + 2) {
      // Show all pages if total is small
      for (let i = 1; i <= totalPages; i++) {
        pages.push(i);
      }
    } else {
      // Always show first page
      pages.push(1);

      // Calculate range around current page
      let start = Math.max(2, currentPage - Math.floor(maxPageButtons / 2));
      let end = Math.min(totalPages - 1, start + maxPageButtons - 1);

      // Adjust start if end is at the limit
      if (end === totalPages - 1) {
        start = Math.max(2, end - maxPageButtons + 1);
      }

      // Add ellipsis before middle pages if needed
      if (start > 2) {
        pages.push('ellipsis');
      }

      // Add middle pages
      for (let i = start; i <= end; i++) {
        pages.push(i);
      }

      // Add ellipsis after middle pages if needed
      if (end < totalPages - 1) {
        pages.push('ellipsis');
      }

      // Always show last page
      pages.push(totalPages);
    }

    return pages;
  };

  const pageNumbers = getPageNumbers();

  // Calculate showing range
  const startItem = pageSize ? (currentPage - 1) * pageSize + 1 : null;
  const endItem = pageSize && totalItems ? Math.min(currentPage * pageSize, totalItems) : null;

  return (
    <div style={containerStyle}>
      {/* Items info */}
      {totalItems !== undefined && pageSize && (
        <span style={infoStyle}>
          {startItem}-{endItem} of {totalItems.toLocaleString()}
        </span>
      )}

      <div style={buttonsContainerStyle}>
        {/* First page button */}
        <button
          onClick={() => onPageChange(1)}
          disabled={currentPage === 1}
          style={currentPage === 1 ? { ...navButtonStyle, ...disabledStyle } : navButtonStyle}
          title="First page"
        >
          <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
            <polyline points="11 17 6 12 11 7" />
            <polyline points="18 17 13 12 18 7" />
          </svg>
        </button>

        {/* Previous button */}
        <button
          onClick={() => onPageChange(currentPage - 1)}
          disabled={currentPage === 1}
          style={currentPage === 1 ? { ...navButtonStyle, ...disabledStyle } : navButtonStyle}
          title="Previous page"
        >
          <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
            <polyline points="15 18 9 12 15 6" />
          </svg>
        </button>

        {/* Page numbers */}
        {showPageNumbers && pageNumbers.map((page, index) =>
          page === 'ellipsis' ? (
            <span key={`ellipsis-${index}`} style={ellipsisStyle}>...</span>
          ) : (
            <button
              key={page}
              onClick={() => onPageChange(page)}
              style={page === currentPage ? { ...pageButtonStyle, ...activePageStyle } : pageButtonStyle}
            >
              {page}
            </button>
          )
        )}

        {/* Next button */}
        <button
          onClick={() => onPageChange(currentPage + 1)}
          disabled={currentPage >= totalPages}
          style={currentPage >= totalPages ? { ...navButtonStyle, ...disabledStyle } : navButtonStyle}
          title="Next page"
        >
          <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
            <polyline points="9 18 15 12 9 6" />
          </svg>
        </button>

        {/* Last page button */}
        <button
          onClick={() => onPageChange(totalPages)}
          disabled={currentPage >= totalPages}
          style={currentPage >= totalPages ? { ...navButtonStyle, ...disabledStyle } : navButtonStyle}
          title="Last page"
        >
          <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
            <polyline points="13 17 18 12 13 7" />
            <polyline points="6 17 11 12 6 7" />
          </svg>
        </button>
      </div>
    </div>
  );
}

// Styles
const containerStyle: CSSProperties = {
  display: 'flex',
  justifyContent: 'center',
  alignItems: 'center',
  gap: '12px',
  flexWrap: 'wrap',
  padding: '6px 0',
};

const infoStyle: CSSProperties = {
  fontSize: '12px',
  color: 'var(--text-secondary)',
};

const buttonsContainerStyle: CSSProperties = {
  display: 'flex',
  alignItems: 'center',
  gap: '3px',
};

const baseButtonStyle: CSSProperties = {
  display: 'flex',
  alignItems: 'center',
  justifyContent: 'center',
  border: '1px solid var(--border-color)',
  backgroundColor: 'var(--bg-card)',
  color: 'var(--text-primary)',
  cursor: 'pointer',
  transition: 'all 0.15s ease',
  borderRadius: '4px',
};

const navButtonStyle: CSSProperties = {
  ...baseButtonStyle,
  width: '28px',
  height: '28px',
  padding: 0,
};

const pageButtonStyle: CSSProperties = {
  ...baseButtonStyle,
  minWidth: '28px',
  height: '28px',
  padding: '0 8px',
  fontSize: '12px',
  fontWeight: 500,
};

const activePageStyle: CSSProperties = {
  backgroundColor: 'var(--accent)',
  borderColor: 'var(--accent)',
  color: 'white',
};

const disabledStyle: CSSProperties = {
  opacity: 0.4,
  cursor: 'not-allowed',
  pointerEvents: 'none',
};

const ellipsisStyle: CSSProperties = {
  padding: '0 3px',
  color: 'var(--text-secondary)',
  fontSize: '12px',
};