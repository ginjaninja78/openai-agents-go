package pod

import (
	"testing"
)

func TestLSystem_InitialState(t *testing.T) {
	l := NewLSystem()
	if got := l.State(); got != StateCreated {
		t.Fatalf("expected StateCreated, got %s", got)
	}
}

func TestLSystem_StateString(t *testing.T) {
	tests := []struct {
		state State
		want  string
	}{
		{StateCreated, "created"},
		{StateStarting, "starting"},
		{StateRunning, "running"},
		{StatePaused, "paused"},
		{StateStopping, "stopping"},
		{StateStopped, "stopped"},
		{State(99), "unknown(99)"},
	}
	for _, tc := range tests {
		if got := tc.state.String(); got != tc.want {
			t.Errorf("State(%d).String() = %q, want %q", int(tc.state), got, tc.want)
		}
	}
}

func TestLSystem_StartStopCycle(t *testing.T) {
	l := NewLSystem()

	if err := l.Start(); err != nil {
		t.Fatalf("Start: %v", err)
	}
	if l.State() != StateRunning {
		t.Fatalf("expected StateRunning after Start, got %s", l.State())
	}

	if err := l.Stop(); err != nil {
		t.Fatalf("Stop: %v", err)
	}
	if l.State() != StateStopped {
		t.Fatalf("expected StateStopped after Stop, got %s", l.State())
	}
}

func TestLSystem_PauseResume(t *testing.T) {
	l := NewLSystem()
	_ = l.Start()

	if err := l.Pause(); err != nil {
		t.Fatalf("Pause: %v", err)
	}
	if l.State() != StatePaused {
		t.Fatalf("expected StatePaused, got %s", l.State())
	}

	if err := l.Resume(); err != nil {
		t.Fatalf("Resume: %v", err)
	}
	if l.State() != StateRunning {
		t.Fatalf("expected StateRunning after Resume, got %s", l.State())
	}
}

func TestLSystem_InvalidTransition(t *testing.T) {
	l := NewLSystem()

	// Cannot pause from Created state.
	err := l.Pause()
	if err == nil {
		t.Fatal("expected error on invalid transition")
	}
	ste, ok := err.(*StateTransitionError)
	if !ok {
		t.Fatalf("expected *StateTransitionError, got %T", err)
	}
	if ste.From != StateCreated || ste.To != StatePaused {
		t.Fatalf("unexpected error fields: %v", ste)
	}
}

func TestLSystem_RestartAfterStop(t *testing.T) {
	l := NewLSystem()
	_ = l.Start()
	_ = l.Stop()

	if err := l.Start(); err != nil {
		t.Fatalf("Start after Stop: %v", err)
	}
	if l.State() != StateRunning {
		t.Fatalf("expected StateRunning, got %s", l.State())
	}
}

func TestLSystem_OnTransitionHook(t *testing.T) {
	l := NewLSystem()
	var transitions [][2]State

	l.OnTransition(func(from, to State) {
		transitions = append(transitions, [2]State{from, to})
	})

	_ = l.Start()
	_ = l.Stop()

	// Expected: Created→Starting, Starting→Running, Running→Stopping, Stopping→Stopped
	if len(transitions) != 4 {
		t.Fatalf("expected 4 transitions, got %d: %v", len(transitions), transitions)
	}
}
