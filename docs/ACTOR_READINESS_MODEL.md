# Actor Readiness and Scheduling Model

## When an Actor is Ready

An actor is **Ready** to take a step if and only if:

### 1. NOT blocked trying to send
```go
func (a *Actor) Ready(w *World) []Step {
    ch := w.ChannelByAddress(target)
    if ch == nil || !ch.CanSend() {
        return nil  // ❌ Blocked on send - NOT ready
    }
    // ... continue checking
}
```

**Blocked**: Channel is full, cannot send

### 2. NOT blocked trying to recv
```go
func (a *Actor) Ready(w *World) []Step {
    if !a.inbox.CanRecv() {
        return nil  // ❌ Blocked on recv - NOT ready
    }
    // ... continue checking
}
```

**Blocked**: Channel is empty, cannot receive

### 3. Variable predicate matches
```go
func (a *Actor) Ready(w *World) []Step {
    // Check variable state (guard/predicate)
    if a.counter >= MAX_COUNT {
        return nil  // ❌ Guard failed - NOT ready
    }
    
    if a.balance < MINIMUM {
        return nil  // ❌ Predicate doesn't match - NOT ready
    }
    
    // ✅ All guards pass - Ready!
    return []Step{ ... }
}
```

**Guard failed**: Variables don't satisfy the predicate

### 4. All Ready() items are possible
```go
func (a *Actor) Ready(w *World) []Step {
    // Can return MULTIPLE possible steps
    return []Step{
        func(w *World) { /* option 1 */ },
        func(w *World) { /* option 2 */ },
        func(w *World) { /* option 3 */ },
    }
}
```

**All steps in the returned slice are possible**

---

## Scheduling Algorithm

### Step 1: Collect Ready Actors
```go
ready := []Step{}

for _, actor := range world.Actors {
    steps := actor.Ready(world)
    if steps != nil && len(steps) > 0 {
        ready = append(ready, steps...)
    }
}
```

Result: List of all possible steps from all actors

### Step 2: Filter Non-Empty
```go
if len(ready) == 0 {
    // No actors ready - system quiesced
    return false
}
```

### Step 3: Pick Uniformly at Random
```go
index := rand.Intn(len(ready))
chosen := ready[index]
```

**Uniform random**: Each possible step has equal probability

### Step 4: Execute Chosen Step
```go
chosen(world)
```

---

## Complete Example

### Actor Definition
```go
type Producer struct {
    id           string
    itemsSent    int
    maxItems     int
}

func (p *Producer) ID() string { return p.id }

func (p *Producer) Ready(w *kripke.World) []kripke.Step {
    // Check 1: Not blocked on send?
    ch := w.ChannelByAddress(kripke.Address{
        ActorID: "consumer", 
        ChannelName: "inbox",
    })
    if ch == nil || !ch.CanSend() {
        return nil  // ❌ Blocked on send
    }
    
    // Check 2: Variable predicate?
    if p.itemsSent >= p.maxItems {
        return nil  // ❌ Guard failed (already sent max)
    }
    
    // ✅ All checks pass - return possible steps
    return []kripke.Step{
        func(w *kripke.World) {
            // Action: send + variable edit
            kripke.SendMessage(w, kripke.Message{
                From:    kripke.Address{ActorID: p.id, ChannelName: "out"},
                To:      kripke.Address{ActorID: "consumer", ChannelName: "inbox"},
                Payload: fmt.Sprintf("item_%d", p.itemsSent),
            })
            p.itemsSent++  // Variable edit
        },
    }
}
```

### Execution Trace

```
Step 1:
  Check all actors:
    Producer.Ready() → [step1]  ✓ Not blocked, guard passes
    Consumer.Ready() → nil      ✗ Blocked on recv (inbox empty)
  
  ready = [producer.step1]
  Pick uniformly random: producer.step1
  Execute: send(item_0); itemsSent = 1

Step 2:
  Check all actors:
    Producer.Ready() → [step1]  ✓ Not blocked, guard passes
    Consumer.Ready() → [step1]  ✓ Not blocked (inbox has items)
  
  ready = [producer.step1, consumer.step1]
  Pick uniformly random: consumer.step1
  Execute: recv(item_0); itemsRecv = 1

Step 3:
  Check all actors:
    Producer.Ready() → [step1]  ✓ Not blocked, guard passes
    Consumer.Ready() → nil      ✗ Blocked on recv (inbox empty)
  
  ready = [producer.step1]
  Pick uniformly random: producer.step1
  Execute: send(item_1); itemsSent = 2

Step N:
  Check all actors:
    Producer.Ready() → nil      ✗ Guard failed (itemsSent >= maxItems)
    Consumer.Ready() → nil      ✗ Blocked on recv (inbox empty)
  
  ready = []
  QUIESCED - system terminates
```

---

## Why "Ready()" Returns []Step

### Not Just Boolean
```go
// ❌ WRONG:
func (a *Actor) Ready() bool {
    return !blocked && guardPasses
}
```

### Returns Possible Steps
```go
// ✅ CORRECT:
func (a *Actor) Ready(w *World) []Step {
    if blocked || !guardPasses {
        return nil  // Not ready
    }
    return []Step{
        func(w *World) { /* do something */ },
    }
}
```

