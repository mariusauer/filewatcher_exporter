# filewatcher-exporter

`filewatcher-exporter` monitors one or more directories for filesystem changes and exposes Prometheus metrics that capture the most recent write activity.

## Features
- Recursive watchers backed by `fsnotify`, including newly created sub-directories.
- Prometheus gauges for last write timestamp and age per top-level directory.
- Simple command-line flags powered by `pflag` for directory selection and listen address.

## Getting Started

### Prerequisites
- Go 1.25.1 or newer

### Build for Linux
```
make build-linux
```
The binary is written to `bin/filewatcher-exporter` with `GOARCH` defaulting to `amd64`. Override by passing `GOARCH=arm64` (or another architecture) to `make`.

### Run Locally
```
go run . --dirs "/tmp:/var/log" --listen ":9100"
```
- `--dirs` accepts a colon-separated list of directories to watch.
- `--listen` controls the HTTP address (defaults to `:9000`).
- Metrics are exposed at `http://localhost:9000/metrics` unless you override the port.

## Releasing
Tagging with the pattern `v*` triggers the GitHub Actions workflow defined in `.github/workflows/release-linux.yml`. The pipeline builds a Linux binary, packages it as `filewatcher-exporter_<tag>_linux_amd64.tgz`, uploads it as an artifact, and publishes a GitHub Release.

## Development Notes
- Keep helper packages beside `main.go` to preserve simple relative imports.
- Format code with `gofmt` (e.g. `go fmt ./...`) before committing.
- Use temporary directories in tests (`*_test.go`) when mocking filesystem events.
