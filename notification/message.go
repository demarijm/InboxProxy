package notification

// Status represents the current state of a notification message.
type Status int

const (
	// StatusPending indicates the message is queued for delivery.
	StatusPending Status = iota
	// StatusComplete indicates the message was delivered successfully.
	StatusComplete
	// StatusFailed indicates delivery ultimately failed.
	StatusFailed
)

// Message represents a single notification payload.
type Message struct {
	ID        string
	Recipient string
	Body      string
	Status    Status
	Retries   int
}
