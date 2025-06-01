# InboxProxy

InboxProxy is a simple SMTP ingestion service used for testing. The server
listens for incoming e‑mails and stores metadata, message bodies and
attachments on the local filesystem.

## Building

```
go build
```

## Running

```
SMTP_PORT=2525 STORAGE_DIR=./data go run .
```

The service exposes SMTP on `SMTP_PORT` and metrics on `METRICS_PORT` (default
`8080`).

