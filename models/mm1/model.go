package mm1

import "github.com/rfielding/kripke-ctl/kripke"

// Model is a minimal stub implementation of kripke.ModelSpec for the M/M/1 example.
// It is intentionally trivial so that the tooling (gendiagrams, docs, etc.) compiles
// cleanly. You can extend BuildGraph, CTLFormulas, and Counters later.
type Model struct{}

// Name returns the model identifier used by tooling (e.g. gendiagrams -name mm1).
func (Model) Name() string { return "mm1" }

// OriginalText should describe the English requirement / scenario this model represents.
func (Model) OriginalText() string {
	return "M/M/1 queue with capacity 10; arrivals and service are exponential; queue length must never exceed 10."
}

// BuildGraph constructs the Kripke graph for this model. For now it returns an empty
// graph and an empty initial state, just to satisfy the interfaces.
func (Model) BuildGraph() (*kripke.SimpleGraph, kripke.NodeID) {
	g := &kripke.SimpleGraph{}
	return g, kripke.NodeID("")
}

// CTLFormulas lists the CTL properties we care about for this model.
// This stub returns nil; add real properties as needed.
func (Model) CTLFormulas() []kripke.CTLSpec {
	return nil
}

// Counters lists any counters you want to track during simulation / diagrams.
// This stub returns nil; add real counters as needed.
func (Model) Counters() []kripke.CounterSpec {
	return nil
}
