package main

import (
	"fmt"
	"strings"

	"github.com/rfielding/kripke-ctl/kripke"
)

// Complete example using the actual kripke package
// This demonstrates:
// 1. Importing github.com/rfielding/kripke-ctl/kripke
// 2. Building a Kripke graph
// 3. Defining CTL properties
// 4. Running model checking
// 5. Generating Mermaid diagrams

func main() {
	fmt.Println("=" + strings.Repeat("=", 78))
	fmt.Println(" Client-Server System - Using kripke package")
	fmt.Println("=" + strings.Repeat("=", 78))
	fmt.Println()

	// Build Kripke Graph
	fmt.Println("SYSTEM DESCRIPTION")
	fmt.Println(strings.Repeat("-", 78))
	fmt.Println(`
A client-server system with timeout:
- Client sends requests to the server
- Server processes requests and sends responses
- If no response arrives within 3 time units, the client times out
- After timeout or response, the client can send a new request
`)

	fmt.Println("BUILDING STATE SPACE")
	fmt.Println(strings.Repeat("-", 78))

	g := kripke.NewGraph()

	// States
	s0 := g.AddState("idle_idle_0", map[string]bool{
		"client_idle": true,
		"server_idle": true,
	})

	g.AddState("waiting_processing_0", map[string]bool{
		"request_sent": true,
		"server_busy":  true,
	})

	g.AddState("waiting_processing_1", map[string]bool{
		"request_sent": true,
		"server_busy":  true,
		"timer_1":      true,
	})

	g.AddState("waiting_processing_2", map[string]bool{
		"request_sent": true,
		"server_busy":  true,
		"timer_2":      true,
	})

	g.AddState("idle_idle_response", map[string]bool{
		"client_idle":  true,
		"server_idle":  true,
		"got_response": true,
	})

	g.AddState("idle_idle_timeout", map[string]bool{
		"client_idle": true,
		"server_idle": true,
		"timed_out":   true,
	})

	// Transitions
	g.AddEdge("idle_idle_0", "waiting_processing_0")
	g.AddEdge("waiting_processing_0", "waiting_processing_1")
	g.AddEdge("waiting_processing_1", "waiting_processing_2")
	g.AddEdge("waiting_processing_1", "idle_idle_response")
	g.AddEdge("waiting_processing_2", "idle_idle_response")
	g.AddEdge("waiting_processing_2", "idle_idle_timeout")
	g.AddEdge("idle_idle_response", "idle_idle_0")
	g.AddEdge("idle_idle_timeout", "idle_idle_0")
	g.AddEdge("idle_idle_0", "idle_idle_0") // Can stay idle

	g.SetInitial("idle_idle_0")

	fmt.Printf("States (%d total):\n", len(g.States()))
	for _, sid := range g.States() {
		fmt.Printf("  %s\n", g.NameOf(sid))
	}
	fmt.Println()

	// CTL Properties
	fmt.Println("CTL PROPERTIES & VERIFICATION")
	fmt.Println(strings.Repeat("-", 78))

	properties := []struct {
		name        string
		formula     kripke.Formula
		description string
	}{
		{
			name:        "P1",
			formula:     kripke.AG(kripke.Or(kripke.Atom("client_idle"), kripke.Atom("request_sent"))),
			description: "Client is always either idle or waiting",
		},
		{
			name:        "P2",
			formula:     kripke.EF(kripke.Atom("got_response")),
			description: "It's possible to get a response",
		},
		{
			name:        "P3",
			formula:     kripke.EF(kripke.Atom("timed_out")),
			description: "It's possible to timeout",
		},
		{
			name:        "P4",
			formula:     kripke.AG(kripke.Not(kripke.And(kripke.Atom("got_response"), kripke.Atom("timed_out")))),
			description: "Never both response and timeout",
		},
		{
			name:        "P5",
			formula:     kripke.AF(kripke.Or(kripke.Atom("got_response"), kripke.Atom("timed_out"))),
			description: "Eventually get response or timeout",
		},
	}

	initialStates := kripke.NewStateSet()
	initialStates.Add(s0)

	for _, p := range properties {
		satisfying := p.formula.Sat(g)

		allInitialsSatisfy := true
		for id := range initialStates {
			if !satisfying.Contains(id) {
				allInitialsSatisfy = false
				break
			}
		}

		status := "✗ FAIL"
		if allInitialsSatisfy {
			status = "✓ PASS"
		}

		fmt.Printf("%s %s: %s\n", status, p.name, p.description)
		
		// Show which states satisfy
		if len(satisfying) > 0 && len(satisfying) <= 6 {
			fmt.Print("    States: ")
			first := true
			for sid := range satisfying {
				if !first {
					fmt.Print(", ")
				}
				fmt.Print(g.NameOf(sid))
				first = false
			}
			fmt.Println()
		}
	}
	fmt.Println()

	// Mermaid Diagram
	fmt.Println("MERMAID DIAGRAM")
	fmt.Println(strings.Repeat("-", 78))
	fmt.Println(generateMermaid(g))
	fmt.Println()

	fmt.Println("=" + strings.Repeat("=", 78))
}

func generateMermaid(g *kripke.Graph) string {
	var sb strings.Builder
	sb.WriteString("stateDiagram-v2\n")
	sb.WriteString("    [*] --> idle_idle_0\n")
	sb.WriteString("    \n")

	for _, sid := range g.States() {
		name := g.NameOf(sid)
		for _, tid := range g.Succ(sid) {
			target := g.NameOf(tid)
			if name != target {
				sb.WriteString(fmt.Sprintf("    %s --> %s\n", name, target))
			}
		}
	}

	sb.WriteString("    \n")
	sb.WriteString("    idle_idle_0: Idle / Ready\n")
	sb.WriteString("    waiting_processing_0: Sent / Timer=0\n")
	sb.WriteString("    waiting_processing_1: Wait / Timer=1\n")
	sb.WriteString("    waiting_processing_2: Wait / Timer=2\n")
	sb.WriteString("    idle_idle_response: Response Received\n")
	sb.WriteString("    idle_idle_timeout: Timeout!\n")

	return sb.String()
}
