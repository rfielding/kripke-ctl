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
	s3 := kripke.NewBasicState("s3", "error")

	// Create a model with initial state
	model := kripke.NewModel(s0)
	model.AddState(s0)
	model.AddState(s1)
	model.AddState(s2)
	model.AddState(s3)

	// Create an actor
	actor := kripke.NewActor("worker", s0)
	actor.AddState(s0)
	actor.AddState(s1)
	actor.AddState(s2)
	actor.AddState(s3)

	// Add transitions (with probabilities and rewards)
	actor.AddTransition(s0, s1, 1.0, "start", 1.0)
	actor.AddTransition(s1, s2, 0.8, "complete", 10.0)
	actor.AddTransition(s1, s3, 0.2, "complete", -5.0)
	actor.AddTransition(s3, s1, 1.0, "retry", 0.0)

	// Add actor to model
	model.AddActor(actor)

	fmt.Println("=== MDP-based State Machine Example ===")
	fmt.Println()

	// Check CTL properties
	fmt.Println("Checking CTL properties:")
	fmt.Println()

	// EF done: Eventually, the system reaches "done" state
	efDone := &kripke.EF{Formula: &kripke.AtomicProp{Name: "done"}}
	result := model.CheckCTL(efDone)
	fmt.Printf("EF(done): %v - There exists a path to eventually reach 'done'\n", result)

	// AF done: All paths eventually reach "done"
	afDone := &kripke.AF{Formula: &kripke.AtomicProp{Name: "done"}}
	result = model.CheckCTL(afDone)
	fmt.Printf("AF(done): %v - All paths eventually reach 'done'\n", result)

	// EX processing: There exists a next state that is "processing"
	exProcessing := &kripke.EX{Formula: &kripke.AtomicProp{Name: "processing"}}
	result = model.CheckCTL(exProcessing)
	fmt.Printf("EX(processing): %v - There exists a next state with 'processing'\n", result)

	// AG ¬error: It's always the case that error doesn't occur
	agNotError := &kripke.AG{Formula: &kripke.Not{Formula: &kripke.AtomicProp{Name: "error"}}}
	result = model.CheckCTL(agNotError)
	fmt.Printf("AG(¬error): %v - Error never occurs on all paths\n", result)

	fmt.Println()
	fmt.Println("=== Simulating State Transitions ===")
	fmt.Println()

	// Simulate state transitions
	fmt.Printf("Initial state: %s\n", actor.CurrentState.ID())

	// Execute transitions
	oldState, reward, err := actor.ExecuteTransition("start")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Transition: %s -> %s (reward: %.2f)\n", oldState.ID(), actor.CurrentState.ID(), reward)
	}

	// Try to complete (probabilistic)
	for i := 0; i < 3; i++ {
		oldState, reward, err := actor.ExecuteTransition("complete")
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			break
		}
		fmt.Printf("Transition: %s -> %s (reward: %.2f)\n", oldState.ID(), actor.CurrentState.ID(), reward)
		
		if actor.CurrentState.ID() == "s3" {
			// Retry from error state
			oldState, reward, err = actor.ExecuteTransition("retry")
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				break
			}
			fmt.Printf("Retry: %s -> %s (reward: %.2f)\n", oldState.ID(), actor.CurrentState.ID(), reward)
		}

		if actor.CurrentState.ID() == "s2" {
			break
		}
	}

	fmt.Println()
	fmt.Println("=== Message Passing Example ===")
	fmt.Println()

	// Create another actor for message passing
	coordinator := kripke.NewActor("coordinator", s0)
	coordinator.AddState(s0)
	coordinator.AddState(s1)
	model.AddActor(coordinator)

	// Send a message (equivalent to MDP reward)
	msg := kripke.Message{
		From:    "coordinator",
		To:      "worker",
		Type:    "task_assigned",
		Payload: "Process job #123",
		Reward:  5.0,
	}

	err = model.SendMessage(msg)
	if err != nil {
		fmt.Printf("Error sending message: %v\n", err)
	} else {
		fmt.Printf("Message sent: %s -> %s (type: %s, reward: %.2f)\n", 
			msg.From, msg.To, msg.Type, msg.Reward)
	}

	// Receive the message
	receivedMsg, ok := actor.TryReceiveMessage()
	if ok {
		fmt.Printf("Message received by %s: %s (reward: %.2f)\n", 
			actor.Name, receivedMsg.Payload, receivedMsg.Reward)
	}

	fmt.Println()
	fmt.Println("=== Model Metrics ===")
	fmt.Println()

	// Get model metrics (free metrics from MDP formalism)
	metrics := model.GetMetrics()
	fmt.Println("Model metrics:")
	for key, value := range metrics {
		fmt.Printf("  %s: %v\n", key, value)
	}

	fmt.Printf("\nTotal accumulated reward: %.2f\n", model.GetTotalReward())
}
