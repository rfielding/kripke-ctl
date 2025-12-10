package main

import (
	"fmt"
	"strings"

	"github.com/rfielding/kripke-ctl/kripke"
)

// This example shows how to:
// 1. Define actors using the kripke engine
// 2. Run the engine to explore behaviors
// 3. Extract a state space (manually for now)
// 4. Verify CTL properties on the extracted graph

// ========== Producer Actor ==========

type Producer struct {
	id string
}

func (p *Producer) ID() string {
	return p.id
}

func (p *Producer) Ready(w *kripke.World) []kripke.Step {
	ch := w.ChannelByAddress(kripke.Address{ActorID: "consumer", ChannelName: "inbox"})
	if ch == nil || !ch.CanSend() {
		return nil
	}

	return []kripke.Step{
		func(w *kripke.World) {
			ch := w.ChannelByAddress(kripke.Address{ActorID: "consumer", ChannelName: "inbox"})
			kripke.SendMessage(w, kripke.Message{
				From:    kripke.Address{ActorID: "producer", ChannelName: "out"},
				To:      kripke.Address{ActorID: "consumer", ChannelName: "inbox"},
				Payload: "data",
			})
		},
	}
}

// ========== Consumer Actor ==========

type Consumer struct {
	id       string
	inbox    *kripke.Channel
	received int
}

func (c *Consumer) ID() string {
	return c.id
}

func (c *Consumer) Ready(w *kripke.World) []kripke.Step {
	if !c.inbox.CanRecv() {
		return nil
	}

	return []kripke.Step{
		func(w *kripke.World) {
			kripke.RecvAndLog(w, c.inbox)
			c.received++
		},
	}
}

// ========== State Space Extraction (Manual for now) ==========

// For a producer-consumer system with buffer capacity 2,
// we can manually enumerate the state space:
// - Buffer can have 0, 1, or 2 items
// - Producer can send if buffer < 2
// - Consumer can receive if buffer > 0

func buildProducerConsumerGraph(capacity int) *kripke.Graph {
	g := kripke.NewGraph()

	// Create states for each buffer size
	for size := 0; size <= capacity; size++ {
		name := fmt.Sprintf("buffer_%d", size)
		labels := map[string]bool{
			"producer_ready": size < capacity,
			"consumer_ready": size > 0,
			"buffer_empty":   size == 0,
			"buffer_full":    size == capacity,
		}
		labels[fmt.Sprintf("buffer_size_%d", size)] = true

		g.AddState(name, labels)
	}

	// Add transitions
	// Producer sends: size -> size+1
	for size := 0; size < capacity; size++ {
		from := fmt.Sprintf("buffer_%d", size)
		to := fmt.Sprintf("buffer_%d", size+1)
		g.AddEdge(from, to)
	}

	// Consumer receives: size -> size-1
	for size := 1; size <= capacity; size++ {
		from := fmt.Sprintf("buffer_%d", size)
		to := fmt.Sprintf("buffer_%d", size-1)
		g.AddEdge(from, to)
	}

	// Initial state: buffer empty
	g.SetInitial("buffer_0")

	return g
}

// ========== Main ==========

