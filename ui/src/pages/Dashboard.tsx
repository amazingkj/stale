import { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { api } from '../api/client';
import type { DependencyStats, Dependency, ScanJob } from '../types';

export function Dashboard() {
  const [stats, setStats] = useState<DependencyStats | null>(null);
  const [outdated, setOutdated] = useState<Dependency[]>([]);
  const [scanning, setScanning] = useState(false);
  const [currentScan, setCurrentScan] = useState<ScanJob | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    loadData();
  }, []);

  useEffect(() => {
    if (!currentScan || currentScan.status === 'completed' || currentScan.status === 'failed') {
      return;
    }

    const interval = setInterval(async () => {
      try {
        const scan = await api.getScan(currentScan.id);
        setCurrentScan(scan);
        if (scan.status === 'completed' || scan.status === 'failed') {
          setScanning(false);
          loadData();
        }
      } catch {
        // Ignore polling errors
      }
    }, 2000);

    return () => clearInterval(interval);
  }, [currentScan]);

  async function loadData() {
    try {
      const [statsData, outdatedData] = await Promise.all([
        api.getDependencyStats(),
        api.getOutdatedDependencies(),
      ]);
      setStats(statsData);
      setOutdated(outdatedData);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load data');
    }
  }

  async function handleScan() {
    setScanning(true);
    setError(null);
    try {
      const scan = await api.triggerScan();
      setCurrentScan(scan);
    } catch (err) {
      setScanning(false);
      setError(err instanceof Error ? err.message : 'Failed to start scan');
    }
  }

  const outdatedPercent = stats && stats.total_dependencies > 0
    ? Math.round((stats.outdated_count / stats.total_dependencies) * 100)
    : 0;

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: '24px' }}>
      {/* Header */}
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <div>
          <h1 style={{ fontSize: '24px', fontWeight: 700, color: 'var(--text-primary)', margin: 0 }}>
            Dashboard
          </h1>
          <p style={{ fontSize: '14px', color: 'var(--text-secondary)', margin: '4px 0 0' }}>
            Overview of dependency versions across repositories
          </p>
        </div>
        <button
          onClick={handleScan}
          disabled={scanning}
          style={{
            padding: '10px 20px',
            borderRadius: '8px',
            border: 'none',
            backgroundColor: 'var(--accent)',
            color: 'white',
            fontSize: '14px',
            fontWeight: 500,
            cursor: scanning ? 'not-allowed' : 'pointer',
            opacity: scanning ? 0.7 : 1,
            display: 'flex',
            alignItems: 'center',
            gap: '8px',
          }}
        >
          {scanning && <Spinner />}
          {scanning ? 'Scanning...' : 'Scan All'}
        </button>
      </div>

      {/* Error */}
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

      {/* Scan Progress */}
      {currentScan && currentScan.status === 'running' && (
        <div style={{
          padding: '12px 16px',
          borderRadius: '8px',
          backgroundColor: 'var(--accent)',
          color: 'white',
          fontSize: '14px',
          display: 'flex',
          alignItems: 'center',
          gap: '8px',
        }}>
          <Spinner />
          Scanning... Found {currentScan.repos_found} repos, {currentScan.deps_found} dependencies
        </div>
      )}

      {/* Stats Grid */}
      <div style={{
        display: 'grid',
        gridTemplateColumns: 'repeat(auto-fit, minmax(200px, 1fr))',
        gap: '16px',
      }}>
        <StatCard
          label="Total Dependencies"
          value={stats?.total_dependencies ?? 0}
          color="accent"
        />
        <StatCard
          label="Outdated"
          value={stats?.outdated_count ?? 0}
          subtitle={stats?.total_dependencies ? `${outdatedPercent}% of total` : undefined}
          color="danger"
        />
        <StatCard
          label="Up to Date"
          value={stats?.up_to_date_count ?? 0}
          color="success"
        />
        <StatCard
          label="Production"
          value={stats?.by_type?.dependency ?? 0}
          color="warning"
        />
      </div>

      {/* Outdated Table */}
      <div style={{
        backgroundColor: 'var(--bg-card)',
        borderRadius: '12px',
        border: '1px solid var(--border-color)',
        overflow: 'hidden',
      }}>
        <div style={{
          padding: '16px 20px',
          borderBottom: '1px solid var(--border-color)',
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
        }}>
          <h2 style={{ fontSize: '16px', fontWeight: 600, color: 'var(--text-primary)', margin: 0 }}>
            Outdated Dependencies
          </h2>
          <Link
            to="/dependencies"
            style={{
              fontSize: '13px',
              color: 'var(--accent)',
              textDecoration: 'none',
            }}
          >
            View all →
          </Link>
        </div>

        {outdated.length === 0 ? (
          <div style={{
            padding: '48px 20px',
            textAlign: 'center',
            color: 'var(--text-muted)',
          }}>
            {stats?.total_dependencies === 0
              ? 'No dependencies found. Add a source and scan to get started.'
              : 'All dependencies are up to date!'}
          </div>
        ) : (
          <div style={{ overflowX: 'auto' }}>
            <table style={{ width: '100%', borderCollapse: 'collapse' }}>
              <thead>
                <tr style={{ backgroundColor: 'var(--bg-secondary)' }}>
                  <th style={thStyle}>Package</th>
                  <th style={thStyle}>Repository</th>
                  <th style={thStyle}>Current</th>
                  <th style={thStyle}>Latest</th>
                </tr>
              </thead>
              <tbody>
                {outdated.slice(0, 10).map((dep) => (
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
                      <span style={{
                        marginLeft: '8px',
                        padding: '2px 6px',
                        borderRadius: '4px',
                        fontSize: '11px',
                        backgroundColor: dep.type === 'dependency' ? 'var(--warning-bg)' : 'var(--bg-hover)',
                        color: dep.type === 'dependency' ? 'var(--warning-text)' : 'var(--text-muted)',
                      }}>
                        {dep.type === 'dependency' ? 'prod' : 'dev'}
                      </span>
                    </td>
                    <td style={{ ...tdStyle, color: 'var(--text-secondary)' }}>
                      {dep.repo_full_name}
                    </td>
                    <td style={tdStyle}>
                      <span style={{
                        padding: '4px 8px',
                        borderRadius: '6px',
                        fontSize: '13px',
                        fontFamily: 'monospace',
                        backgroundColor: 'var(--danger-bg)',
                        color: 'var(--danger-text)',
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
        )}
      </div>
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

function StatCard({ label, value, subtitle, color }: {
  label: string;
  value: number;
  subtitle?: string;
  color: 'accent' | 'success' | 'warning' | 'danger';
}) {
  const colors = {
    accent: { bg: 'var(--bg-hover)', text: 'var(--accent)' },
    success: { bg: 'var(--success-bg)', text: 'var(--success-text)' },
    warning: { bg: 'var(--warning-bg)', text: 'var(--warning-text)' },
    danger: { bg: 'var(--danger-bg)', text: 'var(--danger-text)' },
  };

  return (
    <div style={{
      padding: '20px',
      borderRadius: '12px',
      backgroundColor: colors[color].bg,
    }}>
      <div style={{ fontSize: '13px', color: 'var(--text-secondary)', marginBottom: '8px' }}>
        {label}
      </div>
      <div style={{ fontSize: '32px', fontWeight: 700, color: colors[color].text }}>
        {value.toLocaleString()}
      </div>
      {subtitle && (
        <div style={{ fontSize: '12px', color: 'var(--text-muted)', marginTop: '4px' }}>
          {subtitle}
        </div>
      )}
    </div>
  );
}

function Spinner() {
  return (
    <svg width="16" height="16" viewBox="0 0 24 24" style={{ animation: 'spin 1s linear infinite' }}>
      <style>{`@keyframes spin { from { transform: rotate(0deg); } to { transform: rotate(360deg); } }`}</style>
      <circle cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="3" fill="none" opacity="0.25" />
      <path d="M12 2a10 10 0 0 1 10 10" stroke="currentColor" strokeWidth="3" fill="none" strokeLinecap="round" />
    </svg>
  );
}
