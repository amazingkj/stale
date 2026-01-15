import { useEffect, useState } from 'react';
import { api } from '../api/client';
import type { Repository, Source } from '../types';

export function Repositories() {
  const [repositories, setRepositories] = useState<Repository[]>([]);
  const [sources, setSources] = useState<Source[]>([]);
  const [selectedSource, setSelectedSource] = useState<number | undefined>();
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    loadSources();
  }, []);

  useEffect(() => {
    loadRepositories();
  }, [selectedSource]);

  async function loadSources() {
    try {
      const data = await api.getSources();
      setSources(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load sources');
    }
  }

  async function loadRepositories() {
    setLoading(true);
    try {
      const data = await api.getRepositories(selectedSource);
      setRepositories(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load repositories');
    } finally {
      setLoading(false);
    }
  }

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: '24px' }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <div>
          <h1 style={{ fontSize: '24px', fontWeight: 700, color: 'var(--text-primary)', margin: 0 }}>
            Repositories
          </h1>
          <p style={{ fontSize: '14px', color: 'var(--text-secondary)', margin: '4px 0 0' }}>
            {repositories.length} repositories scanned
          </p>
        </div>
        <select
          value={selectedSource || ''}
          onChange={(e) => setSelectedSource(e.target.value ? Number(e.target.value) : undefined)}
          style={{
            padding: '10px 12px',
            borderRadius: '8px',
            border: '1px solid var(--border-color)',
            backgroundColor: 'var(--bg-card)',
            color: 'var(--text-primary)',
            fontSize: '14px',
            outline: 'none',
            cursor: 'pointer',
          }}
        >
          <option value="">All Sources</option>
          {sources.map((source) => (
            <option key={source.id} value={source.id}>
              {source.name}
            </option>
          ))}
        </select>
      </div>

      {error && (
        <div style={{
          padding: '12px 16px',
          borderRadius: '8px',
          backgroundColor: 'var(--danger-bg)',
          color: 'var(--danger-text)',
          fontSize: '14px',
        }}>
          {error}
        </div>
      )}

      {loading ? (
        <div style={{ textAlign: 'center', padding: '48px', color: 'var(--text-muted)' }}>
          Loading...
        </div>
      ) : repositories.length === 0 ? (
        <div style={{
          backgroundColor: 'var(--bg-card)',
          borderRadius: '12px',
          border: '1px solid var(--border-color)',
          padding: '48px',
          textAlign: 'center',
        }}>
          <div style={{ fontSize: '48px', marginBottom: '16px' }}>📁</div>
          <p style={{ color: 'var(--text-muted)' }}>
            No repositories found. Add a source and run a scan.
          </p>
        </div>
      ) : (
        <div style={{
          backgroundColor: 'var(--bg-card)',
          borderRadius: '12px',
          border: '1px solid var(--border-color)',
          overflow: 'hidden',
        }}>
          <div style={{ overflowX: 'auto' }}>
            <table style={{ width: '100%', borderCollapse: 'collapse' }}>
              <thead>
                <tr style={{ backgroundColor: 'var(--bg-secondary)' }}>
                  <th style={thStyle}>Repository</th>
                  <th style={thStyle}>Branch</th>
                  <th style={thStyle}>package.json</th>
                  <th style={thStyle}>Last Scanned</th>
                </tr>
              </thead>
              <tbody>
                {repositories.map((repo) => (
                  <tr key={repo.id} style={{ borderBottom: '1px solid var(--border-color)' }}>
                    <td style={tdStyle}>
                      <a
                        href={repo.html_url}
                        target="_blank"
                        rel="noopener noreferrer"
                        style={{ color: 'var(--accent)', textDecoration: 'none', fontWeight: 500 }}
                      >
                        {repo.full_name}
                      </a>
                    </td>
                    <td style={{ ...tdStyle, color: 'var(--text-secondary)' }}>
                      <span style={{
                        padding: '4px 8px',
                        borderRadius: '6px',
                        fontSize: '13px',
                        fontFamily: 'monospace',
                        backgroundColor: 'var(--bg-hover)',
                      }}>
                        {repo.default_branch}
                      </span>
                    </td>
                    <td style={tdStyle}>
                      {repo.has_package_json ? (
                        <span style={{
                          padding: '4px 8px',
                          borderRadius: '6px',
                          fontSize: '12px',
                          backgroundColor: 'var(--success-bg)',
                          color: 'var(--success-text)',
                        }}>
                          Yes
                        </span>
                      ) : (
                        <span style={{
                          padding: '4px 8px',
                          borderRadius: '6px',
                          fontSize: '12px',
                          backgroundColor: 'var(--bg-hover)',
                          color: 'var(--text-muted)',
                        }}>
                          No
                        </span>
                      )}
                    </td>
                    <td style={{ ...tdStyle, color: 'var(--text-muted)', fontSize: '13px' }}>
                      {repo.last_scan_at
                        ? new Date(repo.last_scan_at).toLocaleString()
                        : '-'}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}
    </div>
  );
}

const thStyle: React.CSSProperties = {
  padding: '12px 20px',
  textAlign: 'left',
  fontSize: '12px',
  fontWeight: 600,
  color: 'var(--text-muted)',
  textTransform: 'uppercase',
  letterSpacing: '0.5px',
};

const tdStyle: React.CSSProperties = {
  padding: '14px 20px',
  fontSize: '14px',
  color: 'var(--text-primary)',
};
