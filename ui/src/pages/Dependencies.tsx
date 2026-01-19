import { useEffect, useState, useCallback, useRef } from 'react';
import { useSearchParams } from 'react-router-dom';
import { api } from '../api/client';
import { getPackageUrl, getVersionDiff } from '../utils';
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
  VersionDiffBadge,
  EcosystemBadge,
  EmptyState,
  LoadingSpinner,
  ErrorMessage,
} from '../components/common';
import type { PaginatedDependencies } from '../types';

type StatusFilter = 'all' | 'upgradable' | 'uptodate' | 'prod' | 'dev';
type EcosystemFilter = '' | 'npm' | 'maven' | 'gradle' | 'go';

const filterLabels: Record<StatusFilter, string> = {
  all: 'All Status',
  upgradable: 'Upgradable Only',
  uptodate: 'Up to Date Only',
  prod: 'Production Only',
  dev: 'Development Only',
};

const ecosystemLabels: Record<EcosystemFilter, string> = {
  '': 'All Ecosystems',
  npm: 'npm',
  maven: 'Maven',
  gradle: 'Gradle',
  go: 'Go',
};

const PAGE_SIZE = 50;

export function Dependencies() {
  const [searchParams, setSearchParams] = useSearchParams();
  const [paginatedData, setPaginatedData] = useState<PaginatedDependencies | null>(null);
  const [allRepos, setAllRepos] = useState<string[]>([]);
  const [selectedRepo, setSelectedRepo] = useState<string>('all');
  const [selectedEcosystem, setSelectedEcosystem] = useState<EcosystemFilter>('');
  const [search, setSearch] = useState('');
  const [debouncedSearch, setDebouncedSearch] = useState('');
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const searchTimeout = useRef<ReturnType<typeof setTimeout> | null>(null);

  // Read filter and page from URL params
  const statusFilter = (searchParams.get('filter') as StatusFilter) || 'all';
  const currentPage = parseInt(searchParams.get('page') || '1', 10);

  // Load repositories for filter dropdown (using efficient endpoint)
  const loadRepositories = useCallback(async () => {
    try {
      const repos = await api.getRepositoryNames();
      setAllRepos(repos);
    } catch {
      // Ignore - repos dropdown will just be empty
    }
  }, []);

  const loadDependencies = useCallback(async () => {
    setLoading(true);
    try {
      const upgradableOnly = statusFilter === 'upgradable';
      const repoFilter = selectedRepo !== 'all' ? selectedRepo : undefined;
      const ecosystemFilter = selectedEcosystem || undefined;
      const data = await api.getDependenciesPaginated(
        currentPage, PAGE_SIZE, upgradableOnly, repoFilter, ecosystemFilter, debouncedSearch || undefined
      );
      setPaginatedData(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load dependencies');
    } finally {
      setLoading(false);
    }
  }, [currentPage, statusFilter, selectedRepo, selectedEcosystem, debouncedSearch]);

  useEffect(() => {
    loadRepositories();
  }, [loadRepositories]);

  useEffect(() => {
    loadDependencies();
  }, [loadDependencies]);

  // Debounce search input
  useEffect(() => {
    if (searchTimeout.current) {
      clearTimeout(searchTimeout.current);
    }
    searchTimeout.current = setTimeout(() => {
      setDebouncedSearch(search);
      // Reset to page 1 when search changes
      if (search !== debouncedSearch) {
        searchParams.delete('page');
        setSearchParams(searchParams);
      }
    }, 300);
    return () => {
      if (searchTimeout.current) {
        clearTimeout(searchTimeout.current);
      }
    };
  }, [search]);

  // Filter client-side for uptodate/prod/dev (server handles upgradable)
  const filteredDeps = paginatedData?.data?.filter((dep) => {
    switch (statusFilter) {
      case 'uptodate':
        return !dep.is_outdated;
      case 'prod':
        return dep.type === 'dependency';
      case 'dev':
        return dep.type === 'devDependency';
      default:
        return true;
    }
  }) || [];

  const handleSearchChange = useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
    setSearch(e.target.value);
  }, []);

  const handleRepoChange = useCallback((e: React.ChangeEvent<HTMLSelectElement>) => {
    setSelectedRepo(e.target.value);
    searchParams.delete('page');
    setSearchParams(searchParams);
  }, [searchParams, setSearchParams]);

  const handleEcosystemChange = useCallback((e: React.ChangeEvent<HTMLSelectElement>) => {
    setSelectedEcosystem(e.target.value as EcosystemFilter);
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
    searchParams.delete('page');
    setSearchParams(searchParams);
  }, [searchParams, setSearchParams]);

  const handleIgnore = useCallback(async (name: string, ecosystem: string) => {
    try {
      await api.addIgnored({ name, ecosystem });
      loadDependencies();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to ignore dependency');
    }
  }, [loadDependencies]);

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
          value={selectedEcosystem}
          onChange={handleEcosystemChange}
          aria-label="Filter by ecosystem"
          style={selectStyle}
        >
          {Object.entries(ecosystemLabels).map(([value, label]) => (
            <option key={value} value={value}>{label}</option>
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
                <Th width="22%">Package</Th>
                <Th width="20%">Repository</Th>
                <Th width="8%">Ecosystem</Th>
                <Th width="8%">Type</Th>
                <Th width="12%">Current</Th>
                <Th width="18%">Latest</Th>
                <Th width="12%">Actions</Th>
              </TableHead>
              <TableBody>
                {filteredDeps.map((dep) => {
                  const diffType = dep.is_outdated ? getVersionDiff(dep.current_version, dep.latest_version) : 'unknown';
                  return (
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
                        <EcosystemBadge ecosystem={dep.ecosystem} />
                      </Td>
                      <Td>
                        <TypeBadge type={dep.type} />
                      </Td>
                      <Td>
                        <VersionBadge version={dep.current_version} isOutdated={dep.is_outdated} />
                      </Td>
                      <Td>
                        <div style={{ display: 'flex', alignItems: 'center', flexWrap: 'wrap', gap: '4px' }}>
                          <VersionBadge version={dep.latest_version} />
                          {dep.is_outdated && <VersionDiffBadge diffType={diffType} />}
                        </div>
                      </Td>
                      <Td>
                        {dep.is_outdated && (
                          <Button
                            variant="secondary"
                            onClick={() => handleIgnore(dep.name, dep.ecosystem)}
                            style={{ padding: '4px 8px', fontSize: '12px' }}
                          >
                            Ignore
                          </Button>
                        )}
                      </Td>
                    </TableRow>
                  );
                })}
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
