# kripke-ctl
*A tiny communicating-MDP engine with CTL model checking*

`kripke-ctl` is a small but expressive playground for:

- An executable **Kripke-style state machine engine** built out of **actors**, **channels**, and **discrete steps**.
- A compact **CTL (Computation Tree Logic) evaluator** implementing `EF`, `EG`, `AF`, `AG`, `EX`, `AX`, and `EU`.
- A set of small **demos** illustrating:
  - Producer/consumer systems with queue-delay metrics.
  - A tiny business process (“order lifecycle”) verified using CTL.

The goal is to provide a minimal kernel for experimenting with **communicating MDPs**, **temporal logic**, and **actor-based operational semantics**, without committing to a large framework.

---

## Features

### Actors with local state

Each `Process` implements:

```go
ID() string
Ready(w *World) []Step
```

Where `Ready` returns enabled transitions, each a deterministic `Step(*World)`.

### First-class channels

Channels have:

- An owning actor
- A name (addressable)
- A finite capacity (buffered semantics; `cap >= 1`)
- FIFO behavior
- Queue-delay tracking

### Global scheduler

Each logical tick:

- Collect all enabled steps.
- Choose exactly *one* step uniformly at random.
- Execute it atomically.
- Advance logical time.

This produces a **communicating MDP**: deterministic guards → probabilistic step selection.

### Event log for diagrams and metrics

Every received message generates an `Event` containing:

- Timestamps
- Queue delay
- Payload
- Correlation ID
- From/To addresses

The event log can be used to build:

- Sequence diagrams
- Timeline visualizations
- Queueing metrics (latency distributions, throughput)

### CTL evaluator

Evaluate CTL formulas over arbitrary Kripke graphs:

- Path quantifiers: `A` (all futures), `E` (some future)
- Temporal operators:
  - Next: `EX`, `AX`
  - Eventually: `EF`, `AF`
  - Always: `EG`, `AG`
  - Until: `EU`

This enables reasoning over both hand-written and engine-produced state graphs.

---

## Project Layout

```text
kripke-ctl/
├── kripke/
│   ├── engine.go        # Actor engine: World, Scheduler, Channels, Events
│   ├── ctl.go           # CTL AST + evaluator        (existing in your repo)
│   ├── order_model.go   # Order lifecycle model      (existing in your repo)
│   ├── ctl_test.go      # CTL operator tests         (existing in your repo)
│   └── ...
│
├── cmd/
│   ├── demo/            # Producer/Consumer simulation with queue metrics
│   ├── orders/          # Business-style CTL example
│   └── docs/            # Docs and chat server
│
├── docs/                # Markdown sources rendered by the docs server
└── web/                 # Static HTML/JS shell that renders docs and diagrams
```

The new pieces in this snapshot are:

- `cmd/docs/main.go` – a tiny web server that:
  - Serves the documentation tree.
  - Exposes `/api/chat` as an HTTP proxy to OpenAI.
- `web/index.html` – a single-page shell that:
  - Renders Markdown docs via a client-side renderer.
  - Calls `/api/chat` to turn English into Mermaid diagrams.
- `docs/*.md` – initial stubs for a “book-style” documentation set.

---

## Usage

Clone and build:

```bash
git clone https://github.com/rfielding/kripke-ctl
cd kripke-ctl
go mod tidy
```

Run the actor engine demo:

```bash
go run ./cmd/demo
```

Run the order-processing CTL example:

```bash
go run ./cmd/orders
```

Run unit tests:

```bash
go test ./kripke
```

Run the docs + chat server (requires `OPENAI_API_KEY`):

```bash
export OPENAI_API_KEY=sk-********************************
go run ./cmd/docs
```

It will:

- Serve `web/index.html` and associated assets at `/`.
- Serve the `docs/` Markdown files under `/docs/`.
- Provide a `/api/chat` endpoint that proxies to OpenAI and returns JSON
  containing Mermaid diagrams.

Then open:

```text
http://localhost:8080/
```

in a browser.

---

## Philosophy

`kripke-ctl` is a formal methods system that is meant to be usable by normal people doing real business modeling work, while still being capable of computer science protocol analysis and proofs.

The design goals are:

- **Clarity over ceremony** – precise semantics, minimal notation.
- **Executable semantics** – state transitions are concrete code, not just math on a whiteboard.
- **Deterministic guards, probabilistic scheduling** – a clean MDP-style model.
- **Measurement built in** – queue delay and latency are first-class signals.
- **Formal methods without pointless obscurity** – CTL and state-space reasoning made explicit but approachable.
