package notification

import (
	"context"
	"log"
)

// LoggerNotifier writes notifications to the standard logger.
type LoggerNotifier struct{}

// Name returns the notifier name.
func (LoggerNotifier) Name() string { return "logger" }

// Send logs the message body and recipient.
func (LoggerNotifier) Send(ctx context.Context, msg *Message) error {
	log.Printf("sending to %s: %s", msg.Recipient, msg.Body)
	return nil
}
