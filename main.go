package main

import (
	"context"
	"time"

	"github.com/demarijm/inboxproxy/notification"
)

func main() {
	dispatcher := notification.Dispatcher{
		Notifier:   notification.LoggerNotifier{},
		MaxRetries: 3,
		RetryDelay: time.Second,
	}

	msg := &notification.Message{
		ID:        "1",
		Recipient: "user@example.com",
		Body:      "File is ready",
	}

	dispatcher.Dispatch(context.Background(), msg)
}
