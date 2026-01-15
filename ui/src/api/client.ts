import type { Source, SourceInput, Repository, Dependency, ScanJob, DependencyStats } from '../types';

const API_BASE = '/api/v1';

async function request<T>(endpoint: string, options?: RequestInit): Promise<T> {
  const response = await fetch(`${API_BASE}${endpoint}`, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...options?.headers,
    },
  });

  if (!response.ok) {
    const text = await response.text();
    throw new Error(text || `API Error: ${response.statusText}`);
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

  // Dependencies
  getDependencies: (outdated?: boolean) => {
    const params = outdated ? '?outdated=true' : '';
    return request<Dependency[]>(`/dependencies${params}`);
  },
  getOutdatedDependencies: () => request<Dependency[]>('/dependencies/outdated'),
  getDependencyStats: () => request<DependencyStats>('/dependencies/stats'),

  // Scans
  triggerScan: (sourceId?: number) =>
    request<ScanJob>('/scans', {
      method: 'POST',
      body: JSON.stringify(sourceId ? { source_id: sourceId } : {}),
    }),
  getScans: () => request<ScanJob[]>('/scans'),
  getScan: (id: number) => request<ScanJob>(`/scans/${id}`),
};
