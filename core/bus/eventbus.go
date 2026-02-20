package bus

import (
	"context"
	"sync"
	"time"
)

// EventType defines the strict vocabulary of the Swarm.
type EventType string

const (
	EventTaskCreated   EventType = "TASK_CREATED"
	EventPodDispatched EventType = "POD_DISPATCHED"
	Event2PCSuccess    EventType = "2PC_SUCCESS"
	EventSentryTripped EventType = "SENTRY_TRIPPED"
)

// Event is the universal payload moving through the system.
type Event struct {
	ID        string
	Type      EventType
	PodID     string
	Payload   []byte // JSON serialized data
	Timestamp time.Time
}

// GhostObserver defines the interface for the lfm2.5-thinking model.
type GhostObserver interface {
	// Observe receives the firehose. MUST NOT block the bus.
	Observe(ctx context.Context, firehose <-chan Event)
}

// EventBus is the concurrent-safe message broker.
type EventBus struct {
	subscribers map[EventType][]chan<- Event
	ghostStream chan Event
	mu          sync.RWMutex
}

// NewEventBus initializes the bus with a dedicated buffered stream for the Ghost.
func NewEventBus(ghostBuffer int) *EventBus {
	return &EventBus{
		subscribers: make(map[EventType][]chan<- Event),
		// The Ghost gets a massive buffer to prevent it from ever blocking the Swarm.
		ghostStream: make(chan Event, ghostBuffer), 
	}
}

// StartGhost ignites the Animus in a background goroutine.
func (eb *EventBus) StartGhost(ctx context.Context, ghost GhostObserver) {
	go ghost.Observe(ctx, eb.ghostStream)
}

// Subscribe allows Pod Leaders to listen for specific events.
func (eb *EventBus) Subscribe(eventType EventType, bufferSize int) <-chan Event {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	ch := make(chan Event, bufferSize)
	eb.subscribers[eventType] = append(eb.subscribers[eventType], ch)
	return ch
}

// Publish broadcasts an event. It is strictly non-blocking.
func (eb *EventBus) Publish(event Event) {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	// 1. Send to the Subconscious (The Ghost)
	select {
	case eb.ghostStream <- event:
		// Ghost received the thought
	default:
		// Ghost is thinking too slowly; drop the frame to save the Swarm.
	}

	// 2. Send to specific Pod subscribers
	if subs, exists := eb.subscribers[event.Type]; exists {
		for _, ch := range subs {
			select {
			case ch <- event:
			default:
				// Subscriber buffer full; handle backpressure/dead-letter queue here
			}
		}
	}
}