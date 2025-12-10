# Quick Reference: kripke-ctl Updates

## 📦 Package Contents

- **2 library files** (add to `kripke/`)
- **3 example files** (demonstrate usage)
- **5 documentation files** (complete specs)
- **1 README** (this summary)

## 🎯 Key Concepts (5 Minutes)

### 1. States = Actor CODE
```go
type Uploader struct {
    chunksSent     int    // State variable
    totalBytesSent int64  // State variable
}
```
State = `{chunksSent: 5, totalBytesSent: 5242880}`

### 2. Transitions = Guards + Actions
```go
// Guard: predicate that enables
if !ch.CanSend() || count >= MAX {
    return nil  // Not ready
}

// Action: modify state
return []Step{
    func(w *World) {
        SendMessage(...)  // Send
        count++          // Variable edit
    },
}
```

### 3. Actor Ready If:
1. ✅ NOT blocked on send (if trying to send)
2. ✅ NOT blocked on recv (if trying to recv)
3. ✅ Variable guard passes
4. ✅ Returns `[]Step`

### 4. Scheduling: Uniform Random
```
1. Collect all Ready() from actors
2. Pick one uniformly at random
3. Execute chosen step
4. Repeat until quiesced
```

### 5. Chance Nodes (MDP)
```go
if rand() < 0.6 {
    send(1 msg)   // 60%
} else {
    send(2 msgs)  // 40%
}
```

## 🚀 Integration (3 Steps)

### Step 1: Add Library Files
```bash
cd kripke-ctl/kripke
cp library/diagrams.go .    # Diagram generation
cp library/metrics.go .     # Metrics & observability
```

### Step 2: Update Examples
```go
import "github.com/rfielding/kripke-ctl/kripke"

diagram := w.GenerateActorStateMachine()
metrics := w.GenerateMetricsTable()
```

### Step 3: Test
```bash
cd examples
go run example_proper_usage.go
```

## 📚 Documentation Reading Order

1. **COMPLETE_VISION.md** (10 min) - Overview
2. **ACTOR_READINESS_MODEL.md** (5 min) - Actor model
3. **WHY_CHANCE_NODES.md** (5 min) - Design rationale
4. **PROPER_ARCHITECTURE.md** (5 min) - Library structure
5. **INTEGRATION_GUIDE.md** (10 min) - How to integrate

## 🔑 Critical Points

### Why Chance Nodes?
- ❌ Tried: Communicating Markov Chains (only random transitions)
- ✅ Solution: MDP with chance nodes (actions + probability)

### Blocking Semantics
- ONLY blocked if **TRYING** to send/recv
- Go buffered channel rules: `length > 0`

### Messages as Counters
- Line charts: time series data
- Pie charts: categorical distribution

### Library Architecture
- ✅ Diagram generation in `kripke/` package
- ✅ Examples use library methods
- ❌ NO code duplication in examples

## 📊 Library Methods

### Diagrams
```go
g.GenerateStateDiagram()          // State machine
w.GenerateSequenceDiagram(10)     // Message sequence
w.GenerateActorStateMachine()     // Actors with variables
```

### Verification
```go
g.GenerateCTLTable(requirements)  // CTL results
GenerateRequirementsTable(reqs)   // Requirements table
```

### Metrics
```go
mc := NewMetricsCollector()
counter := mc.Counter("bytes", "Total", "bytes")
mc.GenerateMetricsTable()
mc.GenerateMetricsChart(names)
```

### Throughput
```go
tp := CalculateThroughput(bytes, start, end)
GenerateThroughputTable(throughputs)
```

## 🎨 LLM Workflow

```
English: "Create upload system with throughput"
    ↓
LLM generates Go code with kripke import
    ↓
Code calls library methods
    ↓
Library generates diagrams/metrics
    ↓
Complete specification
```

## ✅ Advantages vs TLA+

| Feature | TLA+ | kripke |
|---------|------|--------|
| Actor creation | Hard | Easy |
| Message passing | Manual | Built-in |
| Actors | Global state | Independent |
| Probability | No | Yes (chance nodes) |
| Metrics | No | Yes |
| Diagrams | Manual | Auto-generated |

## 📝 Example Template

```go
package main
import "github.com/rfielding/kripke-ctl/kripke"

type MyActor struct {
    id    string
    count int  // State variable
}

func (a *MyActor) ID() string { return a.id }

func (a *MyActor) Ready(w *kripke.World) []kripke.Step {
    // Check blocking
    if blocked { return nil }
    
    // Check guard
    if a.count >= MAX { return nil }
    
    // Ready!
    return []kripke.Step{
        func(w *kripke.World) {
            // Actions: edit vars, send, recv
            a.count++
            kripke.SendMessage(...)
        },
    }
}

func main() {
    w := kripke.NewWorld(...)
    w.Run()
    
    // Use library methods
    stateDiagram := w.GenerateActorStateMachine()
    metrics := w.GenerateMetricsTable()
}
```

## 🎯 What This Enables

1. ✅ LLM generates Go code from English
2. ✅ Actor-based state machines
3. ✅ First-class message passing
4. ✅ Communicating MDP (not Markov Chains)
5. ✅ Chance nodes for probability
6. ✅ Metrics and observability
7. ✅ Auto-generated diagrams
8. ✅ Proper library architecture

## 📞 Files

### Library (add to kripke/)
- `diagrams.go` - Diagram generation
- `metrics.go` - Metrics, throughput, observability

### Examples
- `example_proper_usage.go` - Basic usage
- `example_llm_workflow.go` - LLM workflow with throughput
- `example_actor_readiness.go` - Readiness model demo

### Documentation
- `COMPLETE_VISION.md` - Full vision
- `ACTOR_READINESS_MODEL.md` - Actor model
- `WHY_CHANCE_NODES.md` - Design evolution
- `PROPER_ARCHITECTURE.md` - Architecture
- `INTEGRATION_GUIDE.md` - Integration steps

---

**Total Time to Integrate**: ~30 minutes  
**Total Time to Read Docs**: ~35 minutes  
**Result**: Complete actor-based model checker with diagrams
