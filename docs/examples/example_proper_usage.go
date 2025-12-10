package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/rfielding/kripke-ctl/kripke"
)

// Producer actor with state variables
type Producer struct {
	id           string
	itemsProduced int
}

func (p *Producer) ID() string { return p.id }

func (p *Producer) Ready(w *kripke.World) []kripke.Step {
	ch := w.ChannelByAddress(kripke.Address{ActorID: "consumer", ChannelName: "inbox"})
	if ch == nil || !ch.CanSend() {
		return nil
	}

	return []kripke.Step{
		func(w *kripke.World) {
			kripke.SendMessage(w, kripke.Message{
				From:    kripke.Address{ActorID: "producer", ChannelName: "out"},
				To:      kripke.Address{ActorID: "consumer", ChannelName: "inbox"},
				Payload: fmt.Sprintf("item_%d", p.itemsProduced),
			})
			p.itemsProduced++
		},
	}
}

// Consumer actor with state variables
type Consumer struct {
	id           string
	inbox        *kripke.Channel
	itemsConsumed int
	totalDelay   int
}

func (c *Consumer) ID() string { return c.id }

func (c *Consumer) Ready(w *kripke.World) []kripke.Step {
	if !c.inbox.CanRecv() {
		return nil
	}

	return []kripke.Step{
		func(w *kripke.World) {
			_, delay := kripke.RecvAndLog(w, c.inbox)
			c.itemsConsumed++
			c.totalDelay += delay
		},
	}
}

// buildGraph constructs the state graph with labels
func buildGraph(capacity int) *kripke.Graph {
	g := kripke.NewGraph()

	for size := 0; size <= capacity; size++ {
		name := fmt.Sprintf("buffer_%d", size)
		labels := map[string]bool{
			"producer_ready": size < capacity,
			"consumer_ready": size > 0,
			"buffer_empty":   size == 0,
			"buffer_full":    size == capacity,
		}
		g.AddState(name, labels)
	}

	// Add transitions
	for size := 0; size < capacity; size++ {
		g.AddEdge(fmt.Sprintf("buffer_%d", size), fmt.Sprintf("buffer_%d", size+1))
	}
	for size := 1; size <= capacity; size++ {
		g.AddEdge(fmt.Sprintf("buffer_%d", size), fmt.Sprintf("buffer_%d", size-1))
	}

	g.SetInitial("buffer_0")
	return g
}

