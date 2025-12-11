# Kripke Architecture: The Correct Model

## Core Concept

**CANDIDATE = guard matches AND not blocked**

Each Process represents ONE state/transition:
- Ready() checks ONE predicate (guard)
- Returns ONE step (or nil)
- If guard matches AND channel wouldn't block → CANDIDATE
- Engine picks one candidate uniformly at random

## The Flow

```
1. For each Process:
   a. Check predicate (application guard)
   b. If false → skip
   c. If true → check if would block (send to full / recv from empty)
   d. If would block → skip (unschedulable)
   e. If not blocked → ADD TO CANDIDATES

2. Pick one candidate uniformly at random

3. Execute the step
```

## Ready() Returns ONE Step

```go
func (a *Actor) Ready(w *World) []Step {
    // Check predicate
    if !guard_matches {
        return nil  // Not a candidate
    }
    
    // Return ONE step (engine checks blocking)
    return []Step{
        func(w *World) {
            // The action for this transition
        },
    }
}
```

Note: Returns `[]Step` with single element for API compatibility, but conceptually it's ONE step.

## Nested States

Multiple predicates can match simultaneously → actor is in MULTIPLE states.

### Example

When x = 5:
- Predicate "x > 0" matches ✓
- Predicate "x < 20" matches ✓
- Predicate "x < 4" doesn't match ✗

The actor is in BOTH "x>0" state AND "x<20" state simultaneously. These are **nested states**.

Both transitions are candidates → engine picks one at random.

## Probabilistic Choice

Pre-rolled dice (R1, R2, ...) are referenced in predicates to select transitions.

Dice are rolled BEFORE checking predicates, and the dice value determines which predicate matches.

Example: 70% path A, 30% path B
- Roll R1 ∈ [0, 100)
- If 0 ≤ R1 < 70 → path A predicate matches
- If 70 ≤ R1 < 100 → path B predicate matches

## Step Actions

A step does ONE of:
1. **Variable update**: `x' = x + 1`
2. **Receive** (assignment): `msg' = recv(ch)`
3. **Send**: `send(ch, msg)`

## Why This Architecture?

### Separation of Concerns
- **Predicates** → availability
- **Engine** → scheduling
- **Actions** → state changes

### Composability
Multiple processes can share state.

### Model Checking
TLC can explore all possible executions.

### Forms MDPs Naturally
Bipartite: Regular → Chance → Regular → Chance

## Summary

**CANDIDATE = guard matches AND not blocked**

Simple, compositional, mathematically sound. 🎯
