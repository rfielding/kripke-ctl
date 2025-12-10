package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/rfielding/kripke-ctl/kripke"
)

// ENGLISH PROMPT (input to LLM):
// """
// Create a file upload system with:
// - Uploader actor that sends chunks
// - Receiver actor that accumulates chunks
// - Track: total upload time, total bytes, throughput
// - Show state machine with actor variables
// - Show interaction diagram from execution
// - Calculate throughput metrics
// """

// LLM GENERATES THIS CODE:

// Uploader actor with state variables
type Uploader struct {
	id             string
	chunksSent     int      // Counter
	totalBytesSent int64    // Counter
	startTime      time.Time
}

func (u *Uploader) ID() string { return u.id }

// Guard: can send
// Action: send chunk, update counters
func (u *Uploader) Ready(w *kripke.World) []kripke.Step {
	ch := w.ChannelByAddress(kripke.Address{ActorID: "receiver", ChannelName: "inbox"})
	if ch == nil || !ch.CanSend() {
		return nil
	}

	// Stop after 10 chunks
	if u.chunksSent >= 10 {
		return nil
	}

	return []kripke.Step{
		func(w *kripke.World) {
			chunkSize := int64(1024 * 1024) // 1 MB chunks

			// Action: send message
			kripke.SendMessage(w, kripke.Message{
				From:    kripke.Address{ActorID: "uploader", ChannelName: "out"},
				To:      kripke.Address{ActorID: "receiver", ChannelName: "inbox"},
				Payload: fmt.Sprintf("chunk_%d", u.chunksSent),
			})

			// Action: update state variables
			u.chunksSent++              // chunksSent' = chunksSent + 1
			u.totalBytesSent += chunkSize // totalBytesSent' = totalBytesSent + chunkSize
		},
	}
}

// Receiver actor with state variables
type Receiver struct {
	id                string
	inbox             *kripke.Channel
	chunksReceived    int      // Counter
	totalBytesReceived int64    // Counter
	totalDelay        int      // Counter
	startTime         time.Time
	endTime           time.Time
}

func (r *Receiver) ID() string { return r.id }

// Guard: has chunks to receive
// Action: receive chunk, update counters
func (r *Receiver) Ready(w *kripke.World) []kripke.Step {
	if !r.inbox.CanRecv() {
		return nil
	}

	return []kripke.Step{
		func(w *kripke.World) {
			// Action: receive message
			_, delay := kripke.RecvAndLog(w, r.inbox)

			// Action: update state variables
			chunkSize := int64(1024 * 1024)
			r.chunksReceived++                  // chunksReceived' = chunksReceived + 1
			r.totalBytesReceived += chunkSize   // totalBytesReceived' = totalBytesReceived + chunkSize
			r.totalDelay += delay               // totalDelay' = totalDelay + delay
			r.endTime = time.Now()
		},
	}
}

