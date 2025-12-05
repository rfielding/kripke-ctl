// Package kripke implements a tiny communicating MDP engine:
//
//   - Processes (actors) with local state and Ready(*World) []Step
//   - Channels owned by actors, with capacity (cap>0 buffered)
//   - A scheduler that picks exactly one enabled Step per tick
//   - A global event log for sequence diagrams and metrics
//
// Rendezvous (cap == 0) is left as a TODO, but the structure is ready for it.
package kripke

import (
	"fmt"
	"math/rand"
	"time"
)

// Step is one atomic transition in the global Kripke graph.
// It may mutate actor-local state, channels, and log events.
type Step func(*World)

// Process is an actor: local state + Ready() to emit enabled Steps.
type Process interface {
	ID() string
	Ready(w *World) []Step
}

// Address identifies a specific inbound channel on a specific actor.
type Address struct {
	ActorID     string
	ChannelName string
}

func (a Address) String() string {
	return fmt.Sprintf("%s.%s", a.ActorID, a.ChannelName)
}

// Message is what flows through channels.
// ReplyTo is optional and can be used for request/response patterns.
type Message struct {
	From    Address
	To      Address
	Payload any
	ReplyTo *Address
}

// Event is an immutable record of what happened at logical time Time.
type Event struct {
	Time     int
	From     Address
	FromChan string
	To       Address
	ToChan   string
	Payload  any
	ReplyTo  *Address
}

// Channel is a FIFO queue owned by a single actor.
// cap > 0  => buffered queue
// cap == 0 => intended rendezvous (TODO, currently treated as cap 1)
type Channel struct {
	OwnerID string // actor that owns (reads) from this channel
	Name    string
	cap     int
	buf     []Message
}

// NewChannel constructs a channel with the given owner, name, and capacity.
// For now, cap == 0 is treated as cap == 1 (see TODO below).
func NewChannel(ownerID, name string, cap int) *Channel {
	if cap < 0 {
		panic("channel capacity must be >= 0")
	}
	// TODO: true rendezvous semantics when cap == 0.
	if cap == 0 {
		cap = 1
	}
	return &Channel{
		OwnerID: ownerID,
		Name:    name,
		cap:     cap,
		buf:     make([]Message, 0),
	}
}

func (ch *Channel) Capacity() int     { return ch.cap }
func (ch *Channel) Len() int          { return len(ch.buf) }
func (ch *Channel) IsEmpty() bool     { return len(ch.buf) == 0 }
func (ch *Channel) IsFull() bool      { return len(ch.buf) >= ch.cap }
func (ch *Channel) CanSend() bool     { return !ch.IsFull() }
func (ch *Channel) CanRecv() bool     { return !ch.IsEmpty() }
func (ch *Channel) Address() Address  { return Address{ActorID: ch.OwnerID, ChannelName: ch.Name} }
func (ch *Channel) String() string    { return ch.Address().String() }

// TrySend enqueues msg if the channel is not full.
// Returns true on success, false if it would block.
func (ch *Channel) TrySend(msg Message) bool {
	if !ch.CanSend() {
		return false
	}
	ch.buf = append(ch.buf, msg)
	return true
}

// TryRecv dequeues the oldest message if present.
// Returns (msg, true) on success, or (zero, false) if empty.
func (ch *Channel) TryRecv() (Message, bool) {
	if !ch.CanRecv() {
		return Message{}, false
	}
	msg := ch.buf[0]
	copy(ch.buf, ch.buf[1:])
	ch.buf = ch.buf[:len(ch.buf)-1]
	return msg, true
}

// World contains all actors, channels, global time, RNG, and event log.
type World struct {
	Time     int
	Procs    []Process
	Channels map[string]*Channel // key = Address.String()
	Events   []Event
	rng      *rand.Rand
}

// NewWorld constructs a world with the given processes and channels.
// rngSeed can be fixed for reproducible runs.
func NewWorld(procs []Process, chans []*Channel, rngSeed int64) *World {
	m := make(map[string]*Channel, len(chans))
	for _, ch := range chans {
		if ch == nil {
			continue
		}
		key := ch.Address().String()
		if _, exists := m[key]; exists {
			panic("duplicate channel address: " + key)
		}
		m[key] = ch
	}
	if rngSeed == 0 {
		rngSeed = time.Now().UnixNano()
	}
	return &World{
		Time:     0,
		Procs:    procs,
		Channels: m,
		Events:   make([]Event, 0),
		rng:      rand.New(rand.NewSource(rngSeed)),
	}
}

// ChannelByAddress returns the channel with the given address, or nil.
func (w *World) ChannelByAddress(a Address) *Channel {
	return w.Channels[a.String()]
}

// LogEvent appends an event to the global event log.
func (w *World) LogEvent(ev Event) {
	w.Events = append(w.Events, ev)
}

// EnabledSteps collects all enabled steps from all actors.
func (w *World) EnabledSteps() []Step {
	var enabled []Step
	for _, p := range w.Procs {
		if p == nil {
			continue
		}
		steps := p.Ready(w)
		if len(steps) == 0 {
			continue
		}
		enabled = append(enabled, steps...)
	}
	return enabled
}

// StepRandom executes exactly one enabled step chosen uniformly at random.
// Returns false if no steps are enabled (quiescent or deadlocked).
func (w *World) StepRandom() bool {
	enabled := w.EnabledSteps()
	if len(enabled) == 0 {
		return false
	}
	idx := w.rng.Intn(len(enabled))
	step := enabled[idx]
	step(w)
	w.Time++
	return true
}

// RunSteps executes up to maxSteps steps or until there are no enabled steps.
func (w *World) RunSteps(maxSteps int) {
	for i := 0; i < maxSteps; i++ {
		if !w.StepRandom() {
			return
		}
	}
}

// Helper for actors: send a message through a known channel.
// This enforces Channel ownership semantics and logs an Event.
//
// Typical usage inside a Step:
//    msg := Message{From: fromAddr, To: toAddr, Payload: payload, ReplyTo: replyTo}
//    if !SendMessage(w, msg) { return } // should not happen if Ready() used CanSend
func SendMessage(w *World, msg Message) bool {
	ch := w.ChannelByAddress(msg.To)
	if ch == nil {
		panic("SendMessage: no such channel: " + msg.To.String())
	}
	if !ch.TrySend(msg) {
		return false
	}

	w.LogEvent(Event{
		Time:     w.Time,
		From:     msg.From,
		FromChan: msg.From.ChannelName,
		To:       msg.To,
		ToChan:   msg.To.ChannelName,
		Payload:  msg.Payload,
		ReplyTo:  msg.ReplyTo,
	})

	return true
}

// Helper for actors: receive from a channel they own.
// Typical usage inside Ready():
//
//    chAddr := Address{ActorID: p.ID(), ChannelName: "in"}
//    ch := w.ChannelByAddress(chAddr)
//    if ch != nil && ch.CanRecv() {
//        steps = append(steps, func(w *World) {
//            msg, ok := ch.TryRecv()
//            if !ok { return } // should be rare
//            // handle msg ...
//        })
//    }
func RecvMessage(ch *Channel) (Message, bool) {
	return ch.TryRecv()
}

