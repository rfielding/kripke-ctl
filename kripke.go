package main

import (
	"fmt"
)

// State represents a state in a Kripke structure
type State string

// Proposition represents an atomic proposition
type Proposition string

// KripkeStructure represents a labeled transition system
type KripkeStructure struct {
	States      []State
	InitialState State
	Transitions map[State][]State
	Labeling    map[State][]Proposition
}

// NewKripkeStructure creates a new Kripke structure
func NewKripkeStructure(initial State) *KripkeStructure {
	return &KripkeStructure{
		States:       []State{initial},
		InitialState: initial,
		Transitions:  make(map[State][]State),
		Labeling:     make(map[State][]Proposition),
	}
}

// AddState adds a new state to the Kripke structure
func (k *KripkeStructure) AddState(s State) {
	for _, existing := range k.States {
		if existing == s {
			return
		}
	}
	k.States = append(k.States, s)
}

// AddTransition adds a transition from one state to another
func (k *KripkeStructure) AddTransition(from, to State) {
	k.AddState(from)
	k.AddState(to)
	k.Transitions[from] = append(k.Transitions[from], to)
}

// AddLabel adds a proposition label to a state
func (k *KripkeStructure) AddLabel(s State, p Proposition) {
	k.AddState(s)
	k.Labeling[s] = append(k.Labeling[s], p)
}

// HasLabel checks if a state has a specific proposition label
func (k *KripkeStructure) HasLabel(s State, p Proposition) bool {
	labels, ok := k.Labeling[s]
	if !ok {
		return false
	}
	for _, label := range labels {
		if label == p {
			return true
		}
	}
	return false
}

// GetSuccessors returns all successor states of a given state
func (k *KripkeStructure) GetSuccessors(s State) []State {
	return k.Transitions[s]
}

// String returns a string representation of the Kripke structure
func (k *KripkeStructure) String() string {
	result := fmt.Sprintf("Kripke Structure:\n")
	result += fmt.Sprintf("  Initial State: %s\n", k.InitialState)
	result += fmt.Sprintf("  States: %v\n", k.States)
	result += fmt.Sprintf("  Transitions:\n")
	for from, tos := range k.Transitions {
		for _, to := range tos {
			result += fmt.Sprintf("    %s -> %s\n", from, to)
		}
	}
	result += fmt.Sprintf("  Labeling:\n")
	for state, props := range k.Labeling {
		result += fmt.Sprintf("    %s: %v\n", state, props)
	}
	return result
}
