import type { Source, SourceInput, Repository, Dependency, ScanJob, DependencyStats, PaginatedDependencies, FilterOptions, Settings, SettingsInput, NextScan, IgnoredDependency, IgnoredDependencyInput } from '../types';

const API_BASE = '/api/v1';

// Custom error class with HTTP status information
export class ApiError extends Error {
  status: number;
  statusText: string;

  constructor(message: string, status: number, statusText: string) {
    super(message);
    this.name = 'ApiError';
    this.status = status;
    this.statusText = statusText;
  }

  get isUnauthorized(): boolean {
    return this.status === 401;
  }

  get isForbidden(): boolean {
    return this.status === 403;
  }

  get isNotFound(): boolean {
    return this.status === 404;
  }

  get isRateLimited(): boolean {
    return this.status === 429;
  }

  get isServerError(): boolean {
    return this.status >= 500;
  }
}

// Get user-friendly error message based on status
function getErrorMessage(status: number, statusText: string, responseText: string): string {
  // Try to parse JSON error response
  if (responseText) {
    try {
      const json = JSON.parse(responseText);
      if (json.error || json.message) {
        return json.error || json.message;
      }
    } catch {
      // Not JSON, use as-is if it's not HTML
      if (!responseText.startsWith('<')) {
        return responseText;
      }
    }
  }

  // Fallback to status-based messages
  switch (status) {
    case 400:
      return 'Invalid request. Please check your input.';
    case 401:
      return 'Authentication required. Please check your API key.';
    case 403:
      return 'Access denied. You do not have permission for this action.';
    case 404:
      return 'Resource not found.';
    case 429:
      return 'Too many requests. Please wait a moment and try again.';
    case 500:
      return 'Server error. Please try again later.';
    case 502:
    case 503:
    case 504:
      return 'Service temporarily unavailable. Please try again later.';
    default:
      return statusText || `Request failed (${status})`;
  }
}

async function request<T>(endpoint: string, options?: RequestInit): Promise<T> {
  let response: Response;

  try {
    response = await fetch(`${API_BASE}${endpoint}`, {
      ...options,
      headers: {
        'Content-Type': 'application/json',
        ...options?.headers,
      },
    });
  } catch (err) {
    // Network error (offline, DNS failure, etc.)
    throw new ApiError(
      'Network error. Please check your connection.',
      0,
      'Network Error'
    );
  }

  if (!response.ok) {
    const text = await response.text();
    const message = getErrorMessage(response.status, response.statusText, text);
    throw new ApiError(message, response.status, response.statusText);
  }

  if (response.status === 204) {
    return undefined as T;
  }

  return response.json();
}

export const api = {
  // Health
  health: () => request<{ status: string }>('/health'),

  // Sources
  getSources: () => request<Source[]>('/sources'),
  getSource: (id: number) => request<Source>(`/sources/${id}`),
  createSource: (data: SourceInput) =>
    request<Source>('/sources', { method: 'POST', body: JSON.stringify(data) }),
  updateSource: (id: number, data: SourceInput) =>
    request<Source>(`/sources/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
  deleteSource: (id: number) =>
    request<void>(`/sources/${id}`, { method: 'DELETE' }),

  // Repositories
  getRepositories: (sourceId?: number) => {
    const params = sourceId ? `?source_id=${sourceId}` : '';
    return request<Repository[]>(`/repositories${params}`);
  },
  getRepository: (id: number) => request<Repository>(`/repositories/${id}`),
  getRepositoryDependencies: (id: number) =>
    request<Dependency[]>(`/repositories/${id}/dependencies`),
  deleteRepository: (id: number) =>
    request<void>(`/repositories/${id}`, { method: 'DELETE' }),
  bulkDeleteRepositories: (ids: number[]) =>
    request<{ deleted: number }>('/repositories/bulk-delete', {
      method: 'POST',
      body: JSON.stringify({ ids }),
    }),

  // Dependencies
  getDependencies: (upgradableOnly?: boolean) => {
    const params = upgradableOnly ? '?outdated=true' : '';
    return request<Dependency[]>(`/dependencies${params}`);
  },
  getRepositoryNames: () => request<string[]>('/dependencies/repos'),
  getPackageNames: () => request<string[]>('/dependencies/packages'),
  getFilterOptions: (repo?: string, ecosystem?: string, status?: string, pkg?: string) => {
    const params = new URLSearchParams();
    if (repo) params.set('repo', repo);
    if (ecosystem) params.set('ecosystem', ecosystem);
    if (status && status !== 'all') params.set('status', status);
    if (pkg) params.set('package', pkg);
    const query = params.toString();
    return request<FilterOptions>(`/dependencies/filter-options${query ? `?${query}` : ''}`);
  },
  getDependenciesPaginated: (page: number = 1, limit: number = 50, status?: string, repo?: string, ecosystem?: string, search?: string) => {
    const params = new URLSearchParams();
    params.set('page', String(page));
    params.set('limit', String(limit));
    if (status && status !== 'all') params.set('status', status);
    if (repo) params.set('repo', repo);
    if (ecosystem) params.set('ecosystem', ecosystem);
    if (search) params.set('search', search);
    return request<PaginatedDependencies>(`/dependencies/paginated?${params.toString()}`);
  },
  getUpgradableDependencies: () => request<Dependency[]>('/dependencies/upgradable'),
  getDependencyStats: () => request<DependencyStats>('/dependencies/stats'),

  // Scans
  triggerScan: (sourceId?: number) =>
    request<ScanJob>('/scans', {
      method: 'POST',
      body: JSON.stringify(sourceId ? { source_id: sourceId } : {}),
    }),
  getScans: () => request<ScanJob[]>('/scans'),
  getScan: (id: number) => request<ScanJob>(`/scans/${id}`),
  getRunningScan: async (): Promise<ScanJob | null> => {
    const response = await fetch(`${API_BASE}/scans/running`, {
      headers: { 'Content-Type': 'application/json' },
    });
    if (!response.ok) return null;
    const text = await response.text();
    if (text === 'null' || !text) return null;
    return JSON.parse(text);
  },
  cancelScan: (id: number) =>
    request<void>(`/scans/${id}/cancel`, { method: 'POST' }),

  // Settings
  getSettings: () => request<Settings>('/settings'),
  updateSettings: (data: SettingsInput) =>
    request<Settings>('/settings', { method: 'PUT', body: JSON.stringify(data) }),
  testEmail: () => request<{ status: string; message: string }>('/settings/test-email', { method: 'POST' }),
  getNextScan: () => request<NextScan>('/settings/next-scan'),

  // Ignored Dependencies
  getIgnored: () => request<IgnoredDependency[]>('/ignored'),
  addIgnored: (data: IgnoredDependencyInput) =>
    request<IgnoredDependency>('/ignored', { method: 'POST', body: JSON.stringify(data) }),
  bulkAddIgnored: (items: IgnoredDependencyInput[]) =>
    request<{ created: number; items: IgnoredDependency[] }>('/ignored/bulk', { method: 'POST', body: JSON.stringify({ items }) }),
  removeIgnored: (id: number) =>
    request<void>(`/ignored/${id}`, { method: 'DELETE' }),
  bulkRemoveIgnored: (ids: number[]) =>
    request<{ deleted: number }>('/ignored/bulk-delete', { method: 'POST', body: JSON.stringify({ ids }) }),
};
