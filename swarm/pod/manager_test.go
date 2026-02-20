package pod

import (
	"testing"
)

func TestNewPod(t *testing.T) {
	p := NewPod("alpha")
	if p.Name != "alpha" {
		t.Fatalf("expected name %q, got %q", "alpha", p.Name)
	}
	if p.State() != StateCreated {
		t.Fatalf("expected StateCreated, got %s", p.State())
	}
}

func TestPod_StartStop(t *testing.T) {
	p := NewPod("beta")
	if err := p.Start(); err != nil {
		t.Fatalf("Start: %v", err)
	}
	if p.State() != StateRunning {
		t.Fatalf("expected StateRunning, got %s", p.State())
	}
	if err := p.Stop(); err != nil {
		t.Fatalf("Stop: %v", err)
	}
	if p.State() != StateStopped {
		t.Fatalf("expected StateStopped, got %s", p.State())
	}
}

func TestManager_AddAndGet(t *testing.T) {
	m := NewManager()
	p := NewPod("gamma")
	if err := m.Add(p); err != nil {
		t.Fatalf("Add: %v", err)
	}
	got, ok := m.Get("gamma")
	if !ok {
		t.Fatal("expected pod to be found")
	}
	if got.Name != "gamma" {
		t.Fatalf("expected pod name %q, got %q", "gamma", got.Name)
	}
}

func TestManager_Add_Duplicate(t *testing.T) {
	m := NewManager()
	p := NewPod("dup")
	_ = m.Add(p)
	err := m.Add(p)
	if err == nil {
		t.Fatal("expected error on duplicate add")
	}
}

func TestManager_Remove(t *testing.T) {
	m := NewManager()
	p := NewPod("delta")
	_ = m.Add(p)

	if err := m.Remove("delta"); err != nil {
		t.Fatalf("Remove: %v", err)
	}
	if _, ok := m.Get("delta"); ok {
		t.Fatal("expected pod to be removed")
	}
}

func TestManager_Remove_NotFound(t *testing.T) {
	m := NewManager()
	if err := m.Remove("ghost"); err == nil {
		t.Fatal("expected error removing unknown pod")
	}
}

func TestManager_List(t *testing.T) {
	m := NewManager()
	_ = m.Add(NewPod("p1"))
	_ = m.Add(NewPod("p2"))

	pods := m.List()
	if len(pods) != 2 {
		t.Fatalf("expected 2 pods, got %d", len(pods))
	}
}

func TestManager_StartAll(t *testing.T) {
	m := NewManager()
	_ = m.Add(NewPod("s1"))
	_ = m.Add(NewPod("s2"))

	errs := m.StartAll()
	if len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	for _, p := range m.List() {
		if p.State() != StateRunning {
			t.Errorf("pod %q expected StateRunning, got %s", p.Name, p.State())
		}
	}
}

func TestManager_StopAll(t *testing.T) {
	m := NewManager()
	p1, p2 := NewPod("t1"), NewPod("t2")
	_ = m.Add(p1)
	_ = m.Add(p2)
	m.StartAll() //nolint:errcheck

	errs := m.StopAll()
	if len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	for _, p := range m.List() {
		if p.State() != StateStopped {
			t.Errorf("pod %q expected StateStopped, got %s", p.Name, p.State())
		}
	}
}

func TestManager_StopAll_AlreadyStopped(t *testing.T) {
	m := NewManager()
	_ = m.Add(NewPod("never-started"))

	// Stopping a pod that was never started should return errors (invalid transition).
	errs := m.StopAll()
	if len(errs) == 0 {
		t.Fatal("expected errors when stopping a never-started pod")
	}
}
