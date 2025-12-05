# Kripke Documentation Set

This document contains the **full generated documentation** for the Kripke engine, formatted as a *book-style manual*. It includes all chapters for all three audiences:

* **Application Authors** – people writing models in English and Go
* **Operators / Deployers** – people running Kripke servers and CI systems
* **Business Stakeholders** – licensing, compliance, and value concerns

Each chapter begins with a header indicating its intended virtual filename inside `docs/`.

---

# docs/index.md

# Kripke Engine — Complete Manual

Kripke is an **executable specification engine** for building, simulating, and verifying communicating actor systems using:

* Go code generated from English
* Message-passing with blocking channels
* Probabilistic transitions (MDP semantics)
* Temporal logic verification (CTL)
* Deterministic simulation with event logs
* Metrics, sequence diagrams, and state diagrams

Kripke’s purpose is to let teams:

* Model hardware/software/business processes
* Verify safety and liveness properties
* Generate diagrams and dashboards automatically
* Use English → Go translation via an LLM

This manual is structured into three sections:

1. **Application Authors** – How to build Kripke models
2. **Operators / Deployers** – How to run Kripke as a service
3. **Business Stakeholders** – Positioning, licensing, compliance

---

# docs/getting-started/overview.md

# Getting Started with Kripke

Kripke provides an *actor‑based, message‑passing* model similar to:

* TLA+ global state semantics
* Go channels
* CSP-style blocking
* Markov decision processes (probabilistic branching)

You write:

* An **English description** of the system
* Kripke (the LLM agent) produces **Go actor structs**
* You run the simulation in a reproducible world

Key ideas:

* Every actor implements:

  ```go
  ID() string
  Ready(w *World) []Step
  ```
* Every channel is **owned** by an actor.
* Every message has:

  ```go
  Message{ID, CorrelationID, From, To, Payload, ReplyTo}
  ```
* The world executes **exactly one enabled step** per tick.
* Blocking is implemented by omitting steps whose channel ops cannot proceed.

This offers complete control over:

* nondeterminism
* concurrency
* visualization
* temporal logic properties

---

# docs/getting-started/installation.md

# Installing Kripke

Requirements:

* Go 1.22+
* Git

```bash
git clone https://github.com/rfielding/kripke-ctl.git
cd kripke-ctl
go mod tidy
```

Run the demo:

```bash
go run ./cmd/demo
```

You should see:

* Producer/Consumer example
* Queueing delays
* Event log
* Basic CTL checks

---

# docs/getting-started/first-model.md

# Writing Your First Model

### 1. Describe the system in English

Example:

> A Factory produces 10 widgets at a time. When inventory ≥ 10 and the truck can accept work, it sends a ShipmentOffer.

### 2. LLM generates Go code

You get a struct:

```go
type Factory struct { ... }
```

And its `Ready()` method.

### 3. Run the world

```go
w := NewWorld(...)
w.Run(200)
```

### 4. Inspect results

* `w.Events` contains all receive events (metrics)
* Sequence diagrams and charts come from post‑processing

---

# docs/getting-started/simulation-basics.md

# How the Simulation Works

### The Scheduler

For each tick:

1. Gather all actors with **at least one enabled step**
2. Choose one uniformly at random
3. Execute exactly one step

### Blocking

* Sending is enabled only if `CanSend()` is true
* Receiving only if `CanRecv()` is true
* Rendezvous (`cap=0`) requires both sides to be ready

### Messages

Messages are enqueued into the *receiver’s* channel, not the sender’s.

### Event Logging

When a message is **received**, Kripke logs an event with:

* MsgID
* CorrelationID
* EnqueueTime
* QueueDelay
* Payload

---

# docs/engine/world.md

# World Semantics

### Structure

```go
type World struct {
    Time      int
    Procs     []Process
    Channels  map[string]*Channel
    Events    []Event
}
```

### Determinism

Kripke is deterministic given a seed; nondeterminism comes only from:

* Actor nondeterministic guards
* RNG inside steps

### Running Steps

```go
func (w *World) RunOneStep() bool
func (w *World) Run(max int)
```

---

# docs/engine/processes.md

# Writing Processes

Each process must implement:

```go
ID() string
Ready(w *World) []Step
```

A `Step` is:

```go
type Step func(w *World)
```

### Guidelines

* Keep steps atomic
* Do not mutate other actors except via messaging
* Use internal predicates for conceptual states (Idle, Busy, Broken…)

---

# docs/engine/channels.md

# Channels

### Definition

```go
type Channel struct {
    OwnerID string
    Name    string
    cap     int
    buf     []Message
}
```

### Blocking Rules

* If full → send disabled
* If empty → receive disabled
* If cap=0 → rendezvous

### Addressing

Each channel is uniquely named by:

```go
Address{ActorID, ChannelName}
```

---

# docs/engine/messages.md

# Messages

