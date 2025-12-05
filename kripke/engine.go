// Package kripke implements a tiny communicating MDP engine:
//
//   - Processes (actors) with local state and Ready(*World) []Step
//   - Channels owned by actors, with capacity (cap > 0 buffered)
//   - A scheduler that picks exactly one enabled Step per tick
//   - A global event log for sequence diagrams and metrics
//
// NOTE: cap == 0 rendezvous semantics are not implemented yet; for now we
// require capacity >= 1 and treat that as a buffered channel.
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
//
// ID:
//   globally unique per message
//
// CorrelationID:
//   - for requests: usually set to the same as ID (self-correlated)
//   - for replies: copied from the original request's CorrelationID
//
// EnqueueTime:
//   logical time when the message was put into a channel.
type Message struct {
	ID            uint64
	CorrelationID uint64
	From          Address
	To            Address
	Payload       any
	ReplyTo       *Address
	EnqueueTime   int
}

// Event is an immutable record of a message being received.
//
// Time:
//   logical time when the message was consumed.
//
// EnqueueTime:
//   when the message entered the queue (from Message.EnqueueTime).
//
// QueueDelay:
//   Time - EnqueueTime (queueing latency in logical ticks).
type Event struct {
	Time          int
	MsgID         uint64
	CorrelationID uint64
	From          Address
	FromChan      string
	To            Address
	ToChan        string
	Payload       any
	ReplyTo       *Address
	EnqueueTime   int
	QueueDelay    int
}

// Channel is a FIFO queue owned by a single actor.
//
// cap > 0  => buffered queue
// cap == 0 => not supported yet (would be rendezvous).
type Channel struct {
	OwnerID string
	Name    string
	cap     int
	buf     []Message
}

// NewChannel constructs a channel with the given owner, name, and capacity.
// For now, cap must be >= 1. Use a separate mechanism later if you want true
// rendezvous (cap == 0) semantics.
func NewChannel(ownerID, name string, cap int) *Channel {
	if cap <= 0 {
		panic("NewChannel: capacity must be >= 1 for now")
	}
	return &Channel{
		OwnerID: ownerID,
		Name:    name,
		cap:     cap,
		buf:     make([]Message, 0),
	}
}

func (ch *Channel) Capacity() int    { return ch.cap }
func (ch *Channel) Len() int         { return len(ch.buf) }
func (ch *Channel) IsEmpty() bool    { return len(ch.buf) == 0 }
func (ch *Channel) IsFull() bool     { return len(ch.buf) >= ch.cap }
func (ch *Channel) CanSend() bool    { return !ch.IsFull() }
func (ch *Channel) CanRecv() bool    { return !ch.IsEmpty() }
func (ch *Channel) Address() Address { return Address{ActorID: ch.OwnerID, ChannelName: ch.Name} }
func (ch *Channel) String() string   { return ch.Address().String() }

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
	Time      int
	Procs     []Process
	Channels  map[string]*Channel // key = Address.String()
	Events    []Event
	rng       *rand.Rand
	nextMsgID uint64
}

// NewWorld constructs a world with the given processes and channels.
// rngSeed can be fixed for reproducible runs (if 0, uses current time).
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
		Time:      0,
		Procs:     procs,
		Channels:  m,
		Events:    make([]Event, 0),
		rng:       rand.New(rand.NewSource(rngSeed)),
		nextMsgID: 1,
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
// Each actor is responsible for using guards + CanSend/CanRecv so that
// returned Steps are actually feasible.
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

// SendMessage enqueues a message into the receiver's channel and assigns IDs.
//
// It assumes the caller has already ensured ch.CanSend() in Ready().
func SendMessage(w *World, msg Message) bool {
	ch := w.ChannelByAddress(msg.To)
	if ch == nil {
		panic("SendMessage: no such channel: " + msg.To.String())
	}
	if !ch.CanSend() {
		return false
	}

	// Assign message ID.
	msg.ID = w.nextMsgID
	w.nextMsgID++

	// Default correlation: request = self-correlation.
	if msg.CorrelationID == 0 {
		msg.CorrelationID = msg.ID
	}

	// Stamp enqueue time.
	msg.EnqueueTime = w.Time

	if !ch.TrySend(msg) {
		return false
	}

	// Note: we log on receive, not send.
	return true
}

// RecvMessage is a low-level helper that just dequeues a message.
// Usually you want RecvAndLog to also create an Event.
func RecvMessage(ch *Channel) (Message, bool) {
	return ch.TryRecv()
}

// RecvAndLog dequeues a message from ch and logs an Event with queue metrics.
//
// This should be used inside a Step body after Ready() has confirmed CanRecv().
func RecvAndLog(w *World, ch *Channel) (Message, bool) {
	msg, ok := ch.TryRecv()
	if !ok {
		return Message{}, false
	}

	ev := Event{
		Time:          w.Time,
		MsgID:         msg.ID,
		CorrelationID: msg.CorrelationID,
		From:          msg.From,
		FromChan:      msg.From.ChannelName,
		To:            msg.To,
		ToChan:        msg.To.ChannelName,
		Payload:       msg.Payload,
		ReplyTo:       msg.ReplyTo,
		EnqueueTime:   msg.EnqueueTime,
		QueueDelay:    w.Time - msg.EnqueueTime,
	}
	w.LogEvent(ev)

	return msg, true
}
