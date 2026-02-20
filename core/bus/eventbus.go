// Package bus provides a thread-safe publish/subscribe event bus for
// inter-agent communication within the OpenAI Agents Go SDK.
package bus

import (
	"sync"
	"sync/atomic"
)

// Handler is a function that handles an event payload.
type Handler func(payload any)

// subID is a monotonically increasing subscription identifier.
var subID atomic.Uint64

// subscription pairs a unique ID with its handler.
type subscription struct {
	id      uint64
	handler Handler
}

// EventBus is a thread-safe publish/subscribe event bus.
// Agents and subsystems can subscribe to named event topics and publish
// events to notify all subscribers.
type EventBus struct {
	mu          sync.RWMutex
	subscribers map[string][]subscription
}

// New creates and returns a new EventBus.
func New() *EventBus {
	return &EventBus{
		subscribers: make(map[string][]subscription),
	}
}

// Subscribe registers a handler for the given event topic.
// The handler will be called each time an event is published on that topic.
// Subscribe returns an unsubscribe function that removes the handler.
func (b *EventBus) Subscribe(topic string, handler Handler) func() {
	id := subID.Add(1)

	b.mu.Lock()
	b.subscribers[topic] = append(b.subscribers[topic], subscription{id: id, handler: handler})
	b.mu.Unlock()

	return func() {
		b.mu.Lock()
		defer b.mu.Unlock()
		subs := b.subscribers[topic]
		for i, s := range subs {
			if s.id == id {
				subs[i] = subs[len(subs)-1]
				subs[len(subs)-1] = subscription{}
				subs = subs[:len(subs)-1]
				break
			}
		}
		if len(subs) == 0 {
			delete(b.subscribers, topic)
		} else {
			b.subscribers[topic] = subs
		}
	}
}

// Publish delivers payload to all handlers subscribed to topic.
// Handlers are invoked synchronously in the order they were registered.
func (b *EventBus) Publish(topic string, payload any) {
	b.mu.RLock()
	subs := make([]subscription, len(b.subscribers[topic]))
	copy(subs, b.subscribers[topic])
	b.mu.RUnlock()

	for _, s := range subs {
		s.handler(payload)
	}
}

// Topics returns the list of topics that have at least one subscriber.
func (b *EventBus) Topics() []string {
	b.mu.RLock()
	defer b.mu.RUnlock()

	topics := make([]string, 0, len(b.subscribers))
	for t := range b.subscribers {
		topics = append(topics, t)
	}
	return topics
}
