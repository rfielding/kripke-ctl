package main

import (
	"fmt"
	"github.com/rfielding/kripke-ctl/kripke"
)

// ============================================================================
// NESTED STATES EXAMPLE
// ============================================================================
//
// This example demonstrates how multiple predicates can match simultaneously,
// creating "nested states". If x=5, both "x>0" and "x<20" are true, so the
// actor is in BOTH states. This means TWO transitions are candidates.
//
// The engine collects all candidates and picks one uniformly at random.
//
// ============================================================================

// Shared state between transitions
type Counter struct {
	Value int
}

// ============================================================================
// TRANSITION 1: Decrement (predicate: x > 0)
// ============================================================================

type DecrementTransition struct {
	IDstr   string
	Counter *Counter
}

func (d *DecrementTransition) ID() string { return d.IDstr }

func (d *DecrementTransition) Ready(w *kripke.World) []kripke.Step {
	// Predicate: x > 0
	if d.Counter.Value <= 0 {
		return nil  // Predicate doesn't match
	}
	
	// Predicate matches - return ONE step
	return []kripke.Step{
		func(w *kripke.World) {
			d.Counter.Value--  // x' = x - 1
			fmt.Printf("DECREMENT: %d → %d\n", d.Counter.Value+1, d.Counter.Value)
		},
	}
}

// ============================================================================
// TRANSITION 2: Increment (predicate: x < 20)
// ============================================================================

type IncrementTransition struct {
	IDstr   string
	Counter *Counter
}

func (i *IncrementTransition) ID() string { return i.IDstr }

func (i *IncrementTransition) Ready(w *kripke.World) []kripke.Step {
	// Predicate: x < 20
	if i.Counter.Value >= 20 {
		return nil  // Predicate doesn't match
	}
	
	// Predicate matches - return ONE step
	return []kripke.Step{
		func(w *kripke.World) {
			i.Counter.Value++  // x' = x + 1
			fmt.Printf("INCREMENT: %d → %d\n", i.Counter.Value-1, i.Counter.Value)
		},
	}
}

// ============================================================================
// TRANSITION 3: Reset (predicate: x == 20)
// ============================================================================

type ResetTransition struct {
	IDstr   string
	Counter *Counter
}

func (r *ResetTransition) ID() string { return r.IDstr }

func (r *ResetTransition) Ready(w *kripke.World) []kripke.Step {
	// Predicate: x == 20
	if r.Counter.Value != 20 {
		return nil  // Predicate doesn't match
	}
	
	// Predicate matches - return ONE step
	return []kripke.Step{
		func(w *kripke.World) {
			fmt.Printf("RESET: %d → 10\n", r.Counter.Value)
			r.Counter.Value = 10  // x' = 10
		},
	}
}

// ============================================================================
// MAIN
// ============================================================================

func main() {
	// Shared state
	counter := &Counter{Value: 5}
	
	// THREE Process objects = THREE transitions
	decrement := &DecrementTransition{IDstr: "decrement", Counter: counter}
	increment := &IncrementTransition{IDstr: "increment", Counter: counter}
	reset := &ResetTransition{IDstr: "reset", Counter: counter}
	
	w := kripke.NewWorld(
		[]kripke.Process{decrement, increment, reset},
		[]*kripke.Channel{},
		42,
	)
	
	fmt.Printf("Starting with x = %d\n\n", counter.Value)
	
	// When x = 5:
	//   - Decrement predicate (x > 0):  TRUE  ✓ CANDIDATE
	//   - Increment predicate (x < 20): TRUE  ✓ CANDIDATE
	//   - Reset predicate (x == 20):    FALSE ✗ NOT A CANDIDATE
	//
	// So TWO transitions are candidates!
	// Engine picks one uniformly at random.
	
	stepCount := 0
	maxSteps := 50
	
	for w.StepRandom() {
		stepCount++
		if stepCount >= maxSteps {
			fmt.Printf("\n⚠️  Reached maximum steps (%d)\n", maxSteps)
			break
		}
	}
	
	fmt.Printf("\nFinal value: x = %d\n", counter.Value)
	fmt.Printf("Total steps: %d\n", stepCount)
}
