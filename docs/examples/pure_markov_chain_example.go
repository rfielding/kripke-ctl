package main

import (
	"fmt"
	"math/rand"
	"github.com/rfielding/kripke-ctl/kripke"
)

// ============================================================================
// PURE MARKOV CHAIN EXAMPLE
// ============================================================================
//
// A pure Markov chain has ONLY probabilistic transitions (chance nodes).
// Every state change is determined by rolling dice.
//
// Key insight: The ONLY way to change which guards match is to modify
// variables (state assignments), because guards check variables!
//
// In this example:
//   State A → State B (40%)
//   State A → State C (60%)
//   State B → State A (50%)
//   State B → State C (50%)
//   State C → State A (100%)
//
// ============================================================================

type MarkovActor struct {
	IDstr string
	State string  // "A", "B", or "C"
	Dice  int     // Pre-rolled dice value
}

func (m *MarkovActor) ID() string { return m.IDstr }

// ============================================================================
// CHANCE NODE 1: Roll dice in state A
// ============================================================================

type RollInA struct {
	IDstr string
	Actor *MarkovActor
}

func (r *RollInA) ID() string { return r.IDstr }

func (r *RollInA) Ready(w *kripke.World) []kripke.Step {
	// Guard: state == "A"
	if r.Actor.State != "A" {
		return nil
	}
	
	return []kripke.Step{
		func(w *kripke.World) {
			// Roll dice
			r.Actor.Dice = rand.Intn(100)
			r.Actor.State = "choosing_from_A"
			fmt.Printf("In A, rolled R1=%d\n", r.Actor.Dice)
		},
	}
}

// ============================================================================
// CHANCE NODE 2: A → B (40% chance, if 0 <= R1 < 40)
// ============================================================================

type A_to_B struct {
	IDstr string
	Actor *MarkovActor
}

func (a *A_to_B) ID() string { return a.IDstr }

func (a *A_to_B) Ready(w *kripke.World) []kripke.Step {
	// Guard: state == "choosing_from_A" AND 0 <= dice < 40
	if a.Actor.State != "choosing_from_A" {
		return nil
	}
	if a.Actor.Dice < 0 || a.Actor.Dice >= 40 {
		return nil
	}
	
	return []kripke.Step{
		func(w *kripke.World) {
			// Variable assignment changes state!
			// This is the ONLY way guards change
			a.Actor.State = "B"
			fmt.Printf("  → Transition A→B (40%%)\n")
		},
	}
}

// ============================================================================
// CHANCE NODE 3: A → C (60% chance, if 40 <= R1 < 100)
// ============================================================================

type A_to_C struct {
	IDstr string
	Actor *MarkovActor
}

func (a *A_to_C) ID() string { return a.IDstr }

func (a *A_to_C) Ready(w *kripke.World) []kripke.Step {
	// Guard: state == "choosing_from_A" AND 40 <= dice < 100
	if a.Actor.State != "choosing_from_A" {
		return nil
	}
	if a.Actor.Dice < 40 || a.Actor.Dice >= 100 {
		return nil
	}
	
	return []kripke.Step{
		func(w *kripke.World) {
			// Variable assignment changes state!
			a.Actor.State = "C"
			fmt.Printf("  → Transition A→C (60%%)\n")
		},
	}
}

// ============================================================================
// CHANCE NODE 4: Roll dice in state B
// ============================================================================

type RollInB struct {
	IDstr string
	Actor *MarkovActor
}

func (r *RollInB) ID() string { return r.IDstr }

func (r *RollInB) Ready(w *kripke.World) []kripke.Step {
	if r.Actor.State != "B" {
		return nil
	}
	
	return []kripke.Step{
		func(w *kripke.World) {
			r.Actor.Dice = rand.Intn(100)
			r.Actor.State = "choosing_from_B"
			fmt.Printf("In B, rolled R1=%d\n", r.Actor.Dice)
		},
	}
}

// ============================================================================
// CHANCE NODE 5: B → A (50% chance)
// ============================================================================

type B_to_A struct {
	IDstr string
	Actor *MarkovActor
}

func (b *B_to_A) ID() string { return b.IDstr }

