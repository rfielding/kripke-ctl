# Pure Markov Chains in Kripke

## The Key Insight

**The ONLY way to change which guards match is to modify variables.**

Guards check variables → variable assignment changes which transitions are available.

## Pure Markov Chain = Only Chance Nodes

If you want a pure Markov chain:
- Every transition is probabilistic (rolls dice)
- No deterministic work (no sends, no complex computations)
- Just: roll dice → check range → assign state variable

## The Pattern

```go
// State variable
type MarkovActor struct {
    State string  // Current state
    Dice  int     // Pre-rolled dice
}

// Step 1: Roll dice in state A
type RollInA struct { Actor *MarkovActor }
func (r *RollInA) Ready(w *World) []Step {
    if r.Actor.State != "A" { return nil }
    return []Step{
        func(w *World) {
            r.Actor.Dice = rand.Intn(100)
            r.Actor.State = "choosing"
        },
    }
}

// Step 2: Transition A→B (40% chance)
type A_to_B struct { Actor *MarkovActor }
func (a *A_to_B) Ready(w *World) []Step {
    if a.Actor.State != "choosing" { return nil }
    if a.Actor.Dice < 0 || a.Actor.Dice >= 40 { return nil }
    return []Step{
        func(w *World) {
            a.Actor.State = "B"  // Variable assignment!
        },
    }
}

// Step 3: Transition A→C (60% chance)
type A_to_C struct { Actor *MarkovActor }
func (a *A_to_C) Ready(w *World) []Step {
    if a.Actor.State != "choosing" { return nil }
    if a.Actor.Dice < 40 || a.Actor.Dice >= 100 { return nil }
    return []Step{
        func(w *World) {
            a.Actor.State = "C"  // Variable assignment!
        },
    }
}
```

## Why This Works

1. **Guards check variables**: `if a.Actor.State != "A"`
2. **Variable assignment changes state**: `a.Actor.State = "B"`
3. **New guards match**: Now `State == "B"` guards become true
4. **Cycle repeats**: Roll dice from B, transition to next state

## Example: Three-State Markov Chain

```
State A: 40% → B, 60% → C
State B: 50% → A, 50% → C  
State C: 100% → A
```

This requires:
- 2 chance nodes for A (roll + choose)
- 2 chance nodes for B (roll + choose)
- 1 deterministic transition for C
- Total: 7 Process objects

Each transition:
```go
if state != "expected" { return nil }  // Guard checks variable
actor.State = "new_state"              // Assignment changes variable
```

## Pure Markov Chain Properties

✅ **Every transition is probabilistic** (except deterministic 100% ones)
✅ **No complex logic** - just state changes
✅ **Variable assignment is the ONLY state change mechanism**
✅ **Guards naturally form the transition matrix**

## In TLA+

```tla
\* Pure Markov chain
A_to_B ==
    /\ state = "A"
    /\ 0 <= R1 < 40
    /\ state' = "B"

A_to_C ==
    /\ state = "A"
    /\ 40 <= R1 < 100
    /\ state' = "C"

\* etc...
```

Clean, simple, pure!

## When to Use This

Use pure Markov chains when:
- Modeling random processes (weather, stock prices)
- Simulating probabilistic state machines
- No complex computation needed
- Just state-to-state transitions

## When NOT to Use This

Don't use pure Markov chains when:
- Need deterministic computation (send/recv messages)
- Need complex logic between states
- Need to maintain additional state beyond "current state"

For those cases, use the **bipartite structure**: Regular states (with logic) ↔ Chance nodes (probabilistic)

## Summary

**Pure Markov Chain in Kripke:**
- Only chance nodes (probabilistic transitions)
- Variable assignment is the ONLY way to change state
- Guards check variables → assignment changes guards
- Simple, clean, mathematically pure

**The key**: Guards check variables, so variable assignment is the mechanism for state changes! 🎯
