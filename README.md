# InboxProxy

A minimal service providing health and metrics endpoints.

## Configuration

Configuration is loaded from `config.json` or the file specified by the `CONFIG_FILE` environment variable. Example:

```json
{
  "port": "8080"
}
```

Send a `SIGHUP` signal to reload the configuration at runtime.

## Usage

```bash
go run main.go
```

Endpoints:

- `/health` - returns `ok` if the service is running.
- `/metrics` - exposes Prometheus-compatible metrics.