func main() {
	var md strings.Builder

	// Title
	md.WriteString("# File Upload System: Complete Specification\n\n")
	md.WriteString("Generated from English prompt by LLM → Go code → kripke library\n\n")
	md.WriteString("---\n\n")

	// Section 0: English Prompt (input)
	md.WriteString("## 0. English Prompt (Input to LLM)\n\n")
	md.WriteString("```\n")
	md.WriteString("Create a file upload system with:\n")
	md.WriteString("- Uploader actor that sends chunks\n")
	md.WriteString("- Receiver actor that accumulates chunks\n")
	md.WriteString("- Track: total upload time, total bytes, throughput\n")
	md.WriteString("- Show state machine with actor variables\n")
	md.WriteString("- Show interaction diagram from execution\n")
	md.WriteString("- Calculate throughput metrics\n")
	md.WriteString("```\n\n")

	// Section 1: Generated Code (Actor Definitions)
	md.WriteString("## 1. Generated Actor Code\n\n")
	md.WriteString("### Uploader\n\n")
	md.WriteString("```go\n")
	md.WriteString("type Uploader struct {\n")
	md.WriteString("    id             string\n")
	md.WriteString("    chunksSent     int      // State variable\n")
	md.WriteString("    totalBytesSent int64    // State variable\n")
	md.WriteString("    startTime      time.Time\n")
	md.WriteString("}\n\n")
	md.WriteString("// Guard: can send AND chunksSent < 10\n")
	md.WriteString("// Action: send(chunk); chunksSent++; totalBytesSent += chunkSize\n")
	md.WriteString("```\n\n")

	md.WriteString("### Receiver\n\n")
	md.WriteString("```go\n")
	md.WriteString("type Receiver struct {\n")
	md.WriteString("    id                string\n")
	md.WriteString("    chunksReceived    int      // State variable\n")
	md.WriteString("    totalBytesReceived int64    // State variable\n")
	md.WriteString("    totalDelay        int      // State variable\n")
	md.WriteString("}\n\n")
	md.WriteString("// Guard: inbox not empty\n")
	md.WriteString("// Action: recv(chunk); chunksReceived++; totalBytesReceived += chunkSize; totalDelay += delay\n")
	md.WriteString("```\n\n")

	// Section 2: Run System
	md.WriteString("## 2. Execution\n\n")

	startTime := time.Now()
	uploader := &Uploader{
		id:        "uploader",
		startTime: startTime,
	}
	inbox := kripke.NewChannel("receiver", "inbox", 3)
	receiver := &Receiver{
		id:        "receiver",
		inbox:     inbox,
		startTime: startTime,
	}

	w := kripke.NewWorld(
		[]kripke.Process{uploader, receiver},
		[]*kripke.Channel{inbox},
		42,
	)

	// Initialize metrics collector
	metrics := kripke.NewMetricsCollector()
	uploadBytes := metrics.Counter("upload_bytes", "Total bytes uploaded", "bytes")
	downloadBytes := metrics.Counter("download_bytes", "Total bytes downloaded", "bytes")
	uploadChunks := metrics.Counter("upload_chunks", "Total chunks uploaded", "chunks")
	downloadChunks := metrics.Counter("download_chunks", "Total chunks downloaded", "chunks")

	// Run until quiesced
	step := 0
	for {
		if !w.StepRandom() {
			break
		}
		step++

		// Update metrics after each step
		uploadBytes.Value = float64(uploader.totalBytesSent)
		downloadBytes.Value = float64(receiver.totalBytesReceived)
		uploadChunks.Value = float64(uploader.chunksSent)
		downloadChunks.Value = float64(receiver.chunksReceived)
	}

	endTime := time.Now()

	md.WriteString(fmt.Sprintf("Executed %d steps\n\n", step))

	// Section 3: Actor States (CODE as state)
	md.WriteString("## 3. Actor States (Code Variables)\n\n")

	actorStates := []kripke.ActorState{
		{
			ActorID: "uploader",
			Variables: map[string]interface{}{
				"chunksSent":     uploader.chunksSent,
				"totalBytesSent": uploader.totalBytesSent,
			},
		},
		{
			ActorID: "receiver",
			Variables: map[string]interface{}{
				"chunksReceived":    receiver.chunksReceived,
				"totalBytesReceived": receiver.totalBytesReceived,
				"totalDelay":        receiver.totalDelay,
			},
		},
	}

	// Use kripke library method
	md.WriteString(kripke.GenerateActorStateTable(actorStates))
	md.WriteString("\n")

	md.WriteString("**State changes happen via variable edits:**\n")
	md.WriteString("- Uploader: `chunksSent' = chunksSent + 1`\n")
	md.WriteString("- Receiver: `chunksReceived' = chunksReceived + 1`\n\n")

	// Section 4: State Machine (with CODE states)
	md.WriteString("## 4. State Machine (Actor Code States)\n\n")

	transitions := []kripke.StateTransition{
		{
			FromActor: "uploader",
			ToActor:   "receiver",
			Guard:     "inbox.CanSend() AND chunksSent < 10",
			Action:    "send(chunk)",
			VariableEdits: []string{
				"chunksSent++",
				"totalBytesSent += 1MB",
			},
		},
		{
			FromActor: "receiver",
			ToActor:   "receiver",
			Guard:     "inbox.CanRecv()",
			Action:    "recv(chunk)",
			VariableEdits: []string{
				"chunksReceived++",
				"totalBytesReceived += 1MB",
				"totalDelay += delay",
			},
		},
	}

	// Use kripke library method
	md.WriteString(kripke.GenerateTransitionTable(transitions))
	md.WriteString("\n")

	md.WriteString("```mermaid\n")
	md.WriteString("stateDiagram-v2\n")
	md.WriteString("    [*] --> uploader\n")
	md.WriteString("    \n")
	md.WriteString("    state uploader {\n")
	md.WriteString("        chunksSent: 0→10\n")
	md.WriteString("        totalBytesSent: 0→10MB\n")
	md.WriteString("    }\n")
	md.WriteString("    \n")
	md.WriteString("    state receiver {\n")
	md.WriteString("        chunksReceived: 0→10\n")
	md.WriteString("        totalBytesReceived: 0→10MB\n")
	md.WriteString("        totalDelay: accumulated\n")
	md.WriteString("    }\n")
	md.WriteString("    \n")
	md.WriteString("    uploader --> receiver: [canSend && chunks<10]<br/>send(chunk)<br/>chunksSent++, bytes+=1MB\n")
	md.WriteString("    receiver --> receiver: [canRecv]<br/>recv(chunk)<br/>chunksRecv++, bytes+=1MB, delay+=d\n")
	md.WriteString("```\n\n")

	// Section 5: Interaction Diagram (from actual execution)
	md.WriteString("## 5. Interaction Diagram (From Execution)\n\n")
	md.WriteString("```mermaid\n")
	// Use kripke library method
	md.WriteString(w.GenerateInteractionDiagram(10))
	md.WriteString("```\n\n")

	md.WriteString("This diagram shows **actual message exchanges** that occurred during execution.\n\n")

	// Section 6: Metrics (Observability Counters)
	md.WriteString("## 6. Metrics (Observability Counters)\n\n")

	// Use kripke library method
	md.WriteString(metrics.GenerateMetricsTable())
	md.WriteString("\n")

	md.WriteString("### Metrics Chart\n\n")
	md.WriteString("```mermaid\n")
	md.WriteString(metrics.GenerateMetricsChart([]string{
		"upload_chunks",
		"download_chunks",
	}))
	md.WriteString("```\n\n")

	// Section 7: Throughput Analysis
	md.WriteString("## 7. Throughput Analysis\n\n")

	uploadThroughput := kripke.CalculateThroughput(
		uploader.totalBytesSent,
		startTime,
		endTime,
	)

	downloadThroughput := kripke.CalculateThroughput(
		receiver.totalBytesReceived,
		startTime,
		endTime,
	)

	throughputs := map[string]kripke.Throughput{
		"upload":   uploadThroughput,
		"download": downloadThroughput,
	}

	// Use kripke library method
	md.WriteString(kripke.GenerateThroughputTable(throughputs))
	md.WriteString("\n")

	md.WriteString(fmt.Sprintf("**Upload Throughput**: %s\n", uploadThroughput.String()))
	md.WriteString(fmt.Sprintf("**Download Throughput**: %s\n\n", downloadThroughput.String()))

	// Section 8: Summary
	md.WriteString("## 8. Summary\n\n")

	md.WriteString("### Workflow\n\n")
	md.WriteString("```\n")
	md.WriteString("1. LLM reads English prompt\n")
	md.WriteString("2. LLM generates Go code with actor definitions\n")
	md.WriteString("3. Go code imports kripke library\n")
	md.WriteString("4. Go code calls library methods:\n")
	md.WriteString("   - GenerateActorStateTable() - Show actor variables\n")
	md.WriteString("   - GenerateTransitionTable() - Show guards + actions\n")
	md.WriteString("   - GenerateInteractionDiagram() - Show message exchanges\n")
	md.WriteString("   - GenerateMetricsTable() - Show counters\n")
	md.WriteString("   - GenerateThroughputTable() - Calculate throughput\n")
	md.WriteString("5. Library generates all diagrams\n")
	md.WriteString("```\n\n")

	md.WriteString("### Key Features\n\n")
	md.WriteString("1. **States = Actor Code**: States are actor instances with variables\n")
	md.WriteString("2. **State Changes = Variable Edits**: `x' = x + 1` changes state\n")
	md.WriteString("3. **Transitions = Guards + Actions**: Predicates enable, actions modify\n")
	md.WriteString("4. **Observability = Counters**: Track bytes, time, throughput\n")
	md.WriteString("5. **First-class Message Passing**: `send()` and `recv()` primitives\n")
	md.WriteString("6. **Separate Actors**: Each actor maintains its own state\n\n")

	md.WriteString("### Why Not TLA+?\n\n")
	md.WriteString("TLA+ lacks:\n")
	md.WriteString("- ❌ Easy instantiation/duplication of actors\n")
	md.WriteString("- ❌ First-class message passing\n")
	md.WriteString("- ❌ Separate actor instances\n")
	md.WriteString("- ❌ Chance nodes (probabilistic internal decisions)\n")
	md.WriteString("- ❌ Observability counters (metrics, throughput)\n\n")

	md.WriteString("This library provides:\n")
	md.WriteString("- ✅ Actor-based state machines\n")
	md.WriteString("- ✅ Message passing primitives\n")
	md.WriteString("- ✅ Independent actor state\n")
	md.WriteString("- ✅ Probabilistic chance nodes\n")
	md.WriteString("- ✅ Built-in metrics/observability\n\n")

	md.WriteString("---\n\n")
	md.WriteString("*All diagrams generated by kripke library methods - no code duplication in examples.*\n")

	// Write file
	filename := "upload-system-specification.md"
	if err := os.WriteFile(filename, []byte(md.String()), 0644); err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		return
	}

	fmt.Printf("✅ Generated: %s\n\n", filename)
	fmt.Println("Workflow:")
	fmt.Println("  1. LLM reads English prompt")
	fmt.Println("  2. LLM generates Go code (this file)")
	fmt.Println("  3. Go code imports kripke library")
	fmt.Println("  4. Go code calls library methods:")
	fmt.Println("     - GenerateActorStateTable()")
	fmt.Println("     - GenerateTransitionTable()")
	fmt.Println("     - GenerateInteractionDiagram()")
	fmt.Println("     - GenerateMetricsTable()")
	fmt.Println("     - GenerateThroughputTable()")
	fmt.Println("  5. Library generates all diagrams")
	fmt.Println()
	fmt.Printf("📊 Results:\n")
	fmt.Printf("   Uploader sent: %d chunks (%d bytes)\n", uploader.chunksSent, uploader.totalBytesSent)
	fmt.Printf("   Receiver got: %d chunks (%d bytes)\n", receiver.chunksReceived, receiver.totalBytesReceived)
	fmt.Printf("   Upload throughput: %s\n", uploadThroughput.String())
	fmt.Printf("   Download throughput: %s\n", downloadThroughput.String())
}