func (b *B_to_A) Ready(w *kripke.World) []kripke.Step {
	if b.Actor.State != "choosing_from_B" {
		return nil
	}
	if b.Actor.Dice < 0 || b.Actor.Dice >= 50 {
		return nil
	}
	
	return []kripke.Step{
		func(w *kripke.World) {
			b.Actor.State = "A"
			fmt.Printf("  → Transition B→A (50%%)\n")
		},
	}
}

// ============================================================================
// CHANCE NODE 6: B → C (50% chance)
// ============================================================================

type B_to_C struct {
	IDstr string
	Actor *MarkovActor
}

func (b *B_to_C) ID() string { return b.IDstr }

func (b *B_to_C) Ready(w *kripke.World) []kripke.Step {
	if b.Actor.State != "choosing_from_B" {
		return nil
	}
	if b.Actor.Dice < 50 || b.Actor.Dice >= 100 {
		return nil
	}
	
	return []kripke.Step{
		func(w *kripke.World) {
			b.Actor.State = "C"
			fmt.Printf("  → Transition B→C (50%%)\n")
		},
	}
}

// ============================================================================
// DETERMINISTIC: C → A (100% chance - no dice needed)
// ============================================================================

type C_to_A struct {
	IDstr string
	Actor *MarkovActor
}

func (c *C_to_A) ID() string { return c.IDstr }

func (c *C_to_A) Ready(w *kripke.World) []kripke.Step {
	if c.Actor.State != "C" {
		return nil
	}
	
	return []kripke.Step{
		func(w *kripke.World) {
			c.Actor.State = "A"
			fmt.Printf("In C → Transition C→A (100%%)\n")
		},
	}
}

// ============================================================================
// MAIN
// ============================================================================

func main() {
	actor := &MarkovActor{IDstr: "markov", State: "A"}
	
	// All transitions are chance nodes (except C→A which is deterministic)
	rollA := &RollInA{IDstr: "roll_A", Actor: actor}
	aToB := &A_to_B{IDstr: "A_to_B", Actor: actor}
	aToC := &A_to_C{IDstr: "A_to_C", Actor: actor}
	rollB := &RollInB{IDstr: "roll_B", Actor: actor}
	bToA := &B_to_A{IDstr: "B_to_A", Actor: actor}
	bToC := &B_to_C{IDstr: "B_to_C", Actor: actor}
	cToA := &C_to_A{IDstr: "C_to_A", Actor: actor}
	
	w := kripke.NewWorld(
		[]kripke.Process{rollA, aToB, aToC, rollB, bToA, bToC, cToA},
		[]*kripke.Channel{},
		42,
	)
	
	fmt.Println("Pure Markov Chain Example")
	fmt.Println("Transition probabilities:")
	fmt.Println("  A → B: 40%")
	fmt.Println("  A → C: 60%")
	fmt.Println("  B → A: 50%")
	fmt.Println("  B → C: 50%")
	fmt.Println("  C → A: 100%")
	fmt.Println()
	
	// Track state visits
	visits := map[string]int{"A": 0, "B": 0, "C": 0}
	
	maxSteps := 100
	stepCount := 0
	
	for stepCount < maxSteps && w.StepRandom() {
		// Count visits to stable states (A, B, C)
		if actor.State == "A" || actor.State == "B" || actor.State == "C" {
			visits[actor.State]++
		}
		stepCount++
	}
	
	fmt.Printf("\n=== RESULTS after %d steps ===\n", stepCount)
	fmt.Printf("Time in state A: %d (%.1f%%)\n", visits["A"], float64(visits["A"])*100/float64(visits["A"]+visits["B"]+visits["C"]))
	fmt.Printf("Time in state B: %d (%.1f%%)\n", visits["B"], float64(visits["B"])*100/float64(visits["A"]+visits["B"]+visits["C"]))
	fmt.Printf("Time in state C: %d (%.1f%%)\n", visits["C"], float64(visits["C"])*100/float64(visits["A"]+visits["B"]+visits["C"]))
	fmt.Printf("\nThis is a PURE Markov chain - only chance nodes!\n")
	fmt.Printf("State changes ONLY through variable assignment.\n")
}
