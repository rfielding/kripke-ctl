# kripke-ctl

A Computational Tree Logic (CTL) model checker with OpenAI integration for converting English descriptions into formal models and temporal logic formulas.

## Features

- **CTL Model Checking Engine**: Complete implementation of CTL model checking algorithm
- **Kripke Structure Modeling**: Define and manipulate labeled transition systems
- **OpenAI Integration**: Convert natural language descriptions into formal models
- **Interactive CLI**: User-friendly command-line interface
- **Pre-built Examples**: Traffic lights, mutual exclusion, and simple systems
- **Graphviz Export**: Visualize Kripke structures (DOT format)

## CTL Operators Supported

### Temporal Operators
- `EX φ` - Exists neXt: There exists a next state where φ holds
- `AX φ` - All neXt: In all next states, φ holds
- `EF φ` - Exists Finally: There exists a path where eventually φ holds
- `AF φ` - All Finally: On all paths, eventually φ holds
- `EG φ` - Exists Globally: There exists a path where always φ holds
- `AG φ` - All Globally: On all paths, always φ holds
- `E[φ U ψ]` - Exists Until: There exists a path where φ holds until ψ holds
- `A[φ U ψ]` - All Until: On all paths, φ holds until ψ holds

### Boolean Operators
- `∧` - Conjunction (AND)
- `∨` - Disjunction (OR)
- `¬` - Negation (NOT)
- `→` - Implication

## Installation

```bash
git clone https://github.com/rfielding/kripke-ctl
cd kripke-ctl
go build -o kripke-ctl
```

## Usage

### Basic Usage

```bash
./kripke-ctl
```

This launches an interactive CLI where you can:
1. Explore predefined examples (Traffic Light, Mutual Exclusion, Simple)
2. Check CTL formulas on models
3. Use OpenAI to generate models from English descriptions (requires API key)

### With OpenAI Integration

Set your OpenAI API key:

```bash
export OPENAI_API_KEY="your-api-key-here"
./kripke-ctl
```

Then you can:
- Generate Kripke structures from natural language descriptions
- Convert English queries into CTL formulas

## Examples

### Traffic Light System

A simple traffic light with three states:
- Red (stop)
- Green (go)
- Yellow (caution)

CTL properties:
- `EF go` - Eventually reach green light (true)
- `AF stop` - Always eventually return to red (true)
- `AG caution` - Always at yellow (false)

### Mutual Exclusion

Two processes competing for a critical section, demonstrating:
- Safety: `AG ¬(critical1 ∧ critical2)` - Never both in critical section
- Liveness: Processes trying eventually enter critical section

### Simple Example

A minimal three-state system for testing basic CTL operators.

## Architecture

### Core Components

1. **kripke.go**: Kripke structure data structures and operations
2. **ctl.go**: CTL formula AST definitions
3. **model_checker.go**: CTL model checking algorithm implementation
4. **openai.go**: OpenAI API integration for natural language processing
5. **examples.go**: Pre-built example systems
6. **graphviz.go**: Visualization export
7. **main.go**: Interactive CLI application

### Testing

Run the test suite:

```bash
go test -v
```

Tests cover:
- All CTL operators (EX, AX, EF, AF, EG, AG, EU, AU)
- Boolean operations (AND, OR, NOT, IMPLIES)
- Pre-built examples validation
- Model checking correctness

## CTL Model Checking Algorithm

The implementation uses fixed-point algorithms for temporal operators:

- **EF/AF**: Computed using EU/AU with trivial left operand
- **EG**: Fixed-point iteration removing states without successors
- **AG**: Dual of EF (¬EF ¬φ)
- **EU/AU**: Fixed-point iteration from goal states backwards

## OpenAI Integration

### Model Generation

Provide a natural language description, and the system:
1. Sends description to GPT-4
2. Receives JSON Kripke structure definition
3. Parses and creates the model
4. Enables model checking on generated structure

### Formula Generation

Describe a property in English:
- "eventually the light is green"
- "always both processes are not in critical section"

The system converts these to CTL formulas using GPT-4.

## Future Enhancements

- [ ] CTL* support (nested path quantifiers)
- [ ] LTL (Linear Temporal Logic) integration
- [ ] Formula parser for arbitrary CTL expressions
- [ ] Counterexample generation
- [ ] State space optimization for large models
- [ ] Web UI with interactive visualization
- [ ] Fairness constraints
- [ ] Bisimulation checking

## License

MIT License

## Contributing

Contributions welcome! This project demonstrates a Turing Award-level task: automated formal verification through natural language processing.

## References

- Clarke, E. M., Emerson, E. A., & Sistla, A. P. (1986). "Automatic verification of finite-state concurrent systems using temporal logic specifications"
- Model Checking (Clarke, Grumberg, Peled)
- Principles of Model Checking (Baier, Katoen)
