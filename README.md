# stale

A dependency version dashboard that scans your repositories and tracks outdated packages across multiple ecosystems.

## Features

- **Multi-Source Support**: Connect to GitHub and GitLab organizations
- **Multi-Ecosystem**: Supports npm, Maven, Gradle, and Go modules
- **Dashboard View**: Visual overview with statistics cards for quick insights
- **Filtering**: Filter dependencies by status (upgradable, up-to-date), type (production, development), and repository
- **Search**: Quick search across all packages and repositories
- **CSV Export**: Export dependency data for external analysis
- **Dark Mode**: Toggle between light and dark themes

## Supported Package Managers

| Ecosystem | Manifest File | Registry |
|-----------|--------------|----------|
| npm | `package.json` | npmjs.com |
| Maven | `pom.xml` | Maven Central |
| Gradle | `build.gradle`, `build.gradle.kts` | Maven Central |
| Go | `go.mod` | pkg.go.dev |

## Quick Start

### Using Docker

```bash
docker pull jiin724/stale:latest

docker run -d \
  -p 8080:8080 \
  -v stale-data:/data \
  jiin724/stale:latest
```

### Building from Source

**Prerequisites:**
- Go 1.21+
- Node.js 18+
- pnpm

```bash
# Clone the repository
git clone https://github.com/jiin/stale.git
cd stale

# Build backend
go build -o stale ./cmd/server

# Build frontend
cd ui
pnpm install
pnpm build
cd ..

# Run
./stale
```

## Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | Server port | `8080` |
| `DATA_DIR` | SQLite database directory | `./data` |
| `STALE_ENCRYPTION_KEY` | Key for encrypting API tokens (recommended for production) | Built-in default |

**Security Note:** API tokens are encrypted using AES-256-GCM before storing in the database. For production deployments, always set a strong `STALE_ENCRYPTION_KEY`:

```bash
docker run -d \
  -p 8080:8080 \
  -v stale-data:/data \
  -e STALE_ENCRYPTION_KEY="your-secure-random-key-here" \
  jiin724/stale:latest
```

### Adding Sources

1. Navigate to **Sources** page in the UI
2. Click **Add Source**
3. Fill in:
   - **Name**: Display name for the source
   - **Type**: GitHub or GitLab
   - **Token**: Personal access token with repo read permissions
   - **Organization**: Organization/group name to scan
   - **URL** (GitLab only): Self-hosted GitLab URL
   - **Repositories** (optional): Comma-separated list to scan specific repos only

## API Endpoints

### Dependencies

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/dependencies` | List all dependencies |
| GET | `/api/v1/dependencies?repo={repo}` | Filter by repository |
| GET | `/api/v1/dependencies/export` | Export as CSV |
| GET | `/api/v1/dependencies/export?outdated=true` | Export outdated only |

### Sources

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/sources` | List all sources |
| POST | `/api/v1/sources` | Add a new source |
| DELETE | `/api/v1/sources/{id}` | Remove a source |

### Repositories

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/repositories` | List all repositories |
| DELETE | `/api/v1/repositories/{id}` | Remove a repository |

### Scanning

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/scan` | Trigger a full scan |
| POST | `/api/v1/scan?source_id={id}` | Scan specific source |
| GET | `/api/v1/scan/{id}` | Get scan status |

## Architecture

```
stale/
├── cmd/server/         # Application entrypoint
├── internal/
│   ├── api/            # HTTP handlers
│   ├── domain/         # Domain models
│   ├── repository/     # Database layer
│   └── service/        # Business logic
│       ├── github/     # GitHub API client
│       ├── gitlab/     # GitLab API client
│       ├── npm/        # npm registry client
│       ├── maven/      # Maven Central client
│       ├── golang/     # pkg.go.dev client
│       └── scanner/    # Repository scanner
└── ui/                 # React frontend
    ├── src/
    │   ├── api/        # API client
    │   ├── components/ # Reusable UI components
    │   ├── pages/      # Page components
    │   └── types/      # TypeScript types
    └── public/         # Static assets
```

## Development

### Backend

```bash
# Run with hot reload
go run ./cmd/server

# Run tests
go test ./...
```

### Frontend

```bash
cd ui

# Install dependencies
pnpm install

# Development server
pnpm dev

# Type checking
pnpm tsc

# Linting
pnpm lint

# Build for production
pnpm build
```

## Docker Build

```bash
# Build image (always use --no-cache for clean builds)
docker build --no-cache -t jiin724/stale:latest .

# Push to Docker Hub
docker push jiin724/stale:latest
```

## License

MIT
