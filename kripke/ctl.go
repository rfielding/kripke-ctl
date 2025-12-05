package kripke

// ---------- Core Kripke Graph Types ----------

type StateID int

// StateSet is a set of states.
type StateSet map[StateID]struct{}

func NewStateSet() StateSet { return make(StateSet) }

func (s StateSet) Add(id StateID) {
	s[id] = struct{}{}
}

func (s StateSet) Contains(id StateID) bool {
	_, ok := s[id]
	return ok
}

func (s StateSet) Clone() StateSet {
	out := make(StateSet, len(s))
	for k := range s {
		out[k] = struct{}{}
	}
	return out
}

func (s StateSet) Equal(other StateSet) bool {
	if len(s) != len(other) {
		return false
	}
	for k := range s {
		if _, ok := other[k]; !ok {
			return false
		}
	}
	return true
}

// Graph is a finite Kripke structure: states, initial states,
// successor edges, and atomic proposition labels.
type Graph struct {
	nextID   int
	nameToID map[string]StateID
	idToName map[StateID]string
	labels   map[StateID]map[string]bool
	succ     map[StateID][]StateID
	init     []StateID
}

// NewGraph constructs an empty Graph.
func NewGraph() *Graph {
	return &Graph{
		nextID:   0,
		nameToID: make(map[string]StateID),
		idToName: make(map[StateID]string),
		labels:   make(map[StateID]map[string]bool),
		succ:     make(map[StateID][]StateID),
		init:     make([]StateID, 0),
	}
}

