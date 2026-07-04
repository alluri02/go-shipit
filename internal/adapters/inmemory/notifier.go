package inmemory

import "fmt"

// Notifier is an in-memory implementation of domain.Notifier.
// Just prints notifications to stdout — useful for local dev.
type Notifier struct{}

func NewNotifier() *Notifier {
	return &Notifier{}
}

// Notify prints the notification to stdout.
func (n *Notifier) Notify(channel, message string) error {
	fmt.Printf("  [notify] %s: %s\n", channel, message)
	return nil
}
