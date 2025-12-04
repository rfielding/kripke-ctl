package kripke

import (
	"testing"
)

func TestAtomicProp(t *testing.T) {
	s := NewBasicState("s1", "ready", "safe")
	model := NewModel(s)

	// Test atomic proposition that exists
	prop := &AtomicProp{Name: "ready"}
	if !prop.Evaluate(s, model) {
		t.Error("Expected 'ready' property to be true")
	}

	// Test atomic proposition that doesn't exist
	prop2 := &AtomicProp{Name: "error"}
	if prop2.Evaluate(s, model) {
		t.Error("Expected 'error' property to be false")
	}
}

func TestNot(t *testing.T) {
	s := NewBasicState("s1", "ready")
	model := NewModel(s)

	prop := &AtomicProp{Name: "ready"}
	notProp := &Not{Formula: prop}

	if notProp.Evaluate(s, model) {
		t.Error("Expected NOT(ready) to be false when ready is true")
	}

	prop2 := &AtomicProp{Name: "error"}
	notProp2 := &Not{Formula: prop2}

	if !notProp2.Evaluate(s, model) {
		t.Error("Expected NOT(error) to be true when error is false")
	}
}

func TestAndOr(t *testing.T) {
	s := NewBasicState("s1", "ready", "safe")
	model := NewModel(s)

	ready := &AtomicProp{Name: "ready"}
	safe := &AtomicProp{Name: "safe"}
	error := &AtomicProp{Name: "error"}

	// Test AND
	andFormula := &And{Left: ready, Right: safe}
	if !andFormula.Evaluate(s, model) {
		t.Error("Expected (ready AND safe) to be true")
	}

	andFormula2 := &And{Left: ready, Right: error}
	if andFormula2.Evaluate(s, model) {
		t.Error("Expected (ready AND error) to be false")
	}

	// Test OR
	orFormula := &Or{Left: ready, Right: error}
	if !orFormula.Evaluate(s, model) {
		t.Error("Expected (ready OR error) to be true")
	}

	orFormula2 := &Or{Left: error, Right: error}
	if orFormula2.Evaluate(s, model) {
		t.Error("Expected (error OR error) to be false")
	}
}

func TestEX(t *testing.T) {
	s0 := NewBasicState("s0", "init")
	s1 := NewBasicState("s1", "processing")
	s2 := NewBasicState("s2", "done")

	model := NewModel(s0)
	model.AddState(s0)
	model.AddState(s1)
	model.AddState(s2)

	actor := NewActor("test", s0)
	actor.AddState(s0)
	actor.AddState(s1)
	actor.AddState(s2)
	actor.AddTransition(s0, s1, 1.0, "start", 0)
	actor.AddTransition(s1, s2, 1.0, "complete", 0)
	model.AddActor(actor)

	// EX processing from s0
	ex := &EX{Formula: &AtomicProp{Name: "processing"}}
	if !ex.Evaluate(s0, model) {
		t.Error("Expected EX(processing) to be true from s0")
	}

	// EX done from s0 (should be false)
	ex2 := &EX{Formula: &AtomicProp{Name: "done"}}
	if ex2.Evaluate(s0, model) {
		t.Error("Expected EX(done) to be false from s0")
	}
}

func TestAX(t *testing.T) {
	s0 := NewBasicState("s0", "init")
	s1 := NewBasicState("s1", "processing")

	model := NewModel(s0)
	model.AddState(s0)
	model.AddState(s1)

	actor := NewActor("test", s0)
	actor.AddState(s0)
	actor.AddState(s1)
	actor.AddTransition(s0, s1, 1.0, "start", 0)
	model.AddActor(actor)

	// AX processing from s0 (only one successor)
	ax := &AX{Formula: &AtomicProp{Name: "processing"}}
	if !ax.Evaluate(s0, model) {
		t.Error("Expected AX(processing) to be true from s0")
	}
}

func TestEF(t *testing.T) {
	s0 := NewBasicState("s0", "init")
	s1 := NewBasicState("s1", "processing")
	s2 := NewBasicState("s2", "done")

	model := NewModel(s0)
	model.AddState(s0)
	model.AddState(s1)
	model.AddState(s2)

	actor := NewActor("test", s0)
	actor.AddState(s0)
	actor.AddState(s1)
	actor.AddState(s2)
	actor.AddTransition(s0, s1, 1.0, "start", 0)
	actor.AddTransition(s1, s2, 1.0, "complete", 0)
	model.AddActor(actor)

	// EF done from s0
	ef := &EF{Formula: &AtomicProp{Name: "done"}}
	if !ef.Evaluate(s0, model) {
		t.Error("Expected EF(done) to be true from s0")
	}
}

