package kripke

import (
	"fmt"
	"math/rand"
	"sync"
)

// State represents a state in the MDP
type State interface {
	ID() string
	HasProperty(prop string) bool
	GetProperties() []string
}

// BasicState is a simple implementation of State
type BasicState struct {
	id         string
	properties map[string]bool
}

func NewBasicState(id string, props ...string) *BasicState {
	s := &BasicState{
		id:         id,
		properties: make(map[string]bool),
	}
	for _, prop := range props {
		s.properties[prop] = true
	}
	return s
}

func (s *BasicState) ID() string {
	return s.id
}

func (s *BasicState) HasProperty(prop string) bool {
	return s.properties[prop]
}

func (s *BasicState) GetProperties() []string {
	props := make([]string, 0, len(s.properties))
	for prop := range s.properties {
		props = append(props, prop)
	}
	return props
}

// Message represents a message passed between actors in the MDP
// Messages are equivalent to MDP rewards/awards
type Message struct {
	From    string
	To      string
	Type    string
	Payload interface{}
	Reward  float64 // MDP reward/award associated with this message
}

// Transition represents a state transition in the MDP
type Transition struct {
	FromState   State
	ToState     State
	Probability float64 // For probabilistic transitions
	Action      string  // Action that triggers this transition
	Reward      float64 // Reward for taking this transition
}

// Actor represents an entity in the MDP that can have states and transitions
type Actor struct {
	Name            string
	CurrentState    State
	States          map[string]State
	Transitions     []Transition
	MessageQueue    chan Message
	RewardAccumulated float64
	mu              sync.RWMutex
}

func NewActor(name string, initialState State) *Actor {
	return &Actor{
		Name:          name,
		CurrentState:  initialState,
		States:        make(map[string]State),
		Transitions:   make([]Transition, 0),
		MessageQueue:  make(chan Message, 100),
		RewardAccumulated: 0,
	}
}

// AddState adds a state to the actor
func (a *Actor) AddState(state State) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.States[state.ID()] = state
}

// AddTransition adds a transition to the actor
func (a *Actor) AddTransition(from, to State, probability float64, action string, reward float64) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.Transitions = append(a.Transitions, Transition{
		FromState:   from,
		ToState:     to,
		Probability: probability,
		Action:      action,
		Reward:      reward,
	})
}

// SendMessage sends a message to another actor
func (a *Actor) SendMessage(msg Message) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.RewardAccumulated += msg.Reward
}

// ReceiveMessage receives a message (blocking)
func (a *Actor) ReceiveMessage() Message {
	return <-a.MessageQueue
}

// TryReceiveMessage tries to receive a message (non-blocking)
func (a *Actor) TryReceiveMessage() (Message, bool) {
	select {
	case msg := <-a.MessageQueue:
		return msg, true
	default:
		return Message{}, false
	}
}

// GetPossibleTransitions returns all possible transitions from the current state
func (a *Actor) GetPossibleTransitions() []Transition {
	a.mu.RLock()
	defer a.mu.RUnlock()
	
	possible := make([]Transition, 0)
	for _, t := range a.Transitions {
		if t.FromState.ID() == a.CurrentState.ID() {
			possible = append(possible, t)
		}
	}
	return possible
}

// ExecuteTransition executes a transition (probabilistic or deterministic)
func (a *Actor) ExecuteTransition(action string) (State, float64, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	possible := make([]Transition, 0)
	for _, t := range a.Transitions {
		if t.FromState.ID() == a.CurrentState.ID() && t.Action == action {
			possible = append(possible, t)
		}
	}

	if len(possible) == 0 {
		return nil, 0, fmt.Errorf("no transition found for action %s from state %s", action, a.CurrentState.ID())
	}

	// Choose transition based on probability
	var chosen Transition
	if len(possible) == 1 {
		chosen = possible[0]
	} else {
		// Probabilistic selection
		r := rand.Float64()
		cumulative := 0.0
		for _, t := range possible {
			cumulative += t.Probability
			if r <= cumulative {
				chosen = t
				break
			}
		}
		if chosen.FromState == nil {
			chosen = possible[len(possible)-1] // Fallback
		}
	}

	oldState := a.CurrentState
	a.CurrentState = chosen.ToState
	a.RewardAccumulated += chosen.Reward

	return oldState, chosen.Reward, nil
}

// GetReward returns the accumulated reward
func (a *Actor) GetReward() float64 {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.RewardAccumulated
}

// Model represents the entire MDP model with multiple actors
type Model struct {
	Actors      map[string]*Actor
	States      map[string]State
	InitialState State
	mu          sync.RWMutex
}

func NewModel(initialState State) *Model {
	return &Model{
		Actors:      make(map[string]*Actor),
		States:      make(map[string]State),
		InitialState: initialState,
	}
}

// AddActor adds an actor to the model
func (m *Model) AddActor(actor *Actor) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Actors[actor.Name] = actor
}

// GetActor retrieves an actor by name
func (m *Model) GetActor(name string) (*Actor, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	actor, ok := m.Actors[name]
	return actor, ok
}

// AddState adds a state to the model
func (m *Model) AddState(state State) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.States[state.ID()] = state
}

// GetState retrieves a state by ID
func (m *Model) GetState(id string) (State, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	state, ok := m.States[id]
	return state, ok
}

// GetSuccessors returns all successor states of a given state
func (m *Model) GetSuccessors(state State) []State {
	m.mu.RLock()
	defer m.mu.RUnlock()

	successors := make([]State, 0)
	seenStates := make(map[string]bool)

	for _, actor := range m.Actors {
		for _, t := range actor.Transitions {
			if t.FromState.ID() == state.ID() {
				if !seenStates[t.ToState.ID()] {
					successors = append(successors, t.ToState)
					seenStates[t.ToState.ID()] = true
				}
			}
		}
	}

	return successors
}

// SendMessage sends a message between actors
func (m *Model) SendMessage(msg Message) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	toActor, ok := m.Actors[msg.To]
	if !ok {
		return fmt.Errorf("actor %s not found", msg.To)
	}

	fromActor, ok := m.Actors[msg.From]
	if ok {
		fromActor.SendMessage(msg)
	}

	select {
	case toActor.MessageQueue <- msg:
		return nil
	default:
		return fmt.Errorf("message queue full for actor %s", msg.To)
	}
}

// GetTotalReward returns the total reward accumulated across all actors
func (m *Model) GetTotalReward() float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	total := 0.0
	for _, actor := range m.Actors {
		total += actor.GetReward()
	}
	return total
}

// GetMetrics returns useful metrics about the model
func (m *Model) GetMetrics() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	metrics := make(map[string]interface{})
	metrics["total_actors"] = len(m.Actors)
	metrics["total_states"] = len(m.States)
	metrics["total_reward"] = m.GetTotalReward()

	actorStates := make(map[string]string)
	actorRewards := make(map[string]float64)
	for name, actor := range m.Actors {
		actorStates[name] = actor.CurrentState.ID()
		actorRewards[name] = actor.GetReward()
	}
	metrics["actor_states"] = actorStates
	metrics["actor_rewards"] = actorRewards

	return metrics
}

// CheckCTL checks a CTL formula against the model starting from the initial state
func (m *Model) CheckCTL(formula CTLFormula) bool {
	return formula.Evaluate(m.InitialState, m)
}

// CheckCTLFromState checks a CTL formula against the model starting from a specific state
func (m *Model) CheckCTLFromState(formula CTLFormula, state State) bool {
	return formula.Evaluate(state, m)
}
