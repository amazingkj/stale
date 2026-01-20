export interface Source {
  id: number;
  name: string;
  type: 'github' | 'gitlab';
  organization?: string;
  url?: string;  // For self-hosted GitLab
  repositories?: string;  // Comma-separated list of repos to scan
  scan_branch?: string;  // Branch to scan (empty = use repo's default branch)
  insecure_skip_verify?: boolean;  // Skip TLS verification for self-hosted instances
  membership_only?: boolean;  // GitLab: only show projects where user is a member
  owner_only?: boolean;  // GitHub: only show repos owned by user
  created_at: string;
  updated_at: string;
  last_scan_at?: string;
}

export interface SourceInput {
  name: string;
  type: 'github' | 'gitlab';
  token: string;
  organization?: string;
  url?: string;  // For self-hosted GitLab
  repositories?: string;  // Comma-separated list of repos to scan
  scan_branch?: string;  // Branch to scan (empty = use repo's default branch)
  insecure_skip_verify?: boolean;  // Skip TLS verification for self-hosted instances
  membership_only?: boolean;  // GitLab: only show projects where user is a member
  owner_only?: boolean;  // GitHub: only show repos owned by user
}

export interface Repository {
  id: number;
  source_id: number;
  name: string;
  full_name: string;
  default_branch: string;
  html_url: string;
  has_package_json: boolean;
  has_pom_xml: boolean;
  has_build_gradle: boolean;
  has_go_mod: boolean;
  created_at: string;
  updated_at: string;
  last_scan_at?: string;
  dependency_count: number;
  outdated_count: number;
}

export interface Dependency {
  id: number;
  repository_id: number;
  name: string;
  current_version: string;
  latest_version: string;
  type: 'dependency' | 'devDependency';
  ecosystem: 'npm' | 'maven' | 'gradle' | 'go';
  is_outdated: boolean;
  updated_at: string;
  // Joined fields
  repo_name?: string;
  repo_full_name?: string;
  source_name?: string;
}

export interface ScanJob {
  id: number;
  source_id?: number;
  status: 'pending' | 'running' | 'completed' | 'failed';
  repos_found: number;
  deps_found: number;
  error?: string;
  started_at?: string;
  finished_at?: string;
  created_at: string;
}

export interface DependencyStats {
  total_dependencies: number;
  outdated_count: number;
  up_to_date_count: number;
  by_type: {
    dependency?: number;
    devDependency?: number;
  };
}

export interface PaginatedDependencies {
  data: Dependency[];
  total: number;
  page: number;
  limit: number;
  total_pages: number;
}

export interface FilterOptions {
  repos: string[];
  packages: string[];
  ecosystems: string[];
}

export interface Settings {
  schedule_enabled: boolean;
  schedule_cron: string;
  email_enabled: boolean;
  email_smtp_host: string;
  email_smtp_port: number;
  email_smtp_user: string;
  email_smtp_pass: string;
  email_from: string;
  email_to: string;
  email_notify_new_outdated: boolean;
}

export interface SettingsInput {
  schedule_enabled?: boolean;
  schedule_cron?: string;
  email_enabled?: boolean;
  email_smtp_host?: string;
  email_smtp_port?: number;
  email_smtp_user?: string;
  email_smtp_pass?: string;
  email_from?: string;
  email_to?: string;
  email_notify_new_outdated?: boolean;
}

export interface NextScan {
  enabled: boolean;
  next_run?: string;
  cron_expr: string;
}

export interface IgnoredDependency {
  id: number;
  name: string;
  ecosystem?: string;
  reason?: string;
  created_at: string;
}

export interface IgnoredDependencyInput {
  name: string;
  ecosystem?: string;
  reason?: string;
}
