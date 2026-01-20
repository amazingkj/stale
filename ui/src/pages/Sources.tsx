import { useEffect, useState, useCallback } from 'react';
import { api } from '../api/client';
import {
  Button,
  Card,
  Modal,
  Input,
  EmptyState,
  LoadingSpinner,
  ErrorMessage,
} from '../components/common';
import type { Source, SourceInput } from '../types';

export function Sources() {
  const [sources, setSources] = useState<Source[]>([]);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [editingSource, setEditingSource] = useState<Source | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const loadSources = useCallback(async () => {
    try {
      const data = await api.getSources();
      setSources(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load sources');
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    loadSources();
  }, [loadSources]);

  const handleDelete = useCallback(async (id: number) => {
    if (!confirm('Are you sure? All associated data will be removed.')) return;
    try {
      await api.deleteSource(id);
      setSources(prev => prev.filter((s) => s.id !== id));
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to delete source');
    }
  }, []);

  const handleCreate = useCallback(async (input: SourceInput) => {
    const source = await api.createSource(input);
    setSources(prev => [source, ...prev]);
    setIsModalOpen(false);
  }, []);

  const handleUpdate = useCallback(async (id: number, input: SourceInput) => {
    const updated = await api.updateSource(id, input);
    setSources(prev => prev.map((s) => (s.id === id ? updated : s)));
    setEditingSource(null);
  }, []);

  if (loading) {
    return <LoadingSpinner fullPage text="Loading..." />;
  }

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: '24px' }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <div>
          <h1 style={{ fontSize: '24px', fontWeight: 700, color: 'var(--text-primary)', margin: 0 }}>
            Sources
          </h1>
          <p style={{ fontSize: '14px', color: 'var(--text-secondary)', margin: '4px 0 0' }}>
            Manage GitHub/GitLab connections
          </p>
        </div>
        <Button onClick={() => setIsModalOpen(true)}>Add Source</Button>
      </div>

      {error && <ErrorMessage message={error} onDismiss={() => setError(null)} />}

      {sources.length === 0 ? (
        <Card>
          <EmptyState
            icon="🔗"
            description="No sources configured yet"
            action={
              <Button onClick={() => setIsModalOpen(true)}>
                Add your first source
              </Button>
            }
          />
        </Card>
      ) : (
        <div style={{ display: 'grid', gap: '12px' }}>
          {sources.map((source) => (
            <SourceCard
              key={source.id}
              source={source}
              onEdit={() => setEditingSource(source)}
              onDelete={() => handleDelete(source.id)}
            />
          ))}
        </div>
      )}

      {isModalOpen && (
        <SourceModal
          onClose={() => setIsModalOpen(false)}
          onSubmit={handleCreate}
        />
      )}

      {editingSource && (
        <SourceModal
          source={editingSource}
          onClose={() => setEditingSource(null)}
          onSubmit={(input) => handleUpdate(editingSource.id, input)}
        />
      )}
    </div>
  );
}

interface SourceCardProps {
  source: Source;
  onEdit: () => void;
  onDelete: () => void;
}

function SourceCard({ source, onEdit, onDelete }: SourceCardProps) {
  return (
    <Card style={{ padding: '20px' }}>
      <div style={{
        display: 'flex',
        justifyContent: 'space-between',
        alignItems: 'center',
      }}>
        <div>
          <div style={{ display: 'flex', alignItems: 'center', gap: '12px' }}>
            <span style={{ fontSize: '24px' }}>
              {source.type === 'github' ? '🐙' : '🦊'}
            </span>
            <div>
              <h3 style={{ fontSize: '16px', fontWeight: 600, color: 'var(--text-primary)', margin: 0 }}>
                {source.name}
              </h3>
              <p style={{ fontSize: '13px', color: 'var(--text-secondary)', margin: '4px 0 0' }}>
                {source.organization || 'Personal repositories'}
                {source.repositories && (
                  <span style={{ marginLeft: '8px', fontSize: '12px', color: 'var(--text-muted)' }}>
                    ({source.repositories.split(',').length} repo{source.repositories.split(',').length > 1 ? 's' : ''} selected)
                  </span>
                )}
                {source.scan_branch && (
                  <span style={{ marginLeft: '8px', fontSize: '12px', color: 'var(--accent)' }}>
                    branch: {source.scan_branch}
                  </span>
                )}
              </p>
            </div>
          </div>
          {source.last_scan_at && (
            <p style={{ fontSize: '12px', color: 'var(--text-muted)', margin: '12px 0 0', paddingLeft: '36px' }}>
              Last scanned: {new Date(source.last_scan_at).toLocaleString()}
            </p>
          )}
        </div>
        <div style={{ display: 'flex', gap: '8px' }}>
          <Button variant="secondary" size="sm" onClick={onEdit}>
            Edit
          </Button>
          <Button
            variant="ghost"
            size="sm"
            onClick={onDelete}
            style={{
              backgroundColor: 'var(--danger-bg)',
              color: 'var(--danger-text)',
            }}
          >
            Delete
          </Button>
        </div>
      </div>
    </Card>
  );
}

