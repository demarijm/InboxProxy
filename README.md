# InboxProxy

This is a small demo application providing a minimal HTTP API for inspecting
captured emails. The implementation uses only Go's standard library to keep
dependencies simple.

## Endpoints

- `GET /emails` – list all captured emails
- `GET /emails/{id}` – fetch a single email
- `POST /hooks` – register a webhook
- `GET /files/{id}` – download an attachment
- `GET /metrics` – simple request counters

Each endpoint requires a token passed via the `Authorization` header:
`Authorization: Bearer <token>`. The token defaults to `secret-token` but can
be overridden with the `API_TOKEN` environment variable.

## Running

```
go run .
```

The server listens on port `8080` by default.

