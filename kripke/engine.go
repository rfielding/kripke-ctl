package kripke

import (
	"math/rand"
	"time"
)

// Step is an atomic transition function for a Process.
type Step func(w *World)

// Process is an actor in the Kripke world.
type Process interface {
	ID() string
	Kind() string
	Ready(w *World) []Step
}

// Endpoint identifies an actor + logical channel.
type Endpoint struct {
	ActorID     string
	ChannelName string
}

// Event is a semantic log entry used for diagrams & metrics.
type Event struct {
	Time int64
	Type string
	From Endpoint
	To   Endpoint
	Data map[string]any
}

// Channel is a bounded FIFO queue owned by an actor.
type Channel[M any] struct {
	Capacity int
	Queue    []M
}

func NewChannel[M any](cap int) *Channel[M] {
	if cap <= 0 {
		cap = 1
	}
	return &Channel[M]{Capacity: cap, Queue: make([]M, 0, cap)}
}

func (ch *Channel[M]) CanSend() bool {
	return len(ch.Queue) < ch.Capacity
}

func (ch *Channel[M]) Send(m M) bool {
	if !ch.CanSend() {
		return false
	}
	ch.Queue = append(ch.Queue, m)
	return true
}

func (ch *Channel[M]) CanRecv() bool {
	return len(ch.Queue) > 0
}

func (ch *Channel[M]) Recv() (M, bool) {
	var zero M
	if len(ch.Queue) == 0 {
		return zero, false
	}
	m := ch.Queue[0]
	ch.Queue = ch.Queue[1:]
	return m, true
}

// World is a single Kripke state + scheduler + event log.
type World struct {
	Procs  map[string]Process
	RNG    *rand.Rand
	Time   int64
	Events []Event
}

// NewWorld creates an empty world with a random-seeded RNG.
func NewWorld() *World {
	return &World{
		Procs: make(map[string]Process),
		RNG:   rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// Add inserts a Process into the world.
func (w *World) Add(p Process) {
	w.Procs[p.ID()] = p
}

// RunOneStep:
//   - collects all READY processes
//   - chooses one at random
//   - chooses one of its steps at random
//   - executes it and advances Time
// returns false if nothing is READY (quiescent / deadlock).
func (w *World) RunOneStep() bool {
	type readyInfo struct {
		p     Process
		steps []Step
	}

	var ready []readyInfo

	for _, p := range w.Procs {
		steps := p.Ready(w)
		if len(steps) > 0 {
			ready = append(ready, readyInfo{p: p, steps: steps})
		}
	}

	if len(ready) == 0 {
		return false
	}

	choice := ready[w.RNG.Intn(len(ready))]
	steps := choice.steps
	step := steps[w.RNG.Intn(len(steps))]
	step(w)
	w.Time++

	return true
}

// Run repeatedly calls RunOneStep up to maxSteps times.
func (w *World) Run(maxSteps int) {
	for i := 0; i < maxSteps; i++ {
		if !w.RunOneStep() {
			return
		}
	}
}