interface SourceModalProps {
  source?: Source;
  onClose: () => void;
  onSubmit: (input: SourceInput) => Promise<void>;
}

function SourceModal({ source, onClose, onSubmit }: SourceModalProps) {
  const isEditing = !!source;
  const [name, setName] = useState(source?.name || '');
  const [type, setType] = useState<'github' | 'gitlab'>(source?.type || 'github');
  const [token, setToken] = useState('');
  const [organization, setOrganization] = useState(source?.organization || '');
  const [url, setUrl] = useState(source?.url || '');
  const [repositories, setRepositories] = useState(source?.repositories || '');
  const [scanBranch, setScanBranch] = useState(source?.scan_branch || '');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleSubmit = useCallback(async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError(null);
    try {
      await onSubmit({
        name,
        type,
        token,
        organization: organization || undefined,
        url: url || undefined,
        repositories: repositories || undefined,
        scan_branch: scanBranch || undefined,
      });
    } catch (err) {
      setError(err instanceof Error ? err.message : (isEditing ? 'Failed to update source' : 'Failed to add source'));
      setLoading(false);
    }
  }, [name, type, token, organization, url, repositories, scanBranch, isEditing, onSubmit]);

  return (
    <Modal
      isOpen={true}
      onClose={onClose}
      title={isEditing ? `Edit ${type === 'github' ? 'GitHub' : 'GitLab'} Source` : 'Add Source'}
      footer={
        <>
          <Button variant="secondary" onClick={onClose}>
            Cancel
          </Button>
          <Button loading={loading} onClick={handleSubmit}>
            {isEditing ? 'Save Changes' : 'Add Source'}
          </Button>
        </>
      }
    >
      <form onSubmit={handleSubmit} style={{ display: 'flex', flexDirection: 'column', gap: '16px' }}>
        {error && <ErrorMessage message={error} onDismiss={() => setError(null)} />}

        {!isEditing && (
          <div style={{ display: 'flex', flexDirection: 'column', gap: '6px' }}>
            <label style={{ fontSize: '14px', fontWeight: 500, color: 'var(--text-primary)' }}>
              Type
            </label>
            <div style={{ display: 'flex', gap: '8px' }}>
              <Button
                type="button"
                variant={type === 'github' ? 'primary' : 'secondary'}
                size="sm"
                onClick={() => setType('github')}
              >
                🐙 GitHub
              </Button>
              <Button
                type="button"
                variant={type === 'gitlab' ? 'primary' : 'secondary'}
                size="sm"
                onClick={() => setType('gitlab')}
              >
                🦊 GitLab
              </Button>
            </div>
          </div>
        )}

        <Input
          label="Name"
          value={name}
          onChange={(e) => setName(e.target.value)}
          placeholder={type === 'github' ? 'My GitHub Org' : 'My GitLab Org'}
          required
        />

        <div style={{ display: 'flex', flexDirection: 'column', gap: '6px' }}>
          <Input
            label="Personal Access Token"
            type="password"
            value={token}
            onChange={(e) => setToken(e.target.value)}
            placeholder={isEditing ? 'Enter new token to update' : (type === 'github' ? 'ghp_xxxxxxxxxxxx' : 'glpat-xxxxxxxxxxxx')}
            required
          />
          <p style={{ fontSize: '12px', color: 'var(--text-muted)', margin: 0 }}>
            {isEditing ? 'Token is required for security verification' : (type === 'github' ? 'Requires repo scope for private repos' : 'Requires read_api and read_repository scopes')}
          </p>
        </div>

        {type === 'gitlab' && (
          <div style={{ display: 'flex', flexDirection: 'column', gap: '6px' }}>
            <Input
              label="GitLab URL (optional)"
              value={url}
              onChange={(e) => setUrl(e.target.value)}
              placeholder="https://gitlab.example.com"
            />
            <p style={{ fontSize: '12px', color: 'var(--text-muted)', margin: 0 }}>
              Leave empty for gitlab.com, or enter your self-hosted GitLab URL
            </p>
          </div>
        )}

        <div style={{ display: 'flex', flexDirection: 'column', gap: '6px' }}>
          <Input
            label="Organization (optional)"
            value={organization}
            onChange={(e) => setOrganization(e.target.value)}
            placeholder={type === 'github' ? 'my-org' : 'my-group'}
          />
          <p style={{ fontSize: '12px', color: 'var(--text-muted)', margin: 0 }}>
            Leave empty to scan personal repos
          </p>
        </div>

        <div style={{ display: 'flex', flexDirection: 'column', gap: '6px' }}>
          <Input
            label="Repositories (optional)"
            value={repositories}
            onChange={(e) => setRepositories(e.target.value)}
            placeholder="repo1, owner/repo2"
          />
          <p style={{ fontSize: '12px', color: 'var(--text-muted)', margin: 0 }}>
            Comma-separated list of repos to scan. Leave empty to scan all repos.
          </p>
        </div>

        <div style={{ display: 'flex', flexDirection: 'column', gap: '6px' }}>
          <Input
            label="Branch (optional)"
            value={scanBranch}
            onChange={(e) => setScanBranch(e.target.value)}
            placeholder="main, develop, release"
          />
          <p style={{ fontSize: '12px', color: 'var(--text-muted)', margin: 0 }}>
            Branch to scan. Leave empty to use each repo's default branch.
          </p>
        </div>

        <HelpSection type={type} />
      </form>
    </Modal>
  );
}

