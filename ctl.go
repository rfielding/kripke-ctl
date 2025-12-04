package main

import "fmt"

// CTLFormula represents a CTL formula
type CTLFormula interface {
	String() string
}

// AtomicProp represents an atomic proposition
type AtomicProp struct {
	Prop Proposition
}

func (a AtomicProp) String() string {
	return string(a.Prop)
}

// Not represents negation
type Not struct {
	Formula CTLFormula
}

func (n Not) String() string {
	return fmt.Sprintf("¬%s", n.Formula)
}

// And represents conjunction
type And struct {
	Left, Right CTLFormula
}

func (a And) String() string {
	return fmt.Sprintf("(%s ∧ %s)", a.Left, a.Right)
}

// Or represents disjunction
type Or struct {
	Left, Right CTLFormula
}

func (o Or) String() string {
	return fmt.Sprintf("(%s ∨ %s)", o.Left, o.Right)
}

// Implies represents implication
type Implies struct {
	Left, Right CTLFormula
}

func (i Implies) String() string {
	return fmt.Sprintf("(%s → %s)", i.Left, i.Right)
}

// EX represents "there exists a next state where"
type EX struct {
	Formula CTLFormula
}

func (e EX) String() string {
	return fmt.Sprintf("EX %s", e.Formula)
}

// AX represents "in all next states"
type AX struct {
	Formula CTLFormula
}

func (a AX) String() string {
	return fmt.Sprintf("AX %s", a.Formula)
}

// EF represents "there exists a path where eventually"
type EF struct {
	Formula CTLFormula
}

func (e EF) String() string {
	return fmt.Sprintf("EF %s", e.Formula)
}

// AF represents "on all paths eventually"
type AF struct {
	Formula CTLFormula
}

func (a AF) String() string {
	return fmt.Sprintf("AF %s", a.Formula)
}

// EG represents "there exists a path where always"
type EG struct {
	Formula CTLFormula
}

func (e EG) String() string {
	return fmt.Sprintf("EG %s", e.Formula)
}

// AG represents "on all paths always"
type AG struct {
	Formula CTLFormula
}

func (a AG) String() string {
	return fmt.Sprintf("AG %s", a.Formula)
}

// EU represents "there exists a path where p until q"
type EU struct {
	Left, Right CTLFormula
}

func (e EU) String() string {
	return fmt.Sprintf("E[%s U %s]", e.Left, e.Right)
}

// AU represents "on all paths p until q"
type AU struct {
	Left, Right CTLFormula
}

func (a AU) String() string {
	return fmt.Sprintf("A[%s U %s]", a.Left, a.Right)
}

// True represents the boolean constant true
type True struct{}

func (t True) String() string {
	return "⊤"
}

// False represents the boolean constant false
type False struct{}

func (f False) String() string {
	return "⊥"
}