func main() {
	fmt.Println("=" + strings.Repeat("=", 78))
	fmt.Println(" Producer-Consumer using kripke engine + CTL verification")
	fmt.Println("=" + strings.Repeat("=", 78))
	fmt.Println()

	// Part 1: Run the engine
	fmt.Println("PART 1: Running the Actor Engine")
	fmt.Println(strings.Repeat("-", 78))

	producer := &Producer{id: "producer"}
	inbox := kripke.NewChannel("consumer", "inbox", 2)
	consumer := &Consumer{id: "consumer", inbox: inbox}

	w := kripke.NewWorld(
		[]kripke.Process{producer, consumer},
		[]*kripke.Channel{inbox},
		42, // seed
	)

	// Run for a few steps to see behavior
	fmt.Println("Running 10 steps...")
	for i := 0; i < 10; i++ {
		if !w.StepRandom() {
			fmt.Println("System quiesced (no enabled steps)")
			break
		}
		fmt.Printf("  Step %d: Time=%d, Buffer=%d, Events=%d\n",
			i+1, w.Time, inbox.Len(), len(w.Events))
	}
	fmt.Printf("\nFinal state: Buffer=%d, Consumer received=%d messages\n",
		inbox.Len(), consumer.received)
	fmt.Println()

	// Part 2: Build state space graph
	fmt.Println("PART 2: Building State Space Graph")
	fmt.Println(strings.Repeat("-", 78))

	capacity := 2
	g := buildProducerConsumerGraph(capacity)

	fmt.Printf("States (%d total):\n", len(g.States()))
	for _, sid := range g.States() {
		name := g.NameOf(sid)
		pReady := g.HasLabel(sid, "producer_ready")
		cReady := g.HasLabel(sid, "consumer_ready")
		fmt.Printf("  %s: P=%v C=%v\n", name, pReady, cReady)
	}
	fmt.Println()

	// Part 3: CTL Verification
	fmt.Println("PART 3: CTL Model Checking")
	fmt.Println(strings.Repeat("-", 78))

	properties := []struct {
		name        string
		formula     kripke.Formula
		description string
	}{
		{
			name:        "Safety",
			formula:     kripke.AG(kripke.Not(kripke.Atom("buffer_overflow"))),
			description: "Buffer never overflows",
		},
		{
			name:        "Liveness-P",
			formula:     kripke.AG(kripke.EF(kripke.Atom("producer_ready"))),
			description: "Producer can always eventually send",
		},
		{
			name:        "Liveness-C",
			formula:     kripke.AG(kripke.EF(kripke.Atom("consumer_ready"))),
			description: "Consumer can always eventually receive",
		},
		{
			name:        "No-Deadlock",
			formula:     kripke.AG(kripke.Or(kripke.Atom("producer_ready"), kripke.Atom("consumer_ready"))),
			description: "System never deadlocks",
		},
		{
			name:        "Reachability-Full",
			formula:     kripke.EF(kripke.Atom("buffer_full")),
			description: "Buffer can become full",
		},
		{
			name:        "Reachability-Empty",
			formula:     kripke.EF(kripke.Atom("buffer_empty")),
			description: "Buffer can become empty",
		},
	}

	initialStates := kripke.NewStateSet()
	for _, initID := range g.InitialStates() {
		initialStates.Add(initID)
	}

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
	}
	fmt.Println()

	// Part 4: Mermaid Diagram
	fmt.Println("PART 4: Mermaid State Diagram")
	fmt.Println(strings.Repeat("-", 78))
	fmt.Println(generateMermaid(g))
	fmt.Println()

	// Part 5: Event Log Analysis
	fmt.Println("PART 5: Event Log (Queue Delay Analysis)")
	fmt.Println(strings.Repeat("-", 78))
	if len(w.Events) > 0 {
		totalDelay := 0
		for _, ev := range w.Events {
			totalDelay += ev.QueueDelay
			fmt.Printf("  Msg %d: Delay=%d ticks, Time=%d\n",
				ev.MsgID, ev.QueueDelay, ev.Time)
		}
		avgDelay := float64(totalDelay) / float64(len(w.Events))
		fmt.Printf("\nAverage queue delay: %.2f ticks\n", avgDelay)
	} else {
		fmt.Println("  (No events logged)")
	}

	fmt.Println()
	fmt.Println("=" + strings.Repeat("=", 78))
}

func generateMermaid(g *kripke.Graph) string {
	var sb strings.Builder
	sb.WriteString("stateDiagram-v2\n")
	
	// Initial state
	if len(g.InitialStates()) > 0 {
		initName := g.NameOf(g.InitialStates()[0])
		sb.WriteString(fmt.Sprintf("    [*] --> %s\n", initName))
	}
	sb.WriteString("\n")

	// Transitions
	for _, sid := range g.States() {
		name := g.NameOf(sid)
		for _, tid := range g.Succ(sid) {
			target := g.NameOf(tid)
			
			// Determine action
			action := ""
			if strings.Contains(target, "_") && strings.Contains(name, "_") {
				// Compare buffer sizes
				if target > name {
					action = ": produce"
				} else {
					action = ": consume"
				}
			}
			
			sb.WriteString(fmt.Sprintf("    %s --> %s%s\n", name, target, action))
		}
	}

	sb.WriteString("\n")

	// State labels
	for _, sid := range g.States() {
		name := g.NameOf(sid)
		pReady := "✓" 
		cReady := "✓"
		if !g.HasLabel(sid, "producer_ready") {
			pReady = "✗"
		}
		if !g.HasLabel(sid, "consumer_ready") {
			cReady = "✗"
		}
		
		sb.WriteString(fmt.Sprintf("    %s: %s\\nP:%s C:%s\n", name, name, pReady, cReady))
	}

	return sb.String()
}
