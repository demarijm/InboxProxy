# InboxProxy

InboxProxy is a minimal SMTP listener that stores incoming messages on disk.

## Features
- Configurable port using `SMTP_PORT` environment variable
- Configurable storage location via `STORAGE_DIR`
- Limits message size with `MAX_MESSAGE_SIZE` (bytes)
- Extracts metadata, text and HTML bodies, and attachments
- Exposes basic metrics on `:8080/metrics`

This project demonstrates a simple ingestion service using
[go-smtp](https://github.com/emersion/go-smtp) and
[go-message](https://github.com/emersion/go-message).
