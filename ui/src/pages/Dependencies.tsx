import { useEffect, useState, useCallback, useRef, useMemo } from 'react';
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
  Pagination,
} from '../components/common';
import type { PaginatedDependencies, IgnoredDependency, FilterOptions } from '../types';

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
  const [filterOptions, setFilterOptions] = useState<FilterOptions>({ repos: [], packages: [], ecosystems: [] });
  const [selectedRepo, setSelectedRepo] = useState<string>(searchParams.get('repo') || 'all');
  const [selectedPackage, setSelectedPackage] = useState<string>('all');
  const [selectedEcosystem, setSelectedEcosystem] = useState<EcosystemFilter>('');
  const [search, setSearch] = useState('');
  const [debouncedSearch, setDebouncedSearch] = useState('');
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [selectedIds, setSelectedIds] = useState<Set<number>>(new Set());
  const [bulkLoading, setBulkLoading] = useState(false);
  const [ignoredDeps, setIgnoredDeps] = useState<IgnoredDependency[]>([]);
  const [showIgnored, setShowIgnored] = useState(false);
  const [selectedIgnoredIds, setSelectedIgnoredIds] = useState<Set<number>>(new Set());
  const [bulkRestoreLoading, setBulkRestoreLoading] = useState(false);
  const searchTimeout = useRef<ReturnType<typeof setTimeout> | null>(null);

  // Read filter and page from URL params
  const statusFilter = (searchParams.get('filter') as StatusFilter) || 'all';
  const currentPage = parseInt(searchParams.get('page') || '1', 10);

  // Load filter options based on current selections (cascading filters)
  const loadFilterOptions = useCallback(async () => {
    try {
      const repoFilter = selectedRepo !== 'all' ? selectedRepo : undefined;
      const ecosystemFilter = selectedEcosystem || undefined;
      const packageFilter = selectedPackage !== 'all' ? selectedPackage : undefined;
      const options = await api.getFilterOptions(repoFilter, ecosystemFilter, statusFilter, packageFilter);
      setFilterOptions(options);
    } catch {
      // Ignore - dropdowns will just be empty
    }
  }, [selectedRepo, selectedEcosystem, statusFilter, selectedPackage]);

  const loadDependencies = useCallback(async () => {
    setLoading(true);
    try {
      const repoFilter = selectedRepo !== 'all' ? selectedRepo : undefined;
      const ecosystemFilter = selectedEcosystem || undefined;
      const data = await api.getDependenciesPaginated(
        currentPage, PAGE_SIZE, statusFilter, repoFilter, ecosystemFilter, debouncedSearch || undefined
      );
      setPaginatedData(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load dependencies');
    } finally {
      setLoading(false);
    }
  }, [currentPage, statusFilter, selectedRepo, selectedEcosystem, debouncedSearch]);

  const loadIgnored = useCallback(async () => {
    try {
      const data = await api.getIgnored();
      setIgnoredDeps(data);
    } catch {
      // Ignore errors for ignored list
    }
  }, []);

  useEffect(() => {
    loadFilterOptions();
    loadIgnored();
  }, [loadFilterOptions, loadIgnored]);

  useEffect(() => {
    loadDependencies();
  }, [loadDependencies]);

  // Reload filter options when page becomes visible (after scan completes in another tab/returning to page)
  useEffect(() => {
    const handleVisibilityChange = () => {
      if (document.visibilityState === 'visible') {
        loadFilterOptions();
      }
    };
    document.addEventListener('visibilitychange', handleVisibilityChange);
    return () => document.removeEventListener('visibilitychange', handleVisibilityChange);
  }, [loadFilterOptions]);

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

  // Dependencies are filtered server-side, with optional client-side package filter
  const filteredDeps = useMemo(() => {
    const data = paginatedData?.data || [];
    if (selectedPackage === 'all') return data;
    return data.filter(dep => dep.name === selectedPackage);
  }, [paginatedData?.data, selectedPackage]);

  // Reset package selection if it's no longer available in options
  useEffect(() => {
    if (selectedPackage !== 'all' && filterOptions.packages.length > 0 && !filterOptions.packages.includes(selectedPackage)) {
      setSelectedPackage('all');
    }
  }, [filterOptions.packages, selectedPackage]);

  const handleSearchChange = useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
    setSearch(e.target.value);
  }, []);

  const handleRepoChange = useCallback((e: React.ChangeEvent<HTMLSelectElement>) => {
    const value = e.target.value;
    setSelectedRepo(value);
    setSelectedPackage('all'); // Reset package when repo changes
    if (value === 'all') {
      searchParams.delete('repo');
    } else {
      searchParams.set('repo', value);
    }
    searchParams.delete('page');
    setSearchParams(searchParams);
  }, [searchParams, setSearchParams]);

  const handleEcosystemChange = useCallback((e: React.ChangeEvent<HTMLSelectElement>) => {
    setSelectedEcosystem(e.target.value as EcosystemFilter);
    setSelectedPackage('all'); // Reset package when ecosystem changes
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
    if (selectedPackage !== 'all') {
      params.set('package', selectedPackage);
    }
    if (selectedEcosystem) {
      params.set('ecosystem', selectedEcosystem);
    }
    if (debouncedSearch) {
      params.set('search', debouncedSearch);
    }
    const queryString = params.toString();
    const url = `/api/v1/dependencies/export${queryString ? `?${queryString}` : ''}`;
    window.open(url, '_blank');
  }, [statusFilter, selectedRepo, selectedPackage, selectedEcosystem, debouncedSearch]);

  // Selection handlers
  const outdatedDeps = useMemo(() =>
    filteredDeps.filter(dep => dep.is_outdated),
    [filteredDeps]
  );

  const handleSelectAll = useCallback((checked: boolean) => {
    if (checked) {
      setSelectedIds(new Set(outdatedDeps.map(dep => dep.id)));
    } else {
      setSelectedIds(new Set());
    }
  }, [outdatedDeps]);

  const handleSelectOne = useCallback((id: number, checked: boolean) => {
    setSelectedIds(prev => {
      const next = new Set(prev);
      if (checked) {
        next.add(id);
      } else {
        next.delete(id);
      }
      return next;
    });
  }, []);

  const handleBulkIgnore = useCallback(async () => {
    if (selectedIds.size === 0) return;

    setBulkLoading(true);
    try {
      const items = filteredDeps
        .filter(dep => selectedIds.has(dep.id))
        .map(dep => ({ name: dep.name, ecosystem: dep.ecosystem }));

      await api.bulkAddIgnored(items);
      setSelectedIds(new Set());
      loadDependencies();
      loadIgnored();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to ignore dependencies');
    } finally {
      setBulkLoading(false);
    }
  }, [selectedIds, filteredDeps, loadDependencies, loadIgnored]);

  const handleClearSelection = useCallback(() => {
    setSelectedIds(new Set());
  }, []);

  // Ignored selection handlers
  const handleSelectAllIgnored = useCallback((checked: boolean) => {
    if (checked) {
      setSelectedIgnoredIds(new Set(ignoredDeps.map(dep => dep.id)));
    } else {
      setSelectedIgnoredIds(new Set());
    }
  }, [ignoredDeps]);

  const handleSelectOneIgnored = useCallback((id: number, checked: boolean) => {
    setSelectedIgnoredIds(prev => {
      const next = new Set(prev);
      if (checked) {
        next.add(id);
      } else {
        next.delete(id);
      }
      return next;
    });
  }, []);

  const handleBulkRestore = useCallback(async () => {
    if (selectedIgnoredIds.size === 0) return;
    setBulkRestoreLoading(true);
    try {
      await api.bulkRemoveIgnored(Array.from(selectedIgnoredIds));
      setSelectedIgnoredIds(new Set());
      loadIgnored();
      loadDependencies();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to restore dependencies');
    } finally {
      setBulkRestoreLoading(false);
    }
  }, [selectedIgnoredIds, loadIgnored, loadDependencies]);

  const handleClearIgnoredSelection = useCallback(() => {
    setSelectedIgnoredIds(new Set());
  }, []);

  const isAllSelected = outdatedDeps.length > 0 && outdatedDeps.every(dep => selectedIds.has(dep.id));
  const isSomeSelected = selectedIds.size > 0;

  const totalPages = paginatedData?.total_pages || 1;

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: '24px' }}>
      {/* Summary row - fixed height to prevent layout shifts */}
      <div style={{ minHeight: '24px', display: 'flex', alignItems: 'center' }}>
        <span style={{ fontSize: '14px', color: 'var(--text-secondary)' }}>
          {paginatedData?.total || 0} dependencies
          {statusFilter === 'upgradable' && ' (upgradable only)'}
          {selectedRepo !== 'all' && ` in ${selectedRepo}`}
        </span>
      </div>

      {/* Filter controls */}
      <div style={{ display: 'flex', justifyContent: 'flex-end', alignItems: 'center', flexWrap: 'wrap', gap: '12px' }}>
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
          <option value="all">All Repositories ({filterOptions.repos.length})</option>
          {filterOptions.repos.map(repo => (
            <option key={repo} value={repo}>{repo}</option>
          ))}
        </select>
        <select
          value={selectedPackage}
          onChange={(e) => setSelectedPackage(e.target.value)}
          aria-label="Filter by package"
          style={{
            ...selectStyle,
            minWidth: '150px',
          }}
        >
          <option value="all">All Packages ({filterOptions.packages.length})</option>
          {filterOptions.packages.map(pkg => (
            <option key={pkg} value={pkg}>{pkg}</option>
          ))}
        </select>
        <select
          value={selectedEcosystem}
          onChange={handleEcosystemChange}
          aria-label="Filter by ecosystem"
          style={selectStyle}
        >
          <option value="">All Ecosystems ({filterOptions.ecosystems.length})</option>
          {filterOptions.ecosystems.map(eco => (
            <option key={eco} value={eco}>{ecosystemLabels[eco as EcosystemFilter] || eco}</option>
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
        <Button
          variant={showIgnored ? 'primary' : 'secondary'}
          onClick={() => setShowIgnored(!showIgnored)}
          style={{ position: 'relative' }}
        >
          <span style={{ marginRight: '6px' }}>🚫</span>
          Ignored
          {ignoredDeps.length > 0 && (
            <span style={{
              marginLeft: '6px',
              backgroundColor: showIgnored ? 'white' : 'var(--accent)',
              color: showIgnored ? 'var(--accent)' : 'white',
              borderRadius: '10px',
              padding: '2px 6px',
              fontSize: '11px',
              fontWeight: 600,
            }}>
              {ignoredDeps.length}
            </span>
          )}
        </Button>
      </div>

      {error && <ErrorMessage message={error} onDismiss={() => setError(null)} />}

      {/* Ignored Dependencies Section */}
      {showIgnored && (
        <Card>
          <div style={{ padding: '16px' }}>
            <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: '12px' }}>
              <h3 style={{ fontSize: '16px', fontWeight: 600, color: 'var(--text-primary)', margin: 0 }}>
                Ignored Dependencies ({ignoredDeps.length})
              </h3>
              {selectedIgnoredIds.size > 0 && (
                <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
                  <span style={{ fontSize: '13px', color: 'var(--accent)' }}>{selectedIgnoredIds.size} selected</span>
                  <Button
                    variant="primary"
                    onClick={handleBulkRestore}
                    disabled={bulkRestoreLoading}
                    style={{ padding: '4px 10px', fontSize: '12px' }}
                  >
                    {bulkRestoreLoading ? 'Restoring...' : 'Restore Selected'}
                  </Button>
                  <Button
                    variant="secondary"
                    onClick={handleClearIgnoredSelection}
                    style={{ padding: '4px 10px', fontSize: '12px' }}
                  >
                    Clear
                  </Button>
                </div>
              )}
            </div>
            {ignoredDeps.length === 0 ? (
              <p style={{ color: 'var(--text-muted)', fontSize: '14px', margin: 0 }}>
                No ignored dependencies. Select outdated dependencies and click "Ignore Selected" to add them here.
              </p>
            ) : (
              <div style={{ display: 'flex', flexDirection: 'column', gap: '8px' }}>
                <div style={{ display: 'flex', alignItems: 'center', gap: '8px', padding: '4px 0' }}>
                  <input
                    type="checkbox"
                    checked={ignoredDeps.length > 0 && ignoredDeps.every(dep => selectedIgnoredIds.has(dep.id))}
                    onChange={(e) => handleSelectAllIgnored(e.target.checked)}
                    style={{ width: '16px', height: '16px', cursor: 'pointer' }}
                  />
                  <span style={{ fontSize: '12px', color: 'var(--text-muted)' }}>Select all</span>
                </div>
                {ignoredDeps.map((dep) => (
                  <div
                    key={dep.id}
                    style={{
                      display: 'flex',
                      alignItems: 'center',
                      gap: '12px',
                      padding: '10px 14px',
                      backgroundColor: selectedIgnoredIds.has(dep.id) ? 'var(--accent-light)' : 'var(--bg-secondary)',
                      borderRadius: '6px',
                    }}
                  >
                    <input
                      type="checkbox"
                      checked={selectedIgnoredIds.has(dep.id)}
                      onChange={(e) => handleSelectOneIgnored(dep.id, e.target.checked)}
                      style={{ width: '16px', height: '16px', cursor: 'pointer' }}
                    />
                    <div style={{ flex: 1 }}>
                      <span style={{ fontWeight: 500, color: 'var(--text-primary)' }}>{dep.name}</span>
                      <span style={{ marginLeft: '8px', fontSize: '12px', color: 'var(--text-muted)' }}>
                        {dep.ecosystem || 'all ecosystems'}
                      </span>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>
        </Card>
      )}


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
                <Th width="4%" noEllipsis>
                  <input
                    type="checkbox"
                    checked={isAllSelected}
                    onChange={(e) => handleSelectAll(e.target.checked)}
                    style={{ width: '16px', height: '16px', cursor: 'pointer' }}
                    title="Select all outdated"
                  />
                </Th>
                <Th width="4%" noEllipsis>#</Th>
                <Th width="22%">Package</Th>
                <Th width="24%">Repository</Th>
                <Th width="10%">Ecosystem</Th>
                <Th width="10%">Type</Th>
                <Th width="13%">Current</Th>
                <Th width="13%">Latest</Th>
              </TableHead>
              <TableBody>
                {filteredDeps.map((dep, index) => {
                  const diffType = dep.is_outdated ? getVersionDiff(dep.current_version, dep.latest_version) : 'unknown';
                  const isSelected = selectedIds.has(dep.id);
                  const rowNumber = (currentPage - 1) * PAGE_SIZE + index + 1;
                  return (
                    <TableRow key={dep.id} style={isSelected ? { backgroundColor: 'var(--accent-light)' } : undefined}>
                      <Td noEllipsis>
                        {dep.is_outdated ? (
                          <input
                            type="checkbox"
                            checked={isSelected}
                            onChange={(e) => handleSelectOne(dep.id, e.target.checked)}
                            style={{ width: '16px', height: '16px', cursor: 'pointer' }}
                          />
                        ) : (
                          <span style={{ width: '16px', display: 'inline-block' }} />
                        )}
                      </Td>
                      <Td noEllipsis muted style={{ fontSize: '12px' }}>{rowNumber}</Td>
                      <Td>
                        {(() => {
                          const url = getPackageUrl(dep.name, dep.ecosystem);
                          return url && url !== '#' ? (
                            <a
                              href={url}
                              target="_blank"
                              rel="noopener noreferrer"
                              style={{ color: 'var(--accent)', textDecoration: 'none', fontWeight: 500 }}
                            >
                              {dep.name}
                            </a>
                          ) : (
                            <span style={{ fontWeight: 500, color: 'var(--text-primary)' }}>{dep.name}</span>
                          );
                        })()}
                      </Td>
                      <Td secondary>{dep.repo_full_name}</Td>
                      <Td noEllipsis>
                        <EcosystemBadge ecosystem={dep.ecosystem} />
                      </Td>
                      <Td noEllipsis>
                        <TypeBadge type={dep.type} />
                      </Td>
                      <Td noEllipsis>
                        <VersionBadge version={dep.current_version} isOutdated={dep.is_outdated} />
                      </Td>
                      <Td noEllipsis>
                        <div style={{ display: 'flex', alignItems: 'center', flexWrap: 'wrap', gap: '4px' }}>
                          <VersionBadge version={dep.latest_version} />
                          {dep.is_outdated && <VersionDiffBadge diffType={diffType} />}
                        </div>
                      </Td>
                    </TableRow>
                  );
                })}
              </TableBody>
            </Table>
          </Card>

          <Pagination
            currentPage={currentPage}
            totalPages={totalPages}
            totalItems={paginatedData?.total}
            pageSize={PAGE_SIZE}
            onPageChange={handlePageChange}
          />
        </>
      )}

      {/* Floating Action Bar */}
      {isSomeSelected && (
        <div style={{
          position: 'fixed',
          bottom: '24px',
          left: '50%',
          transform: 'translateX(-50%)',
          display: 'flex',
          alignItems: 'center',
          gap: '12px',
          padding: '12px 20px',
          backgroundColor: 'var(--bg-card)',
          borderRadius: '12px',
          boxShadow: '0 4px 20px rgba(0, 0, 0, 0.15)',
          border: '1px solid var(--border-color)',
          zIndex: 100,
        }}>
          <span style={{ fontSize: '14px', fontWeight: 600, color: 'var(--accent)' }}>
            {selectedIds.size} selected
          </span>
          <Button
            variant="primary"
            onClick={handleBulkIgnore}
            disabled={bulkLoading}
          >
            {bulkLoading ? 'Ignoring...' : 'Ignore Selected'}
          </Button>
          <Button
            variant="secondary"
            onClick={handleClearSelection}
          >
            Clear
          </Button>
        </div>
      )}
    </div>
  );
}
