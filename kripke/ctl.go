package kripke

// CTL evaluator over a finite Kripke graph.
// Independent from World/Process; you just map your states to StateID
// and build Graph.Succ accordingly.

type StateID string

// Graph is a finite Kripke structure: states + successor relation.
type Graph struct {
	States []StateID
	Succ   map[StateID][]StateID // R(s) = Succ[s]
}

// ----- State sets -----

type StateSet map[StateID]struct{}

func NewStateSet() StateSet                                { return make(StateSet) }
func (s StateSet) Has(id StateID) bool                     { _, ok := s[id]; return ok }
func (s StateSet) Add(id StateID)                          { s[id] = struct{}{} }
func (s StateSet) Copy() StateSet                          { out := NewStateSet(); for k := range s { out[k] = struct{}{} }; return out }
func (s StateSet) Size() int                               { return len(s) }
func (s StateSet) ToSlice() []StateID                      { out := make([]StateID, 0, len(s)); for k := range s { out = append(out, k) }; return out }
func (s StateSet) Equals(other StateSet) bool              { if len(s) != len(other) { return false }; for k := range s { if !other.Has(k) { return false } }; return true }
func (s StateSet) Intersect(other StateSet) StateSet       { out := NewStateSet(); for k := range s { if other.Has(k) { out.Add(k) } }; return out }
func (s StateSet) Union(other StateSet) StateSet           { out := s.Copy(); for k := range other { out.Add(k) }; return out }
func (s StateSet) Difference(other StateSet) StateSet      { out := NewStateSet(); for k := range s { if !other.Has(k) { out.Add(k) } }; return out }

// Universe builds a set containing all states in the graph.
func Universe(g *Graph) StateSet {
	u := NewStateSet()
	for _, s := range g.States {
		u.Add(s)
	}
	return u
}

// Pre_E returns predecessors with SOME successor in W:
// Pre_E(W) = { s | ∃ s' . R(s,s') ∧ s' ∈ W }
func Pre_E(W StateSet, g *Graph) StateSet {
	out := NewStateSet()
	for s, succs := range g.Succ {
		for _, s2 := range succs {
			if W.Has(s2) {
				out.Add(s)
				break
			}
		}
	}
	return out
}

// Pre_A returns predecessors whose ALL successors are in W.
// Pre_A(W) = { s | Succ(s) ⊆ W } (vacuously true if Succ(s) is empty).
func Pre_A(W StateSet, g *Graph) StateSet {
	out := NewStateSet()
	for s, succs := range g.Succ {
		if len(succs) == 0 {
			// Convention: dead-ends satisfy AX φ for all φ (or treat differently if desired).
			out.Add(s)
			continue
		}
		all := true
		for _, s2 := range succs {
			if !W.Has(s2) {
				all = false
				break
			}
		}
		if all {
			out.Add(s)
		}
	}
	return out
}

// ----- CTL Formula AST -----

// Formula is a CTL state formula.
// Sat(g) returns the set of states satisfying the formula in graph g.
type Formula interface {
	Sat(g *Graph) StateSet
}

// Atom: an atomic proposition is represented as a set of states
// where it holds. You map your own predicates to this set.
type Atom struct {
	States StateSet
}

func (a Atom) Sat(g *Graph) StateSet {
	// We trust caller to give a subset of g.States; no check here.
	return a.States.Copy()
}

// Not: ¬φ
type Not struct {
	F Formula
}

func (n Not) Sat(g *Graph) StateSet {
	all := Universe(g)
	satF := n.F.Sat(g)
	return all.Difference(satF)
}

// And: (φ ∧ ψ)
type And struct {
	Left, Right Formula
}

func (a And) Sat(g *Graph) StateSet {
	l := a.Left.Sat(g)
	r := a.Right.Sat(g)
	return l.Intersect(r)
}

// Or: (φ ∨ ψ)
type Or struct {
	Left, Right Formula
}

func (o Or) Sat(g *Graph) StateSet {
	l := o.Left.Sat(g)
	r := o.Right.Sat(g)
	return l.Union(r)
}

// EX φ: "there exists a next state where φ holds"
type EX struct {
	F Formula
}

func (e EX) Sat(g *Graph) StateSet {
	satF := e.F.Sat(g)
	return Pre_E(satF, g)
}

// AX φ: "for all next states, φ holds"
type AX struct {
	F Formula
}

func (a AX) Sat(g *Graph) StateSet {
	satF := a.F.Sat(g)
	return Pre_A(satF, g)
}

// EU(p, q): "there exists a path where p holds UNTIL q holds"
type EU struct {
	P, Q Formula
}

func (eu EU) Sat(g *Graph) StateSet {
	satP := eu.P.Sat(g)
	satQ := eu.Q.Sat(g)

	// Least fixpoint:
	// W0 = Sat(Q)
	// W_{i+1} = W_i ∪ (Sat(P) ∩ Pre_E(W_i))
	W := satQ.Copy()
	for {
		pre := Pre_E(W, g).Intersect(satP)
		next := W.Union(pre)
		if next.Equals(W) {
			return W
		}
		W = next
	}
}

// EG φ: "there exists a path where φ holds globally (forever)"
type EG struct {
	F Formula
}

func (eg EG) Sat(g *Graph) StateSet {
	satP := eg.F.Sat(g)

	// Greatest fixpoint:
	// Start with all states where φ holds,
	// iteratively remove states that cannot stay in φ forever
	// (i.e., have no successor in the current set).
	Z := satP.Copy()
	for {
		next := NewStateSet()
		for s := range Z {
			succs := g.Succ[s]
			// keep s if it has at least one successor in Z
			keep := false
			for _, s2 := range succs {
				if Z.Has(s2) {
					keep = true
					break
				}
			}
			if keep {
				next.Add(s)
			}
		}
		if next.Equals(Z) {
			return Z
		}
		Z = next
	}
}

// EF φ: "there exists a path where EVENTUALLY φ" (derived from EU)
type EF struct {
	F Formula
}

func (ef EF) Sat(g *Graph) StateSet {
	// EF φ ≡ E[ true U φ ]
	trueSet := Universe(g)
	return EU{P: Atom{States: trueSet}, Q: ef.F}.Sat(g)
}

// AF φ: "for all paths, EVENTUALLY φ"
// AF φ ≡ ¬EG ¬φ
type AF struct {
	F Formula
}

func (af AF) Sat(g *Graph) StateSet {
	notF := Not{F: af.F}
	egNotF := EG{F: notF}
	bad := egNotF.Sat(g)
	all := Universe(g)
	return all.Difference(bad)
}

// AG φ: "for all paths, φ holds globally"
// AG φ ≡ ¬EF ¬φ
type AG struct {
	F Formula
}

func (ag AG) Sat(g *Graph) StateSet {
	notF := Not{F: ag.F}
	efNotF := EF{F: notF}
	bad := efNotF.Sat(g)
	all := Universe(g)
	return all.Difference(bad)
}

// SatIn evaluates a formula and asks if a given initial state satisfies it.
func SatIn(f Formula, g *Graph, init StateID) bool {
	satSet := f.Sat(g)
	return satSet.Has(init)
}

