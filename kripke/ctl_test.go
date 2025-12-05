package kripke

import (
	"reflect"
	"testing"
)

// Helpers for tests

func stateSet(states ...StateID) StateSet {
	ss := NewStateSet()
	for _, s := range states {
		ss.Add(s)
	}
	return ss
}

func equalStateSets(a, b StateSet) bool {
	return a.Equals(b)
}

func mustEqualStateSets(t *testing.T, name string, got, want StateSet) {
	t.Helper()
	if !equalStateSets(got, want) {
		t.Fatalf("%s: got %v, want %v", name, got.ToSlice(), want.ToSlice())
	}
}

func mustBool(t *testing.T, name string, got, want bool) {
	t.Helper()
	if got != want {
		t.Fatalf("%s: got %v, want %v", name, got, want)
	}
}

// --- Basic Boolean connectives ---

func TestCTL_BooleanConnectives(t *testing.T) {
	// Graph with three states, no edges (successors don't matter for boolean connectives)
	s0 := StateID("s0")
	s1 := StateID("s1")
	s2 := StateID("s2")

	g := &Graph{
		States: []StateID{s0, s1, s2},
		Succ: map[StateID][]StateID{
			s0: {},
			s1: {},
			s2: {},
		},
	}

	pStates := stateSet(s0, s1) // p holds at s0, s1
	qStates := stateSet(s1, s2) // q holds at s1, s2
	p := Atom{States: pStates}
	q := Atom{States: qStates}

	// ¬p
	notP := Not{F: p}
	wantNotP := stateSet(s2)
	gotNotP := notP.Sat(g)
	mustEqualStateSets(t, "Not p", gotNotP, wantNotP)

	// p ∧ q = {s1}
	andPQ := And{Left: p, Right: q}
	wantAnd := stateSet(s1)
	gotAnd := andPQ.Sat(g)
	mustEqualStateSets(t, "p AND q", gotAnd, wantAnd)

	// p ∨ q = {s0, s1, s2}
	orPQ := Or{Left: p, Right: q}
	wantOr := stateSet(s0, s1, s2)
	gotOr := orPQ.Sat(g)
	mustEqualStateSets(t, "p OR q", gotOr, wantOr)
}

// --- EX / AX ---

func TestCTL_EX_AX(t *testing.T) {
	// Chain: s0 -> s1 -> s2 -> s2
	s0 := StateID("s0")
	s1 := StateID("s1")
	s2 := StateID("s2")

	g := &Graph{
		States: []StateID{s0, s1, s2},
		Succ: map[StateID][]StateID{
			s0: {s1},
			s1: {s2},
			s2: {s2},
		},
	}

	// p holds at s1, s2
	pStates := stateSet(s1, s2)
	p := Atom{States: pStates}

	// EX p: states that have SOME successor where p holds
	exP := EX{F: p}
	wantEX := stateSet(s0, s1, s2)
	gotEX := exP.Sat(g)
	mustEqualStateSets(t, "EX p", gotEX, wantEX)

	// AX p: states where ALL successors satisfy p
	axP := AX{F: p}
	wantAX := stateSet(s0, s1, s2)
	gotAX := axP.Sat(g)
	mustEqualStateSets(t, "AX p", gotAX, wantAX)
}

// --- EF / AF ---

func TestCTL_EF_AF(t *testing.T) {
	// Chain: s0 -> s1 -> s2 -> s2
	s0 := StateID("s0")
	s1 := StateID("s1")
	s2 := StateID("s2")

	g := &Graph{
		States: []StateID{s0, s1, s2},
		Succ: map[StateID][]StateID{
			s0: {s1},
			s1: {s2},
			s2: {s2},
		},
	}

	// p holds only at s2
	pStates := stateSet(s2)
	p := Atom{States: pStates}

	// EF p: there exists a path where p eventually holds.
	efP := EF{F: p}
	wantEF := stateSet(s0, s1, s2)
	gotEF := efP.Sat(g)
	mustEqualStateSets(t, "EF p", gotEF, wantEF)

	// AF p: along all paths, p eventually holds.
	afP := AF{F: p}
	wantAF := stateSet(s0, s1, s2)
	gotAF := afP.Sat(g)
	mustEqualStateSets(t, "AF p", gotAF, wantAF)
}

// --- EG / AG ---

