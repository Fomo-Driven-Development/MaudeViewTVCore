package relay

import (
	"sync"
	"sync/atomic"
)

const subscriberBufSize = 256

// Event represents a single relay event to be sent via SSE.
type Event struct {
	Feed    string
	Payload string
}

// Broker fans out events to all subscribed SSE clients.
type Broker struct {
	mu          sync.RWMutex
	subscribers map[int64]chan Event
	nextID      atomic.Int64
}

// NewBroker creates a new SSE event broker.
func NewBroker() *Broker {
	return &Broker{
		subscribers: make(map[int64]chan Event),
	}
}

// Subscribe registers a new client. Returns the subscriber ID and a channel
// to receive events on. The channel is buffered; slow consumers will have
// events dropped.
func (b *Broker) Subscribe() (int64, <-chan Event) {
	id := b.nextID.Add(1)
	ch := make(chan Event, subscriberBufSize)
	b.mu.Lock()
	b.subscribers[id] = ch
	b.mu.Unlock()
	return id, ch
}

// Unsubscribe removes a subscriber and closes its channel.
func (b *Broker) Unsubscribe(id int64) {
	b.mu.Lock()
	ch, ok := b.subscribers[id]
	if ok {
		delete(b.subscribers, id)
		close(ch)
	}
	b.mu.Unlock()
}

// Publish sends an event to all subscribers. Non-blocking: slow clients
// have events dropped.
func (b *Broker) Publish(evt Event) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	for _, ch := range b.subscribers {
		select {
		case ch <- evt:
		default:
		}
	}
}

// ClientCount returns the number of active subscribers.
func (b *Broker) ClientCount() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.subscribers)
}
