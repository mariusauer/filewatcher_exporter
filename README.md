# filewatcher_exporter

`filewatcher_exporter` monitors one or more directories for filesystem changes and exposes Prometheus metrics that capture the most recent write activity.
It serves metrics via the Prometheus exporter-toolkit, so you can enable TLS, basic auth, and other hardening features with the standard toolkit config.

## Features

- Recursive watchers backed by `fsnotify`, including newly created sub-directories.
- Prometheus gauges for last write timestamp and age per top-level directory.

## Getting Started

### Build for Linux

#### Prerequisites

- Go 1.24 or newer

```bash
make build
```

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

## Flags

- `--dirs` (string, required): colon-separated directories to watch.
- `--recursive` (bool, default `false`): whether to watch directories recursively.
- `--web.listen-address` (string, default `:9150`): exporter listen address.
- `--web.config.file` (string): path to exporter-toolkit web config file.
