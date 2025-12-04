# kripke-ctl

Minimal Go playground for:

- An executable **Kripke engine** (`World`, `Process`, `Channel`, `Events`)
- A small **CTL evaluator** (`EF`, `EG`, `AF`, `AG`, `EX`, `AX`, `EU`)
- A demo that runs a simple process and a tiny CTL check

## Layout

- `kripke/engine.go` – World + scheduler + channels + events
- `kripke/ctl.go` – CTL AST + `sat(g)` evaluator
- `cmd/demo/main.go` – demo Process + CTL example

## Usage

```bash
git clone ...
cd kripke-ctl
go mod tidy
go run ./cmd/demo
```


<!-- REFACTOR_REQUEST:
- Tighten the definition of “admissible implication”.
- Add a short “Why CTL, not LTL” section.
- Add a concrete example with AG(p -> AF(q)).
-->

