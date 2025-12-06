
package kripke

// CTLSpec describes one named CTL formula attached to a model.
type CTLSpec struct {
    Name        string // e.g. "AF delivered"
    Description string // human meaning
    Formula     string // textual CTL syntax for now
}

// CounterSpec describes a named counter for simulations/charts.
type CounterSpec struct {
    Name        string
    Description string
}

// ModelSpec is the small API that model packages implement.
type ModelSpec interface {
    Name() string
    OriginalText() string
    BuildGraph() (*SimpleGraph, NodeID)
    CTLFormulas() []CTLSpec
    Counters() []CounterSpec
}
