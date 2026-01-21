# stale

A dependency version dashboard that scans your repositories and tracks outdated packages across multiple ecosystems.

## Features

- **Multi-Source Support**: GitHub and GitLab organizations
- **Multi-Ecosystem**: npm, Maven, Gradle, Go modules
- **Multi-Module**: Monorepo and multi-module project support
- **Dashboard**: Visual overview with filtering, search, and CSV export
- **Scheduled Scans**: Cron-based automatic scanning
- **Dark Mode**: Light and dark themes

## Quick Start

```bash
docker run -d \
  -p 8080:8080 \
  -v stale-data:/data \
  jiin724/stale:latest
```

Open http://localhost:8080 and add your first source.

## Documentation

See the [Wiki](https://github.com/amazingkj/stale/wiki) for detailed documentation:

- [Installation](https://github.com/amazingkj/stale/wiki/Installation)
- [Configuration](https://github.com/amazingkj/stale/wiki/Configuration)
- [API Reference](https://github.com/amazingkj/stale/wiki/API-Reference)
- [Development](https://github.com/amazingkj/stale/wiki/Development)

## License

MIT