func TestCTL_EG_AG(t *testing.T) {
	// Graph:
	//   s0 -> s0   (self-loop)
	//   s1 -> s2
	//   s2 -> s2   (self-loop)
	s0 := StateID("s0")
	s1 := StateID("s1")
	s2 := StateID("s2")

	g := &Graph{
		States: []StateID{s0, s1, s2},
		Succ: map[StateID][]StateID{
			s0: {s0},
			s1: {s2},
			s2: {s2},
		},
	}

	// p holds at s0 and s2
	pStates := stateSet(s0, s2)
	p := Atom{States: pStates}

	// EG p: there exists a path where p holds globally.
	egP := EG{F: p}
	wantEG := stateSet(s0, s2)
	gotEG := egP.Sat(g)
	mustEqualStateSets(t, "EG p", gotEG, wantEG)

	// AG p: we'll check via Sat and Has directly, to avoid assuming SatIn exists.
	agP := AG{F: p}
	satAG := agP.Sat(g)

	mustBool(t, "AG p at s0 (self-loop p only)", satAG.Has(s0), true)
	mustBool(t, "AG p at s2 (self-loop p only)", satAG.Has(s2), true)
	mustBool(t, "AG p at s1 (immediate ¬p)", satAG.Has(s1), false)
}

// --- EU ---

func TestCTL_EU(t *testing.T) {
	// Graph:
	//   s0 -> s1
	//   s1 -> s2
	//   s2 -> s2
	s0 := StateID("s0")
	s1 := StateID("s1")
	s2 := StateID("s2")

	g := &Graph{
		States: []StateID{s0, s1, s2},
		Succ: map[StateID][]StateID{
			s0: {s1},
			s1: {s2},
			s2: {s2},
		},
	}

	// p holds at s0, s1
	// q holds at s2
	pStates := stateSet(s0, s1)
	qStates := stateSet(s2)
	p := Atom{States: pStates}
	q := Atom{States: qStates}

	// EU(p, q): exists a path where p holds UNTIL q holds.
	eu := EU{P: p, Q: q}
	wantEU := stateSet(s0, s1, s2)
	gotEU := eu.Sat(g)
	mustEqualStateSets(t, "EU(p, q)", gotEU, wantEU)
}

// --- SatIn-like sanity without SatIn() dependence ---

func TestCTL_EF_SatPerState(t *testing.T) {
	// Simple graph with two states s0 -> s1
	s0 := StateID("s0")
	s1 := StateID("s1")

	g := &Graph{
		States: []StateID{s0, s1},
		Succ: map[StateID][]StateID{
			s0: {s1},
			s1: {s1},
		},
	}

	pStates := stateSet(s1)
	p := Atom{States: pStates}

	efP := EF{F: p} // there exists a path eventually p
	satSet := efP.Sat(g)

	mustBool(t, "EF p should hold at s0", satSet.Has(s0), true)
	mustBool(t, "EF p should hold at s1", satSet.Has(s1), true)
}

// --- Meta-test: Universe/Pre_E/Pre_A sanity ---

func TestCTL_PreOperators(t *testing.T) {
	// Graph:
	//   a -> b, c
	//   b -> c
	//   c -> (none)
	a := StateID("a")
	b := StateID("b")
	c := StateID("c")

	g := &Graph{
		States: []StateID{a, b, c},
		Succ: map[StateID][]StateID{
			a: {b, c},
			b: {c},
			c: {},
		},
	}

	W := stateSet(c)

	preE := Pre_E(W, g)
	// Pre_E({c}) = {a,b} because:
	//   a has succ c
	//   b has succ c
	//   c has no succ in W
	wantPreE := stateSet(a, b)
	if !reflect.DeepEqual(preE.ToSlice(), wantPreE.ToSlice()) && !preE.Equals(wantPreE) {
		t.Fatalf("Pre_E: got %v, want %v", preE.ToSlice(), wantPreE.ToSlice())
	}

	preA := Pre_A(W, g)
	// Pre_A({c}) = {b,c} because:
	//   a has succ {b,c} but b ∉ W -> not all successors in W
	//   b has succ {c} and c ∈ W  -> included
	//   c has no successors; by convention, vacuously included
	wantPreA := stateSet(b, c)
	if !preA.Equals(wantPreA) {
		t.Fatalf("Pre_A: got %v, want %v", preA.ToSlice(), wantPreA.ToSlice())
	}
}

