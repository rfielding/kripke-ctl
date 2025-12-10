package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/rfielding/kripke-ctl/kripke"
)

// Example demonstrating the Actor Readiness Model:
// 1. NOT blocked on send
// 2. NOT blocked on recv
// 3. Variable predicate matches (guard)
// 4. Uniform random scheduling

type Producer struct {
	id        string
	itemsSent int
	maxItems  int
}

func (p *Producer) ID() string { return p.id }

func (p *Producer) Ready(w *kripke.World) []kripke.Step {
	// Readiness Check 1: NOT blocked on send?
	ch := w.ChannelByAddress(kripke.Address{ActorID: "consumer", ChannelName: "inbox"})
	if ch == nil || !ch.CanSend() {
		// ❌ Blocked on send - channel full
		return nil
	}

	// Readiness Check 2: Variable predicate (guard)?
	if p.itemsSent >= p.maxItems {
		// ❌ Guard failed - already sent max items
		return nil
	}

	// ✅ All checks pass - return possible steps
	return []kripke.Step{
		func(w *kripke.World) {
			// Action: send message + edit variables
			kripke.SendMessage(w, kripke.Message{
				From:    kripke.Address{ActorID: p.id, ChannelName: "out"},
				To:      kripke.Address{ActorID: "consumer", ChannelName: "inbox"},
				Payload: fmt.Sprintf("item_%d", p.itemsSent),
			})
			p.itemsSent++ // Variable edit: itemsSent' = itemsSent + 1
		},
	}
}

type Consumer struct {
	id           string
	inbox        *kripke.Channel
	itemsRecv    int
	totalDelay   int
	processingSlow bool // Variable affecting behavior
}

func (c *Consumer) ID() string { return c.id }

func (c *Consumer) Ready(w *kripke.World) []kripke.Step {
	// Readiness Check 1: NOT blocked on recv?
	if !c.inbox.CanRecv() {
		// ❌ Blocked on recv - channel empty
		return nil
	}

	// Readiness Check 2: Variable predicate?
	// (Consumer always processes if items available)
	// Could add: if c.processingSlow { return nil }

	// ✅ All checks pass - return possible steps
	return []kripke.Step{
		func(w *kripke.World) {
			// Action: recv message + edit variables
			_, delay := kripke.RecvAndLog(w, c.inbox)
			c.itemsRecv++        // Variable edit
			c.totalDelay += delay // Variable edit
		},
	}
}

type Logger struct {
	id        string
	logsWritten int
	enabled   bool // Variable guard
}

func (l *Logger) ID() string { return l.id }

func (l *Logger) Ready(w *kripke.World) []kripke.Step {
	// Readiness Check: Variable predicate only
	if !l.enabled {
		// ❌ Guard failed - logging disabled
		return nil
	}

	if l.logsWritten >= 5 {
		// ❌ Guard failed - enough logs
		return nil
	}

	// ✅ Guard passes - return possible step
	return []kripke.Step{
		func(w *kripke.World) {
			// Action: variable edit only (no send/recv)
			l.logsWritten++
		},
	}
}

