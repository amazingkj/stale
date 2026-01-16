import { useEffect, useState } from 'react';
import { api } from '../../api/client';
import type { Source, SourceInput, Settings } from '../../types';

interface Props {
  isOpen: boolean;
  onClose: () => void;
}

type SettingsTab = 'sources' | 'schedule' | 'email';

export function SettingsPanel({ isOpen, onClose }: Props) {
  const [activeTab, setActiveTab] = useState<SettingsTab>('sources');
  const [sources, setSources] = useState<Source[]>([]);
  const [isAddingSource, setIsAddingSource] = useState(false);
  const [editingSource, setEditingSource] = useState<Source | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [settings, setSettings] = useState<Settings | null>(null);

  useEffect(() => {
    if (isOpen) {
      loadSources();
      loadSettings();
    }
  }, [isOpen]);

  async function loadSettings() {
    try {
      const data = await api.getSettings();
      setSettings(data);
    } catch (err) {
      console.error('Failed to load settings:', err);
    }
  }

  async function loadSources() {
    setLoading(true);
    try {
      const data = await api.getSources();
      setSources(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load sources');
    } finally {
      setLoading(false);
    }
  }

  async function handleDelete(id: number) {
    if (!confirm('Are you sure? All associated data will be removed.')) return;
    try {
      await api.deleteSource(id);
      setSources(sources.filter((s) => s.id !== id));
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to delete source');
    }
  }

  async function handleCreate(input: SourceInput) {
    const source = await api.createSource(input);
    setSources([source, ...sources]);
    setIsAddingSource(false);
  }

  async function handleUpdate(id: number, input: SourceInput) {
    const updated = await api.updateSource(id, input);
    setSources(sources.map((s) => (s.id === id ? updated : s)));
    setEditingSource(null);
  }

  async function handleSettingsUpdate(updates: Partial<Settings>) {
    try {
      const updated = await api.updateSettings(updates);
      setSettings(updated);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to update settings');
    }
  }

  async function handleTestEmail() {
    try {
      const result = await api.testEmail();
      alert(result.message);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to send test email');
    }
  }

  if (!isOpen) return null;

  return (
    <>
      {/* Backdrop */}
      <div
        style={{
          position: 'fixed',
          inset: 0,
          backgroundColor: 'rgba(0, 0, 0, 0.5)',
          zIndex: 100,
          transition: 'opacity 0.2s',
        }}
        onClick={onClose}
      />

      {/* Panel */}
      <div
        style={{
          position: 'fixed',
          top: 0,
          right: 0,
          bottom: 0,
          width: '100%',
          maxWidth: '420px',
          backgroundColor: 'var(--bg-card)',
          boxShadow: '-4px 0 20px rgba(0, 0, 0, 0.15)',
          zIndex: 101,
          display: 'flex',
          flexDirection: 'column',
          animation: 'slideIn 0.2s ease-out',
        }}
      >
        <style>{`
          @keyframes slideIn {
            from { transform: translateX(100%); }
            to { transform: translateX(0); }
          }
        `}</style>

        {/* Header */}
        <div style={{
          padding: '20px 24px',
          borderBottom: '1px solid var(--border-color)',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
        }}>
          <h2 style={{ fontSize: '18px', fontWeight: 600, color: 'var(--text-primary)', margin: 0 }}>
            Settings
          </h2>
          <button
            onClick={onClose}
            style={{
              padding: '8px',
              borderRadius: '8px',
              border: 'none',
              backgroundColor: 'var(--bg-hover)',
              color: 'var(--text-secondary)',
              cursor: 'pointer',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
            }}
          >
            <CloseIcon />
          </button>
        </div>

        {/* Tabs */}
        <div style={{ display: 'flex', borderBottom: '1px solid var(--border-color)' }}>
          {(['sources', 'schedule', 'email'] as const).map((tab) => (
            <button
              key={tab}
              onClick={() => setActiveTab(tab)}
              style={{
                flex: 1,
                padding: '12px',
                border: 'none',
                backgroundColor: 'transparent',
                color: activeTab === tab ? 'var(--accent)' : 'var(--text-secondary)',
                fontSize: '13px',
                fontWeight: 500,
                cursor: 'pointer',
                borderBottom: activeTab === tab ? '2px solid var(--accent)' : '2px solid transparent',
                marginBottom: '-1px',
              }}
            >
              {tab === 'sources' ? 'Sources' : tab === 'schedule' ? 'Schedule' : 'Email'}
            </button>
          ))}
        </div>

        {/* Content */}
        <div style={{ flex: 1, overflow: 'auto', padding: '24px' }}>
          {error && (
            <div style={{
              padding: '12px',
              borderRadius: '8px',
              backgroundColor: 'var(--danger-bg)',
              color: 'var(--danger-text)',
              fontSize: '13px',
              marginBottom: '16px',
            }}>
              {error}
              <button onClick={() => setError(null)} style={{ float: 'right', background: 'none', border: 'none', cursor: 'pointer', color: 'inherit' }}>×</button>
            </div>
          )}

          {activeTab === 'sources' && (
            <SourcesTab
              sources={sources}
              loading={loading}
              onAddSource={() => setIsAddingSource(true)}
              onEditSource={setEditingSource}
              onDeleteSource={handleDelete}
            />
          )}

          {activeTab === 'schedule' && settings && (
            <ScheduleTab settings={settings} onUpdate={handleSettingsUpdate} />
          )}

          {activeTab === 'email' && settings && (
            <EmailTab settings={settings} onUpdate={handleSettingsUpdate} onTestEmail={handleTestEmail} />
          )}
        </div>
      </div>

      {/* Add Source Modal */}
      {isAddingSource && (
        <SourceModal
          onClose={() => setIsAddingSource(false)}
          onSubmit={handleCreate}
        />
      )}

      {/* Edit Source Modal */}
      {editingSource && (
        <SourceModal
          source={editingSource}
          onClose={() => setEditingSource(null)}
          onSubmit={(input) => handleUpdate(editingSource.id, input)}
        />
      )}
    </>
  );
}

function SourceModal({ source, onClose, onSubmit }: {
  source?: Source;
  onClose: () => void;
  onSubmit: (input: SourceInput) => Promise<void>;
}) {
  const isEditing = !!source;
  const [sourceType, setSourceType] = useState<'github' | 'gitlab'>(source?.type || 'github');
  const [name, setName] = useState(source?.name || '');
  const [token, setToken] = useState('');
  const [organization, setOrganization] = useState(source?.organization || '');
  const [repositories, setRepositories] = useState(source?.repositories || '');
  const [url, setUrl] = useState(source?.url || '');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setLoading(true);
    setError(null);
    try {
      await onSubmit({
        name,
        type: sourceType,
        token,
        organization: organization || undefined,
        repositories: repositories || undefined,
        url: sourceType === 'gitlab' && url ? url : undefined,
      });
    } catch (err) {
      setError(err instanceof Error ? err.message : (isEditing ? 'Failed to update source' : 'Failed to add source'));
      setLoading(false);
    }
  }

  return (
    <div style={{
      position: 'fixed',
      inset: 0,
      zIndex: 200,
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'center',
      padding: '16px',
    }}>
      <div
        style={{
          position: 'absolute',
          inset: 0,
          backgroundColor: 'rgba(0, 0, 0, 0.5)',
        }}
        onClick={onClose}
      />
      <div style={{
        position: 'relative',
        backgroundColor: 'var(--bg-card)',
        borderRadius: '16px',
        width: '100%',
        maxWidth: '400px',
        boxShadow: '0 25px 50px -12px rgba(0, 0, 0, 0.25)',
      }}>
        <div style={{
          padding: '16px 20px',
          borderBottom: '1px solid var(--border-color)',
        }}>
          <h2 style={{ fontSize: '16px', fontWeight: 600, color: 'var(--text-primary)', margin: 0 }}>
            {isEditing ? 'Edit Source' : 'Add Source'}
          </h2>
        </div>

        <form onSubmit={handleSubmit} style={{ padding: '20px' }}>
          {error && (
            <div style={{
              padding: '10px 12px',
              borderRadius: '6px',
              backgroundColor: 'var(--danger-bg)',
              color: 'var(--danger-text)',
              fontSize: '13px',
              marginBottom: '16px',
            }}>
              {error}
            </div>
          )}

          {/* Source Type Selector */}
          <div style={{ marginBottom: '14px' }}>
            <label style={{ display: 'block', fontSize: '13px', fontWeight: 500, color: 'var(--text-primary)', marginBottom: '6px' }}>
              Type
            </label>
            <div style={{ display: 'flex', gap: '8px' }}>
              <button
                type="button"
                onClick={() => setSourceType('github')}
                style={{
                  flex: 1,
                  padding: '10px',
                  borderRadius: '8px',
                  border: `2px solid ${sourceType === 'github' ? 'var(--accent)' : 'var(--border-color)'}`,
                  backgroundColor: sourceType === 'github' ? 'var(--bg-hover)' : 'transparent',
                  color: 'var(--text-primary)',
                  fontSize: '13px',
                  fontWeight: 500,
                  cursor: 'pointer',
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  gap: '6px',
                }}
              >
                🐙 GitHub
              </button>
              <button
                type="button"
                onClick={() => setSourceType('gitlab')}
                style={{
                  flex: 1,
                  padding: '10px',
                  borderRadius: '8px',
                  border: `2px solid ${sourceType === 'gitlab' ? 'var(--accent)' : 'var(--border-color)'}`,
                  backgroundColor: sourceType === 'gitlab' ? 'var(--bg-hover)' : 'transparent',
                  color: 'var(--text-primary)',
                  fontSize: '13px',
                  fontWeight: 500,
                  cursor: 'pointer',
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  gap: '6px',
                }}
              >
                🦊 GitLab
              </button>
            </div>
          </div>

          <div style={{ marginBottom: '14px' }}>
            <label style={{ display: 'block', fontSize: '13px', fontWeight: 500, color: 'var(--text-primary)', marginBottom: '6px' }}>
              Name
            </label>
            <input
              type="text"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder={sourceType === 'gitlab' ? 'My GitLab Group' : 'My GitHub Org'}
              required
              style={inputStyle}
            />
          </div>

          {sourceType === 'gitlab' && (
            <div style={{ marginBottom: '14px' }}>
              <label style={{ display: 'block', fontSize: '13px', fontWeight: 500, color: 'var(--text-primary)', marginBottom: '6px' }}>
                GitLab URL (optional)
              </label>
              <input
                type="url"
                value={url}
                onChange={(e) => setUrl(e.target.value)}
                placeholder="https://gitlab.com"
                style={inputStyle}
              />
              <p style={{ fontSize: '11px', color: 'var(--text-muted)', marginTop: '4px' }}>
                Leave empty for gitlab.com
              </p>
            </div>
          )}

          <div style={{ marginBottom: '14px' }}>
            <label style={{ display: 'block', fontSize: '13px', fontWeight: 500, color: 'var(--text-primary)', marginBottom: '6px' }}>
              Personal Access Token
            </label>
            <input
              type="password"
              value={token}
              onChange={(e) => setToken(e.target.value)}
              placeholder={isEditing ? 'Enter new token to update' : (sourceType === 'gitlab' ? 'glpat-xxxxxxxxxxxx' : 'ghp_xxxxxxxxxxxx')}
              required
              style={inputStyle}
            />
            <p style={{ fontSize: '11px', color: 'var(--text-muted)', marginTop: '4px' }}>
              {isEditing
                ? 'Token is required for security verification'
                : (sourceType === 'gitlab'
                  ? 'Requires read_api scope'
                  : 'Requires repo scope for private repos')}
            </p>
          </div>

          <div style={{ marginBottom: '14px' }}>
            <label style={{ display: 'block', fontSize: '13px', fontWeight: 500, color: 'var(--text-primary)', marginBottom: '6px' }}>
              {sourceType === 'gitlab' ? 'Group (optional)' : 'Organization (optional)'}
            </label>
            <input
              type="text"
              value={organization}
              onChange={(e) => setOrganization(e.target.value)}
              placeholder={sourceType === 'gitlab' ? 'my-group' : 'my-org'}
              style={inputStyle}
            />
            <p style={{ fontSize: '11px', color: 'var(--text-muted)', marginTop: '4px' }}>
              {sourceType === 'gitlab'
                ? 'Leave empty to scan all accessible projects'
                : 'Leave empty to scan personal repos'}
            </p>
          </div>

          <div style={{ marginBottom: '14px' }}>
            <label style={{ display: 'block', fontSize: '13px', fontWeight: 500, color: 'var(--text-primary)', marginBottom: '6px' }}>
              Repositories (optional)
            </label>
            <input
              type="text"
              value={repositories}
              onChange={(e) => setRepositories(e.target.value)}
              placeholder="repo1, owner/repo2"
              style={inputStyle}
            />
            <p style={{ fontSize: '11px', color: 'var(--text-muted)', marginTop: '4px' }}>
              Comma-separated. Leave empty to scan all repos.
            </p>
          </div>

          <div style={{
            padding: '10px',
            borderRadius: '6px',
            backgroundColor: 'var(--bg-primary)',
            border: '1px solid var(--border-color)',
            marginBottom: '20px',
          }}>
            <p style={{ fontSize: '11px', fontWeight: 600, color: 'var(--text-primary)', margin: '0 0 6px' }}>
              {sourceType === 'github' ? 'GitHub Token' : 'GitLab Token'}
            </p>
            <ol style={{ fontSize: '11px', color: 'var(--text-secondary)', margin: '0 0 10px', paddingLeft: '14px', lineHeight: 1.5 }}>
              {sourceType === 'github' ? (
                <>
                  <li>Go to GitHub → Settings → Developer settings</li>
                  <li>Personal access tokens → Tokens (classic)</li>
                  <li>Generate new token → Check <strong>repo</strong> scope</li>
                </>
              ) : (
                <>
                  <li>Go to GitLab → Preferences → Access Tokens</li>
                  <li>Add new token → Check <strong>read_api</strong> scope</li>
                </>
              )}
            </ol>
            <p style={{ fontSize: '11px', fontWeight: 600, color: 'var(--text-primary)', margin: '0 0 6px' }}>
              Repositories Filter
            </p>
            <ul style={{ fontSize: '11px', color: 'var(--text-secondary)', margin: 0, paddingLeft: '14px', lineHeight: 1.5 }}>
              <li><strong>All repos:</strong> Leave Repositories empty</li>
              <li><strong>Specific repo:</strong> <code style={{ backgroundColor: 'var(--bg-card)', padding: '1px 3px', borderRadius: '2px' }}>my-repo</code> or <code style={{ backgroundColor: 'var(--bg-card)', padding: '1px 3px', borderRadius: '2px' }}>owner/my-repo</code></li>
              <li><strong>Multiple:</strong> <code style={{ backgroundColor: 'var(--bg-card)', padding: '1px 3px', borderRadius: '2px' }}>repo1, repo2</code></li>
            </ul>
          </div>

          <div style={{ display: 'flex', gap: '10px', justifyContent: 'flex-end' }}>
            <button
              type="button"
              onClick={onClose}
              style={{
                padding: '8px 16px',
                borderRadius: '6px',
                border: '1px solid var(--border-color)',
                backgroundColor: 'transparent',
                color: 'var(--text-primary)',
                fontSize: '13px',
                fontWeight: 500,
                cursor: 'pointer',
              }}
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={loading}
              style={{
                padding: '8px 16px',
                borderRadius: '6px',
                border: 'none',
                backgroundColor: 'var(--accent)',
                color: 'white',
                fontSize: '13px',
                fontWeight: 500,
                cursor: loading ? 'not-allowed' : 'pointer',
                opacity: loading ? 0.7 : 1,
              }}
            >
              {loading ? (isEditing ? 'Saving...' : 'Adding...') : (isEditing ? 'Save Changes' : 'Add Source')}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}

function SourcesTab({ sources, loading, onAddSource, onEditSource, onDeleteSource }: {
  sources: Source[];
  loading: boolean;
  onAddSource: () => void;
  onEditSource: (source: Source) => void;
  onDeleteSource: (id: number) => void;
}) {
  return (
    <div>
      <div style={{
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'space-between',
        marginBottom: '16px',
      }}>
        <h3 style={{ fontSize: '14px', fontWeight: 600, color: 'var(--text-primary)', margin: 0 }}>
          Git Sources
        </h3>
        <button
          onClick={onAddSource}
          style={{
            padding: '6px 12px',
            borderRadius: '6px',
            border: 'none',
            backgroundColor: 'var(--accent)',
            color: 'white',
            fontSize: '13px',
            fontWeight: 500,
            cursor: 'pointer',
          }}
        >
          + Add
        </button>
      </div>

      {loading ? (
        <div style={{ textAlign: 'center', padding: '24px', color: 'var(--text-muted)' }}>
          Loading...
        </div>
      ) : sources.length === 0 ? (
        <div style={{
          padding: '32px 16px',
          textAlign: 'center',
          backgroundColor: 'var(--bg-primary)',
          borderRadius: '8px',
          border: '1px solid var(--border-color)',
        }}>
          <div style={{ fontSize: '32px', marginBottom: '12px' }}>🔗</div>
          <p style={{ color: 'var(--text-muted)', fontSize: '13px', margin: 0 }}>
            No sources configured yet
          </p>
        </div>
      ) : (
        <div style={{ display: 'flex', flexDirection: 'column', gap: '8px' }}>
          {sources.map((source) => (
            <div
              key={source.id}
              style={{
                padding: '12px 16px',
                borderRadius: '8px',
                backgroundColor: 'var(--bg-primary)',
                border: '1px solid var(--border-color)',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'space-between',
              }}
            >
              <div style={{ display: 'flex', alignItems: 'center', gap: '12px' }}>
                <span style={{ fontSize: '20px' }}>
                  {source.type === 'gitlab' ? '🦊' : '🐙'}
                </span>
                <div>
                  <div style={{ fontSize: '14px', fontWeight: 500, color: 'var(--text-primary)' }}>
                    {source.name}
                  </div>
                  <div style={{ fontSize: '12px', color: 'var(--text-muted)' }}>
                    {source.type === 'gitlab' ? 'GitLab' : 'GitHub'} · {source.organization || 'Personal'}
                  </div>
                </div>
              </div>
              <div style={{ display: 'flex', gap: '6px' }}>
                <button
                  onClick={() => onEditSource(source)}
                  style={{
                    padding: '6px 10px',
                    borderRadius: '6px',
                    border: '1px solid var(--border-color)',
                    backgroundColor: 'transparent',
                    color: 'var(--text-primary)',
                    fontSize: '12px',
                    fontWeight: 500,
                    cursor: 'pointer',
                  }}
                >
                  Edit
                </button>
                <button
                  onClick={() => onDeleteSource(source.id)}
                  style={{
                    padding: '6px 10px',
                    borderRadius: '6px',
                    border: 'none',
                    backgroundColor: 'var(--danger-bg)',
                    color: 'var(--danger-text)',
                    fontSize: '12px',
                    fontWeight: 500,
                    cursor: 'pointer',
                  }}
                >
                  Delete
                </button>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}

function ScheduleTab({ settings, onUpdate }: { settings: Settings; onUpdate: (updates: Partial<Settings>) => void }) {
  const [cron, setCron] = useState(settings.schedule_cron);

  const cronPresets = [
    { label: 'Daily at 9 AM', value: '0 9 * * *' },
    { label: 'Daily at 6 PM', value: '0 18 * * *' },
    { label: 'Every 12 hours', value: '0 */12 * * *' },
    { label: 'Weekly (Mon 9 AM)', value: '0 9 * * 1' },
    { label: 'Every 6 hours', value: '0 */6 * * *' },
  ];

  return (
    <div>
      <h3 style={{ fontSize: '14px', fontWeight: 600, color: 'var(--text-primary)', margin: '0 0 16px' }}>
        Scheduled Scans
      </h3>

      <div style={{ marginBottom: '20px' }}>
        <label style={{ display: 'flex', alignItems: 'center', gap: '10px', cursor: 'pointer' }}>
          <input
            type="checkbox"
            checked={settings.schedule_enabled}
            onChange={(e) => onUpdate({ schedule_enabled: e.target.checked })}
            style={{ width: '18px', height: '18px', cursor: 'pointer' }}
          />
          <span style={{ fontSize: '14px', color: 'var(--text-primary)' }}>Enable scheduled scans</span>
        </label>
      </div>

      <div style={{ marginBottom: '16px' }}>
        <label style={{ display: 'block', fontSize: '13px', fontWeight: 500, color: 'var(--text-primary)', marginBottom: '8px' }}>
          Schedule Presets
        </label>
        <div style={{ display: 'flex', flexWrap: 'wrap', gap: '8px' }}>
          {cronPresets.map((preset) => (
            <button
              key={preset.value}
              onClick={() => {
                setCron(preset.value);
                onUpdate({ schedule_cron: preset.value });
              }}
              style={{
                padding: '6px 12px',
                borderRadius: '6px',
                border: `1px solid ${cron === preset.value ? 'var(--accent)' : 'var(--border-color)'}`,
                backgroundColor: cron === preset.value ? 'var(--accent-light)' : 'transparent',
                color: cron === preset.value ? 'var(--accent)' : 'var(--text-secondary)',
                fontSize: '12px',
                fontWeight: 500,
                cursor: 'pointer',
              }}
            >
              {preset.label}
            </button>
          ))}
        </div>
      </div>

      <div style={{ marginBottom: '16px' }}>
        <label style={{ display: 'block', fontSize: '13px', fontWeight: 500, color: 'var(--text-primary)', marginBottom: '6px' }}>
          Cron Expression
        </label>
        <div style={{ display: 'flex', gap: '8px' }}>
          <input
            type="text"
            value={cron}
            onChange={(e) => setCron(e.target.value)}
            placeholder="0 9 * * *"
            style={{ ...inputStyle, flex: 1 }}
          />
          <button
            onClick={() => onUpdate({ schedule_cron: cron })}
            style={{
              padding: '10px 16px',
              borderRadius: '8px',
              border: 'none',
              backgroundColor: 'var(--accent)',
              color: 'white',
              fontSize: '13px',
              fontWeight: 500,
              cursor: 'pointer',
            }}
          >
            Save
          </button>
        </div>
        <p style={{ fontSize: '11px', color: 'var(--text-muted)', marginTop: '6px' }}>
          Format: minute hour day month weekday (e.g., "0 9 * * *" = every day at 9 AM)
        </p>
      </div>

      <div style={{
        padding: '12px',
        borderRadius: '8px',
        backgroundColor: 'var(--bg-primary)',
        border: '1px solid var(--border-color)',
      }}>
        <p style={{ fontSize: '12px', color: 'var(--text-secondary)', margin: 0 }}>
          <strong>Current:</strong> {settings.schedule_enabled ? `Enabled (${settings.schedule_cron})` : 'Disabled'}
        </p>
      </div>
    </div>
  );
}

function EmailTab({ settings, onUpdate, onTestEmail }: {
  settings: Settings;
  onUpdate: (updates: Partial<Settings>) => void;
  onTestEmail: () => void;
}) {
  const [form, setForm] = useState({
    email_smtp_host: settings.email_smtp_host,
    email_smtp_port: settings.email_smtp_port,
    email_smtp_user: settings.email_smtp_user,
    email_smtp_pass: settings.email_smtp_pass,
    email_from: settings.email_from,
    email_to: settings.email_to,
  });
  const [saving, setSaving] = useState(false);
  const [testing, setTesting] = useState(false);

  async function handleSave() {
    setSaving(true);
    try {
      await onUpdate(form);
    } finally {
      setSaving(false);
    }
  }

  async function handleTest() {
    setTesting(true);
    try {
      await onTestEmail();
    } finally {
      setTesting(false);
    }
  }

  return (
    <div>
      <h3 style={{ fontSize: '14px', fontWeight: 600, color: 'var(--text-primary)', margin: '0 0 16px' }}>
        Email Notifications
      </h3>

      <div style={{ marginBottom: '20px' }}>
        <label style={{ display: 'flex', alignItems: 'center', gap: '10px', cursor: 'pointer' }}>
          <input
            type="checkbox"
            checked={settings.email_enabled}
            onChange={(e) => onUpdate({ email_enabled: e.target.checked })}
            style={{ width: '18px', height: '18px', cursor: 'pointer' }}
          />
          <span style={{ fontSize: '14px', color: 'var(--text-primary)' }}>Enable email notifications</span>
        </label>
      </div>

      <div style={{ marginBottom: '20px' }}>
        <label style={{ display: 'flex', alignItems: 'center', gap: '10px', cursor: 'pointer' }}>
          <input
            type="checkbox"
            checked={settings.email_notify_new_outdated}
            onChange={(e) => onUpdate({ email_notify_new_outdated: e.target.checked })}
            style={{ width: '18px', height: '18px', cursor: 'pointer' }}
          />
          <span style={{ fontSize: '14px', color: 'var(--text-primary)' }}>Notify on new outdated dependencies</span>
        </label>
      </div>

      <div style={{ marginBottom: '14px' }}>
        <label style={{ display: 'block', fontSize: '13px', fontWeight: 500, color: 'var(--text-primary)', marginBottom: '6px' }}>
          SMTP Host
        </label>
        <input
          type="text"
          value={form.email_smtp_host}
          onChange={(e) => setForm({ ...form, email_smtp_host: e.target.value })}
          placeholder="smtp.gmail.com"
          style={inputStyle}
        />
      </div>

      <div style={{ marginBottom: '14px' }}>
        <label style={{ display: 'block', fontSize: '13px', fontWeight: 500, color: 'var(--text-primary)', marginBottom: '6px' }}>
          SMTP Port
        </label>
        <input
          type="number"
          value={form.email_smtp_port}
          onChange={(e) => setForm({ ...form, email_smtp_port: parseInt(e.target.value) || 587 })}
          placeholder="587"
          style={inputStyle}
        />
      </div>

      <div style={{ marginBottom: '14px' }}>
        <label style={{ display: 'block', fontSize: '13px', fontWeight: 500, color: 'var(--text-primary)', marginBottom: '6px' }}>
          SMTP Username
        </label>
        <input
          type="text"
          value={form.email_smtp_user}
          onChange={(e) => setForm({ ...form, email_smtp_user: e.target.value })}
          placeholder="your-email@gmail.com"
          style={inputStyle}
        />
      </div>

      <div style={{ marginBottom: '14px' }}>
        <label style={{ display: 'block', fontSize: '13px', fontWeight: 500, color: 'var(--text-primary)', marginBottom: '6px' }}>
          SMTP Password
        </label>
        <input
          type="password"
          value={form.email_smtp_pass}
          onChange={(e) => setForm({ ...form, email_smtp_pass: e.target.value })}
          placeholder="App password or SMTP password"
          style={inputStyle}
        />
        <p style={{ fontSize: '11px', color: 'var(--text-muted)', marginTop: '4px' }}>
          For Gmail, use an App Password
        </p>
      </div>

      <div style={{ marginBottom: '14px' }}>
        <label style={{ display: 'block', fontSize: '13px', fontWeight: 500, color: 'var(--text-primary)', marginBottom: '6px' }}>
          From Address
        </label>
        <input
          type="email"
          value={form.email_from}
          onChange={(e) => setForm({ ...form, email_from: e.target.value })}
          placeholder="stale@example.com"
          style={inputStyle}
        />
      </div>

      <div style={{ marginBottom: '20px' }}>
        <label style={{ display: 'block', fontSize: '13px', fontWeight: 500, color: 'var(--text-primary)', marginBottom: '6px' }}>
          To Address(es)
        </label>
        <input
          type="text"
          value={form.email_to}
          onChange={(e) => setForm({ ...form, email_to: e.target.value })}
          placeholder="team@example.com, dev@example.com"
          style={inputStyle}
        />
        <p style={{ fontSize: '11px', color: 'var(--text-muted)', marginTop: '4px' }}>
          Comma-separated for multiple recipients
        </p>
      </div>

      <div style={{ display: 'flex', gap: '10px' }}>
        <button
          onClick={handleSave}
          disabled={saving}
          style={{
            flex: 1,
            padding: '10px 16px',
            borderRadius: '8px',
            border: 'none',
            backgroundColor: 'var(--accent)',
            color: 'white',
            fontSize: '13px',
            fontWeight: 500,
            cursor: saving ? 'not-allowed' : 'pointer',
            opacity: saving ? 0.7 : 1,
          }}
        >
          {saving ? 'Saving...' : 'Save Settings'}
        </button>
        <button
          onClick={handleTest}
          disabled={testing || !settings.email_smtp_host}
          style={{
            padding: '10px 16px',
            borderRadius: '8px',
            border: '1px solid var(--border-color)',
            backgroundColor: 'transparent',
            color: 'var(--text-primary)',
            fontSize: '13px',
            fontWeight: 500,
            cursor: testing || !settings.email_smtp_host ? 'not-allowed' : 'pointer',
            opacity: testing || !settings.email_smtp_host ? 0.7 : 1,
          }}
        >
          {testing ? 'Sending...' : 'Send Test'}
        </button>
      </div>
    </div>
  );
}

function CloseIcon() {
  return (
    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
      <line x1="18" y1="6" x2="6" y2="18" />
      <line x1="6" y1="6" x2="18" y2="18" />
    </svg>
  );
}

const inputStyle: React.CSSProperties = {
  width: '100%',
  padding: '10px 12px',
  borderRadius: '8px',
  border: '1px solid var(--border-color)',
  backgroundColor: 'var(--bg-primary)',
  color: 'var(--text-primary)',
  fontSize: '14px',
  outline: 'none',
  boxSizing: 'border-box',
};
