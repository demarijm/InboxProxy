# InboxProxy

InboxProxy is a simple SMTP listener that stores incoming email metadata and content to disk.

## Configuration

The server is controlled via environment variables:

- `SMTP_ADDR` - address to listen on (default `:2525`)
- `STORAGE_DIR` - directory where emails are stored (default `data`)
- `MAX_FILE_SIZE` - maximum bytes to read from an incoming message (default `10485760`)

## Running

```
SMTP_ADDR=":2525" STORAGE_DIR="data" go run ./...
```

Each received email is stored in a timestamped folder under `STORAGE_DIR` with the raw message, metadata and any attachments.
