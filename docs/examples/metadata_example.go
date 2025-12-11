package main

import (
	"fmt"
	"github.com/rfielding/kripke-ctl/kripke"
)

// ============================================================================
// PROCESS METADATA EXAMPLE
// ============================================================================
//
// Shows how to add visualization metadata to processes so FSM diagrams
// can show guards and actions.
//
// ============================================================================

// ProcessMetadata is an optional interface for visualization
type ProcessMetadata interface {
	kripke.Process
	Guard() string      // Guard/predicate: "x > 0", "state = idle"
	Action() string     // Action: "x' = x + 1", "send(ch, msg)", "recv(ch)"
	FromState() string  // Source state name
	ToState() string    // Target state name
}

// ============================================================================
// COUNTER EXAMPLE WITH METADATA
// ============================================================================

type Counter struct {
	Value int
}

type IncrementTransition struct {
	IDstr   string
	Counter *Counter
}

func (i *IncrementTransition) ID() string { return i.IDstr }

// Metadata for visualization
func (i *IncrementTransition) Guard() string { return "x < 20" }
func (i *IncrementTransition) Action() string { return "x' = x + 1" }
func (i *IncrementTransition) FromState() string { return "Active" }
func (i *IncrementTransition) ToState() string { return "Active" }

// Actual behavior
func (i *IncrementTransition) Ready(w *kripke.World) []kripke.Step {
	if i.Counter.Value >= 20 {
		return nil
	}
	return []kripke.Step{
		func(w *kripke.World) {
			i.Counter.Value++
		},
	}
}

type DecrementTransition struct {
	IDstr   string
	Counter *Counter
}

func (d *DecrementTransition) ID() string { return d.IDstr }

// Metadata
func (d *DecrementTransition) Guard() string { return "x > 0" }
func (d *DecrementTransition) Action() string { return "x' = x - 1" }
func (d *DecrementTransition) FromState() string { return "Active" }
func (d *DecrementTransition) ToState() string { return "Active" }

// Behavior
func (d *DecrementTransition) Ready(w *kripke.World) []kripke.Step {
	if d.Counter.Value <= 0 {
		return nil
	}
	return []kripke.Step{
		func(w *kripke.World) {
			d.Counter.Value--
		},
	}
}

type ResetTransition struct {
	IDstr   string
	Counter *Counter
}

func (r *ResetTransition) ID() string { return r.IDstr }

// Metadata
func (r *ResetTransition) Guard() string { return "x = 20" }
func (r *ResetTransition) Action() string { return "x' = 0" }
func (r *ResetTransition) FromState() string { return "Active" }
func (r *ResetTransition) ToState() string { return "Reset" }

// Behavior
func (r *ResetTransition) Ready(w *kripke.World) []kripke.Step {
	if r.Counter.Value != 20 {
		return nil
	}
	return []kripke.Step{
		func(w *kripke.World) {
			r.Counter.Value = 0
		},
	}
}

// ============================================================================
// FSM DIAGRAM GENERATOR
// ============================================================================

func GenerateFSMDiagram(processes []kripke.Process) string {
	var diagram string
	diagram += "```mermaid\n"
	diagram += "stateDiagram-v2\n"
	
	// Collect states and transitions
	states := make(map[string]string)  // state name -> guard
	
	for _, process := range processes {
		if pm, ok := process.(ProcessMetadata); ok {
			from := pm.FromState()
			to := pm.ToState()
			guard := pm.Guard()
			action := pm.Action()
			
			// Record state with its guard
			if _, exists := states[from]; !exists {
				states[from] = guard
			}
			
			// Draw transition with action
			diagram += fmt.Sprintf("    %s --> %s: %s\n", from, to, action)
		}
	}
	
	diagram += "```\n"
	return diagram
}

// ============================================================================
// MAIN
// ============================================================================

func main() {
	counter := &Counter{Value: 5}
	
	inc := &IncrementTransition{IDstr: "increment", Counter: counter}
	dec := &DecrementTransition{IDstr: "decrement", Counter: counter}
	reset := &ResetTransition{IDstr: "reset", Counter: counter}
	
	processes := []kripke.Process{inc, dec, reset}
	
	// Generate FSM diagram
	fmt.Println("# Counter State Machine\n")
	fmt.Println(GenerateFSMDiagram(processes))
	fmt.Println()
	
	// Show metadata
	fmt.Println("## Process Metadata\n")
	for _, process := range processes {
		if pm, ok := process.(ProcessMetadata); ok {
			fmt.Printf("**%s**\n", process.ID())
			fmt.Printf("- Guard: `%s`\n", pm.Guard())
			fmt.Printf("- Action: `%s`\n", pm.Action())
			fmt.Printf("- Transition: %s → %s\n", pm.FromState(), pm.ToState())
			fmt.Println()
		}
	}
	
	// Run the model
	w := kripke.NewWorld(processes, []*kripke.Channel{}, 42)
	
	fmt.Println("## Execution Trace\n")
	for i := 0; i < 20 && w.StepRandom(); i++ {
		fmt.Printf("Step %d: x = %d\n", i+1, counter.Value)
	}
}
