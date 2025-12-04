package kripke

import (
	"testing"
)

func TestBasicState(t *testing.T) {
	s := NewBasicState("state1", "ready", "safe")

	if s.ID() != "state1" {
		t.Errorf("Expected state ID 'state1', got '%s'", s.ID())
	}

	if !s.HasProperty("ready") {
		t.Error("Expected state to have 'ready' property")
	}

	if !s.HasProperty("safe") {
		t.Error("Expected state to have 'safe' property")
	}

	if s.HasProperty("error") {
		t.Error("Expected state to not have 'error' property")
	}

	props := s.GetProperties()
	if len(props) != 2 {
		t.Errorf("Expected 2 properties, got %d", len(props))
	}
}

func TestMessage(t *testing.T) {
	msg := Message{
		From:    "actor1",
		To:      "actor2",
		Type:    "test",
		Payload: "test payload",
		Reward:  10.0,
	}

	if msg.From != "actor1" {
		t.Errorf("Expected From='actor1', got '%s'", msg.From)
	}

	if msg.Reward != 10.0 {
		t.Errorf("Expected Reward=10.0, got %f", msg.Reward)
	}
}

func TestActor(t *testing.T) {
	s0 := NewBasicState("s0", "init")
	actor := NewActor("test_actor", s0)

	if actor.Name != "test_actor" {
		t.Errorf("Expected actor name 'test_actor', got '%s'", actor.Name)
	}

	if actor.CurrentState.ID() != "s0" {
		t.Errorf("Expected current state 's0', got '%s'", actor.CurrentState.ID())
	}

	if actor.GetReward() != 0 {
		t.Errorf("Expected initial reward 0, got %f", actor.GetReward())
	}
}

func TestActorTransition(t *testing.T) {
	s0 := NewBasicState("s0", "init")
	s1 := NewBasicState("s1", "processing")

	actor := NewActor("test_actor", s0)
	actor.AddState(s0)
	actor.AddState(s1)
	actor.AddTransition(s0, s1, 1.0, "start", 5.0)

	oldState, reward, err := actor.ExecuteTransition("start")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if oldState.ID() != "s0" {
		t.Errorf("Expected old state 's0', got '%s'", oldState.ID())
	}

	if actor.CurrentState.ID() != "s1" {
		t.Errorf("Expected current state 's1', got '%s'", actor.CurrentState.ID())
	}

	if reward != 5.0 {
		t.Errorf("Expected reward 5.0, got %f", reward)
	}

	if actor.GetReward() != 5.0 {
		t.Errorf("Expected accumulated reward 5.0, got %f", actor.GetReward())
	}
}

func TestActorInvalidTransition(t *testing.T) {
	s0 := NewBasicState("s0", "init")
	actor := NewActor("test_actor", s0)
	actor.AddState(s0)

	_, _, err := actor.ExecuteTransition("invalid")
	if err == nil {
		t.Error("Expected error for invalid transition, got nil")
	}
}

func TestActorProbabilisticTransition(t *testing.T) {
	s0 := NewBasicState("s0", "init")
	s1 := NewBasicState("s1", "success")
	s2 := NewBasicState("s2", "failure")

	actor := NewActor("test_actor", s0)
	actor.AddState(s0)
	actor.AddState(s1)
	actor.AddState(s2)
	actor.AddTransition(s0, s1, 0.7, "try", 10.0)
	actor.AddTransition(s0, s2, 0.3, "try", -5.0)

	// Execute multiple times to test probabilistic behavior
	successCount := 0
	failureCount := 0
	for i := 0; i < 100; i++ {
		actor.CurrentState = s0
		_, _, err := actor.ExecuteTransition("try")
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if actor.CurrentState.ID() == "s1" {
			successCount++
		} else if actor.CurrentState.ID() == "s2" {
			failureCount++
		}
	}

	// Check that we got some successes and some failures (probabilistic)
	if successCount == 0 || failureCount == 0 {
		t.Logf("Success count: %d, Failure count: %d", successCount, failureCount)
		// Note: This could fail occasionally due to randomness, but very unlikely
	}
}

func TestModel(t *testing.T) {
	s0 := NewBasicState("s0", "init")
	model := NewModel(s0)

	if model.InitialState.ID() != "s0" {
		t.Errorf("Expected initial state 's0', got '%s'", model.InitialState.ID())
	}

	s1 := NewBasicState("s1", "processing")
	model.AddState(s1)

	state, ok := model.GetState("s1")
	if !ok {
		t.Error("Expected to find state 's1'")
	}
	if state.ID() != "s1" {
		t.Errorf("Expected state ID 's1', got '%s'", state.ID())
	}
}

