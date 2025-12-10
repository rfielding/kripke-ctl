package main

import (
	"fmt"
	"strings"
	"https://github.com/rfielding/kripke-ctl/kripke"
)

// Producer-Consumer Example with Bounded Buffer
//
// This example demonstrates how to extract a Kripke graph from
// an actor-based system and verify CTL properties.

// ========== SIMPLIFIED CTL (same as example 1) ==========

type StateID int
type StateSet map[StateID]struct{}

func NewStateSet() StateSet              { return make(StateSet) }
func (s StateSet) Add(id StateID)        { s[id] = struct{}{} }
func (s StateSet) Contains(id StateID) bool { _, ok := s[id]; return ok }

type Graph struct {
	nextID   int
	nameToID map[string]StateID
	idToName map[StateID]string
	labels   map[StateID]map[string]bool
	succ     map[StateID][]StateID
	init     []StateID
}

func NewGraph() *Graph {
	return &Graph{
		nameToID: make(map[string]StateID),
		idToName: make(map[StateID]string),
		labels:   make(map[StateID]map[string]bool),
		succ:     make(map[StateID][]StateID),
		init:     make([]StateID, 0),
	}
}

func (g *Graph) AddState(name string, lbls map[string]bool) StateID {
	if id, ok := g.nameToID[name]; ok {
		return id
	}
	id := StateID(g.nextID)
	g.nextID++
	g.nameToID[name] = id
	g.idToName[id] = name
	if lbls == nil {
		lbls = make(map[string]bool)
	}
	g.labels[id] = lbls
	return id
}

func (g *Graph) AddEdge(fromName, toName string) {
	from := g.nameToID[fromName]
	to := g.nameToID[toName]
	g.succ[from] = append(g.succ[from], to)
}

func (g *Graph) SetInitial(name string) {
	id := g.nameToID[name]
	g.init = append(g.init, id)
}

func (g *Graph) States() []StateID {
	out := make([]StateID, 0, len(g.idToName))
	for id := range g.idToName {
		out = append(out, id)
	}
	return out
}

func (g *Graph) Succ(s StateID) []StateID      { return g.succ[s] }
func (g *Graph) HasLabel(s StateID, p string) bool { return g.labels[s][p] }
func (g *Graph) NameOf(s StateID) string       { return g.idToName[s] }

type Formula interface {
	Sat(g *Graph) StateSet
}

type AtomFormula struct{ Prop string }

func Atom(prop string) Formula { return AtomFormula{Prop: prop} }
func (a AtomFormula) Sat(g *Graph) StateSet {
	res := NewStateSet()
	for _, s := range g.States() {
		if g.HasLabel(s, a.Prop) {
			res.Add(s)
		}
	}
	return res
}

type NotFormula struct{ Inner Formula }

func Not(inner Formula) Formula { return NotFormula{Inner: inner} }
func (n NotFormula) Sat(g *Graph) StateSet {
	inner := n.Inner.Sat(g)
	res := NewStateSet()
	for _, s := range g.States() {
		if !inner.Contains(s) {
			res.Add(s)
		}
	}
	return res
}

type OrFormula struct{ Left, Right Formula }

func Or(l, r Formula) Formula { return OrFormula{Left: l, Right: r} }
func (o OrFormula) Sat(g *Graph) StateSet {
	L := o.Left.Sat(g)
	R := o.Right.Sat(g)
	res := NewStateSet()
	for _, s := range g.States() {
		if L.Contains(s) || R.Contains(s) {
			res.Add(s)
		}
	}
	return res
}

type AGFormula struct{ Inner Formula }

func AG(inner Formula) Formula { return AGFormula{Inner: inner} }
func (f AGFormula) Sat(g *Graph) StateSet {
	return Not(EF(Not(f.Inner))).Sat(g)
}

type EFFormula struct{ Inner Formula }

func EF(inner Formula) Formula { return EFFormula{Inner: inner} }
func (f EFFormula) Sat(g *Graph) StateSet {
	target := f.Inner.Sat(g)
	X := target.Clone()
	changed := true
	for changed {
		changed = false
		for _, s := range g.States() {
			if X.Contains(s) {
				continue
			}
			for _, t := range g.Succ(s) {
				if X.Contains(t) {
					X.Add(s)
					changed = true
					break
				}
			}
		}
	}
	return X
}

// ========== MAIN ==========

