# kripke-ctl

A minimal Go playground for:

- An executable **Kripke engine** (`World`, `Process`, `Channel`, `Message`, `Event`)
- A small **CTL evaluator** (`EF`, `EG`, `AF`, `AG`, `EX`, `AX`, `EU`)
- Demos that run a simple actor system and a tiny business-style CTL check

## Layout

- `kripke/engine.go` – World + scheduler + channels + events
- `kripke/ctl.go` – CTL AST + `Sat(g)` evaluator
- `kripke/order_model.go` – abstract order lifecycle as a Kripke graph
- `kripke/ctl_test.go` – basic unit tests for CTL operators
- `cmd/demo/main.go` – Producer/Consumer demo with queue-delay metrics
- `cmd/orders/main.go` – order-processing CTL example

## Usage

```bash
git clone https://github.com/rfielding/kripke-ctl
cd kripke-ctl
go mod tidy

# Run the producer/consumer engine demo
go run ./cmd/demo

# Run the order CTL demo
go run ./cmd/orders

# Run CTL unit tests
go test ./kripke
