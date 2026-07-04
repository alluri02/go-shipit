package inmemory

import "fmt"

// Queue is an in-memory implementation of domain.QueuePublisher.
// Just prints messages — useful for local dev.
type Queue struct{}

func NewQueue() *Queue {
	return &Queue{}
}

// Publish prints the message to stdout (no real queue in local dev).
func (q *Queue) Publish(queueName string, message []byte) error {
	fmt.Printf("  [queue] → %s: %s\n", queueName, string(message))
	return nil
}