// AddState adds a state with the given name and AP labels.
// Returns its StateID.
func (g *Graph) AddState(name string, lbls map[string]bool) StateID {
	if _, exists := g.nameToID[name]; exists {
		panic("AddState: duplicate state name " + name)
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

func (g *Graph) ensureState(name string) StateID {
	if id, ok := g.nameToID[name]; ok {
		return id
	}
	return g.AddState(name, nil)
}

// AddEdge adds a transition from 'fromName' to 'toName'.
// States are auto-created if not already present.
func (g *Graph) AddEdge(fromName, toName string) {
	from := g.ensureState(fromName)
	to := g.ensureState(toName)
	g.succ[from] = append(g.succ[from], to)
}

// SetInitial marks a named state as initial.
func (g *Graph) SetInitial(name string) {
	id := g.ensureState(name)
	g.init = append(g.init, id)
}

// InitialStates returns the initial state IDs.
func (g *Graph) InitialStates() []StateID {
	out := make([]StateID, len(g.init))
	copy(out, g.init)
	return out
}

// States returns all defined states.
func (g *Graph) States() []StateID {
	out := make([]StateID, 0, len(g.idToName))
	for id := range g.idToName {
		out = append(out, id)
	}
	return out
}

// Succ returns the successors of a state.
func (g *Graph) Succ(s StateID) []StateID {
	return g.succ[s]
}

// HasLabel checks if state s has atomic proposition 'prop'.
func (g *Graph) HasLabel(s StateID, prop string) bool {
	lbls := g.labels[s]
	return lbls[prop]
}

// NameOf returns the human-readable name of a state.
func (g *Graph) NameOf(s StateID) string {
	return g.idToName[s]
}

// ---------- CTL formulas ----------

type Formula interface {
	Sat(g *Graph) StateSet
}

// ----- atomic proposition -----

type AtomFormula struct {
	Prop string
}

func Atom(prop string) Formula {
	return AtomFormula{Prop: prop}
}

func (a AtomFormula) Sat(g *Graph) StateSet {
	res := NewStateSet()
	for _, s := range g.States() {
		if g.HasLabel(s, a.Prop) {
			res.Add(s)
		}
	}
	return res
}

// ----- boolean connectives -----

type NotFormula struct {
	Inner Formula
}

func Not(inner Formula) Formula {
	return NotFormula{Inner: inner}
}

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

type AndFormula struct {
	Left, Right Formula
}

func And(l, r Formula) Formula {
	return AndFormula{Left: l, Right: r}
}

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

type OrFormula struct {
	Left, Right Formula
}

func Or(l, r Formula) Formula {
	return OrFormula{Left: l, Right: r}
}

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

// Implies(p, q) = Or(Not(p), q).
func Implies(p, q Formula) Formula {
	return Or(Not(p), q)
}

// ----- EX and AX -----

type EXFormula struct {
	Inner Formula
}

func EX(inner Formula) Formula {
	return EXFormula{Inner: inner}
}

func (e EXFormula) Sat(g *Graph) StateSet {
	target := e.Inner.Sat(g)
	res := NewStateSet()
	for _, s := range g.States() {
		for _, t := range g.Succ(s) {
			if target.Contains(t) {
				res.Add(s)
				break
			}
		}
	}
	return res
}

type AXFormula struct {
	Inner Formula
}

func AX(inner Formula) Formula {
	return AXFormula{Inner: inner}
}

func (a AXFormula) Sat(g *Graph) StateSet {
	target := a.Inner.Sat(g)
	res := NewStateSet()
	for _, s := range g.States() {
		succs := g.Succ(s)
		if len(succs) == 0 {
			// Vacuum truth: if no successors, AX φ holds.
			res.Add(s)
			continue
		}
		allGood := true
		for _, t := range succs {
			if !target.Contains(t) {
				allGood = false
				break
			}
		}
		if allGood {
			res.Add(s)
		}
	}
	return res
}

// ----- EU (E[φ U ψ]) -----

type EUFormula struct {
	Phi Formula
	Psi Formula
}

func EU(phi, psi Formula) Formula {
	return EUFormula{Phi: phi, Psi: psi}
}

func (f EUFormula) Sat(g *Graph) StateSet {
	phiSet := f.Phi.Sat(g)
	psiSet := f.Psi.Sat(g)

	// Standard least fixpoint algorithm:
	// X0 = ψ
	// Xi+1 = Xi ∪ { s | s ∈ φ and ∃ succ in Xi }
	X := psiSet.Clone()
	changed := true
	for changed {
		changed = false
		for _, s := range g.States() {
			if X.Contains(s) {
				continue
			}
			if !phiSet.Contains(s) {
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

// ----- EF (exists eventually φ) -----

type EFFormula struct {
	Inner Formula
}

func EF(inner Formula) Formula {
	return EFFormula{Inner: inner}
}

func (f EFFormula) Sat(g *Graph) StateSet {
	target := f.Inner.Sat(g)

	// Backward reachability:
	// X0 = target
	// Xi+1 = Xi ∪ { s | ∃ succ in Xi }
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

// ----- EG (exists globally φ) -----

type EGFormula struct {
	Inner Formula
}

func EG(inner Formula) Formula {
	return EGFormula{Inner: inner}
}

func (f EGFormula) Sat(g *Graph) StateSet {
	phiSet := f.Inner.Sat(g)

	// Standard algorithm:
	// Start with Z = { s | s satisfies φ }
	// Repeatedly remove any s in Z that has no successor in Z.
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
			// No successors in Z => cannot sustain φ globally along a path
			if !hasSuccInZ {
				delete(Z, s)
				changed = true
			}
		}
	}
	return Z
}

// ----- AF and AG via dualities -----

type AFFormula struct {
	Inner Formula
}

func AF(inner Formula) Formula {
	return AFFormula{Inner: inner}
}

func (f AFFormula) Sat(g *Graph) StateSet {
	// AF φ = ¬ EG ¬φ
	return Not(EG(Not(f.Inner))).Sat(g)
}

type AGFormula struct {
	Inner Formula
}

func AG(inner Formula) Formula {
	return AGFormula{Inner: inner}
}

func (f AGFormula) Sat(g *Graph) StateSet {
	// AG φ = ¬ EF ¬φ
	return Not(EF(Not(f.Inner))).Sat(g)
}
