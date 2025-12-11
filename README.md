# kripke-ctl

A framework for modeling concurrent systems as communicating state machines with probabilistic behavior.

## The Canonical Example: Bakery Simulation

This framework was built to model a real bakery business:

**Actors**:
- **Production**: Makes bread (dough → kneading → baking → cooling)
- **Truck**: Transports bread (loading → driving → unloading)
- **Storefront**: Manages inventory and sales
- **Customers**: Arrive and purchase bread

**Communication**: Message passing (load bread, deliver bread, purchase bread)

**Business Metrics**: Automatically tracked
- Costs (hourly rates for workers)
- Revenue (sales from customers)
- Waste (unsold inventory)
- Popularity (which breads sell most)

**Business Questions Answered**:
1. What are our profits?
2. How much waste do we have?
3. What are the most popular breads?

See `docs/examples/bakery_example.go` for the complete implementation.

## Architecture

### Core Concept

**CANDIDATE = guard matches AND not blocked**

- Each `Process` represents ONE transition (one predicate check)
- `Ready()` returns ONE step (or nil)
- If guard matches AND channel wouldn't block → CANDIDATE
- Engine picks one candidate uniformly at random

## Step Pattern

**Correct Pattern**:

```go
type Producer struct {
    State string
    Count int
}

func (p *Producer) ID() string { return "producer" }

// Ready() checks ONE predicate and returns ONE step
func (p *Producer) Ready(w *World) []Step {
    // Guard: check application logic
    if p.Count >= 10 {
        return nil  // Predicate doesn't match
    }
    
    // Return ONE step
    // Engine will check if channels would block
    return []Step{
        func(w *World) {
            // The action for this transition
            SendMessage(w, Message{...})
            p.Count++
        },
    }
}
```

**Key points**:
- ✅ Ready() checks ONE predicate
- ✅ Returns ONE step (single-element array)
- ✅ Guard checks application logic
- ✅ Engine checks channel blocking
- ✅ Step contains the action (send/recv/assign)

## Examples

See `docs/examples/` for complete working examples:
- `bakery_example.go` - The canonical example (business metrics)
- `nested_states_example.go` - Shows nested states
- `probabilistic_choice_example.go` - Shows 70/30 probabilistic split
- `pure_markov_chain_example.go` - Shows pure Markov chains

## Documentation

- `docs/architecture/ARCHITECTURE.md` - Core architecture
- `docs/architecture/PURE_MARKOV_CHAINS.md` - Markov chain pattern
- `cmd/docs/README.md` - Web interface

## Key Insights

1. **CANDIDATE = guard matches AND not blocked**
2. **Nested states**: Multiple predicates can match simultaneously
3. **Variable assignment**: The only way to change which guards match
4. **Natural metrics**: Business metrics captured through state machine execution

## License

[Your License Here]