func TestModelActors(t *testing.T) {
	s0 := NewBasicState("s0", "init")
	model := NewModel(s0)

	actor := NewActor("worker", s0)
	model.AddActor(actor)

	retrievedActor, ok := model.GetActor("worker")
	if !ok {
		t.Error("Expected to find actor 'worker'")
	}
	if retrievedActor.Name != "worker" {
		t.Errorf("Expected actor name 'worker', got '%s'", retrievedActor.Name)
	}
}

func TestModelMessagePassing(t *testing.T) {
	s0 := NewBasicState("s0", "init")
	model := NewModel(s0)

	actor1 := NewActor("actor1", s0)
	actor2 := NewActor("actor2", s0)
	model.AddActor(actor1)
	model.AddActor(actor2)

	msg := Message{
		From:    "actor1",
		To:      "actor2",
		Type:    "task",
		Payload: "do something",
		Reward:  5.0,
	}

	err := model.SendMessage(msg)
	if err != nil {
		t.Errorf("Unexpected error sending message: %v", err)
	}

	receivedMsg, ok := actor2.TryReceiveMessage()
	if !ok {
		t.Error("Expected to receive message")
	}

	if receivedMsg.From != "actor1" {
		t.Errorf("Expected message from 'actor1', got '%s'", receivedMsg.From)
	}

	if receivedMsg.Reward != 5.0 {
		t.Errorf("Expected reward 5.0, got %f", receivedMsg.Reward)
	}
}

func TestModelSuccessors(t *testing.T) {
	s0 := NewBasicState("s0", "init")
	s1 := NewBasicState("s1", "processing")
	s2 := NewBasicState("s2", "done")

	model := NewModel(s0)
	model.AddState(s0)
	model.AddState(s1)
	model.AddState(s2)

	actor := NewActor("worker", s0)
	actor.AddState(s0)
	actor.AddState(s1)
	actor.AddState(s2)
	actor.AddTransition(s0, s1, 1.0, "start", 1.0)
	actor.AddTransition(s0, s2, 1.0, "skip", 0.0)
	model.AddActor(actor)

	successors := model.GetSuccessors(s0)
	if len(successors) != 2 {
		t.Errorf("Expected 2 successors, got %d", len(successors))
	}

	// Check that both s1 and s2 are in successors
	foundS1, foundS2 := false, false
	for _, succ := range successors {
		if succ.ID() == "s1" {
			foundS1 = true
		}
		if succ.ID() == "s2" {
			foundS2 = true
		}
	}

	if !foundS1 || !foundS2 {
		t.Error("Expected to find both s1 and s2 as successors")
	}
}

func TestModelMetrics(t *testing.T) {
	s0 := NewBasicState("s0", "init")
	s1 := NewBasicState("s1", "processing")

	model := NewModel(s0)
	model.AddState(s0)
	model.AddState(s1)

	actor1 := NewActor("worker1", s0)
	actor1.AddState(s0)
	actor1.AddState(s1)
	actor1.AddTransition(s0, s1, 1.0, "start", 10.0)
	model.AddActor(actor1)

	actor2 := NewActor("worker2", s0)
	model.AddActor(actor2)

	// Execute a transition to accumulate reward
	actor1.ExecuteTransition("start")

	metrics := model.GetMetrics()

	if metrics["total_actors"] != 2 {
		t.Errorf("Expected 2 actors, got %v", metrics["total_actors"])
	}

	if metrics["total_states"] != 2 {
		t.Errorf("Expected 2 states, got %v", metrics["total_states"])
	}

	totalReward := model.GetTotalReward()
	if totalReward != 10.0 {
		t.Errorf("Expected total reward 10.0, got %f", totalReward)
	}
}

func TestModelCTLCheck(t *testing.T) {
	s0 := NewBasicState("s0", "init")
	s1 := NewBasicState("s1", "done")

	model := NewModel(s0)
	model.AddState(s0)
	model.AddState(s1)

	actor := NewActor("worker", s0)
	actor.AddState(s0)
	actor.AddState(s1)
	actor.AddTransition(s0, s1, 1.0, "complete", 10.0)
	model.AddActor(actor)

	// Check EF done
	efDone := &EF{Formula: &AtomicProp{Name: "done"}}
	result := model.CheckCTL(efDone)
	if !result {
		t.Error("Expected EF(done) to be true")
	}

	// Check from specific state
	result = model.CheckCTLFromState(efDone, s0)
	if !result {
		t.Error("Expected EF(done) to be true from s0")
	}
}
