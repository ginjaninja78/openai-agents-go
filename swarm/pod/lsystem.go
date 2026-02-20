// Package pod provides lifecycle management for swarm agent pods.
package pod

import (
	"fmt"
	"sync"
)

// State represents the lifecycle state of a pod.
type State int

const (
	// StateCreated indicates the pod has been created but not yet started.
	StateCreated State = iota
	// StateStarting indicates the pod is in the process of starting.
	StateStarting
	// StateRunning indicates the pod is actively running.
	StateRunning
	// StatePaused indicates the pod has been paused.
	StatePaused
	// StateStopping indicates the pod is in the process of stopping.
	StateStopping
	// StateStopped indicates the pod has been stopped.
	StateStopped
)

// String returns the human-readable name of the State.
func (s State) String() string {
	switch s {
	case StateCreated:
		return "created"
	case StateStarting:
		return "starting"
	case StateRunning:
		return "running"
	case StatePaused:
		return "paused"
	case StateStopping:
		return "stopping"
	case StateStopped:
		return "stopped"
	default:
		return fmt.Sprintf("unknown(%d)", int(s))
	}
}

// StateTransitionError is returned when a lifecycle transition is not allowed
// from the current state.
type StateTransitionError struct {
	From State
	To   State
}

// Error implements the error interface.
func (e *StateTransitionError) Error() string {
	return fmt.Sprintf("pod: invalid state transition from %s to %s", e.From, e.To)
}

// LSystem is the pod lifecycle system.
// It governs valid state transitions and notifies registered hooks.
type LSystem struct {
	mu    sync.Mutex
	state State
	hooks []func(from, to State)
}

// NewLSystem creates a new LSystem in the StateCreated state.
func NewLSystem() *LSystem {
	return &LSystem{state: StateCreated}
}

// State returns the current lifecycle state.
func (l *LSystem) State() State {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.state
}

// OnTransition registers a hook that is called after every successful state
// transition. The hook receives the previous and new states.
func (l *LSystem) OnTransition(hook func(from, to State)) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.hooks = append(l.hooks, hook)
}

// transitionLocked validates and applies a state transition while the caller
// holds l.mu. It returns the previous state so callers can fire hooks after
// releasing the lock.
func (l *LSystem) transitionLocked(target State) (prev State, err error) {
	if !isValidTransition(l.state, target) {
		return l.state, &StateTransitionError{From: l.state, To: target}
	}
	prev = l.state
	l.state = target
	return prev, nil
}

// fireHooks invokes each registered hook with the previous and next states.
// Must be called without holding l.mu.
func (l *LSystem) fireHooks(events [][2]State) {
	l.mu.Lock()
	hooks := make([]func(from, to State), len(l.hooks))
	copy(hooks, l.hooks)
	l.mu.Unlock()

	for _, e := range events {
		for _, h := range hooks {
			h(e[0], e[1])
		}
	}
}

// Start transitions the pod from Created or Stopped → Starting → Running.
func (l *LSystem) Start() error {
	l.mu.Lock()
	prev1, err := l.transitionLocked(StateStarting)
	if err != nil {
		l.mu.Unlock()
		return err
	}
	prev2, err := l.transitionLocked(StateRunning)
	l.mu.Unlock()
	if err != nil {
		return err
	}
	l.fireHooks([][2]State{{prev1, StateStarting}, {prev2, StateRunning}})
	return nil
}

// Pause transitions the pod from Running → Paused.
func (l *LSystem) Pause() error {
	l.mu.Lock()
	prev, err := l.transitionLocked(StatePaused)
	l.mu.Unlock()
	if err != nil {
		return err
	}
	l.fireHooks([][2]State{{prev, StatePaused}})
	return nil
}

// Resume transitions the pod from Paused → Running.
func (l *LSystem) Resume() error {
	l.mu.Lock()
	prev, err := l.transitionLocked(StateRunning)
	l.mu.Unlock()
	if err != nil {
		return err
	}
	l.fireHooks([][2]State{{prev, StateRunning}})
	return nil
}

// Stop transitions the pod from Running or Paused → Stopping → Stopped.
func (l *LSystem) Stop() error {
	l.mu.Lock()
	prev1, err := l.transitionLocked(StateStopping)
	if err != nil {
		l.mu.Unlock()
		return err
	}
	prev2, err := l.transitionLocked(StateStopped)
	l.mu.Unlock()
	if err != nil {
		return err
	}
	l.fireHooks([][2]State{{prev1, StateStopping}, {prev2, StateStopped}})
	return nil
}

// isValidTransition reports whether moving from src to dst is permitted.
func isValidTransition(src, dst State) bool {
	allowed := map[State][]State{
		StateCreated:  {StateStarting},
		StateStarting: {StateRunning},
		StateRunning:  {StatePaused, StateStopping},
		StatePaused:   {StateRunning, StateStopping},
		StateStopping: {StateStopped},
		StateStopped:  {StateStarting},
	}
	for _, s := range allowed[src] {
		if s == dst {
			return true
		}
	}
	return false
}
