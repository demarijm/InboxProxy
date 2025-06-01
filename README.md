# InboxProxy

InboxProxy is a modular SMTP listener that stores incoming messages and their metadata on disk.

## Features
- Configurable port using `SMTP_ADDR` environment variable
- Configurable storage location via `STORAGE_DIR`
- Limits message size with `MAX_FILE_SIZE` (bytes)
- Extracts metadata, text and HTML bodies, and attachments
- Stores each email in a timestamped folder with raw message, metadata, and attachments

This project demonstrates a simple, extensible ingestion service using
[go-smtp](https://github.com/emersion/go-smtp) and
[go-message](https://github.com/emersion/go-message).

## Project Structure

```
inboxproxy/
  cmd/
    inboxproxy/
      main.go         # Entry point, configures and starts the SMTP server
  internal/
    smtpserver/
      server.go       # SMTP server logic (Backend, Session, NewServer)
      parser.go       # Email parsing logic (ParseEmail, types)
      types.go        # Type aliases for re-export
    storage/
      storage.go      # SaveMetadata for storing parsed email as JSON
  go.mod
  go.sum
  README.md
```

## Configuration

The server is controlled via environment variables:

- `SMTP_ADDR` - address to listen on (default `:2525`)
- `STORAGE_DIR` - directory where emails are stored (default `data`)
- `MAX_FILE_SIZE` - maximum bytes to read from an incoming message (default `10485760`)

## Running

```
SMTP_ADDR=":2525" STORAGE_DIR="data" go run ./cmd/inboxproxy
```

Each received email is stored in a timestamped folder under `STORAGE_DIR` with:
- `raw.eml`: the raw RFC822 message
- `meta.json`: parsed metadata (from, to, subject, text, html, attachments)
- Attachments: each file is saved in the same folder

## Extending
- Add new features (e.g., HTTP API, metrics, filtering) by creating new packages in `internal/` and wiring them up in `cmd/inboxproxy/main.go`.

---
MIT License