function HelpSection({ type }: { type: 'github' | 'gitlab' }) {
  return (
    <div style={{
      padding: '12px',
      borderRadius: '8px',
      backgroundColor: 'var(--bg-primary)',
      border: '1px solid var(--border-color)',
    }}>
      <p style={{ fontSize: '12px', fontWeight: 600, color: 'var(--text-primary)', margin: '0 0 8px' }}>
        {type === 'github' ? 'GitHub Token' : 'GitLab Token'}
      </p>
      {type === 'github' ? (
        <ol style={{ fontSize: '12px', color: 'var(--text-secondary)', margin: '0 0 12px', paddingLeft: '16px', lineHeight: 1.6 }}>
          <li>Go to GitHub → Settings → Developer settings</li>
          <li>Personal access tokens → Tokens (classic)</li>
          <li>Generate new token → Check <strong>repo</strong> scope</li>
        </ol>
      ) : (
        <ol style={{ fontSize: '12px', color: 'var(--text-secondary)', margin: '0 0 12px', paddingLeft: '16px', lineHeight: 1.6 }}>
          <li>Go to GitLab → Preferences → Access Tokens</li>
          <li>Create a personal access token</li>
          <li>Select <strong>read_api</strong> and <strong>read_repository</strong> scopes</li>
        </ol>
      )}
      <p style={{ fontSize: '12px', fontWeight: 600, color: 'var(--text-primary)', margin: '0 0 8px' }}>
        Repositories Filter
      </p>
      <ul style={{ fontSize: '12px', color: 'var(--text-secondary)', margin: 0, paddingLeft: '16px', lineHeight: 1.6 }}>
        <li><strong>Scan all repos:</strong> Leave Repositories empty</li>
        <li><strong>Scan specific repo:</strong> <code style={{ backgroundColor: 'var(--bg-card)', padding: '1px 4px', borderRadius: '3px' }}>my-repo</code> or <code style={{ backgroundColor: 'var(--bg-card)', padding: '1px 4px', borderRadius: '3px' }}>owner/my-repo</code></li>
        <li><strong>Scan multiple repos:</strong> <code style={{ backgroundColor: 'var(--bg-card)', padding: '1px 4px', borderRadius: '3px' }}>repo1, repo2</code></li>
      </ul>
    </div>
  );
}