func TestAF(t *testing.T) {
	s0 := NewBasicState("s0", "init")
	s1 := NewBasicState("s1", "processing")
	s2 := NewBasicState("s2", "done")

	model := NewModel(s0)
	model.AddState(s0)
	model.AddState(s1)
	model.AddState(s2)

	actor := NewActor("test", s0)
	actor.AddState(s0)
	actor.AddState(s1)
	actor.AddState(s2)
	actor.AddTransition(s0, s1, 1.0, "start", 0)
	actor.AddTransition(s1, s2, 1.0, "complete", 0)
	model.AddActor(actor)

	// AF done from s0
	af := &AF{Formula: &AtomicProp{Name: "done"}}
	if !af.Evaluate(s0, model) {
		t.Error("Expected AF(done) to be true from s0")
	}
}

func TestEG(t *testing.T) {
	s0 := NewBasicState("s0", "init", "safe")
	s1 := NewBasicState("s1", "processing", "safe")

	model := NewModel(s0)
	model.AddState(s0)
	model.AddState(s1)

	actor := NewActor("test", s0)
	actor.AddState(s0)
	actor.AddState(s1)
	// Create a cycle s0 -> s1 -> s0
	actor.AddTransition(s0, s1, 1.0, "step", 0)
	actor.AddTransition(s1, s0, 1.0, "back", 0)
	model.AddActor(actor)

	// EG safe from s0 (cycle where safe always holds)
	eg := &EG{Formula: &AtomicProp{Name: "safe"}}
	if !eg.Evaluate(s0, model) {
		t.Error("Expected EG(safe) to be true from s0")
	}
}

func TestAG(t *testing.T) {
	s0 := NewBasicState("s0", "init", "safe")
	s1 := NewBasicState("s1", "processing", "safe")

	model := NewModel(s0)
	model.AddState(s0)
	model.AddState(s1)

	actor := NewActor("test", s0)
	actor.AddState(s0)
	actor.AddState(s1)
	actor.AddTransition(s0, s1, 1.0, "step", 0)
	model.AddActor(actor)

	// AG safe from s0
	ag := &AG{Formula: &AtomicProp{Name: "safe"}}
	if !ag.Evaluate(s0, model) {
		t.Error("Expected AG(safe) to be true from s0")
	}
}

func TestEU(t *testing.T) {
	s0 := NewBasicState("s0", "init", "ready")
	s1 := NewBasicState("s1", "processing", "ready")
	s2 := NewBasicState("s2", "done")

	model := NewModel(s0)
	model.AddState(s0)
	model.AddState(s1)
	model.AddState(s2)

	actor := NewActor("test", s0)
	actor.AddState(s0)
	actor.AddState(s1)
	actor.AddState(s2)
	actor.AddTransition(s0, s1, 1.0, "start", 0)
	actor.AddTransition(s1, s2, 1.0, "complete", 0)
	model.AddActor(actor)

	// E[ready U done] from s0
	eu := &EU{
		Left:  &AtomicProp{Name: "ready"},
		Right: &AtomicProp{Name: "done"},
	}
	if !eu.Evaluate(s0, model) {
		t.Error("Expected E[ready U done] to be true from s0")
	}
}

func TestAU(t *testing.T) {
	s0 := NewBasicState("s0", "init", "safe")
	s1 := NewBasicState("s1", "processing", "safe")
	s2 := NewBasicState("s2", "done")

	model := NewModel(s0)
	model.AddState(s0)
	model.AddState(s1)
	model.AddState(s2)

	actor := NewActor("test", s0)
	actor.AddState(s0)
	actor.AddState(s1)
	actor.AddState(s2)
	actor.AddTransition(s0, s1, 1.0, "start", 0)
	actor.AddTransition(s1, s2, 1.0, "complete", 0)
	model.AddActor(actor)

	// A[safe U done] from s0
	au := &AU{
		Left:  &AtomicProp{Name: "safe"},
		Right: &AtomicProp{Name: "done"},
	}
	if !au.Evaluate(s0, model) {
		t.Error("Expected A[safe U done] to be true from s0")
	}
}
