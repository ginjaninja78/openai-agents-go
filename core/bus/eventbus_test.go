package bus

import (
	"sync"
	"testing"
)

func TestNew(t *testing.T) {
	b := New()
	if b == nil {
		t.Fatal("expected non-nil EventBus")
	}
	if len(b.Topics()) != 0 {
		t.Fatalf("expected 0 topics, got %d", len(b.Topics()))
	}
}

func TestSubscribeAndPublish(t *testing.T) {
	b := New()
	var got any

	b.Subscribe("test", func(payload any) {
		got = payload
	})
	b.Publish("test", "hello")

	if got != "hello" {
		t.Fatalf("expected %q, got %v", "hello", got)
	}
}

func TestPublish_NoSubscribers(t *testing.T) {
	b := New()
	// Should not panic
	b.Publish("noop", "value")
}

func TestUnsubscribe(t *testing.T) {
	b := New()
	var count int

	unsub := b.Subscribe("tick", func(_ any) { count++ })
	b.Publish("tick", nil)
	unsub()
	b.Publish("tick", nil)

	if count != 1 {
		t.Fatalf("expected handler called once, got %d", count)
	}
}

func TestTopics(t *testing.T) {
	b := New()
	b.Subscribe("alpha", func(_ any) {})
	b.Subscribe("beta", func(_ any) {})

	topics := b.Topics()
	if len(topics) != 2 {
		t.Fatalf("expected 2 topics, got %d", len(topics))
	}
}

func TestTopics_EmptyAfterUnsubscribe(t *testing.T) {
	b := New()
	unsub := b.Subscribe("temp", func(_ any) {})
	unsub()

	topics := b.Topics()
	if len(topics) != 0 {
		t.Fatalf("expected 0 topics after unsubscribe, got %d", len(topics))
	}
}

func TestPublish_MultipleSubscribers(t *testing.T) {
	b := New()
	var mu sync.Mutex
	var received []int

	for i := range 3 {
		i := i
		b.Subscribe("multi", func(_ any) {
			mu.Lock()
			received = append(received, i)
			mu.Unlock()
		})
	}
	b.Publish("multi", nil)

	mu.Lock()
	defer mu.Unlock()
	if len(received) != 3 {
		t.Fatalf("expected 3 calls, got %d", len(received))
	}
}

func TestPublish_ConcurrentSafe(t *testing.T) {
	b := New()
	var wg sync.WaitGroup

	for range 10 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			b.Subscribe("concurrent", func(_ any) {})
			b.Publish("concurrent", "payload")
		}()
	}
	wg.Wait()
}
