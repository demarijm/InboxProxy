package notification

import (
	"context"
	"time"
)

// Dispatcher handles sending messages with retry logic.
type Dispatcher struct {
	Notifier   Notifier
	MaxRetries int
	RetryDelay time.Duration
}

// Dispatch attempts to send the message using the configured notifier.
func (d *Dispatcher) Dispatch(ctx context.Context, msg *Message) {
	msg.Status = StatusPending
	for attempt := 0; attempt <= d.MaxRetries; attempt++ {
		err := d.Notifier.Send(ctx, msg)
		if err == nil {
			msg.Status = StatusComplete
			return
		}
		msg.Retries++
		if attempt < d.MaxRetries {
			select {
			case <-ctx.Done():
				msg.Status = StatusFailed
				return
			case <-time.After(d.RetryDelay):
			}
		}
	}
	msg.Status = StatusFailed
}
