import { useEffect, useState, useCallback, useMemo } from 'react';
import { Link } from 'react-router-dom';
import { api } from '../api/client';
import { selectStyle, inputStyle } from '../constants/styles';
import {
  Button,
  Card,
  Table,
  TableHead,
  TableBody,
  TableRow,
  Th,
  Td,
  EcosystemBadge,
  EmptyState,
  LoadingSpinner,
  ErrorMessage,
} from '../components/common';
import type { Repository, Source } from '../types';

type ViewMode = 'list' | 'grouped';
type EcosystemFilter = '' | 'npm' | 'maven' | 'gradle' | 'go';
type OutdatedFilter = '' | 'outdated' | 'uptodate';

export function Repositories() {
  const [repositories, setRepositories] = useState<Repository[]>([]);
  const [sources, setSources] = useState<Source[]>([]);
  const [selectedSource, setSelectedSource] = useState<number | undefined>();
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [selectedIds, setSelectedIds] = useState<Set<number>>(new Set());
  const [deleting, setDeleting] = useState(false);
  const [viewMode, setViewMode] = useState<ViewMode>('grouped');
  const [collapsedGroups, setCollapsedGroups] = useState<Set<number>>(new Set());
  const [search, setSearch] = useState('');
  const [ecosystemFilter, setEcosystemFilter] = useState<EcosystemFilter>('');
  const [outdatedFilter, setOutdatedFilter] = useState<OutdatedFilter>('');

  const loadSources = useCallback(async () => {
    try {
      const data = await api.getSources();
      setSources(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load sources');
    }
  }, []);

  const loadRepositories = useCallback(async () => {
    setLoading(true);
    try {
      const data = await api.getRepositories(selectedSource);
      setRepositories(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load repositories');
    } finally {
      setLoading(false);
    }
  }, [selectedSource]);

  useEffect(() => {
    loadSources();
  }, [loadSources]);

  useEffect(() => {
    loadRepositories();
  }, [loadRepositories]);

  const handleSourceChange = useCallback((e: React.ChangeEvent<HTMLSelectElement>) => {
    setSelectedSource(e.target.value ? Number(e.target.value) : undefined);
    setSelectedIds(new Set());
    setSearch('');
    setEcosystemFilter('');
    setOutdatedFilter('');
  }, []);

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

  const handleBulkDelete = useCallback(async () => {
    if (selectedIds.size === 0) return;
    if (!confirm(`Remove ${selectedIds.size} repositories from dashboard?\n(This will not delete the actual repositories)`)) return;

    setDeleting(true);
    try {
      await api.bulkDeleteRepositories(Array.from(selectedIds));
      setRepositories(prev => prev.filter(r => !selectedIds.has(r.id)));
      setSelectedIds(new Set());
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to remove repositories');
    } finally {
      setDeleting(false);
    }
  }, [selectedIds]);

  const toggleGroup = useCallback((sourceId: number) => {
    setCollapsedGroups(prev => {
      const next = new Set(prev);
      if (next.has(sourceId)) {
        next.delete(sourceId);
      } else {
        next.add(sourceId);
      }
      return next;
    });
  }, []);

  // Filter repositories
  const filteredRepositories = useMemo(() => {
    let filtered = repositories;

    // Search filter
    if (search) {
      const searchLower = search.toLowerCase();
      filtered = filtered.filter(repo =>
        repo.full_name.toLowerCase().includes(searchLower) ||
        repo.name.toLowerCase().includes(searchLower)
      );
    }

    // Ecosystem filter
    if (ecosystemFilter) {
      filtered = filtered.filter(repo => {
        switch (ecosystemFilter) {
          case 'npm': return repo.has_package_json;
          case 'go': return repo.has_go_mod;
          case 'maven': return repo.has_pom_xml;
          case 'gradle': return repo.has_build_gradle;
          default: return true;
        }
      });
    }

    // Outdated filter
    if (outdatedFilter) {
      filtered = filtered.filter(repo => {
        if (outdatedFilter === 'outdated') return repo.outdated_count > 0;
        if (outdatedFilter === 'uptodate') return repo.outdated_count === 0;
        return true;
      });
    }

    return filtered;
  }, [repositories, search, ecosystemFilter, outdatedFilter]);

  // Group repositories by source (using filtered repositories)
  // Keep all groups visible even when empty (for filter visibility)
  const groupedRepositories = useMemo(() => {
    const groups = new Map<number, { source: Source; repos: Repository[]; totalRepos: number }>();

    for (const source of sources) {
      // Count total repos for this source (before filtering)
      const totalRepos = repositories.filter(r => r.source_id === source.id).length;
      groups.set(source.id, { source, repos: [], totalRepos });
    }

    for (const repo of filteredRepositories) {
      const group = groups.get(repo.source_id);
      if (group) {
        group.repos.push(repo);
      }
    }

    // Keep all groups that have repos (even if filtered to 0), sort by source name
    return Array.from(groups.values())
      .filter(g => g.totalRepos > 0)
      .sort((a, b) => a.source.name.localeCompare(b.source.name));
  }, [filteredRepositories, sources, repositories]);

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: '24px' }}>
      <div style={{ display: 'flex', justifyContent: 'flex-end', alignItems: 'center', gap: '12px', flexWrap: 'wrap' }}>
        <span style={{ fontSize: '14px', color: 'var(--text-secondary)', marginRight: 'auto' }}>
          {filteredRepositories.length === repositories.length
            ? `${repositories.length} repositories`
            : `${filteredRepositories.length} of ${repositories.length} repositories`}
          {selectedIds.size > 0 && ` (${selectedIds.size} selected)`}
        </span>
        {selectedIds.size > 0 && (
          <Button
            variant="danger"
            size="sm"
            onClick={handleBulkDelete}
            loading={deleting}
          >
            Remove {selectedIds.size} Selected
          </Button>
        )}
        {/* Search input */}
        <input
          type="text"
          placeholder="Search repositories..."
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          style={{ ...inputStyle, width: '180px' }}
        />
        {/* Ecosystem filter */}
        <select
          value={ecosystemFilter}
          onChange={(e) => setEcosystemFilter(e.target.value as EcosystemFilter)}
          style={selectStyle}
        >
          <option value="">All Ecosystems</option>
          <option value="npm">npm</option>
          <option value="go">Go</option>
          <option value="maven">Maven</option>
          <option value="gradle">Gradle</option>
        </select>
        {/* Outdated filter */}
        <select
          value={outdatedFilter}
          onChange={(e) => setOutdatedFilter(e.target.value as OutdatedFilter)}
          style={selectStyle}
        >
          <option value="">All Status</option>
          <option value="outdated">Has Outdated</option>
          <option value="uptodate">Up to Date</option>
        </select>
        {/* View Mode Toggle */}
        <div style={{ display: 'flex', borderRadius: '6px', overflow: 'hidden', border: '1px solid var(--border-color)' }}>
          <button
            onClick={() => {
              setViewMode('grouped');
              // Clear source filter when switching to grouped view
              // since grouped view shows all sources by design
              setSelectedSource(undefined);
            }}
            style={{
              padding: '6px 12px',
              fontSize: '12px',
              fontWeight: 500,
              border: 'none',
              cursor: 'pointer',
              backgroundColor: viewMode === 'grouped' ? 'var(--accent)' : 'var(--bg-card)',
              color: viewMode === 'grouped' ? 'white' : 'var(--text-secondary)',
            }}
          >
            Grouped
          </button>
          <button
            onClick={() => setViewMode('list')}
            style={{
              padding: '6px 12px',
              fontSize: '12px',
              fontWeight: 500,
              border: 'none',
              borderLeft: '1px solid var(--border-color)',
              cursor: 'pointer',
              backgroundColor: viewMode === 'list' ? 'var(--accent)' : 'var(--bg-card)',
              color: viewMode === 'list' ? 'white' : 'var(--text-secondary)',
            }}
          >
            List
          </button>
        </div>
        {viewMode === 'list' && (
          <select
            value={selectedSource || ''}
            onChange={handleSourceChange}
            aria-label="Filter by source"
            style={selectStyle}
          >
            <option value="">All Sources</option>
            {sources.map((source) => (
              <option key={source.id} value={source.id}>
                {source.name}
              </option>
            ))}
          </select>
        )}
      </div>

      {error && <ErrorMessage message={error} onDismiss={() => setError(null)} />}

      {loading ? (
        <LoadingSpinner fullPage text="Loading..." />
      ) : repositories.length === 0 ? (
        <Card>
          <EmptyState
            icon="📁"
            description="No repositories found. Add a source and run a scan."
          />
        </Card>
      ) : viewMode === 'grouped' ? (
        /* Grouped View */
        <div style={{ display: 'flex', flexDirection: 'column', gap: '16px' }}>
          {groupedRepositories.map(({ source, repos, totalRepos }) => (
            <Card key={source.id}>
              {/* Group Header */}
              <div
                onClick={() => toggleGroup(source.id)}
                style={{
                  display: 'flex',
                  alignItems: 'center',
                  gap: '12px',
                  padding: '12px 16px',
                  cursor: 'pointer',
                  borderBottom: collapsedGroups.has(source.id) || repos.length === 0 ? 'none' : '1px solid var(--border-color)',
                  backgroundColor: 'var(--bg-secondary)',
                  borderRadius: collapsedGroups.has(source.id) || repos.length === 0 ? '8px' : '8px 8px 0 0',
                }}
              >
                <span style={{ fontSize: '18px' }}>
                  {source.type === 'gitlab' ? '🦊' : '🐙'}
                </span>
                <div style={{ flex: 1 }}>
                  <span style={{ fontWeight: 600, color: 'var(--text-primary)' }}>
                    {source.name}
                  </span>
                  <span style={{ marginLeft: '8px', fontSize: '13px', color: 'var(--text-muted)' }}>
                    {repos.length === totalRepos
                      ? `${repos.length} repositories`
                      : `${repos.length} of ${totalRepos} repositories`}
                  </span>
                  {repos.reduce((sum, r) => sum + r.outdated_count, 0) > 0 && (
                    <span style={{
                      marginLeft: '8px',
                      padding: '2px 6px',
                      borderRadius: '4px',
                      fontSize: '11px',
                      fontWeight: 600,
                      backgroundColor: 'var(--danger-bg)',
                      color: 'var(--danger)',
                    }}>
                      {repos.reduce((sum, r) => sum + r.outdated_count, 0)} outdated
                    </span>
                  )}
                </div>
                <span style={{ color: 'var(--text-muted)', fontSize: '12px' }}>
                  {collapsedGroups.has(source.id) ? '▶' : '▼'}
                </span>
              </div>

              {/* Group Content */}
              {!collapsedGroups.has(source.id) && repos.length === 0 && (
                <div style={{ padding: '24px', textAlign: 'center', color: 'var(--text-muted)', fontSize: '14px' }}>
                  No repositories matching the current filter
                </div>
              )}
              {!collapsedGroups.has(source.id) && repos.length > 0 && (
                <Table fixed>
                  <TableHead>
                    <Th width="4%" noEllipsis>
                      <input
                        type="checkbox"
                        checked={repos.every(r => selectedIds.has(r.id))}
                        onChange={(e) => {
                          if (e.target.checked) {
                            setSelectedIds(prev => new Set([...prev, ...repos.map(r => r.id)]));
                          } else {
                            setSelectedIds(prev => {
                              const next = new Set(prev);
                              repos.forEach(r => next.delete(r.id));
                              return next;
                            });
                          }
                        }}
                        style={{ cursor: 'pointer', width: '16px', height: '16px' }}
                      />
                    </Th>
                    <Th width="4%" noEllipsis>#</Th>
                    <Th width="36%">Repository</Th>
                    <Th width="12%">Branch</Th>
                    <Th width="12%">Manifest</Th>
                    <Th width="18%">Dependencies</Th>
                    <Th width="14%">Last Scanned</Th>
                  </TableHead>
                  <TableBody>
                    {repos.map((repo, index) => (
                      <TableRow key={repo.id} style={selectedIds.has(repo.id) ? { backgroundColor: 'var(--accent-light)' } : undefined}>
                        <Td noEllipsis>
                          <input
                            type="checkbox"
                            checked={selectedIds.has(repo.id)}
                            onChange={(e) => handleSelectOne(repo.id, e.target.checked)}
                            style={{ cursor: 'pointer', width: '16px', height: '16px' }}
                          />
                        </Td>
                        <Td noEllipsis muted style={{ fontSize: '12px' }}>{index + 1}</Td>
                        <Td>
                          <Link
                            to={`/dependencies?repo=${encodeURIComponent(repo.full_name)}`}
                            style={{ color: 'var(--accent)', textDecoration: 'none', fontWeight: 500 }}
                          >
                            {repo.full_name}
                          </Link>
                        </Td>
                        <Td noEllipsis>
                          <span style={{
                            padding: '4px 8px',
                            borderRadius: '4px',
                            fontSize: '12px',
                            fontWeight: 600,
                            fontFamily: "'JetBrains Mono', 'Fira Code', monospace",
                            backgroundColor: 'var(--bg-primary)',
                            color: 'var(--text-primary)',
                            border: '1px solid var(--border-color)',
                          }}>
                            {repo.default_branch}
                          </span>
                        </Td>
                        <Td noEllipsis>
                          <div style={{ display: 'flex', gap: '4px', flexWrap: 'wrap' }}>
                            {repo.has_package_json && <EcosystemBadge ecosystem="npm" />}
                            {repo.has_go_mod && <EcosystemBadge ecosystem="go" />}
                            {repo.has_pom_xml && <EcosystemBadge ecosystem="maven" />}
                            {repo.has_build_gradle && <EcosystemBadge ecosystem="gradle" />}
                            {!repo.has_package_json && !repo.has_go_mod && !repo.has_pom_xml && !repo.has_build_gradle && (
                              <span style={{ color: 'var(--text-muted)', fontSize: '12px' }}>-</span>
                            )}
                          </div>
                        </Td>
                        <Td noEllipsis>
                          {repo.dependency_count > 0 ? (
                            <div style={{ display: 'flex', alignItems: 'center', gap: '6px' }}>
                              <span style={{ fontWeight: 500 }}>{repo.dependency_count}</span>
                              {repo.outdated_count > 0 && (
                                <span style={{
                                  padding: '2px 6px',
                                  borderRadius: '4px',
                                  fontSize: '11px',
                                  fontWeight: 600,
                                  backgroundColor: 'var(--danger-bg)',
                                  color: 'var(--danger)',
                                }}>
                                  {repo.outdated_count} outdated
                                </span>
                              )}
                            </div>
                          ) : (
                            <span style={{ color: 'var(--text-muted)', fontSize: '12px' }}>-</span>
                          )}
                        </Td>
                        <Td muted style={{ fontSize: '13px' }}>
                          {repo.last_scan_at
                            ? new Date(repo.last_scan_at).toLocaleString()
                            : '-'}
                        </Td>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              )}
            </Card>
          ))}
        </div>
      ) : (
        /* List View */
        <Card>
          <Table fixed>
            <TableHead>
              <Th width="4%" noEllipsis>
                <input
                  type="checkbox"
                  checked={selectedIds.size === filteredRepositories.length && filteredRepositories.length > 0}
                  onChange={(e) => {
                    if (e.target.checked) {
                      setSelectedIds(new Set(filteredRepositories.map(r => r.id)));
                    } else {
                      setSelectedIds(new Set());
                    }
                  }}
                  style={{ cursor: 'pointer', width: '16px', height: '16px' }}
                />
              </Th>
              <Th width="4%" noEllipsis>#</Th>
              <Th width="36%">Repository</Th>
              <Th width="12%">Branch</Th>
              <Th width="12%">Manifest</Th>
              <Th width="18%">Dependencies</Th>
              <Th width="14%">Last Scanned</Th>
            </TableHead>
            <TableBody>
              {filteredRepositories.map((repo, index) => (
                <TableRow key={repo.id} style={selectedIds.has(repo.id) ? { backgroundColor: 'var(--accent-light)' } : undefined}>
                  <Td noEllipsis>
                    <input
                      type="checkbox"
                      checked={selectedIds.has(repo.id)}
                      onChange={(e) => handleSelectOne(repo.id, e.target.checked)}
                      style={{ cursor: 'pointer', width: '16px', height: '16px' }}
                    />
                  </Td>
                  <Td noEllipsis muted style={{ fontSize: '12px' }}>{index + 1}</Td>
                  <Td>
                    <Link
                      to={`/dependencies?repo=${encodeURIComponent(repo.full_name)}`}
                      style={{ color: 'var(--accent)', textDecoration: 'none', fontWeight: 500 }}
                    >
                      {repo.full_name}
                    </Link>
                  </Td>
                  <Td noEllipsis>
                    <span style={{
                      padding: '5px 10px',
                      borderRadius: '6px',
                      fontSize: '13px',
                      fontWeight: 600,
                      fontFamily: "'JetBrains Mono', 'Fira Code', monospace",
                      backgroundColor: 'var(--bg-secondary)',
                      color: 'var(--text-primary)',
                      border: '1px solid var(--border-color)',
                    }}>
                      {repo.default_branch}
                    </span>
                  </Td>
                  <Td noEllipsis>
                    <div style={{ display: 'flex', gap: '4px', flexWrap: 'wrap' }}>
                      {repo.has_package_json && <EcosystemBadge ecosystem="npm" />}
                      {repo.has_go_mod && <EcosystemBadge ecosystem="go" />}
                      {repo.has_pom_xml && <EcosystemBadge ecosystem="maven" />}
                      {repo.has_build_gradle && <EcosystemBadge ecosystem="gradle" />}
                      {!repo.has_package_json && !repo.has_go_mod && !repo.has_pom_xml && !repo.has_build_gradle && (
                        <span style={{ color: 'var(--text-muted)', fontSize: '12px' }}>-</span>
                      )}
                    </div>
                  </Td>
                  <Td noEllipsis>
                    {repo.dependency_count > 0 ? (
                      <div style={{ display: 'flex', alignItems: 'center', gap: '6px' }}>
                        <span style={{ fontWeight: 500 }}>{repo.dependency_count}</span>
                        {repo.outdated_count > 0 && (
                          <span style={{
                            padding: '2px 6px',
                            borderRadius: '4px',
                            fontSize: '11px',
                            fontWeight: 600,
                            backgroundColor: 'var(--danger-bg)',
                            color: 'var(--danger)',
                          }}>
                            {repo.outdated_count} outdated
                          </span>
                        )}
                      </div>
                    ) : (
                      <span style={{ color: 'var(--text-muted)', fontSize: '12px' }}>-</span>
                    )}
                  </Td>
                  <Td muted style={{ fontSize: '13px' }}>
                    {repo.last_scan_at
                      ? new Date(repo.last_scan_at).toLocaleString()
                      : '-'}
                  </Td>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </Card>
      )}
    </div>
  );
}