func main() {
	var md strings.Builder

	md.WriteString("# Actor Readiness Model: Complete Example\n\n")
	md.WriteString("Demonstrates:\n")
	md.WriteString("1. Blocking on send/recv\n")
	md.WriteString("2. Variable predicates (guards)\n")
	md.WriteString("3. Uniform random scheduling\n\n")
	md.WriteString("---\n\n")

	// Section 1: Actor Definitions
	md.WriteString("## 1. Actor Definitions\n\n")

	md.WriteString("### Producer\n\n")
	md.WriteString("```go\n")
	md.WriteString("func (p *Producer) Ready(w *World) []Step {\n")
	md.WriteString("    // Check 1: NOT blocked on send?\n")
	md.WriteString("    if !ch.CanSend() {\n")
	md.WriteString("        return nil  // Blocked\n")
	md.WriteString("    }\n")
	md.WriteString("    \n")
	md.WriteString("    // Check 2: Variable guard?\n")
	md.WriteString("    if p.itemsSent >= p.maxItems {\n")
	md.WriteString("        return nil  // Guard failed\n")
	md.WriteString("    }\n")
	md.WriteString("    \n")
	md.WriteString("    // Ready!\n")
	md.WriteString("    return []Step{...}\n")
	md.WriteString("}\n")
	md.WriteString("```\n\n")

	md.WriteString("**Readiness conditions:**\n")
	md.WriteString("- ✅ Channel not full (`ch.CanSend()`)\n")
	md.WriteString("- ✅ Under item limit (`itemsSent < maxItems`)\n\n")

	md.WriteString("### Consumer\n\n")
	md.WriteString("```go\n")
	md.WriteString("func (c *Consumer) Ready(w *World) []Step {\n")
	md.WriteString("    // Check 1: NOT blocked on recv?\n")
	md.WriteString("    if !c.inbox.CanRecv() {\n")
	md.WriteString("        return nil  // Blocked\n")
	md.WriteString("    }\n")
	md.WriteString("    \n")
	md.WriteString("    // Ready!\n")
	md.WriteString("    return []Step{...}\n")
	md.WriteString("}\n")
	md.WriteString("```\n\n")

	md.WriteString("**Readiness conditions:**\n")
	md.WriteString("- ✅ Channel not empty (`inbox.CanRecv()`)\n\n")

	// Section 2: Execution Trace
	md.WriteString("## 2. Execution Trace (Scheduler Decisions)\n\n")

	producer := &Producer{id: "producer", maxItems: 10}
	inbox := kripke.NewChannel("consumer", "inbox", 2) // Capacity 2
	consumer := &Consumer{id: "consumer", inbox: inbox}
	logger := &Logger{id: "logger", enabled: true}

	w := kripke.NewWorld(
		[]kripke.Process{producer, consumer, logger},
		[]*kripke.Channel{inbox},
		42,
	)

	md.WriteString("| Step | Buffer | P.Ready? | C.Ready? | L.Ready? | Possible | Chosen | Action |\n")
	md.WriteString("|------|--------|----------|----------|----------|----------|--------|--------|\n")

	for step := 1; step <= 20; step++ {
		bufSize := inbox.Len()

		// Check readiness
		pReady := producer.Ready(w) != nil
		cReady := consumer.Ready(w) != nil
		lReady := logger.Ready(w) != nil

		// Count ready actors
		readyActors := []string{}
		if pReady {
			readyActors = append(readyActors, "P")
		}
		if cReady {
			readyActors = append(readyActors, "C")
		}
		if lReady {
			readyActors = append(readyActors, "L")
		}

		if len(readyActors) == 0 {
			md.WriteString(fmt.Sprintf("| %d | %d | ❌ | ❌ | ❌ | none | - | QUIESCED |\n",
				step, bufSize))
			break
		}

		possible := strings.Join(readyActors, ",")

		// Execute step
		beforeBuf := inbox.Len()
		beforeSent := producer.itemsSent
		beforeRecv := consumer.itemsRecv
		beforeLogs := logger.logsWritten

		if !w.StepRandom() {
			break
		}

		afterBuf := inbox.Len()

		// Determine what happened
		var chosen, action string
		if producer.itemsSent > beforeSent {
			chosen = "P"
			action = fmt.Sprintf("send (sent=%d)", producer.itemsSent)
		} else if consumer.itemsRecv > beforeRecv {
			chosen = "C"
			action = fmt.Sprintf("recv (recv=%d)", consumer.itemsRecv)
		} else if logger.logsWritten > beforeLogs {
			chosen = "L"
			action = fmt.Sprintf("log (logs=%d)", logger.logsWritten)
		}

		pReadyStr := "❌"
		if pReady {
			pReadyStr = "✅"
		}
		cReadyStr := "❌"
		if cReady {
			cReadyStr = "✅"
		}
		lReadyStr := "❌"
		if lReady {
			lReadyStr = "✅"
		}

		md.WriteString(fmt.Sprintf("| %d | %d | %s | %s | %s | %s | %s | %s |\n",
			step, beforeBuf, pReadyStr, cReadyStr, lReadyStr, possible, chosen, action))
	}

	md.WriteString("\n")

	// Section 3: Analysis
	md.WriteString("## 3. Analysis\n\n")

	md.WriteString("### Blocking Behavior\n\n")
	md.WriteString("**Producer blocked when:**\n")
	md.WriteString("- Buffer full (size = 2)\n")
	md.WriteString("- OR sent max items (itemsSent >= 10)\n\n")

	md.WriteString("**Consumer blocked when:**\n")
	md.WriteString("- Buffer empty (size = 0)\n\n")

	md.WriteString("**Logger blocked when:**\n")
	md.WriteString("- Disabled (enabled = false)\n")
	md.WriteString("- OR written enough logs (logsWritten >= 5)\n\n")

	md.WriteString("### Uniform Random Scheduling\n\n")
	md.WriteString("At each step:\n")
	md.WriteString("1. Check all actors: `Ready()` for each\n")
	md.WriteString("2. Collect possible steps (non-nil returns)\n")
	md.WriteString("3. Pick one uniformly at random: `choice = ready[rand.Intn(len(ready))]`\n")
	md.WriteString("4. Execute chosen step\n\n")

	md.WriteString("**Example**: If P, C, L all ready:\n")
	md.WriteString("- P(Producer chosen) = 1/3\n")
	md.WriteString("- P(Consumer chosen) = 1/3\n")
	md.WriteString("- P(Logger chosen) = 1/3\n\n")

	// Section 4: Readiness Summary
	md.WriteString("## 4. Readiness Summary\n\n")

	md.WriteString("| Actor | Condition | Check |\n")
	md.WriteString("|-------|-----------|-------|\n")
	md.WriteString("| Producer | NOT blocked send | `ch.CanSend()` |\n")
	md.WriteString("| Producer | Guard (count) | `itemsSent < maxItems` |\n")
	md.WriteString("| Consumer | NOT blocked recv | `inbox.CanRecv()` |\n")
	md.WriteString("| Logger | Guard (enabled) | `enabled == true` |\n")
	md.WriteString("| Logger | Guard (count) | `logsWritten < 5` |\n\n")

	md.WriteString("### Final State\n\n")
	md.WriteString(fmt.Sprintf("- **Producer**: sent %d items\n", producer.itemsSent))
	md.WriteString(fmt.Sprintf("- **Consumer**: received %d items\n", consumer.itemsRecv))
	md.WriteString(fmt.Sprintf("- **Logger**: wrote %d logs\n", logger.logsWritten))
	md.WriteString(fmt.Sprintf("- **Buffer**: %d items remaining\n\n", inbox.Len()))

	// Section 5: Code Template
	md.WriteString("## 5. Actor Readiness Template\n\n")
	md.WriteString("```go\n")
	md.WriteString("func (a *Actor) Ready(w *kripke.World) []kripke.Step {\n")
	md.WriteString("    // Check 1: NOT blocked on send?\n")
	md.WriteString("    if need_to_send {\n")
	md.WriteString("        ch := w.ChannelByAddress(...)\n")
	md.WriteString("        if ch == nil || !ch.CanSend() {\n")
	md.WriteString("            return nil  // Blocked on send\n")
	md.WriteString("        }\n")
	md.WriteString("    }\n")
	md.WriteString("    \n")
	md.WriteString("    // Check 2: NOT blocked on recv?\n")
	md.WriteString("    if need_to_recv {\n")
	md.WriteString("        if !ch.CanRecv() {\n")
	md.WriteString("            return nil  // Blocked on recv\n")
	md.WriteString("        }\n")
	md.WriteString("    }\n")
	md.WriteString("    \n")
	md.WriteString("    // Check 3: Variable predicates (guards)?\n")
	md.WriteString("    if !guard_condition {\n")
	md.WriteString("        return nil  // Guard failed\n")
	md.WriteString("    }\n")
	md.WriteString("    \n")
	md.WriteString("    // All checks pass - return possible steps\n")
	md.WriteString("    return []kripke.Step{\n")
	md.WriteString("        func(w *kripke.World) {\n")
	md.WriteString("            // Actions:\n")
	md.WriteString("            // 1. Variable edits: a.x++\n")
	md.WriteString("            // 2. Send message: SendMessage(...)\n")
	md.WriteString("            // 3. Recv message: RecvAndLog(...)\n")
	md.WriteString("        },\n")
	md.WriteString("    }\n")
	md.WriteString("}\n")
	md.WriteString("```\n\n")

	md.WriteString("---\n\n")
	md.WriteString("**Key Points:**\n")
	md.WriteString("1. Actor ready ⟺ NOT blocked AND guards pass\n")
	md.WriteString("2. `Ready()` returns `nil` if not ready\n")
	md.WriteString("3. `Ready()` returns `[]Step` if ready\n")
	md.WriteString("4. Scheduler picks uniformly at random from all ready actors\n")

	// Write file
	filename := "actor-readiness-example.md"
	if err := os.WriteFile(filename, []byte(md.String()), 0644); err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		return
	}

	fmt.Printf("✅ Generated: %s\n\n", filename)
	fmt.Println("Actor Readiness Model:")
	fmt.Println("  1. NOT blocked on send")
	fmt.Println("  2. NOT blocked on recv")
	fmt.Println("  3. Variable predicates (guards) pass")
	fmt.Println("  4. Uniform random scheduling")
	fmt.Println()
	fmt.Printf("Results:\n")
	fmt.Printf("  Producer: sent %d/%d items\n", producer.itemsSent, producer.maxItems)
	fmt.Printf("  Consumer: received %d items (delay=%d)\n", consumer.itemsRecv, consumer.totalDelay)
	fmt.Printf("  Logger: wrote %d logs\n", logger.logsWritten)
}
