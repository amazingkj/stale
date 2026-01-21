import { useEffect, useState, useMemo, useCallback, useRef } from 'react';
import { Link } from 'react-router-dom';
import { api } from '../api/client';
import { getPackageUrl } from '../utils';
import { selectStyle } from '../constants/styles';
import {
  Button,
  Card,
  CardHeader,
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
  StatCard,
} from '../components/common';
import type { Dependency, ScanJob } from '../types';

type CardFilter = 'all' | 'upgradable' | 'uptodate' | 'prod' | 'dev';

const filterLabels: Record<CardFilter, string> = {
  all: 'All Dependencies',
  upgradable: 'Upgradable Dependencies',
  uptodate: 'Up to Date Dependencies',
  prod: 'Production Dependencies',
  dev: 'Development Dependencies',
};

export function Dashboard() {
  const [allDeps, setAllDeps] = useState<Dependency[]>([]);
  const [repositories, setRepositories] = useState<string[]>([]);
  const [scanning, setScanning] = useState(false);
  const [currentScan, setCurrentScan] = useState<ScanJob | null>(null);
  const [lastScanTime, setLastScanTime] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [selectedRepo, setSelectedRepo] = useState<string>('all');
  const [cardFilter, setCardFilter] = useState<CardFilter>('all');
  const [search, setSearch] = useState('');
  const [displayLimit, setDisplayLimit] = useState(20);

  // Filter dependencies by selected repo
  const repoFilteredDeps = useMemo(() => {
    if (selectedRepo === 'all') return allDeps;
    return allDeps.filter(dep => dep.repo_full_name === selectedRepo);
  }, [allDeps, selectedRepo]);

  // Compute stats from repo-filtered deps
  const stats = useMemo(() => {
    const total = repoFilteredDeps.length;
    const outdated = repoFilteredDeps.filter(d => d.is_outdated).length;
    const byType: Record<string, number> = {};
    repoFilteredDeps.forEach(dep => {
      byType[dep.type] = (byType[dep.type] || 0) + 1;
    });
    return {
      total_dependencies: total,
      outdated_count: outdated,
      up_to_date_count: total - outdated,
      by_type: byType,
    };
  }, [repoFilteredDeps]);

  // Apply card filter and search
  const displayDeps = useMemo(() => {
    let deps = repoFilteredDeps;

    // Apply card filter
    switch (cardFilter) {
      case 'upgradable':
        deps = deps.filter(d => d.is_outdated);
        break;
      case 'uptodate':
        deps = deps.filter(d => !d.is_outdated);
        break;
      case 'prod':
        deps = deps.filter(d => d.type === 'dependency');
        break;
      case 'dev':
        deps = deps.filter(d => d.type === 'devDependency');
        break;
    }

    // Apply search
    if (search) {
      const searchLower = search.toLowerCase();
      deps = deps.filter(d =>
        d.name.toLowerCase().includes(searchLower) ||
        d.repo_full_name?.toLowerCase().includes(searchLower)
      );
    }

    return deps;
  }, [repoFilteredDeps, cardFilter, search]);

  const loadData = useCallback(async () => {
    try {
      const [deps, scans, runningScan, repos] = await Promise.all([
        api.getDependencies(),
        api.getScans(),
        api.getRunningScan(),
        api.getRepositoryNames(),
      ]);
      setAllDeps(deps);
      setRepositories(repos);

      // Check if there's a running scan
      if (runningScan) {
        setCurrentScan(runningScan);
        setScanning(true);
      }

      // Find the last completed scan
      const completedScans = scans.filter(s => s.status === 'completed' && s.finished_at);
      if (completedScans.length > 0) {
        const lastScan = completedScans.reduce((latest, scan) => {
          if (!latest.finished_at) return scan;
          if (!scan.finished_at) return latest;
          return new Date(scan.finished_at) > new Date(latest.finished_at) ? scan : latest;
        });
        setLastScanTime(lastScan.finished_at || null);
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load data');
    }
  }, []);

  useEffect(() => {
    loadData();
  }, [loadData]);

  // Use ref to track polling interval for exponential backoff
  const pollIntervalRef = useRef(2000); // Start at 2 seconds
  const pollTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  useEffect(() => {
    if (!currentScan || currentScan.status === 'completed' || currentScan.status === 'failed') {
      pollIntervalRef.current = 2000; // Reset interval when scan is done
      return;
    }

    // Poll with exponential backoff (2s -> 4s -> 8s -> max 15s)
    const poll = async () => {
      try {
        const scan = await api.getScan(currentScan.id);
        setCurrentScan(scan);
        if (scan.status === 'completed') {
          setScanning(false);
          pollIntervalRef.current = 2000; // Reset
          if (scan.repos_found === 0) {
            setError('Scan completed but no repositories with manifest files were found. Check your sources configuration.');
          }
          loadData();
          return;
        } else if (scan.status === 'failed') {
          setScanning(false);
          pollIntervalRef.current = 2000; // Reset
          setError(scan.error || 'Scan failed');
          loadData();
          return;
        }
        // Increase interval with exponential backoff, max 15 seconds
        pollIntervalRef.current = Math.min(pollIntervalRef.current * 1.5, 15000);
      } catch {
        // On error, slow down polling more aggressively
        pollIntervalRef.current = Math.min(pollIntervalRef.current * 2, 15000);
      }
      // Schedule next poll
      pollTimeoutRef.current = setTimeout(poll, pollIntervalRef.current);
    };

    // Start polling
    pollTimeoutRef.current = setTimeout(poll, pollIntervalRef.current);

    return () => {
      if (pollTimeoutRef.current) {
        clearTimeout(pollTimeoutRef.current);
      }
    };
  }, [currentScan, loadData]);

  const handleScan = useCallback(async () => {
    if (scanning) return; // Prevent double-click
    setScanning(true);
    setError(null);
    try {
      const scan = await api.triggerScan();
      setCurrentScan(scan);
    } catch (err) {
      setScanning(false);
      const message = err instanceof Error ? err.message : 'Failed to start scan';
      // Check if it's "already running" error - try to find the running scan
      if (message.includes('already running')) {
        const runningScan = await api.getRunningScan();
        if (runningScan) {
          setCurrentScan(runningScan);
          setScanning(true);
          return;
        }
      }
      setError(message);
    }
  }, [scanning]);

  const handleCancelScan = useCallback(async () => {
    if (!currentScan) return;
    try {
      await api.cancelScan(currentScan.id);
      setScanning(false);
      setCurrentScan(null);
      loadData();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to cancel scan');
    }
  }, [currentScan, loadData]);

  const handleSearchChange = useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
    setSearch(e.target.value);
  }, []);

  const handleRepoChange = useCallback((e: React.ChangeEvent<HTMLSelectElement>) => {
    setSelectedRepo(e.target.value);
    setDisplayLimit(20); // Reset limit when filter changes
  }, []);

  const handleCardFilterChange = useCallback((filter: CardFilter) => {
    setCardFilter(filter);
    setDisplayLimit(20); // Reset limit when filter changes
  }, []);

  const outdatedPercent = stats.total_dependencies > 0
    ? Math.round((stats.outdated_count / stats.total_dependencies) * 100)
    : 0;

  const formatLastScanTime = (dateStr: string | null) => {
    if (!dateStr) return null;
    const date = new Date(dateStr);
    const now = new Date();
    const diffMs = now.getTime() - date.getTime();
    const diffMins = Math.floor(diffMs / 60000);
    const diffHours = Math.floor(diffMs / 3600000);
    const diffDays = Math.floor(diffMs / 86400000);

    if (diffMins < 1) return 'just now';
    if (diffMins < 60) return `${diffMins}m ago`;
    if (diffHours < 24) return `${diffHours}h ago`;
    if (diffDays < 7) return `${diffDays}d ago`;
    return date.toLocaleDateString();
  };

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: '24px' }}>
      {/* Controls */}
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
            cursor: 'text',
          }}
        />
        <select
          value={selectedRepo}
          onChange={handleRepoChange}
          aria-label="Filter by repository"
          style={{
            ...selectStyle,
            minWidth: '200px',
          }}
        >
          <option value="all">All Repositories ({repositories.length})</option>
          {repositories.map(repo => (
            <option key={repo} value={repo}>{repo}</option>
          ))}
        </select>
        {lastScanTime && !scanning && (
          <span style={{ fontSize: '13px', color: 'var(--text-muted)' }}>
            Last scan: {formatLastScanTime(lastScanTime)}
          </span>
        )}
        <Button onClick={handleScan} loading={scanning}>
          {scanning ? 'Scanning...' : 'Scan All'}
        </Button>
      </div>

      {/* Error */}
      {error && <ErrorMessage message={error} onDismiss={() => setError(null)} />}

      {/* Scan Progress */}
      {currentScan && (currentScan.status === 'pending' || currentScan.status === 'running') && (
        <div style={{
          padding: '14px 20px',
          borderRadius: 'var(--radius-lg)',
          background: 'var(--accent-gradient)',
          color: 'white',
          fontSize: '14px',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          boxShadow: '0 4px 16px -4px rgba(124, 181, 149, 0.4)',
        }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: '10px', fontWeight: 500 }}>
            <LoadingSpinner size="sm" />
            {currentScan.status === 'pending'
              ? 'Starting scan...'
              : `Scanning... Found ${currentScan.repos_found} repos, ${currentScan.deps_found} dependencies`}
          </div>
          <button
            onClick={handleCancelScan}
            style={{
              background: 'rgba(255,255,255,0.2)',
              border: 'none',
              color: 'white',
              padding: '6px 14px',
              borderRadius: 'var(--radius-full)',
              cursor: 'pointer',
              fontSize: '13px',
              fontWeight: 600,
              transition: 'background 0.2s ease',
            }}
            onMouseEnter={(e) => e.currentTarget.style.background = 'rgba(255,255,255,0.3)'}
            onMouseLeave={(e) => e.currentTarget.style.background = 'rgba(255,255,255,0.2)'}
          >
            Cancel
          </button>
        </div>
      )}

      {/* Stats Grid */}
      <div
        className="animate-stagger"
        style={{
          display: 'grid',
          gridTemplateColumns: 'repeat(auto-fit, minmax(170px, 1fr))',
          gap: '18px',
        }}
      >
        <StatCard
          label="Total"
          value={stats.total_dependencies}
          color="accent"
          active={cardFilter === 'all'}
          onClick={() => handleCardFilterChange('all')}
        />
        <StatCard
          label="Upgradable"
          value={stats.outdated_count}
          subtitle={stats.total_dependencies ? `${outdatedPercent}%` : undefined}
          color="danger"
          active={cardFilter === 'upgradable'}
          onClick={() => handleCardFilterChange('upgradable')}
        />
        <StatCard
          label="Up to Date"
          value={stats.up_to_date_count}
          color="success"
          active={cardFilter === 'uptodate'}
          onClick={() => handleCardFilterChange('uptodate')}
        />
        <StatCard
          label="Production"
          value={stats.by_type?.dependency ?? 0}
          color="warning"
          active={cardFilter === 'prod'}
          onClick={() => handleCardFilterChange('prod')}
        />
        <StatCard
          label="Development"
          value={stats.by_type?.devDependency ?? 0}
          color="accent"
          active={cardFilter === 'dev'}
          onClick={() => handleCardFilterChange('dev')}
        />
      </div>

      {/* Dependencies Table */}
      <Card minHeight={400}>
        <CardHeader style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: '12px' }}>
            <h2 style={{ fontSize: '16px', fontWeight: 600, color: 'var(--text-primary)', margin: 0 }}>
              {filterLabels[cardFilter]}
            </h2>
            <span style={{ fontSize: '13px', color: 'var(--text-muted)' }}>
              ({displayDeps.length})
            </span>
          </div>
          <Link
            to={(() => {
              const params = new URLSearchParams();
              if (cardFilter !== 'all') params.set('filter', cardFilter);
              if (selectedRepo !== 'all') params.set('repo', selectedRepo);
              const query = params.toString();
              return query ? `/dependencies?${query}` : '/dependencies';
            })()}
            style={{
              fontSize: '13px',
              color: 'var(--accent)',
              textDecoration: 'none',
            }}
          >
            View all {displayDeps.length} →
          </Link>
        </CardHeader>

        {displayDeps.length === 0 ? (
          <EmptyState
            icon={stats.total_dependencies === 0 ? '📦' : undefined}
            description={
              stats.total_dependencies === 0
                ? 'No dependencies found.\nAdd a source and scan to get started.'
                : search
                  ? 'No dependencies matching your search.'
                  : 'No dependencies in this category.'
            }
          />
        ) : (
          <>
            <Table fixed minHeight={300}>
              <TableHead>
                <Th width="35%">Package</Th>
                <Th width="30%">Repository</Th>
                <Th width="17.5%">Current</Th>
                <Th width="17.5%">Latest</Th>
              </TableHead>
              <TableBody>
                {displayDeps.slice(0, displayLimit).map((dep) => (
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
                      <span style={{ marginLeft: '8px' }}>
                        <TypeBadge type={dep.type} />
                      </span>
                    </Td>
                    <Td secondary>{dep.repo_full_name}</Td>
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
            {displayDeps.length > displayLimit && (
              <div style={{
                padding: '12px 20px',
                textAlign: 'center',
                color: 'var(--text-muted)',
                fontSize: '13px',
                borderTop: '1px solid var(--border-color)',
                display: 'flex',
                justifyContent: 'center',
                alignItems: 'center',
                gap: '16px',
              }}>
                <span>Showing {displayLimit} of {displayDeps.length}</span>
                <button
                  onClick={() => setDisplayLimit(prev => Math.min(prev + 50, displayDeps.length))}
                  style={{
                    padding: '4px 12px',
                    fontSize: '13px',
                    fontWeight: 500,
                    color: 'var(--accent)',
                    backgroundColor: 'var(--accent-light)',
                    border: 'none',
                    borderRadius: '4px',
                    cursor: 'pointer',
                  }}
                >
                  +50 more
                </button>
                <Link to={(() => {
                  const params = new URLSearchParams();
                  if (cardFilter !== 'all') params.set('filter', cardFilter);
                  if (selectedRepo !== 'all') params.set('repo', selectedRepo);
                  const query = params.toString();
                  return query ? `/dependencies?${query}` : '/dependencies';
                })()} style={{ color: 'var(--accent)' }}>View all →</Link>
              </div>
            )}
          </>
        )}
      </Card>
    </div>
  );
}
