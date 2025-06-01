# InboxProxy

This repository contains simple examples related to file delivery
notifications. It demonstrates how to dispatch notifications and retry
failed deliveries.

## Notification System

Notifications are sent using a `Notifier` interface. The repository
includes a `LoggerNotifier` that writes messages to the standard log.

A `Dispatcher` coordinates sending messages and sets a status flag based
on whether the notification succeeds.

Statuses include:

- `StatusPending` – the message is waiting to be delivered.
- `StatusComplete` – the notification succeeded.
- `StatusFailed` – retries have been exhausted and the message failed.

The dispatcher retries a notification up to a configurable number of
attempts with a delay between tries.

## Example

```
go run ./
```

The example sends a single message using the logger notifier. You can
adapt the `Notifier` interface to integrate with email, webhooks or
Slack.
