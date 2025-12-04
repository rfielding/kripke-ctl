# kripke-ctl

English to Requirements OpenAI service, featuring Temporal Logic Engine, graphing tools.

## Overview

`kripke-ctl` is a formal verification service that aims to surpass tools like TLA+, Prism, and Circal by using a flexible and powerful formalism: **Markov Decision Processes (MDPs) that communicate via messages**.

## Formalism: MDPs with Message-Based Communication

This service uses Markov Decision Processes as the foundation for modeling state machines because:

1. **Flexibility**: MDPs allow modeling of systems where state changes are not purely random but can be internally triggered
2. **Not Pure Markov Chains**: Actors in the system cannot be modeled as simple Markov Chains because not all state transitions are probabilistic
3. **Message-Based Communication**: Actors communicate via messages, which are equivalent to MDP rewards/awards
4. **Free Metrics**: The MDP formalism naturally yields useful metrics for system analysis

## Features

- **CTL Model Checking**: Computational Tree Logic operators for verifying temporal properties
- **Flexible State Machines**: Define complex state machines with both probabilistic and deterministic transitions
- **Message Passing**: Actors communicate through messages that influence state transitions
- **Automatic Metrics**: Get useful system metrics automatically from the MDP formalism

## Components

- **kripke/ctl.go**: Implementation of CTL (Computational Tree Logic) operators and model checking
- **kripke/engine.go**: MDP engine for executing state machines with message-based communication

## Why MDPs?

Traditional Markov Chains assume all transitions are probabilistic, but real systems often have:
- Deterministic internal triggers
- External events
- Message-based coordination between components

MDPs provide the flexibility to model both probabilistic and non-deterministic behaviors while maintaining the mathematical rigor needed for formal verification.

## Usage Example

```go
package main

import (
    "fmt"
    "github.com/rfielding/kripke-ctl/kripke"
)

func main() {
    // Create states
    s0 := kripke.NewBasicState("s0", "init")
    s1 := kripke.NewBasicState("s1", "processing")
    s2 := kripke.NewBasicState("s2", "done")

    // Create a model with initial state
    model := kripke.NewModel(s0)
    model.AddState(s0)
    model.AddState(s1)
    model.AddState(s2)

    // Create an actor (MDP agent)
    actor := kripke.NewActor("worker", s0)
    actor.AddState(s0)
    actor.AddState(s1)
    actor.AddState(s2)

    // Add transitions with probabilities and rewards
    actor.AddTransition(s0, s1, 1.0, "start", 1.0)
    actor.AddTransition(s1, s2, 0.8, "complete", 10.0)

    model.AddActor(actor)

    // Check CTL properties
    efDone := &kripke.EF{Formula: &kripke.AtomicProp{Name: "done"}}
    result := model.CheckCTL(efDone)
    fmt.Printf("EF(done): %v\n", result) // true - eventually reaches "done"

    // Execute state transitions
    actor.ExecuteTransition("start")
    actor.ExecuteTransition("complete")

    // Get accumulated metrics
    fmt.Printf("Total reward: %.2f\n", model.GetTotalReward())
}
```

See the `example/` directory for a complete working example demonstrating:
- State machine creation with probabilistic transitions
- CTL model checking
- Message passing between actors
- Automatic metric collection
