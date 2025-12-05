// kripke/ctl_test.go
package kripke

import "testing"

// simpleGraph builds a tiny Kripke graph:
//
//   s0: {}
//   s1: {p}
//   s2: {q}
//
// Edges:
//   s0 -> s1
//   s1 -> s2
//   s2 -> s2
//
// Initial state: s0
func simpleGraph() *Graph {
	g := NewGraph()

	g.AddState("s0", map[string]bool{})
	g.AddState("s1", map[string]bool{"p": true})
	g.AddState("s2", map[string]bool{"q": true})

	g.AddEdge("s0", "s1")
	g.AddEdge("s1", "s2")
	g.AddEdge("s2", "s2")

	g.SetInitial("s0")
	return g
}

// stateNames is a helper to turn a StateSet into a name->true map
// so failures are easier to read.
func stateNames(g *Graph, ss StateSet) map[string]bool {
	out := make(map[string]bool)
	for s := range ss {
		out[g.NameOf(s)] = true
	}
	return out
}

// TestEF checks EF(p) on the simple line graph.
func TestEF(t *testing.T) {
	g := simpleGraph()
	p := Atom("p")

	// EF(p) means: "there exists a path where eventually p holds".
	//
	// From s0: s0 -> s1, and s1 has p -> EF(p) true at s0.
	// From s1: already at p, so EF(p) true at s1.
	// From s2: you can never reach p, so EF(p) false at s2.
	res := EF(p).Sat(g)
	names := stateNames(g, res)

	if !names["s0"] || !names["s1"] {
		t.Fatalf("expected EF(p) to hold at s0 and s1, got %#v", names)
	}
	if names["s2"] {
		t.Fatalf("expected EF(p) NOT to hold at s2, got %#v", names)
	}
}

// TestAG checks AG(q) on the simple line graph.
func TestAG(t *testing.T) {
	g := simpleGraph()
	q := Atom("q")

	// AG(q) means: "on all paths, at all times, q holds".
	//
	// s0: {}      -> q false at s0 -> AG(q) false at s0
	// s1: {p}     -> q false at s1 -> AG(q) false at s1
	// s2: {q}     -> only successor is s2, where q still holds
	//               -> AG(q) true at s2
	res := AG(q).Sat(g)
	names := stateNames(g, res)

	if !names["s2"] {
		t.Fatalf("expected AG(q) to hold at s2, got %#v", names)
	}
	if names["s0"] || names["s1"] {
		t.Fatalf("expected AG(q) to be false at s0 and s1, got %#v", names)
	}
}

// TestEU checks EU(p, q) on the simple line graph.
func TestEU(t *testing.T) {
	g := simpleGraph()
	p := Atom("p")
	q := Atom("q")

	// EU(p,q) means:
	//   "there exists a path where eventually q holds, and
	//    p holds on all states strictly before the first q".
	//
	// Standard CTL semantics allow a length-0 path when q is already
	// true in the current state. So:
	//   s1: p true, path s1->s2 where q holds -> EU(p,q) true
	//   s2: q already true, length-0 path     -> EU(p,q) true
	//   s0: neither p nor q, can't start a valid path -> EU(p,q) false
	res := EU(p, q).Sat(g)
	names := stateNames(g, res)

	if !names["s1"] || !names["s2"] {
		t.Fatalf("expected EU(p,q) to hold at s1 and s2, got %#v", names)
	}
	if names["s0"] {
		t.Fatalf("expected EU(p,q) NOT to hold at s0, got %#v", names)
	}
}
