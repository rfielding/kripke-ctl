Scheduler rules
=============

interaction of communicating actors in bakery scenario

```mermaid
sequenceDiagram
    participant B as Bakery
    participant T as Truck
    participant S as Storefront
    participant C as Customer

    Note over B,T: Bakery writes into Truck.breadRack (owned by Truck)
    B->>T: B.c[breadRack] ! loaf
    T-->>T: B.c[breadRack] ? x / load loaf into cargo

    Note over T,S: Truck delivers loaves into Storefront.stock
    T->>S: S.c[stock] ! loaf
    S-->>S: S.c[stock] ? x / increase inventory

    Note over C,S: Customer buys bread from Storefront
    C->>S: order
    S->>C: receipt
```

flowchard of steps in scheduler

```mermaid
flowchart TD
    A[Current world state S]
    B[For each actor]
    C[Collect candidate steps]
    D[Filter: guards true and channels not blocked]
    E{Any enabled steps?}
    F[Quiescent or deadlock]
    G[Pick one enabled step at random]
    H[Execute chosen step and update S to S']

    A --> B --> C --> D --> E
    E -->|No| F
    E -->|Yes| G --> H --> A

```

FSM
====================

The state machines and interactions are related because reads and writes happen based on states we are in. Roughly each state is a machine instruction for a single varible change, send, recv, or write.

```mermaid
stateDiagram-v2
    [*] --> Init

    state "Init\n(init: x = 0)" as Init

    state NeedsAttention {
        [*] --> NotFull

        state "NotFull\n(needsAttention: x > 0 ∧ x < cap)" as NotFull
        state "IsFull\n(isFull: x >= cap)" as IsFull

        NotFull --> IsFull: enqueue\nx' = x + 1
        IsFull --> NotFull: dequeue\nx' = x - 1
    }

    Init --> NeedsAttention: first enqueue\nx' = x + 1
    NeedsAttention --> Init: drain to zero\nx' = 0
```
