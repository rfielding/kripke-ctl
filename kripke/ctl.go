package kripke

import (
	"fmt"
)

// CTLFormula represents a Computational Tree Logic formula
type CTLFormula interface {
	Evaluate(state State, model *Model) bool
	String() string
}

// AtomicProp represents an atomic proposition
type AtomicProp struct {
	Name string
}

func (a *AtomicProp) Evaluate(state State, model *Model) bool {
	return state.HasProperty(a.Name)
}

func (a *AtomicProp) String() string {
	return a.Name
}

// Not represents negation: ¬φ
type Not struct {
	Formula CTLFormula
}

func (n *Not) Evaluate(state State, model *Model) bool {
	return !n.Formula.Evaluate(state, model)
}

func (n *Not) String() string {
	return fmt.Sprintf("¬(%s)", n.Formula.String())
}

// And represents conjunction: φ ∧ ψ
type And struct {
	Left, Right CTLFormula
}

func (a *And) Evaluate(state State, model *Model) bool {
	return a.Left.Evaluate(state, model) && a.Right.Evaluate(state, model)
}

func (a *And) String() string {
	return fmt.Sprintf("(%s ∧ %s)", a.Left.String(), a.Right.String())
}

// Or represents disjunction: φ ∨ ψ
type Or struct {
	Left, Right CTLFormula
}

func (o *Or) Evaluate(state State, model *Model) bool {
	return o.Left.Evaluate(state, model) || o.Right.Evaluate(state, model)
}

func (o *Or) String() string {
	return fmt.Sprintf("(%s ∨ %s)", o.Left.String(), o.Right.String())
}

// EX represents "Exists neXt": EX φ
// There exists a path where φ holds in the next state
type EX struct {
	Formula CTLFormula
}

func (ex *EX) Evaluate(state State, model *Model) bool {
	successors := model.GetSuccessors(state)
	for _, successor := range successors {
		if ex.Formula.Evaluate(successor, model) {
			return true
		}
	}
	return false
}

func (ex *EX) String() string {
	return fmt.Sprintf("EX(%s)", ex.Formula.String())
}

// AX represents "All neXt": AX φ
// For all paths, φ holds in the next state
type AX struct {
	Formula CTLFormula
}

func (ax *AX) Evaluate(state State, model *Model) bool {
	successors := model.GetSuccessors(state)
	if len(successors) == 0 {
		return false
	}
	for _, successor := range successors {
		if !ax.Formula.Evaluate(successor, model) {
			return false
		}
	}
	return true
}

func (ax *AX) String() string {
	return fmt.Sprintf("AX(%s)", ax.Formula.String())
}

// EF represents "Exists Finally": EF φ
// There exists a path where φ eventually holds
type EF struct {
	Formula CTLFormula
}

func (ef *EF) Evaluate(state State, model *Model) bool {
	visited := make(map[string]bool)
	return ef.evaluateHelper(state, model, visited)
}

func (ef *EF) evaluateHelper(state State, model *Model, visited map[string]bool) bool {
	stateID := state.ID()
	if visited[stateID] {
		return false
	}
	visited[stateID] = true

	if ef.Formula.Evaluate(state, model) {
		return true
	}

	successors := model.GetSuccessors(state)
	for _, successor := range successors {
		if ef.evaluateHelper(successor, model, visited) {
			return true
		}
	}
	return false
}

func (ef *EF) String() string {
	return fmt.Sprintf("EF(%s)", ef.Formula.String())
}

// AF represents "All Finally": AF φ
// For all paths, φ eventually holds
type AF struct {
	Formula CTLFormula
}

func (af *AF) Evaluate(state State, model *Model) bool {
	visited := make(map[string]bool)
	return af.evaluateHelper(state, model, visited)
}

func (af *AF) evaluateHelper(state State, model *Model, visited map[string]bool) bool {
	stateID := state.ID()
	if visited[stateID] {
		return false
	}
	visited[stateID] = true

	if af.Formula.Evaluate(state, model) {
		return true
	}

	successors := model.GetSuccessors(state)
	if len(successors) == 0 {
		return false
	}

	for _, successor := range successors {
		if !af.evaluateHelper(successor, model, visited) {
			return false
		}
	}
	return true
}

