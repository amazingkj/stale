import { useEffect, useState } from 'react';
import { api } from '../api/client';
import type { Dependency } from '../types';

export function Dependencies() {
  const [dependencies, setDependencies] = useState<Dependency[]>([]);
  const [filter, setFilter] = useState<'all' | 'outdated'>('all');
  const [search, setSearch] = useState('');
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    loadDependencies();
  }, [filter]);

  async function loadDependencies() {
    setLoading(true);
    try {
      const data = filter === 'outdated'
        ? await api.getOutdatedDependencies()
        : await api.getDependencies();
      setDependencies(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load dependencies');
    } finally {
      setLoading(false);
    }
  }

  const filteredDeps = dependencies.filter((dep) =>
    dep.name.toLowerCase().includes(search.toLowerCase()) ||
    dep.repo_full_name?.toLowerCase().includes(search.toLowerCase())
  );

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: '24px' }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', flexWrap: 'wrap', gap: '16px' }}>
        <div>
          <h1 style={{ fontSize: '24px', fontWeight: 700, color: 'var(--text-primary)', margin: 0 }}>
            Dependencies
          </h1>
          <p style={{ fontSize: '14px', color: 'var(--text-secondary)', margin: '4px 0 0' }}>
            {filteredDeps.length} dependencies {filter === 'outdated' && '(outdated only)'}
          </p>
        </div>
        <div style={{ display: 'flex', gap: '12px' }}>
          <input
            type="text"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder="Search packages..."
            style={{
              padding: '10px 12px',
              borderRadius: '8px',
              border: '1px solid var(--border-color)',
              backgroundColor: 'var(--bg-card)',
              color: 'var(--text-primary)',
              fontSize: '14px',
              outline: 'none',
              width: '200px',
            }}
          />
          <select
            value={filter}
            onChange={(e) => setFilter(e.target.value as 'all' | 'outdated')}
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
            <option value="all">All</option>
            <option value="outdated">Outdated Only</option>
          </select>
        </div>
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
      ) : filteredDeps.length === 0 ? (
        <div style={{
          backgroundColor: 'var(--bg-card)',
          borderRadius: '12px',
          border: '1px solid var(--border-color)',
          padding: '48px',
          textAlign: 'center',
        }}>
          <div style={{ fontSize: '48px', marginBottom: '16px' }}>📦</div>
          <p style={{ color: 'var(--text-muted)' }}>
            {search ? 'No dependencies matching your search.' : 'No dependencies found.'}
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
                  <th style={thStyle}>Package</th>
                  <th style={thStyle}>Repository</th>
                  <th style={thStyle}>Type</th>
                  <th style={thStyle}>Current</th>
                  <th style={thStyle}>Latest</th>
                </tr>
              </thead>
              <tbody>
                {filteredDeps.map((dep) => (
                  <tr key={dep.id} style={{ borderBottom: '1px solid var(--border-color)' }}>
                    <td style={tdStyle}>
                      <a
                        href={`https://www.npmjs.com/package/${dep.name}`}
                        target="_blank"
                        rel="noopener noreferrer"
                        style={{ color: 'var(--accent)', textDecoration: 'none', fontWeight: 500 }}
                      >
                        {dep.name}
                      </a>
                    </td>
                    <td style={{ ...tdStyle, color: 'var(--text-secondary)' }}>
                      {dep.repo_full_name}
                    </td>
                    <td style={tdStyle}>
                      <span style={{
                        padding: '4px 8px',
                        borderRadius: '6px',
                        fontSize: '12px',
                        backgroundColor: dep.type === 'dependency' ? 'var(--warning-bg)' : 'var(--bg-hover)',
                        color: dep.type === 'dependency' ? 'var(--warning-text)' : 'var(--text-muted)',
                      }}>
                        {dep.type === 'dependency' ? 'prod' : 'dev'}
                      </span>
                    </td>
                    <td style={tdStyle}>
                      <span style={{
                        padding: '4px 8px',
                        borderRadius: '6px',
                        fontSize: '13px',
                        fontFamily: 'monospace',
                        backgroundColor: dep.is_outdated ? 'var(--danger-bg)' : 'var(--success-bg)',
                        color: dep.is_outdated ? 'var(--danger-text)' : 'var(--success-text)',
                      }}>
                        {dep.current_version}
                      </span>
                    </td>
                    <td style={tdStyle}>
                      <span style={{
                        padding: '4px 8px',
                        borderRadius: '6px',
                        fontSize: '13px',
                        fontFamily: 'monospace',
                        backgroundColor: 'var(--success-bg)',
                        color: 'var(--success-text)',
                      }}>
                        {dep.latest_version}
                      </span>
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
