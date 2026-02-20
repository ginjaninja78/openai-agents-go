package pod

import (
	"context"
	"fmt"
	"sync"
)

// HardwareMonitor checks the local Ollama queue.
type HardwareMonitor interface {
	IsLocalQueueFull(ctx context.Context) bool
}

// Manager spins up the GEA Pods.
type Manager struct {
	hardware HardwareMonitor
	active   sync.WaitGroup
}

// Dispatch evaluates the hardware and spawns the Pod.
func (m *Manager) Dispatch(ctx context.Context, domainLease string, tasks []TaskNode) error {
	m.active.Add(1)
	
	go func() {
		defer m.active.Done()

		var provider string
		if m.hardware.IsLocalQueueFull(ctx) {
			provider = "Cloud_Groq_Llama3" // Backpressure tripped, route to cloud
			fmt.Printf("[Dispatcher] Local VRAM full. Routing Pod for %s to %s\n", domainLease, provider)
		} else {
			provider = "Local_Ollama_Llama3" // Local execution
			fmt.Printf("[Dispatcher] Local VRAM available. Spinning up local Pod for %s\n", domainLease)
		}

		// 1. Acquire Domain Lease (Mutex) for `domainLease`
		// 2. Execute TaskNodes sequentially or in parallel based on Dependencies
		// 3. Trigger Sandbox 2PC on completion
	}()

	return nil
}

// Wait blocks until all Pods have completed their epochs.
func (m *Manager) Wait() {
	m.active.Wait()
}