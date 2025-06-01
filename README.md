# InboxProxy

This prototype demonstrates a small asynchronous job queue used to save files and execute hooks without blocking the main program. Jobs are retried on failure and failures are recorded in `failures.log`.

Run `go run ./` to see a simple example.