```go
type Message struct {
    ID, CorrelationID uint64
    From, To          Address
    Payload           any
    ReplyTo           *Address
    EnqueueTime       int
}
```

### Recommendations

* Use `Payload` as a tagged object: `{kind: "Order", qty: 3}`
* Use `CorrelationID` for request/response matching
* Use `ReplyTo` for asynchronous callbacks

---

# docs/engine/events.md

# Events and Metrics

Kripke logs events **only on message receive**.

```go
type Event struct {
    Time          int
    MsgID         uint64
    CorrelationID uint64
    From, To      Address
    Payload       any
    ReplyTo       *Address
    EnqueueTime   int
    QueueDelay    int
}
```

### Common Metrics

* Throughput = number of deliveries per time window
* Latency = QueueDelay distribution
* Output = sum of product quantities
* Revenue = derived from sale payloads

---

# docs/engine/scheduler.md

# Scheduler and Readiness

### READY Definition

An actor is READY if **any** step in `Ready()` is enabled *and does not violate channel blocking rules*.

### Selecting One Step

If multiple actors and steps are ready:

* Choose actor uniformly at random
* Choose one of its steps uniformly

### Why Only One Step?

This produces:

* A tree of possible executions
* A well‑defined transition relation for CTL

---

# docs/engine/temporal-logic.md

# CTL and Temporal Logic

Kripke supports CTL operators (in a separate package):

* **AG**, **AF**, **EG**, **EF**, **AX**, **EX**, **EU**

Examples:

* *Safety*: `AG(not(Error))`
* *Liveness*: `AF(Delivered)`
* *Possibility*: `EF(Completed && Balance>=0)`

Application authors typically express:

```go
eg.Sat(g)
```

Or ask the LLM:

> Prove that every order is eventually delivered or cancelled.

---

# docs/domain-modeling/english-to-go.md

# English → Go Translation

The LLM acts as a structured compiler:

1. Parse English descriptions
2. Produce actor structs
3. Produce channels and wiring
4. Assign payload structures
5. Emit diagrams after simulation

### Example Input

> Customers send orders to storefronts; storefronts fulfill them if inventory permits.

### Example Output

* `customer.go`
* `storefront.go`
* Channel wiring in `main.go`
* Metrics and diagrams

---

# docs/domain-modeling/actor-design.md

# Designing Actors

Guidelines:

* Make predicates explicit (e.g. `IsIdle()`, `IsWaiting()`)
* Include internal counters
* Separate message I/O from internal processing
* Avoid mutation of other actors except by messaging

---

# docs/domain-modeling/message-design.md

# Message Schema Design

Use structured payloads:

```go
type OrderPayload struct {
    Product string
    Qty     int
}
```

Use `kind` tags:

```go
{"kind":"Order","product":"Widget","qty":3}
```

---

# docs/domain-modeling/metrics-and-diagrams.md

# Metrics, Sequence Diagrams, and State Diagrams

### Sequence Diagrams

Derived from chronological `Event` entries.

### State Diagrams

Derived from internal predicate transitions.

### Charts

* Throughput
* Revenue
* Latency distributions

Used via chart tools in the Kripke agent.

---

# docs/deployment/running-server.md

# Running Kripke as a Server

### Capabilities

* Accept models via API
* Run simulations
* Return CTL results
* Generate diagrams

### API Endpoints (example)

* `POST /model` – upload actors
* `POST /simulate` – run world
* `GET /events` – retrieve event stream

---

# docs/deployment/CI-integration.md

# CI Integration

Use CI to:

* Verify that safety properties still hold
* Generate nightly simulations
* Export diagrams for documentation

Example GitHub Actions:

```yaml
- name: Run simulation
  run: go run ./cmd/demo
```

---

# docs/deployment/scaling-models.md

# Scaling Simulations

Techniques:

* Multiple worlds in parallel
* Seeded RNG for reproducibility
* Distributed job runners

---

# docs/business/value-proposition.md

# Value Proposition

Kripke:

* Reduces systemic risk
* Catches deadlocks and starvation early
* Allows quick prototyping of concurrency behaviors
* Provides formal guarantees (via CTL)

Industries:

* Logistics
* Finance
* Hardware design
* Safety-critical automation

---

# docs/business/security-and-compliance.md

# Security and Compliance

Kripke offers:

* Deterministic replay
* Sandboxed execution
* Fully auditable event logs
* Static code generation from human-readable specs

---

# docs/business/licensing.md

# Licensing

Possible models:

* Open core
* Commercial license
* Support subscriptions
* Audit/compliance packages

---

# docs/business/integration-case-studies.md

# Integration Case Studies

### Logistics Example

A warehouse → truck → storefront model verifies:

* Orders are eventually served
* No infinite backlog occurs
* Revenue correlates with throughput

### Financial Example

Settlement model verifies:

* No double-spend paths exist
* AF(Balanced) holds

---

# End of Full Documentation
