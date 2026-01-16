import { useEffect, useState, useCallback } from 'react';
import { api } from '../api/client';
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
  EcosystemBadge,
  EmptyState,
  LoadingSpinner,
  ErrorMessage,
} from '../components/common';
import type { Repository, Source } from '../types';

export function Repositories() {
  const [repositories, setRepositories] = useState<Repository[]>([]);
  const [sources, setSources] = useState<Source[]>([]);
  const [selectedSource, setSelectedSource] = useState<number | undefined>();
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [selectedIds, setSelectedIds] = useState<Set<number>>(new Set());
  const [deleting, setDeleting] = useState(false);

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

  const handleRemove = useCallback(async (id: number, name: string) => {
    if (!confirm(`Remove "${name}" from dashboard?\n(This will not delete the actual repository)`)) return;
    try {
      await api.deleteRepository(id);
      setRepositories(prev => prev.filter(r => r.id !== id));
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to remove repository');
    }
  }, []);

  const handleSourceChange = useCallback((e: React.ChangeEvent<HTMLSelectElement>) => {
    setSelectedSource(e.target.value ? Number(e.target.value) : undefined);
    setSelectedIds(new Set()); // Clear selection when source changes
  }, []);

  const handleSelectAll = useCallback((checked: boolean) => {
    if (checked) {
      setSelectedIds(new Set(repositories.map(r => r.id)));
    } else {
      setSelectedIds(new Set());
    }
  }, [repositories]);

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

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: '24px' }}>
      <div style={{ display: 'flex', justifyContent: 'flex-end', alignItems: 'center', gap: '12px' }}>
        <span style={{ fontSize: '14px', color: 'var(--text-secondary)', marginRight: 'auto' }}>
          {repositories.length} repositories scanned
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
      ) : (
        <Card>
          <Table fixed>
            <TableHead>
              <Th width="5%">
                <input
                  type="checkbox"
                  checked={selectedIds.size === repositories.length && repositories.length > 0}
                  onChange={(e) => handleSelectAll(e.target.checked)}
                  style={{ cursor: 'pointer', width: '16px', height: '16px' }}
                />
              </Th>
              <Th width="33%">Repository</Th>
              <Th width="12%">Branch</Th>
              <Th width="20%">Manifest</Th>
              <Th width="22%">Last Scanned</Th>
              <Th width="8%"></Th>
            </TableHead>
            <TableBody>
              {repositories.map((repo) => (
                <TableRow key={repo.id}>
                  <Td>
                    <input
                      type="checkbox"
                      checked={selectedIds.has(repo.id)}
                      onChange={(e) => handleSelectOne(repo.id, e.target.checked)}
                      style={{ cursor: 'pointer', width: '16px', height: '16px' }}
                    />
                  </Td>
                  <Td>
                    <a
                      href={repo.html_url}
                      target="_blank"
                      rel="noopener noreferrer"
                      style={{ color: 'var(--accent)', textDecoration: 'none', fontWeight: 500 }}
                    >
                      {repo.full_name}
                    </a>
                  </Td>
                  <Td>
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
                  <Td>
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
                  <Td muted style={{ fontSize: '13px' }}>
                    {repo.last_scan_at
                      ? new Date(repo.last_scan_at).toLocaleString()
                      : '-'}
                  </Td>
                  <Td>
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={() => handleRemove(repo.id, repo.full_name)}
                      aria-label={`Remove ${repo.full_name} from dashboard`}
                      title="Remove from dashboard"
                    >
                      ✕
                    </Button>
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
