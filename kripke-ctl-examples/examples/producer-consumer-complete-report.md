# Producer-Consumer: Complete Analysis Report

**Generated**: 2025-12-10 05:45:02

**Tool**: kripke-ctl (CTL model checker + actor engine)

---

## 1. System Description

### English Specification

A producer-consumer system with bounded buffer:

- **Producer**: Creates items and sends them to consumer's inbox
- **Consumer**: Receives items from inbox and processes them
- **Buffer**: FIFO channel with capacity = 2
- **Blocking**: Producer waits when full, consumer waits when empty

## 2. Engine Execution

Running actor engine for 20 steps...

| Step | Time | Action | Buffer Size | Events |
|------|------|--------|-------------|--------|
| 1 | 1 | produce | 1 | 0 |
| 2 | 2 | consume | 0 | 1 |
| 3 | 3 | produce | 1 | 1 |
| 4 | 4 | produce | 2 | 1 |
| 5 | 5 | consume | 1 | 2 |
| 6 | 6 | consume | 0 | 3 |
| 7 | 7 | produce | 1 | 3 |
| 8 | 8 | produce | 2 | 3 |
| 9 | 9 | consume | 1 | 4 |
| 10 | 10 | consume | 0 | 5 |
| 11 | 11 | produce | 1 | 5 |
| 12 | 12 | consume | 0 | 6 |
| 13 | 13 | produce | 1 | 6 |
| 14 | 14 | produce | 2 | 6 |
| 15 | 15 | consume | 1 | 7 |
| 16 | 16 | consume | 0 | 8 |
| 17 | 17 | produce | 1 | 8 |
| 18 | 18 | consume | 0 | 9 |
| 19 | 19 | produce | 1 | 9 |
| 20 | 20 | produce | 2 | 9 |

## 3. Event Log Analysis

**Total messages**: 9
**Average queue delay**: 1.67 ticks
**Maximum queue delay**: 2 ticks
**Consumer processed**: 9 items

### Event Timeline

```mermaid
sequenceDiagram
    participant P as Producer
    participant B as Buffer
    participant C as Consumer
    P->>B: send (t=0)
    B->>C: recv (t=1, delay=1)
    P->>B: send (t=2)
    B->>C: recv (t=4, delay=2)
    P->>B: send (t=3)
    B->>C: recv (t=5, delay=2)
    P->>B: send (t=6)
    B->>C: recv (t=8, delay=2)
    P->>B: send (t=7)
    B->>C: recv (t=9, delay=2)
    P->>B: send (t=10)
    B->>C: recv (t=11, delay=1)
    P->>B: send (t=12)
    B->>C: recv (t=14, delay=2)
    P->>B: send (t=13)
    B->>C: recv (t=15, delay=2)
    P->>B: send (t=16)
    B->>C: recv (t=17, delay=1)
```

## 4. State Space Model

**States**: 3
**Transitions**: 4

### States

- **buffer_0**: [empty, P:ready]
- **buffer_1**: [P:ready, C:ready]
- **buffer_2**: [full, C:ready]

### State Diagram

```mermaid
stateDiagram-v2
    [*] --> buffer_0

    buffer_1 --> buffer_2: produce
    buffer_1 --> buffer_0: consume
    buffer_2 --> buffer_1: consume
    buffer_0 --> buffer_1: produce

    buffer_0: Empty (P:ready C:blocked)
    buffer_1: Partial (P:ready C:ready)
    buffer_2: Full (P:blocked C:ready)
```

## 5. CTL Property Verification

| Property | Formula | Result | Description |
|----------|---------|--------|-------------|
| Safety | `AG(¬overflow)` | ✅ | Buffer never overflows |
| Liveness-P | `AG(EF(producer_ready))` | ✅ | Producer can always eventually send |
| Liveness-C | `AG(EF(consumer_ready))` | ✅ | Consumer can always eventually receive |
| No-Deadlock | `AG(P ∨ C)` | ✅ | At least one actor can always progress |
| Reachable-Full | `EF(full)` | ✅ | Buffer can become full |
| Reachable-Empty | `EF(empty)` | ✅ | Buffer can become empty |

## 6. Conclusion

**Verification Summary**: 6/6 properties verified

### Key Findings

1. ✅ **Safety**: The buffer never overflows (capacity constraint respected)
2. ✅ **Liveness**: Both producer and consumer can always make progress eventually
3. ✅ **Deadlock-freedom**: The system never reaches a state where no progress is possible
4. ✅ **Reachability**: All possible buffer states (empty, partial, full) are reachable

### Performance Metrics

- Average queue delay: **1.67 ticks**
- Messages processed: **9**
- Execution time: **142.225µs**

