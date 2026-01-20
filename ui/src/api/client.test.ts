import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { api } from './client';

describe('api', () => {
  const mockFetch = vi.fn();
  const originalFetch = global.fetch;

  beforeEach(() => {
    global.fetch = mockFetch;
    mockFetch.mockClear();
  });

  afterEach(() => {
    global.fetch = originalFetch;
  });

  describe('health', () => {
    it('calls the health endpoint', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve({ status: 'ok' }),
      });

      const result = await api.health();

      expect(mockFetch).toHaveBeenCalledWith('/api/v1/health', expect.objectContaining({
        headers: { 'Content-Type': 'application/json' },
      }));
      expect(result).toEqual({ status: 'ok' });
    });
  });

  describe('getSources', () => {
    it('returns sources array', async () => {
      const mockSources = [{ id: 1, name: 'GitHub' }];
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve(mockSources),
      });

      const result = await api.getSources();

      expect(mockFetch).toHaveBeenCalledWith('/api/v1/sources', expect.any(Object));
      expect(result).toEqual(mockSources);
    });
  });

  describe('getSource', () => {
    it('fetches a single source by id', async () => {
      const mockSource = { id: 1, name: 'GitHub' };
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve(mockSource),
      });

      const result = await api.getSource(1);

      expect(mockFetch).toHaveBeenCalledWith('/api/v1/sources/1', expect.any(Object));
      expect(result).toEqual(mockSource);
    });
  });

  describe('createSource', () => {
    it('creates a new source', async () => {
      const newSource = { name: 'GitLab', type: 'gitlab', token: 'token123' };
      const createdSource = { id: 1, ...newSource };
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve(createdSource),
      });

      const result = await api.createSource(newSource as any);

      expect(mockFetch).toHaveBeenCalledWith('/api/v1/sources', expect.objectContaining({
        method: 'POST',
        body: JSON.stringify(newSource),
      }));
      expect(result).toEqual(createdSource);
    });
  });

  describe('deleteSource', () => {
    it('deletes a source', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        status: 204,
        json: () => Promise.resolve(undefined),
      });

      await api.deleteSource(1);

      expect(mockFetch).toHaveBeenCalledWith('/api/v1/sources/1', expect.objectContaining({
        method: 'DELETE',
      }));
    });
  });

  describe('getRepositories', () => {
    it('fetches all repositories', async () => {
      const mockRepos = [{ id: 1, name: 'repo1' }];
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve(mockRepos),
      });

      const result = await api.getRepositories();

      expect(mockFetch).toHaveBeenCalledWith('/api/v1/repositories', expect.any(Object));
      expect(result).toEqual(mockRepos);
    });

    it('filters by source_id', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve([]),
      });

      await api.getRepositories(5);

      expect(mockFetch).toHaveBeenCalledWith('/api/v1/repositories?source_id=5', expect.any(Object));
    });
  });

  describe('getDependenciesPaginated', () => {
    it('includes all parameters in query string', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve({ data: [], total: 0 }),
      });

      await api.getDependenciesPaginated(2, 25, 'upgradable', 'owner/repo', 'npm', 'react');

      const calledUrl = mockFetch.mock.calls[0][0];
      expect(calledUrl).toContain('page=2');
      expect(calledUrl).toContain('limit=25');
      expect(calledUrl).toContain('status=upgradable');
      expect(calledUrl).toContain('repo=owner%2Frepo');
      expect(calledUrl).toContain('ecosystem=npm');
      expect(calledUrl).toContain('search=react');
    });

    it('skips status=all', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve({ data: [], total: 0 }),
      });

      await api.getDependenciesPaginated(1, 50, 'all');

      const calledUrl = mockFetch.mock.calls[0][0];
      expect(calledUrl).not.toContain('status=');
    });
  });

  describe('triggerScan', () => {
    it('triggers scan without source_id', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve({ id: 1, status: 'running' }),
      });

      await api.triggerScan();

      expect(mockFetch).toHaveBeenCalledWith('/api/v1/scans', expect.objectContaining({
        method: 'POST',
        body: JSON.stringify({}),
      }));
    });

    it('triggers scan with source_id', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve({ id: 1, status: 'running' }),
      });

      await api.triggerScan(5);

      expect(mockFetch).toHaveBeenCalledWith('/api/v1/scans', expect.objectContaining({
        method: 'POST',
        body: JSON.stringify({ source_id: 5 }),
      }));
    });
  });

  describe('getRunningScan', () => {
    it('returns null when no running scan', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        text: () => Promise.resolve('null'),
      });

      const result = await api.getRunningScan();

      expect(result).toBeNull();
    });

    it('returns scan job when running', async () => {
      const runningScan = { id: 1, status: 'running' };
      mockFetch.mockResolvedValueOnce({
        ok: true,
        text: () => Promise.resolve(JSON.stringify(runningScan)),
      });

      const result = await api.getRunningScan();

      expect(result).toEqual(runningScan);
    });

    it('returns null on error', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: false,
      });

      const result = await api.getRunningScan();

      expect(result).toBeNull();
    });
  });

  describe('error handling', () => {
    it('throws error on non-ok response', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: false,
        statusText: 'Not Found',
        text: () => Promise.resolve('Resource not found'),
      });

      await expect(api.getSources()).rejects.toThrow('Resource not found');
    });

    it('throws statusText when no error text', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: false,
        statusText: 'Internal Server Error',
        text: () => Promise.resolve(''),
      });

      await expect(api.getSources()).rejects.toThrow('API Error: Internal Server Error');
    });
  });

  describe('getFilterOptions', () => {
    it('builds query string correctly', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve({ repos: [], packages: [], ecosystems: [] }),
      });

      await api.getFilterOptions('owner/repo', 'npm', 'upgradable', 'react');

      const calledUrl = mockFetch.mock.calls[0][0];
      expect(calledUrl).toContain('repo=owner%2Frepo');
      expect(calledUrl).toContain('ecosystem=npm');
      expect(calledUrl).toContain('status=upgradable');
      expect(calledUrl).toContain('package=react');
    });
  });

  describe('bulkDeleteRepositories', () => {
    it('sends array of ids', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve({ deleted: 3 }),
      });

      const result = await api.bulkDeleteRepositories([1, 2, 3]);

      expect(mockFetch).toHaveBeenCalledWith('/api/v1/repositories/bulk-delete', expect.objectContaining({
        method: 'POST',
        body: JSON.stringify({ ids: [1, 2, 3] }),
      }));
      expect(result).toEqual({ deleted: 3 });
    });
  });

  describe('ignored dependencies', () => {
    it('gets ignored list', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve([{ id: 1, name: 'lodash' }]),
      });

      const result = await api.getIgnored();

      expect(mockFetch).toHaveBeenCalledWith('/api/v1/ignored', expect.any(Object));
      expect(result).toHaveLength(1);
    });

    it('adds ignored dependency', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve({ id: 1, name: 'lodash', reason: 'test' }),
      });

      await api.addIgnored({ name: 'lodash', reason: 'test' });

      expect(mockFetch).toHaveBeenCalledWith('/api/v1/ignored', expect.objectContaining({
        method: 'POST',
      }));
    });
  });
});
