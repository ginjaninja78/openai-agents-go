package pod

import (
	"fmt"
	"sync"
)

// Pod represents a named unit of execution in a swarm, combining an identity
// with its own lifecycle system.
type Pod struct {
	Name    string
	lsystem *LSystem
}

// NewPod creates a Pod with the given name in the StateCreated state.
func NewPod(name string) *Pod {
	return &Pod{Name: name, lsystem: NewLSystem()}
}

// Start starts the pod.
func (p *Pod) Start() error { return p.lsystem.Start() }

// Stop stops the pod.
func (p *Pod) Stop() error { return p.lsystem.Stop() }

// Pause pauses the pod.
func (p *Pod) Pause() error { return p.lsystem.Pause() }

// Resume resumes a paused pod.
func (p *Pod) Resume() error { return p.lsystem.Resume() }

// State returns the current lifecycle state of the pod.
func (p *Pod) State() State { return p.lsystem.State() }

// Manager maintains a collection of named Pods and provides
// fleet-level lifecycle operations.
type Manager struct {
	mu   sync.RWMutex
	pods map[string]*Pod
}

// NewManager creates an empty Manager.
func NewManager() *Manager {
	return &Manager{pods: make(map[string]*Pod)}
}

// Add registers a pod with the manager.
// Returns an error if a pod with the same name already exists.
func (m *Manager) Add(pod *Pod) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.pods[pod.Name]; exists {
		return fmt.Errorf("pod %q already registered", pod.Name)
	}
	m.pods[pod.Name] = pod
	return nil
}

// Remove deregisters the pod with the given name.
// Returns an error if no pod with that name is registered.
func (m *Manager) Remove(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.pods[name]; !exists {
		return fmt.Errorf("pod %q not found", name)
	}
	delete(m.pods, name)
	return nil
}

// Get returns the pod registered under name.
// Returns nil and false if the pod is not registered.
func (m *Manager) Get(name string) (*Pod, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	p, ok := m.pods[name]
	return p, ok
}

// List returns a snapshot of all registered pods.
func (m *Manager) List() []*Pod {
	m.mu.RLock()
	defer m.mu.RUnlock()

	pods := make([]*Pod, 0, len(m.pods))
	for _, p := range m.pods {
		pods = append(pods, p)
	}
	return pods
}

// StartAll starts all registered pods, collecting any errors.
func (m *Manager) StartAll() []error {
	m.mu.RLock()
	pods := m.snapshot()
	m.mu.RUnlock()

	var errs []error
	for _, p := range pods {
		if err := p.Start(); err != nil {
			errs = append(errs, fmt.Errorf("pod %q: %w", p.Name, err))
		}
	}
	return errs
}

// StopAll stops all registered pods, collecting any errors.
func (m *Manager) StopAll() []error {
	m.mu.RLock()
	pods := m.snapshot()
	m.mu.RUnlock()

	var errs []error
	for _, p := range pods {
		if err := p.Stop(); err != nil {
			errs = append(errs, fmt.Errorf("pod %q: %w", p.Name, err))
		}
	}
	return errs
}

// snapshot returns a copy of all pods without holding a lock.
// Callers must hold at least a read lock before calling this.
func (m *Manager) snapshot() []*Pod {
	pods := make([]*Pod, 0, len(m.pods))
	for _, p := range m.pods {
		pods = append(pods, p)
	}
	return pods
}