func main() {
	fmt.Println("=" + strings.Repeat("=", 78))
	fmt.Println(" Producer-Consumer with Bounded Buffer - Complete Example")
	fmt.Println("=" + strings.Repeat("=", 78))
	fmt.Println()

	// English Description
	fmt.Println("SYSTEM DESCRIPTION")
	fmt.Println(strings.Repeat("-", 78))
	fmt.Println(`
Producer-Consumer with bounded buffer (capacity = 2):
- Producer creates items and sends them to a buffer
- Consumer receives items from the buffer
- When buffer is full (2 items), producer must wait
- When buffer is empty, consumer must wait
`)

	// Build Kripke Graph (manually modeling the state space)
	fmt.Println("STATE SPACE")
	fmt.Println(strings.Repeat("-", 78))

	g := NewGraph()

	// States: (producer_can_send, consumer_can_recv, buffer_size)
	// Buffer capacity is 2

	// Buffer size 0 (empty)
	s0 := g.AddState("B0", map[string]bool{
		"producer_ready": true,
		"consumer_ready": false,
		"buffer_empty":   true,
		"buffer_size_0":  true,
	})

	// Buffer size 1
	_ = g.AddState("B1", map[string]bool{
		"producer_ready": true,
		"consumer_ready": true,
		"buffer_size_1":  true,
	})

	// Buffer size 2 (full)
	_ = g.AddState("B2", map[string]bool{
		"producer_ready": false,
		"consumer_ready": true,
		"buffer_full":    true,
		"buffer_size_2":  true,
	})

	// Transitions
	g.AddEdge("B0", "B1") // Producer sends (0 → 1)
	g.AddEdge("B1", "B2") // Producer sends (1 → 2)
	g.AddEdge("B1", "B0") // Consumer receives (1 → 0)
	g.AddEdge("B2", "B1") // Consumer receives (2 → 1)

	g.SetInitial("B0")

	fmt.Printf("States (%d total):\n", len(g.States()))
	for _, sid := range g.States() {
		name := g.NameOf(sid)
		labels := []string{}
		if g.HasLabel(sid, "buffer_empty") {
			labels = append(labels, "empty")
		}
		if g.HasLabel(sid, "buffer_full") {
			labels = append(labels, "full")
		}
		if g.HasLabel(sid, "producer_ready") {
			labels = append(labels, "P-ready")
		}
		if g.HasLabel(sid, "consumer_ready") {
			labels = append(labels, "C-ready")
		}
		fmt.Printf("  %s: [%s]\n", name, strings.Join(labels, ", "))
	}
	fmt.Println()

	// CTL Properties
	fmt.Println("CTL PROPERTIES")
	fmt.Println(strings.Repeat("-", 78))

	properties := []struct {
		name        string
		formula     Formula
		description string
	}{
		{
			name:        "P1",
			formula:     AG(Not(Atom("buffer_overflow"))),
			description: "Buffer never overflows (should pass)",
		},
		{
			name:        "P2",
			formula:     AG(EF(Atom("producer_ready"))),
			description: "Always possible to eventually produce",
		},
		{
			name:        "P3",
			formula:     AG(EF(Atom("consumer_ready"))),
			description: "Always possible to eventually consume",
		},
		{
			name:        "P4",
			formula:     AG(Or(Atom("producer_ready"), Atom("consumer_ready"))),
			description: "System never deadlocks",
		},
		{
			name:        "P5",
			formula:     EF(Atom("buffer_full")),
			description: "Buffer can become full",
		},
		{
			name:        "P6",
			formula:     EF(Atom("buffer_empty")),
			description: "Buffer can become empty",
		},
	}

	initialStates := NewStateSet()
	initialStates.Add(s0)

	for _, p := range properties {
		satisfying := p.formula.Sat(g)

		allInitialsSatisfy := true
		for id := range initialStates {
			if !satisfying.Contains(id) {
				allInitialsSatisfy = false
				break
			}
		}

		status := "✗ FAIL"
		if allInitialsSatisfy {
			status = "✓ PASS"
		}

		fmt.Printf("%s %s: %s\n", status, p.name, p.description)
	}
	fmt.Println()

	// Mermaid Diagram
	fmt.Println("MERMAID DIAGRAM")
	fmt.Println(strings.Repeat("-", 78))
	fmt.Println(generateMermaid(g))
	fmt.Println()

	fmt.Println("=" + strings.Repeat("=", 78))
}

func generateMermaid(g *Graph) string {
	var sb strings.Builder
	sb.WriteString("stateDiagram-v2\n")
	sb.WriteString("    [*] --> B0\n")
	sb.WriteString("    \n")

	for _, sid := range g.States() {
		name := g.NameOf(sid)
		for _, tid := range g.Succ(sid) {
			target := g.NameOf(tid)
			action := ""
			if g.HasLabel(sid, "buffer_empty") || g.HasLabel(sid, "buffer_size_1") {
				// From B0 or B1, going up means produce
				targetSize := g.NameOf(tid)
				if targetSize > name {
					action = ": produce"
				} else {
					action = ": consume"
				}
			} else {
				action = ": consume"
			}
			sb.WriteString(fmt.Sprintf("    %s --> %s%s\n", name, target, action))
		}
	}

	sb.WriteString("    \n")
	sb.WriteString("    B0: Empty\\nP:ready C:blocked\n")
	sb.WriteString("    B1: Half\\nP:ready C:ready\n")
	sb.WriteString("    B2: Full\\nP:blocked C:ready\n")

	return sb.String()
}
