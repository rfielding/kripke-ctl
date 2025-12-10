package main

import (
	"fmt"
	"strings"
)

// Complete End-to-End Example: Client-Server with Timeout
//
// This demonstrates:
// 1. English description → Kripke Graph
// 2. CTL property definition
// 3. Model checking and verification
// 4. Mermaid diagram generation

// ========== CTL Implementation ==========

type StateID int
type StateSet map[StateID]struct{}

func NewStateSet() StateSet              { return make(StateSet) }
func (s StateSet) Add(id StateID)        { s[id] = struct{}{} }
func (s StateSet) Contains(id StateID) bool { _, ok := s[id]; return ok }
func (s StateSet) Clone() StateSet {
	out := make(StateSet, len(s))
	for k := range s {
		out[k] = struct{}{}
	}
	return out
}

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
	from, fok := g.nameToID[fromName]
	to, tok := g.nameToID[toName]
	if !fok || !tok {
		panic("AddEdge: unknown state")
	}
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

type AndFormula struct{ Left, Right Formula }

func And(l, r Formula) Formula { return AndFormula{Left: l, Right: r} }
func (a AndFormula) Sat(g *Graph) StateSet {
	L := a.Left.Sat(g)
	R := a.Right.Sat(g)
	res := NewStateSet()
	for _, s := range g.States() {
		if L.Contains(s) && R.Contains(s) {
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

type AFFormula struct{ Inner Formula }

func AF(inner Formula) Formula { return AFFormula{Inner: inner} }
func (f AFFormula) Sat(g *Graph) StateSet {
	return Not(EG(Not(f.Inner))).Sat(g)
}

type EGFormula struct{ Inner Formula }

func EG(inner Formula) Formula { return EGFormula{Inner: inner} }
func (f EGFormula) Sat(g *Graph) StateSet {
	phiSet := f.Inner.Sat(g)
	Z := phiSet.Clone()
	changed := true
	for changed {
		changed = false
		for s := range Z {
			succs := g.Succ(s)
			hasSuccInZ := false
			for _, t := range succs {
				if Z.Contains(t) {
					hasSuccInZ = true
					break
				}
			}
			if !hasSuccInZ {
				delete(Z, s)
				changed = true
			}
		}
	}
	return Z
}

// ========== MAIN ==========

func main() {
	fmt.Println("=" + strings.Repeat("=", 78))
	fmt.Println(" Client-Server System with Timeout - Complete Example")
	fmt.Println("=" + strings.Repeat("=", 78))
	fmt.Println()

	// English Description
	fmt.Println("SYSTEM DESCRIPTION")
	fmt.Println(strings.Repeat("-", 78))
	fmt.Println(`
A client-server system with timeout:
- Client sends requests to the server
- Server processes requests and sends responses
- If no response arrives within 3 time units, the client times out
- After timeout or response, the client can send a new request
`)

	// Build Kripke Graph
	fmt.Println("STATE SPACE")
	fmt.Println(strings.Repeat("-", 78))

	g := NewGraph()

	// States
	s0 := g.AddState("idle_idle_0", map[string]bool{
		"client_idle": true,
		"server_idle": true,
	})

	_ = g.AddState("waiting_processing_0", map[string]bool{
		"request_sent": true,
		"server_busy":  true,
	})

	_ = g.AddState("waiting_processing_1", map[string]bool{
		"request_sent": true,
		"server_busy":  true,
		"timer_1":      true,
	})

	_ = g.AddState("waiting_processing_2", map[string]bool{
		"request_sent": true,
		"server_busy":  true,
		"timer_2":      true,
	})

	_ = g.AddState("idle_idle_response", map[string]bool{
		"client_idle":  true,
		"server_idle":  true,
		"got_response": true,
	})

	_ = g.AddState("idle_idle_timeout", map[string]bool{
		"client_idle": true,
		"server_idle": true,
		"timed_out":   true,
	})

	// Transitions
	g.AddEdge("idle_idle_0", "waiting_processing_0")
	g.AddEdge("waiting_processing_0", "waiting_processing_1")
	g.AddEdge("waiting_processing_1", "waiting_processing_2")
	g.AddEdge("waiting_processing_1", "idle_idle_response")
	g.AddEdge("waiting_processing_2", "idle_idle_response")
	g.AddEdge("waiting_processing_2", "idle_idle_timeout")
	g.AddEdge("idle_idle_response", "idle_idle_0")
	g.AddEdge("idle_idle_timeout", "idle_idle_0")
	g.AddEdge("idle_idle_0", "idle_idle_0") // Can stay idle

	g.SetInitial("idle_idle_0")

	fmt.Printf("States (%d total):\n", len(g.States()))
	for _, sid := range g.States() {
		fmt.Printf("  %s\n", g.NameOf(sid))
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
			formula:     AG(Or(Atom("client_idle"), Atom("request_sent"))),
			description: "Client is always either idle or waiting",
		},
		{
			name:        "P2",
			formula:     EF(Atom("got_response")),
			description: "It's possible to get a response",
		},
		{
			name:        "P3",
			formula:     EF(Atom("timed_out")),
			description: "It's possible to timeout",
		},
		{
			name:        "P4",
			formula:     AG(Not(And(Atom("got_response"), Atom("timed_out")))),
			description: "Never both response and timeout",
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
	sb.WriteString("    [*] --> idle_idle_0\n")
	sb.WriteString("    \n")

	for _, sid := range g.States() {
		name := g.NameOf(sid)
		for _, tid := range g.Succ(sid) {
			target := g.NameOf(tid)
			if name != target {
				sb.WriteString(fmt.Sprintf("    %s --> %s\n", name, target))
			}
		}
	}

	sb.WriteString("    \n")
	sb.WriteString("    idle_idle_0: Idle / Ready\n")
	sb.WriteString("    waiting_processing_0: Sent / Timer=0\n")
	sb.WriteString("    waiting_processing_1: Wait / Timer=1\n")
	sb.WriteString("    waiting_processing_2: Wait / Timer=2\n")
	sb.WriteString("    idle_idle_response: Response Received\n")
	sb.WriteString("    idle_idle_timeout: Timeout!\n")

	return sb.String()
}