**Why?**
1. **Action attached**: Ready() includes WHAT to do, not just "can do"
2. **Multiple options**: Actor might have several possible actions
3. **Closure over state**: Step function can access actor's variables

---

## Blocking Conditions

### Send Blocking
```go
// Blocked if:
// - Channel doesn't exist
// - Channel is full (len(ch.queue) >= ch.capacity)

ch := w.ChannelByAddress(addr)
if ch == nil || !ch.CanSend() {
    return nil  // Blocked on send
}
```

### Recv Blocking
```go
// Blocked if:
// - Channel is empty (len(ch.queue) == 0)

if !ch.CanRecv() {
    return nil  // Blocked on recv
}
```

### Variable Predicate
```go
// Not ready if guard/predicate fails

if a.count >= MAX {
    return nil  // Guard failed
}

if a.buffer.IsFull() {
    return nil  // Predicate doesn't match
}

if !a.hasCredits() {
    return nil  // Condition not met
}
```

---

## Multiple Ready Actors

### Example: 3 Actors Ready
```go
// State: inbox has 1 item, outbox can send

Producer.Ready()  → [send_step]       ✓ Ready
Consumer.Ready()  → [recv_step]       ✓ Ready  
Logger.Ready()    → [log_step]        ✓ Ready

ready = [send_step, recv_step, log_step]

// Scheduler picks uniformly at random:
// P(send_step) = 1/3
// P(recv_step) = 1/3
// P(log_step) = 1/3

chosen = ready[rand.Intn(3)]
```

### Example: 1 Actor Ready
```go
// State: inbox empty, outbox can send

Producer.Ready()  → [send_step]       ✓ Ready
Consumer.Ready()  → nil               ✗ Blocked on recv
Logger.Ready()    → nil               ✗ Guard failed

ready = [send_step]

// Only one choice:
chosen = send_step
```

---

## Uniform Random Scheduling

### Properties

1. **Fairness**: Each ready actor has equal chance
2. **Nondeterministic**: Can't predict which actor runs
3. **State-space exploration**: Explores all interleavings
4. **Models concurrency**: Represents arbitrary scheduling

### Implementation
```go
func (w *World) StepRandom() bool {
    // Collect all possible steps
    var ready []Step
    for _, proc := range w.Processes {
        steps := proc.Ready(w)
        ready = append(ready, steps...)
    }
    
    // Check if any ready
    if len(ready) == 0 {
        return false  // Quiesced
    }
    
    // Pick uniformly at random
    index := w.Rand.Intn(len(ready))
    chosen := ready[index]
    
    // Execute
    chosen(w)
    w.CurrentStep++
    
    return true
}
```

---

## Guard vs Blocking

### Guard (Variable Predicate)
```go
if a.counter >= MAX {
    return nil  // Guard failed
}
```
- **Depends on**: Actor's internal variables
- **Controlled by**: Actor's own state
- **Example**: "sent enough messages", "balance too low"

### Blocking (Communication)
```go
if !ch.CanSend() {
    return nil  // Blocked
}
```
- **Depends on**: External channel state
- **Controlled by**: Other actors (who recv/send)
- **Example**: "channel full", "no messages available"

---

## Summary

### Actor Ready Conditions (ALL must be true)
1. ✅ NOT blocked on send (`ch.CanSend()`)
2. ✅ NOT blocked on recv (`ch.CanRecv()`)
3. ✅ Variable predicate matches (guard passes)
4. ✅ Returns non-empty `[]Step`

### Scheduler Behavior
1. Collect all `Ready()` from all actors
2. Filter to non-blocked actors
3. Pick one step uniformly at random
4. Execute chosen step
5. Repeat until no actors ready (quiesced)

### Key Insight
```
Ready() = NOT blocked + Guard passes + Returns []Step

Scheduler = Uniform random choice from all Ready()
```

This is the **communicating MDP** with:
- **Nondeterminism**: Scheduler choice (uniform random)
- **Communication**: Send/recv blocking
- **Guards**: Variable predicates
- **Actions**: Variable edits, sends, recvs

---

## Example Summary Table

| Actor | Blocked Send? | Blocked Recv? | Guard? | Ready? | Steps |
|-------|---------------|---------------|--------|--------|-------|
| Producer | ✅ No | N/A | ✅ Yes (count<10) | ✅ Yes | [send] |
| Consumer | N/A | ❌ Yes (empty) | ✅ Yes | ❌ No | nil |

**Scheduler picks**: producer.send (only option)

---

## Code Template

```go
func (a *Actor) Ready(w *kripke.World) []kripke.Step {
    // Check blocking conditions
    if blocked_on_send || blocked_on_recv {
        return nil
    }
    
    // Check variable predicates (guards)
    if !guard_passes {
        return nil
    }
    
    // Return possible steps (all are possible)
    return []kripke.Step{
        func(w *kripke.World) {
            // Actions:
            // 1. Variable edits: a.x++
            // 2. Send message: SendMessage(...)
            // 3. Recv message: RecvAndLog(...)
        },
    }
}
```

**All checks pass** → Actor is Ready  
**Any check fails** → Actor is NOT Ready  
**Scheduler** → Picks uniformly at random from ready actors
