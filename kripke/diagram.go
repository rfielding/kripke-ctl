
package kripke

import (
    "fmt"
    "io"
)

// NodeID is a simple identifier for states in diagrams.
// It is intentionally independent from the CTL StateID type.
type NodeID string

// SimpleGraph is a minimal explicit graph representation for diagrams.
type SimpleGraph struct {
    States []NodeID
    Succ   map[NodeID][]NodeID
}

// WriteMermaidStateDiagram writes a Mermaid stateDiagram-v2 representation
// of the given graph to w. "initial" is the starting state.
func WriteMermaidStateDiagram(g *SimpleGraph, initial NodeID, w io.Writer) error {
    fmt.Fprintln(w, "stateDiagram-v2")
    fmt.Fprintf(w, "  [*] --> %s\n\n", initial)

    seen := make(map[string]bool)
    for _, from := range g.States {
        for _, to := range g.Succ[from] {
            key := string(from) + "->" + string(to)
            if seen[key] {
                continue
            }
            seen[key] = true
            fmt.Fprintf(w, "  %s --> %s\n", from, to)
        }
    }
    return nil
}
