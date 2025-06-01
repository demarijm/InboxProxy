package notification

import "context"

// Notifier defines how to send a message to a recipient.
type Notifier interface {
	Send(ctx context.Context, msg *Message) error
	Name() string
}
