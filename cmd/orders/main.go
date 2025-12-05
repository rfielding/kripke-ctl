package main

import (
	"fmt"

	"github.com/rfielding/kripke-ctl/kripke"
)

func main() {
	g := kripke.OrderGraph()

	// Atomic propositions
	accepted := kripke.Atom("accepted")
	delivered := kripke.Atom("delivered")
	cancelled := kripke.Atom("cancelled")

	// R1: Every accepted order is eventually delivered or cancelled.
	// AG(accepted -> AF(delivered ∨ cancelled))
	r1 := kripke.AG(
		kripke.Implies(
			accepted,
			kripke.AF(kripke.Or(delivered, cancelled)),
		),
	)

	// R2: No state is both delivered and cancelled.
	// AG ¬(delivered ∧ cancelled)
	r2 := kripke.AG(
		kripke.Not(
			kripke.And(delivered, cancelled),
		),
	)

	// R3: It is possible to deliver an order.
	// EF delivered
	r3 := kripke.EF(delivered)

	// R4: It is possible to cancel an order.
	// EF cancelled
	r4 := kripke.EF(cancelled)

	check("R1: accepted eventually resolved", r1, g)
	check("R2: no delivered & cancelled simultaneously", r2, g)
	check("R3: delivery is possible", r3, g)
	check("R4: cancellation is possible", r4, g)
}

func check(name string, f kripke.Formula, g *kripke.Graph) {
	satisfying := f.Sat(g)
	allGood := true
	for _, s := range g.InitialStates() {
		if !satisfying.Contains(s) {
			allGood = false
			break
		}
	}
	if allGood {
		fmt.Printf("PASS: %s\n", name)
	} else {
		fmt.Printf("FAIL: %s\n", name)
		fmt.Printf("  (At least one initial state violates the formula)\n")
	}
}
