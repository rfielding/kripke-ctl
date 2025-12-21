# BoundedLISP

A tool for specifying and verifying multi-party protocols through conversation.

## What It Does

You describe a distributed system or protocol in plain English. The tool helps you:

1. **Sketch ideas** on a shared whiteboard (LaTeX, state machines, message flows)
2. **Formalize** those sketches into executable actor specifications
3. **Verify** properties using CTL model checking
4. **Visualize** with automatically generated diagrams

Think of it as pair-programming for protocol design. You and the AI are in a conference room with a whiteboard, working toward a formal specification.

## Quick Start

```bash
# Run tests
cat prologue.lisp tests.lisp | go run main.go -repl

# Start interactive server
export ANTHROPIC_API_KEY=sk-ant-...
go run main.go

# Open http://localhost:8080 in browser
```

## Usage Modes

### Web UI (default)
```bash
go run main.go
```
Opens a web interface with:
- **Chat panel** - describe your system, ask questions
- **Whiteboard** - sketch ideas before formalizing (LaTeX, diagrams)
- **Specification** - rendered markdown with diagrams
- **LISP** - the executable code

### Console + Server
When the server is running, you can also type in the terminal. Useful for quick queries without switching to the browser.

### REPL Only
```bash
go run main.go -repl
```
Pure LISP interpreter, no server.

### File Execution
```bash
go run main.go myspec.lisp
```

## The Actor Model

Actors are the source of truth. Each actor:
- Has a mailbox (bounded queue)
- Processes messages sequentially
- Uses `become` to carry state forward

```lisp
(define (server-loop request-count)
  (let msg (receive!)
    (send-to! (first msg) 'ack)
    (list 'become (list 'server-loop (+ request-count 1)))))

(spawn-actor 'server 16 '(server-loop 0))
```

## Key Primitives

| Function | Purpose |
|----------|---------|
| `spawn-actor` | Create an actor with mailbox |
| `send-to!` | Send message to actor |
| `receive!` | Block until message arrives |
| `(list 'become ...)` | Continue with new state |
| `'done` | Actor terminates |

## Properties (CTL)

Specify what must be true:

```lisp
; Every request eventually gets a response
(defproperty 'responsive
  (AG (ctl-implies (prop 'request) (AF (prop 'response)))))

; No deadlocks - always can make progress
(defproperty 'no-deadlock
  (AG (EX (prop 'true))))
```

## Probability Distributions

For modeling stochastic systems:

```lisp
(exponential u rate)        ; Inter-arrival times
(normal u1 u2 mean stddev)  ; Gaussian
(bernoulli u p)             ; Success/failure
(discrete-uniform u min max) ; Random integer
```

## Workflow

1. **Sketch** - Draw rough ideas on the whiteboard
   - "Client sends request, server responds or times out"
   - State machine sketches
   - Message sequence ideas

2. **Discuss** - Refine with the AI
   - "What if the server crashes mid-request?"
   - "Add retry logic"

3. **Formalize** - Convert to LISP
   - AI generates actor code
   - Properties are defined

4. **Verify** - Check properties
   - Model checking runs
   - Counterexamples shown if failed

5. **Iterate** - Refine and repeat

## Files

| File | Purpose |
|------|---------|
| `main.go` | Interpreter, scheduler, web server |
| `prologue.lisp` | Runtime library (actors, CTL, distributions) |
| `tests.lisp` | 155 unit tests |

## Environment Variables

| Variable | Purpose |
|----------|---------|
| `ANTHROPIC_API_KEY` | Claude API key |
| `OPENAI_API_KEY` | GPT-4 API key (alternative) |
| `KRIPKE_PORT` | Server port (default: 8080) |

## Running Tests

```bash
make test
# or
cat prologue.lisp tests.lisp | go run main.go -repl
```

## License

MIT
