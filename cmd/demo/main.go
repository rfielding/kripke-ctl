package main

import (
	"encoding/json"
	"fmt"

	"github.com/rfielding/kripke-ctl/kripke"
)

// ----- Demo Process: Counter -----

// Counter is a trivial demo Process that increments a value.
type Counter struct {
	IDValue string
	Value   int
	Max     int
}

func NewCounter(id string, max int) *Counter {
	return &Counter{
		IDValue: id,
		Max:     max,
	}
}

func (c *Counter) ID() string   { return c.IDValue }
func (c *Counter) Kind() string { return "Counter" }

func (c *Counter) Ready(w *kripke.World) []kripke.Step {
	if c.Value >= c.Max {
		return nil
	}
	return []kripke.Step{c.stepInc()}
}

func (c *Counter) stepInc() kripke.Step {
	return func(w *kripke.World) {
		c.Value++
		w.Events = append(w.Events, kripke.Event{
			Time: w.Time,
			Type: "Increment",
			From: kripke.Endpoint{ActorID: c.IDValue, ChannelName: "internal"},
			Data: map[string]any{
				"value": c.Value,
			},
		})
	}
}

func main() {
	fmt.Println("=== Kripke engine demo (Counter) ===")

	w := kripke.NewWorld()
	counter := NewCounter("C1", 5)
	w.Add(counter)

	w.Run(20)

	fmt.Println("Counter final value:", counter.Value)
	fmt.Println("Total events:", len(w.Events))

	for _, ev := range w.Events {
		b, _ := json.Marshal(ev)
		fmt.Println(string(b))
	}

	// ----- Tiny CTL demo over a hand-built graph -----

	fmt.Println("\n=== CTL demo: EG(p) on a 3-state graph ===")

	// States: s0 -> s1 -> s2 -> s2 (self-loop)
	s0 := kripke.StateID("s0")
	s1 := kripke.StateID("s1")
	s2 := kripke.StateID("s2")

	g := &kripke.Graph{
		States: []kripke.StateID{s0, s1, s2},
		Succ: map[kripke.StateID][]kripke.StateID{
			s0: {s1},
			s1: {s2},
			s2: {s2},
		},
	}

	// Atomic proposition p holds in s1 and s2.
	pStates := kripke.NewStateSet()
	pStates.Add(s1)
	pStates.Add(s2)

	p := kripke.Atom{States: pStates}
	eg := kripke.EG{F: p}

	satisfying := eg.Sat(g)
	fmt.Println("States satisfying EG(p):", satisfying.ToSlice())

	fmt.Println("s0 |= EG(p)?", kripke.SatIn(eg, g, s0))
	fmt.Println("s1 |= EG(p)?", kripke.SatIn(eg, g, s1))
	fmt.Println("s2 |= EG(p)?", kripke.SatIn(eg, g, s2))
}

