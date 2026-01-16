import { useState, useMemo, useCallback } from 'react';

interface UseFilteringOptions<T> {
  /** Initial data to filter */
  data: T[];
  /** Function to get searchable text from an item */
  getSearchText?: (item: T) => string;
}

interface UseFilteringResult<T, F extends string> {
  /** Current search query */
  search: string;
  /** Set search query */
  setSearch: (value: string) => void;
  /** Current filter value */
  filter: F;
  /** Set filter value */
  setFilter: (value: F) => void;
  /** Filtered data */
  filteredData: T[];
  /** Register a filter function */
  registerFilter: (name: F, filterFn: (item: T) => boolean) => void;
}

/**
 * Custom hook for filtering data with search and category filters
 */
export function useFiltering<T, F extends string = string>(
  options: UseFilteringOptions<T>,
  defaultFilter: F
): UseFilteringResult<T, F> {
  const { data, getSearchText } = options;

  const [search, setSearch] = useState('');
  const [filter, setFilter] = useState<F>(defaultFilter);
  const [filters, setFilters] = useState<Map<F, (item: T) => boolean>>(new Map());

  const registerFilter = useCallback((name: F, filterFn: (item: T) => boolean) => {
    setFilters((prev) => {
      const next = new Map(prev);
      next.set(name, filterFn);
      return next;
    });
  }, []);

  const filteredData = useMemo(() => {
    let result = data;

    // Apply category filter
    const filterFn = filters.get(filter);
    if (filterFn) {
      result = result.filter(filterFn);
    }

    // Apply search filter
    if (search && getSearchText) {
      const searchLower = search.toLowerCase();
      result = result.filter((item) =>
        getSearchText(item).toLowerCase().includes(searchLower)
      );
    }

    return result;
  }, [data, filter, filters, search, getSearchText]);

  return {
    search,
    setSearch,
    filter,
    setFilter,
    filteredData,
    registerFilter,
  };
}
