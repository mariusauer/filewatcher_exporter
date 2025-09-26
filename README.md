# filewatcher-exporter

`filewatcher-exporter` monitors one or more directories for filesystem changes and exposes Prometheus metrics that capture the most recent write activity.

## Features
- Recursive watchers backed by `fsnotify`, including newly created sub-directories.
- Prometheus gauges for last write timestamp and age per top-level directory.
- Simple command-line flags powered by `pflag` for directory selection and listen address.

## Getting Started

### Build for Linux

#### Prerequisites
- Go 1.24 or newer

```
make build-linux
```
This builds with `CGO_ENABLED=0` for a static binary. The binary is written to `bin/filewatcher-exporter` with `GOARCH` defaulting to `amd64`. Override by passing `GOARCH=arm64` (or another architecture) to `make`.

### Run Locally
```
CGO_ENABLED=0 go run . --dirs "/tmp:/var/log" --web.listen-address ":9100" --web.config.file web-config.yml
```
- `--dirs` accepts a colon-separated list of directories to watch.
- `--web.listen-address` controls the HTTP address (defaults to `:9000`).
- `--web.config.file` enables TLS/basic-auth/etc. through exporter-toolkit.
- Metrics are exposed at `http://localhost:9000/metrics` unless you override the port.

### Example web-config.yml
```
tls_server_config:
  cert_file: /path/to/server.crt
  key_file: /path/to/server.key

basic_auth_users:
  admin: $2y$05$2b22E... # bcrypt hash
```

## Releasing
Tagging with the pattern `v*` triggers the GitHub Actions workflow defined in `.github/workflows/release-linux.yml`. The pipeline builds a Linux binary, packages it as `filewatcher-exporter_<tag>_linux_amd64.tgz`, uploads it as an artifact, and publishes a GitHub Release.

## Development Notes
- Keep helper packages beside `main.go` to preserve simple relative imports.
- Format code with `gofmt` (e.g. `go fmt ./...`) before committing.
- Use temporary directories in tests (`*_test.go`) when mocking filesystem events.

## Flags
- `--dirs` (string, default `/tmp`): colon-separated directories to watch.
- `--web.listen-address` (string, default `:9000`): exporter listen address.
- `--web.config.file` (string): path to exporter-toolkit web config file.
