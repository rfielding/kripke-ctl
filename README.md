# kripke-ctl

A formal methods framework for modeling and verifying concurrent systems, leveraging LLMs to make formal verification more accessible than traditional approaches.

## What is kripke-ctl?

Kripke-ctl bridges the gap between informal system descriptions and formal verification. It provides:

- **Executable specifications** - Models that run and can be verified
- **LLM-friendly** - Generate models from English descriptions
- **TLA+ integration** - Automatic TLA+ spec generation with KripkeLib operators
- **Probabilistic modeling** - Handle chance nodes and MDPs (which vanilla TLA+ struggles with)
- **Message passing** - Clean CSP/π-calculus-style communication
- **Multiple actor instances** - Easy to maintain separate actor instances

## Why kripke-ctl?

Traditional TLA+ is powerful but has challenges:
- **Probability**: No native support for probabilistic transitions
- **Message passing**: Verbose and unintuitive
- **Multiple instances**: Managing many actor instances is tedious
- **Learning curve**: Steep for most developers

Kripke-ctl makes formal methods practical by:
- Providing intuitive Go-based model definitions
- Generating TLA+ specs automatically
- Handling probability natively with pre-rolled dice
- Making message passing clean with process calculus semantics
- Letting LLMs bridge English → formal model

## Core Architecture

### CANDIDATE = guard matches AND not blocked

Each Process represents ONE transition:
- Ready() checks ONE predicate (guard)
- Returns ONE step (or nil)
- Engine picks one candidate uniformly at random

```go
type Actor struct {
    State string
    Count int
}

func (a *Actor) Ready(w *World) []Step {
    // Guard: application logic
    if a.Count >= 10 {
        return nil
    }
    
    // Return ONE step
    return []Step{
        func(w *World) {
            // Action: send/recv/assign
            SendMessage(w, Message{...})
            a.Count++
        },
    }
}
```

### Key Concepts

**Nested States**: Multiple predicates can match simultaneously

**Probabilistic Choice**: Pre-rolled dice in predicates

**Variable Assignment**: Only way to change which guards match

## Examples

See `docs/examples/` for complete working examples:

- **nested_states_example.go** - Shows nested state spaces
- **probabilistic_choice_example.go** - 70/30 split with dice
- **pure_markov_chain_example.go** - Pure probabilistic transitions
- **bakery_example.go** - Real-world business process (production workflow, logistics, sales)

The bakery example was the motivation for adding probability support - business processes need chance nodes! But kripke-ctl is a general formal methods tool.

## TLA+ Integration

### KripkeLib Operators

```tla
---- MODULE YourSystem ----
EXTENDS Naturals, Sequences, KripkeLib

Producer ==
    /\ count < 10
    /\ can_send(channel, capacity)
    /\ channel' = snd(channel, msg)     \* Process calculus: channel ! msg
    /\ count' = count + 1

Consumer ==
    LET result == rcv(channel) IN       \* Process calculus: channel ? msg
    /\ can_recv(channel)
    /\ channel' = result.channel
    /\ count' = count + 1

\* Probabilistic choice (0-100 percentages)
PathChoice ==
    \/ choice(0, 70, TRUE, state' = "A")    \* 70% 
    \/ choice(70, 100, TRUE, state' = "B")  \* 30%
====
```

Process calculus semantics that vanilla TLA+ lacks.

## Web Interface

Generate comprehensive requirements documents from English:

```bash
cd cmd/docs
go run .
# Visit http://localhost:8080
```

English description → executable Go model + TLA+ spec + diagrams + metrics.

## Installation

```bash
go get github.com/rfielding/kripke-ctl
```

## Documentation

- `docs/architecture/ARCHITECTURE.md` - Core concepts
- `docs/architecture/PURE_MARKOV_CHAINS.md` - Markov chain pattern
- `docs/examples/` - Working examples

## Use Cases

- **Distributed systems** - Consensus protocols, replication, failures
- **Concurrent algorithms** - Locks, barriers, data structures
- **Communication protocols** - TCP, gossip, message ordering
- **Business processes** - Workflows with probabilistic outcomes
- **Robotics** - Planning with uncertainty
- **Game theory** - Strategic interactions

## Philosophy

**Never lose progress. Requirements before code.**

The requirements document is the source of truth.

## Key Innovations

1. **Process calculus semantics** - Clean message passing (`snd`, `rcv`)
2. **Native probability** - Pre-rolled dice for model checking
3. **LLM integration** - English → formal model
4. **TLA+ generation** - Automatic spec creation
5. **Simple architecture** - CANDIDATE = guard AND not blocked

## License

[Your License Here]
