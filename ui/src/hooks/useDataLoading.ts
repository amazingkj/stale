import { useState, useEffect, useCallback } from 'react';

interface UseDataLoadingOptions {
  /** Whether to load data immediately on mount */
  loadOnMount?: boolean;
}

interface UseDataLoadingResult<T> {
  data: T | null;
  loading: boolean;
  error: string | null;
  reload: () => Promise<void>;
  setData: React.Dispatch<React.SetStateAction<T | null>>;
  clearError: () => void;
}

/**
 * Custom hook for loading data with loading and error states
 */
export function useDataLoading<T>(
  fetchFn: () => Promise<T>,
  options: UseDataLoadingOptions = {}
): UseDataLoadingResult<T> {
  const { loadOnMount = true } = options;

  const [data, setData] = useState<T | null>(null);
  const [loading, setLoading] = useState(loadOnMount);
  const [error, setError] = useState<string | null>(null);

  const reload = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const result = await fetchFn();
      setData(result);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'An error occurred');
    } finally {
      setLoading(false);
    }
  }, [fetchFn]);

  const clearError = useCallback(() => {
    setError(null);
  }, []);

  useEffect(() => {
    if (loadOnMount) {
      reload();
    }
  }, [loadOnMount, reload]);

  return { data, loading, error, reload, setData, clearError };
}
