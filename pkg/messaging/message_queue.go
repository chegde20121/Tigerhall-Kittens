package messaging

import "sync"

// PubSub represents the pub-sub system
type PubSub struct {
	mu          sync.RWMutex
	subscribers map[chan any]struct{}
}

// NewPubSub creates a new instance of the PubSub system
func NewPubSub() *PubSub {
	return &PubSub{
		subscribers: make(map[chan any]struct{}),
	}
}

// Subscribe adds a new subscriber to the PubSub system
func (ps *PubSub) Subscribe() chan any {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	subscriber := make(chan any, 10) // Buffering to avoid blocking publishers
	ps.subscribers[subscriber] = struct{}{}

	return subscriber
}

// Unsubscribe removes a subscriber from the PubSub system
func (ps *PubSub) Unsubscribe(subscriber chan any) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	delete(ps.subscribers, subscriber)
	close(subscriber)
}

// Publish sends a message to all subscribers
func (ps *PubSub) Publish(message any) {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	for subscriber := range ps.subscribers {
		go func(s chan any) {
			s <- message
		}(subscriber)
	}
}