func main() {
	var md strings.Builder

	// Title
	md.WriteString("# Producer-Consumer Specification\n\n")
	md.WriteString("Generated using kripke package diagram methods\n\n")
	md.WriteString("---\n\n")

	// Section 1: Requirements
	md.WriteString("## 1. Requirements\n\n")
	
	requirements := []kripke.Requirement{
		{
			ID:            "REQ-SAF-01",
			Category:      "Safety",
			Description:   "Buffer SHALL never overflow",
			FormulaString: "AG(¬buffer_full ∨ ¬producer_ready)",
			EnglishRef:    "Buffer capacity constraint",
			Rationale:     "Prevents memory overflow",
		},
		{
			ID:            "REQ-LIVE-01",
			Category:      "Liveness",
			Description:   "Producer SHALL always eventually be ready",
			FormulaString: "AG(EF(producer_ready))",
			EnglishRef:    "No producer starvation",
			Rationale:     "Ensures system makes progress",
		},
		{
			ID:            "REQ-LIVE-02",
			Category:      "Liveness",
			Description:   "Consumer SHALL always eventually be ready",
			FormulaString: "AG(EF(consumer_ready))",
			EnglishRef:    "No consumer starvation",
			Rationale:     "Ensures data is consumed",
		},
	}

	// Use kripke package method to generate requirements table
	md.WriteString(kripke.GenerateRequirementsTable(requirements))
	md.WriteString("\n")

	// Section 2: Run system
	md.WriteString("## 2. Execution\n\n")

	producer := &Producer{id: "producer", itemsProduced: 0}
	inbox := kripke.NewChannel("consumer", "inbox", 2)
	consumer := &Consumer{id: "consumer", inbox: inbox, itemsConsumed: 0}

	w := kripke.NewWorld(
		[]kripke.Process{producer, consumer},
		[]*kripke.Channel{inbox},
		42,
	)

	// Run for some steps
	for i := 0; i < 10; i++ {
		if !w.StepRandom() {
			break
		}
	}

	md.WriteString(fmt.Sprintf("Executed %d steps\n\n", len(w.Events)))
	md.WriteString(fmt.Sprintf("- Producer sent: %d items\n", producer.itemsProduced))
	md.WriteString(fmt.Sprintf("- Consumer received: %d items\n", consumer.itemsConsumed))
	md.WriteString(fmt.Sprintf("- Buffer contains: %d items\n\n", inbox.Len()))

	// Section 3: Sequence Diagram - USE KRIPKE PACKAGE METHOD
	md.WriteString("## 3. Message Sequence\n\n")
	md.WriteString("```mermaid\n")
	md.WriteString(w.GenerateSequenceDiagram(10)) // Use package method!
	md.WriteString("```\n\n")

	// Section 4: State Machine - USE KRIPKE PACKAGE METHOD
	md.WriteString("## 4. State Machine\n\n")
	
	g := buildGraph(2)

	// Custom state describer
	stateDescriber := func(sid kripke.StateID, g *kripke.Graph) string {
		name := g.NameOf(sid)
		var size int
		fmt.Sscanf(name, "buffer_%d", &size)
		return fmt.Sprintf("buffer_size=%d", size)
	}

	// Custom edge labeler showing guards and actions
	edgeLabeler := func(from, to kripke.StateID, g *kripke.Graph) string {
		fromName := g.NameOf(from)
		toName := g.NameOf(to)
		
		var fromSize, toSize int
		fmt.Sscanf(fromName, "buffer_%d", &fromSize)
		fmt.Sscanf(toName, "buffer_%d", &toSize)
		
		if toSize > fromSize {
			// produce transition
			return "[buffer_size < 2] produce: send(msg); itemsProduced++"
		} else {
			// consume transition
			return "[buffer_size > 0] consume: recv(msg); itemsConsumed++"
		}
	}

	md.WriteString("```mermaid\n")
	// Use package method with custom options!
	md.WriteString(g.GenerateStateDiagram(
		kripke.WithStateDescriber(stateDescriber),
		kripke.WithEdgeLabeler(edgeLabeler),
	))
	md.WriteString("```\n\n")

	// Section 5: CTL Verification - USE KRIPKE PACKAGE METHOD
	md.WriteString("## 5. CTL Verification\n\n")

	// Build formulas
	producerReady := kripke.Atom("producer_ready")
	consumerReady := kripke.Atom("consumer_ready")
	bufferFull := kripke.Atom("buffer_full")

	requirements[0].Formula = kripke.AG(kripke.Or(kripke.Not(bufferFull), kripke.Not(producerReady)))
	requirements[1].Formula = kripke.AG(kripke.EF(producerReady))
	requirements[2].Formula = kripke.AG(kripke.EF(consumerReady))

	// Use package method to generate CTL table!
	md.WriteString(g.GenerateCTLTable(requirements))
	md.WriteString("\n")

	// Section 6: Summary
	md.WriteString("## 6. Summary\n\n")
	md.WriteString("This specification was generated using kripke package methods:\n\n")
	md.WriteString("- `kripke.GenerateRequirementsTable()` - Requirements table\n")
	md.WriteString("- `world.GenerateSequenceDiagram()` - Message sequence diagram\n")
	md.WriteString("- `graph.GenerateStateDiagram()` - State machine with custom labels\n")
	md.WriteString("- `graph.GenerateCTLTable()` - CTL verification results\n\n")

	md.WriteString("All diagram generation is handled by the kripke package, not duplicated in examples.\n\n")

	// Write file
	filename := "producer-consumer-proper.md"
	if err := os.WriteFile(filename, []byte(md.String()), 0644); err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		return
	}

	fmt.Printf("✅ Generated: %s\n\n", filename)
	fmt.Println("Uses kripke package methods:")
	fmt.Println("  ✓ GenerateRequirementsTable()")
	fmt.Println("  ✓ GenerateSequenceDiagram()")
	fmt.Println("  ✓ GenerateStateDiagram()")
	fmt.Println("  ✓ GenerateCTLTable()")
	fmt.Println()
	fmt.Println("No diagram generation code duplicated in example!")
}
