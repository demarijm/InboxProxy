# InboxProxy

This is a simple server used for demonstrating configuration and observability features.

## Usage

```bash
go run . -config config.json
```

The server exposes:

- `/health` – health check endpoint returning JSON
- `/metrics` – Prometheus style metrics

Configuration is loaded from a JSON file and environment variables.
Reload the configuration by sending `SIGHUP` to the process.
