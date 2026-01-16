export interface Source {
  id: number;
  name: string;
  type: 'github' | 'gitlab';
  organization?: string;
  url?: string;  // For self-hosted GitLab
  repositories?: string;  // Comma-separated list of repos to scan
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
