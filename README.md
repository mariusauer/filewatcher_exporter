# filewatcher-exporter

`filewatcher-exporter` monitors one or more directories for filesystem changes and exposes Prometheus metrics that capture the most recent write activity.
It serves metrics via the Prometheus exporter-toolkit, so you can enable TLS, basic auth, and other hardening features with the standard toolkit config.

## Features

- Recursive watchers backed by `fsnotify`, including newly created sub-directories.
- Prometheus gauges for last write timestamp and age per top-level directory.

## Getting Started

### Build for Linux

#### Prerequisites

- Go 1.24 or newer

```bash
make build-linux
```

This builds with `CGO_ENABLED=0` for a static binary. The binary is written to `bin/filewatcher-exporter` with `GOARCH` defaulting to `amd64`. Override by passing `GOARCH=arm64` (or another architecture) to `make`.

### Run Locally

```bash
CGO_ENABLED=0 go run . --dirs "/tmp:/var/log" --web.listen-address ":9100" --web.config.file web-config.yml
```

- `--dirs` accepts a colon-separated list of directories to watch.
- `--web.listen-address` controls the HTTP address (defaults to `:9000`).
- `--web.config.file` enables TLS/basic-auth/etc. through exporter-toolkit.
- Metrics are exposed at `http://localhost:9000/metrics` unless you override the port.

### Example web-config.yml

```yaml
tls_server_config:
  cert_file: /path/to/server.crt
  key_file: /path/to/server.key

basic_auth_users:
  admin: $2y$05$2b22E... # bcrypt hash
```

### Example output

```
# HELP file_last_write_timestamp_seconds Last modification timestamp of any file in directory (event-driven)
# TYPE file_last_write_timestamp_seconds gauge
file_last_write_timestamp_seconds{directory="/opt/test"} 1.759264902e+09
file_last_write_timestamp_seconds{directory="/opt/test2"} 1.759264914e+09
```

## Releasing

Tagging with the pattern `v*` triggers the GitHub Actions workflow defined in `.github/workflows/release-linux.yml`.

## Flags

- `--dirs` (string, required): colon-separated directories to watch.
- `--recursive` (bool, default `false`): whether to watch directories recursively.
- `--web.listen-address` (string, default `:9000`): exporter listen address.
- `--web.config.file` (string): path to exporter-toolkit web config file.
