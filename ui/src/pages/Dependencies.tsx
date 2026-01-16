import { useEffect, useState, useMemo, useCallback } from 'react';
import { useSearchParams } from 'react-router-dom';
import { api } from '../api/client';
import { getPackageUrl } from '../utils';
import { selectStyle } from '../constants/styles';
import {
  Button,
  Card,
  Table,
  TableHead,
  TableBody,
  TableRow,
  Th,
  Td,
  TypeBadge,
  VersionBadge,
  EmptyState,
  LoadingSpinner,
  ErrorMessage,
} from '../components/common';
import type { PaginatedDependencies } from '../types';

type StatusFilter = 'all' | 'upgradable' | 'uptodate' | 'prod' | 'dev';

const filterLabels: Record<StatusFilter, string> = {
  all: 'All Status',
  upgradable: 'Upgradable Only',
  uptodate: 'Up to Date Only',
  prod: 'Production Only',
  dev: 'Development Only',
};

const PAGE_SIZE = 50;

export function Dependencies() {
  const [searchParams, setSearchParams] = useSearchParams();
  const [paginatedData, setPaginatedData] = useState<PaginatedDependencies | null>(null);
  const [allRepos, setAllRepos] = useState<string[]>([]);
  const [selectedRepo, setSelectedRepo] = useState<string>('all');
  const [search, setSearch] = useState('');
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // Read filter and page from URL params
  const statusFilter = (searchParams.get('filter') as StatusFilter) || 'all';
  const currentPage = parseInt(searchParams.get('page') || '1', 10);

  // Load repositories for filter dropdown (from non-paginated endpoint)
  const loadRepositories = useCallback(async () => {
    try {
      const data = await api.getDependencies();
      const repoSet = new Set<string>();
      data.forEach(dep => {
        if (dep.repo_full_name) {
          repoSet.add(dep.repo_full_name);
        }
      });
      setAllRepos(Array.from(repoSet).sort());
    } catch {
      // Ignore - repos dropdown will just be empty
    }
  }, []);

  const loadDependencies = useCallback(async () => {
    setLoading(true);
    try {
      const upgradableOnly = statusFilter === 'upgradable';
      const repoFilter = selectedRepo !== 'all' ? selectedRepo : undefined;
      const data = await api.getDependenciesPaginated(currentPage, PAGE_SIZE, upgradableOnly, repoFilter);
      setPaginatedData(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load dependencies');
    } finally {
      setLoading(false);
    }
  }, [currentPage, statusFilter, selectedRepo]);

  useEffect(() => {
    loadRepositories();
  }, [loadRepositories]);

  useEffect(() => {
    loadDependencies();
  }, [loadDependencies]);

  // Client-side filtering for search and uptodate/prod/dev filters
  const filteredDeps = useMemo(() => {
    if (!paginatedData?.data) return [];
    return paginatedData.data.filter((dep) => {
      // Additional status filters (server handles 'upgradable' and 'all')
      switch (statusFilter) {
        case 'uptodate':
          if (dep.is_outdated) return false;
          break;
        case 'prod':
          if (dep.type !== 'dependency') return false;
          break;
        case 'dev':
          if (dep.type !== 'devDependency') return false;
          break;
      }
      // Search filter
      if (search) {
        const searchLower = search.toLowerCase();
        return dep.name.toLowerCase().includes(searchLower) ||
          dep.repo_full_name?.toLowerCase().includes(searchLower);
      }
      return true;
    });
  }, [paginatedData, statusFilter, search]);

  const handleSearchChange = useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
    setSearch(e.target.value);
  }, []);

  const handleRepoChange = useCallback((e: React.ChangeEvent<HTMLSelectElement>) => {
    setSelectedRepo(e.target.value);
    // Reset to page 1 when changing repo filter
    searchParams.delete('page');
    setSearchParams(searchParams);
  }, [searchParams, setSearchParams]);

  const handleStatusChange = useCallback((e: React.ChangeEvent<HTMLSelectElement>) => {
    const value = e.target.value;
    if (value === 'all') {
      searchParams.delete('filter');
    } else {
      searchParams.set('filter', value);
    }
    // Reset to page 1 when changing status filter
    searchParams.delete('page');
    setSearchParams(searchParams);
  }, [searchParams, setSearchParams]);

  const handlePageChange = useCallback((page: number) => {
    if (page === 1) {
      searchParams.delete('page');
    } else {
      searchParams.set('page', String(page));
    }
    setSearchParams(searchParams);
  }, [searchParams, setSearchParams]);

  const handleExport = useCallback(() => {
    const params = new URLSearchParams();
    if (statusFilter !== 'all') {
      params.set('filter', statusFilter);
    }
    if (selectedRepo !== 'all') {
      params.set('repo', selectedRepo);
    }
    const queryString = params.toString();
    const url = `/api/v1/dependencies/export${queryString ? `?${queryString}` : ''}`;
    window.open(url, '_blank');
  }, [statusFilter, selectedRepo]);

  const totalPages = paginatedData?.total_pages || 1;

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: '24px' }}>
      <div style={{ display: 'flex', justifyContent: 'flex-end', alignItems: 'center', flexWrap: 'wrap', gap: '12px' }}>
        <span style={{ fontSize: '14px', color: 'var(--text-secondary)', marginRight: 'auto' }}>
          {paginatedData?.total || 0} dependencies
          {statusFilter === 'upgradable' && ' (upgradable only)'}
          {selectedRepo !== 'all' && ` in ${selectedRepo}`}
        </span>
        <input
          type="text"
          value={search}
          onChange={handleSearchChange}
          placeholder="Search packages..."
          aria-label="Search packages"
          style={{
            ...selectStyle,
            width: '200px',
          }}
        />
        <select
          value={selectedRepo}
          onChange={handleRepoChange}
          aria-label="Filter by repository"
          style={{
            ...selectStyle,
            minWidth: '180px',
          }}
        >
          <option value="all">All Repositories</option>
          {allRepos.map(repo => (
            <option key={repo} value={repo}>{repo}</option>
          ))}
        </select>
        <select
          value={statusFilter}
          onChange={handleStatusChange}
          aria-label="Filter by status"
          style={selectStyle}
        >
          {Object.entries(filterLabels).map(([value, label]) => (
            <option key={value} value={value}>{label}</option>
          ))}
        </select>
        <Button variant="secondary" onClick={handleExport}>
          <span style={{ marginRight: '6px' }}>⬇</span> Export CSV
        </Button>
      </div>

      {error && <ErrorMessage message={error} onDismiss={() => setError(null)} />}

      {loading ? (
        <LoadingSpinner fullPage text="Loading..." />
      ) : filteredDeps.length === 0 ? (
        <Card>
          <EmptyState
            icon="📦"
            description={search ? 'No dependencies matching your search.' : 'No dependencies found.'}
          />
        </Card>
      ) : (
        <>
          <Card>
            <Table fixed>
              <TableHead>
                <Th width="30%">Package</Th>
                <Th width="28%">Repository</Th>
                <Th width="12%">Type</Th>
                <Th width="15%">Current</Th>
                <Th width="15%">Latest</Th>
              </TableHead>
              <TableBody>
                {filteredDeps.map((dep) => (
                  <TableRow key={dep.id}>
                    <Td>
                      <a
                        href={getPackageUrl(dep.name, dep.ecosystem)}
                        target="_blank"
                        rel="noopener noreferrer"
                        style={{ color: 'var(--accent)', textDecoration: 'none', fontWeight: 500 }}
                      >
                        {dep.name}
                      </a>
                    </Td>
                    <Td secondary>{dep.repo_full_name}</Td>
                    <Td>
                      <TypeBadge type={dep.type} />
                    </Td>
                    <Td>
                      <VersionBadge version={dep.current_version} isOutdated={dep.is_outdated} />
                    </Td>
                    <Td>
                      <VersionBadge version={dep.latest_version} />
                    </Td>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </Card>

          {totalPages > 1 && (
            <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', gap: '8px' }}>
              <Button
                variant="secondary"
                onClick={() => handlePageChange(currentPage - 1)}
                disabled={currentPage === 1}
                style={{ padding: '8px 12px' }}
              >
                Previous
              </Button>
              <span style={{ color: 'var(--text-secondary)', fontSize: '14px', minWidth: '100px', textAlign: 'center' }}>
                Page {currentPage} of {totalPages}
              </span>
              <Button
                variant="secondary"
                onClick={() => handlePageChange(currentPage + 1)}
                disabled={currentPage >= totalPages}
                style={{ padding: '8px 12px' }}
              >
                Next
              </Button>
            </div>
          )}
        </>
      )}
    </div>
  );
}
