package main

import (
	"fmt"
	"math/rand"
	"github.com/rfielding/kripke-ctl/kripke"
)

// ============================================================================
// PROBABILISTIC CHOICE EXAMPLE
// ============================================================================
//
// This example demonstrates how to use pre-rolled dice (R1, R2, etc.) to
// make probabilistic choices. The dice are rolled BEFORE checking predicates,
// and transitions reference them to decide if they're available.
//
// Example: 70% small request, 30% large request
//
// ============================================================================

type Client struct {
	IDstr    string
	State    string
	Choice   int  // Pre-rolled dice value (0-99)
	Requests int
}

func (c *Client) ID() string { return c.IDstr }

// ============================================================================
// TRANSITION 1: Roll dice (always ready)
// ============================================================================

func (c *Client) Ready(w *kripke.World) []kripke.Step {
	if c.State != "ready" {
		return nil
	}
	
	return []kripke.Step{
		func(w *kripke.World) {
			// Pre-roll the dice
			c.Choice = rand.Intn(100)  // R1 ∈ [0, 100)
			c.State = "choosing"
			fmt.Printf("ROLLED: R1 = %d\n", c.Choice)
		},
	}
}

// ============================================================================
// TRANSITION 2: Send small request (predicate: 0 <= R1 < 70)
// ============================================================================

type SmallRequestTransition struct {
	IDstr  string
	Client *Client
}

func (s *SmallRequestTransition) ID() string { return s.IDstr }

func (s *SmallRequestTransition) Ready(w *kripke.World) []kripke.Step {
	// Predicate: state is choosing AND 0 <= R1 < 70
	if s.Client.State != "choosing" {
		return nil
	}
	if s.Client.Choice < 0 || s.Client.Choice >= 70 {
		return nil  // Dice says: take other path
	}
	
	// Predicate matches (70% chance)
	return []kripke.Step{
		func(w *kripke.World) {
			fmt.Printf("SMALL REQUEST (R1=%d in [0,70))\n", s.Client.Choice)
			s.Client.Requests++
			s.Client.State = "ready"
		},
	}
}

// ============================================================================
// TRANSITION 3: Send large request (predicate: 70 <= R1 < 100)
// ============================================================================

type LargeRequestTransition struct {
	IDstr  string
	Client *Client
}

func (l *LargeRequestTransition) ID() string { return l.IDstr }

func (l *LargeRequestTransition) Ready(w *kripke.World) []kripke.Step {
	// Predicate: state is choosing AND 70 <= R1 < 100
	if l.Client.State != "choosing" {
		return nil
	}
	if l.Client.Choice < 70 || l.Client.Choice >= 100 {
		return nil  // Dice says: take other path
	}
	
	// Predicate matches (30% chance)
	return []kripke.Step{
		func(w *kripke.World) {
			fmt.Printf("LARGE REQUEST (R1=%d in [70,100))\n", l.Client.Choice)
			l.Client.Requests++
			l.Client.State = "ready"
		},
	}
}

// ============================================================================
// MAIN
// ============================================================================

func main() {
	client := &Client{IDstr: "client", State: "ready"}
	smallTrans := &SmallRequestTransition{IDstr: "small", Client: client}
	largeTrans := &LargeRequestTransition{IDstr: "large", Client: client}
	
	w := kripke.NewWorld(
		[]kripke.Process{client, smallTrans, largeTrans},
		[]*kripke.Channel{},
		42,
	)
	
	fmt.Println("Probabilistic choice example: 70% small, 30% large\n")
	
	smallCount := 0
	largeCount := 0
	maxRequests := 100
	
	for client.Requests < maxRequests && w.StepRandom() {
		// Count results
		if client.State == "ready" && client.Requests > 0 {
			// Just completed a request
			if client.Choice < 70 {
				smallCount++
			} else {
				largeCount++
			}
		}
	}
	
	fmt.Printf("\n=== RESULTS ===\n")
	fmt.Printf("Small requests: %d (%.1f%%)\n", smallCount, float64(smallCount)*100/float64(maxRequests))
	fmt.Printf("Large requests: %d (%.1f%%)\n", largeCount, float64(largeCount)*100/float64(maxRequests))
	fmt.Printf("Total: %d\n", smallCount+largeCount)
}
