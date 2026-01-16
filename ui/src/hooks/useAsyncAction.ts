import { useState, useCallback } from 'react';

interface UseAsyncActionResult<T, Args extends unknown[]> {
  execute: (...args: Args) => Promise<T | undefined>;
  loading: boolean;
  error: string | null;
  clearError: () => void;
}

/**
 * Custom hook for handling async actions with loading and error states
 */
export function useAsyncAction<T, Args extends unknown[]>(
  asyncFn: (...args: Args) => Promise<T>,
  options?: {
    onSuccess?: (result: T) => void;
    onError?: (error: Error) => void;
  }
): UseAsyncActionResult<T, Args> {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const execute = useCallback(
    async (...args: Args): Promise<T | undefined> => {
      setLoading(true);
      setError(null);
      try {
        const result = await asyncFn(...args);
        options?.onSuccess?.(result);
        return result;
      } catch (err) {
        const error = err instanceof Error ? err : new Error('An error occurred');
        setError(error.message);
        options?.onError?.(error);
        return undefined;
      } finally {
        setLoading(false);
      }
    },
    [asyncFn, options]
  );

  const clearError = useCallback(() => {
    setError(null);
  }, []);

  return { execute, loading, error, clearError };
}