func (af *AF) String() string {
	return fmt.Sprintf("AF(%s)", af.Formula.String())
}

// EG represents "Exists Globally": EG φ
// There exists a path where φ always holds
type EG struct {
	Formula CTLFormula
}

func (eg *EG) Evaluate(state State, model *Model) bool {
	visited := make(map[string]bool)
	return eg.evaluateHelper(state, model, visited)
}

func (eg *EG) evaluateHelper(state State, model *Model, visited map[string]bool) bool {
	stateID := state.ID()
	if !eg.Formula.Evaluate(state, model) {
		return false
	}

	if visited[stateID] {
		return true // Cycle detected where formula holds
	}
	visited[stateID] = true

	successors := model.GetSuccessors(state)
	if len(successors) == 0 {
		return true
	}

	for _, successor := range successors {
		if eg.evaluateHelper(successor, model, visited) {
			return true
		}
	}
	return false
}

func (eg *EG) String() string {
	return fmt.Sprintf("EG(%s)", eg.Formula.String())
}

// AG represents "All Globally": AG φ
// For all paths, φ always holds
type AG struct {
	Formula CTLFormula
}

func (ag *AG) Evaluate(state State, model *Model) bool {
	visited := make(map[string]bool)
	return ag.evaluateHelper(state, model, visited)
}

func (ag *AG) evaluateHelper(state State, model *Model, visited map[string]bool) bool {
	stateID := state.ID()
	if !ag.Formula.Evaluate(state, model) {
		return false
	}

	if visited[stateID] {
		return true // Cycle detected where formula holds
	}
	visited[stateID] = true

	successors := model.GetSuccessors(state)
	for _, successor := range successors {
		if !ag.evaluateHelper(successor, model, visited) {
			return false
		}
	}
	return true
}

func (ag *AG) String() string {
	return fmt.Sprintf("AG(%s)", ag.Formula.String())
}

// EU represents "Exists Until": E[φ U ψ]
// There exists a path where φ holds until ψ holds
type EU struct {
	Left, Right CTLFormula
}

func (eu *EU) Evaluate(state State, model *Model) bool {
	visited := make(map[string]bool)
	return eu.evaluateHelper(state, model, visited)
}

func (eu *EU) evaluateHelper(state State, model *Model, visited map[string]bool) bool {
	stateID := state.ID()
	if visited[stateID] {
		return false
	}
	visited[stateID] = true

	if eu.Right.Evaluate(state, model) {
		return true
	}

	if !eu.Left.Evaluate(state, model) {
		return false
	}

	successors := model.GetSuccessors(state)
	for _, successor := range successors {
		if eu.evaluateHelper(successor, model, visited) {
			return true
		}
	}
	return false
}

func (eu *EU) String() string {
	return fmt.Sprintf("E[%s U %s]", eu.Left.String(), eu.Right.String())
}

// AU represents "All Until": A[φ U ψ]
// For all paths, φ holds until ψ holds
type AU struct {
	Left, Right CTLFormula
}

func (au *AU) Evaluate(state State, model *Model) bool {
	visited := make(map[string]bool)
	return au.evaluateHelper(state, model, visited)
}

func (au *AU) evaluateHelper(state State, model *Model, visited map[string]bool) bool {
	stateID := state.ID()
	if visited[stateID] {
		return false
	}
	visited[stateID] = true

	if au.Right.Evaluate(state, model) {
		return true
	}

	if !au.Left.Evaluate(state, model) {
		return false
	}

	successors := model.GetSuccessors(state)
	if len(successors) == 0 {
		return false
	}

	for _, successor := range successors {
		if !au.evaluateHelper(successor, model, visited) {
			return false
		}
	}
	return true
}

func (au *AU) String() string {
	return fmt.Sprintf("A[%s U %s]", au.Left.String(), au.Right.String())
}
