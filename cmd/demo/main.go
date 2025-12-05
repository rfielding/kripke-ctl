package main

import (
	"fmt"

	"github.com/rfielding/kripke-ctl/kripke"
)

// Producer sends N integer messages to a given channel.
type Producer struct {
	id        string
	target    kripke.Address
	nextValue int
	maxValue  int
}

func NewProducer(id string, target kripke.Address, max int) *Producer {
	return &Producer{
		id:        id,
		target:    target,
		nextValue: 1,
		maxValue:  max,
	}
}

func (p *Producer) ID() string { return p.id }

// Ready generates at most one Step: "send nextValue to target"
// if we still have work to do and the channel is not full.
func (p *Producer) Ready(w *kripke.World) []kripke.Step {
	var steps []kripke.Step

	if p.nextValue > p.maxValue {
		return steps // done
	}

	ch := w.ChannelByAddress(p.target)
	if ch == nil || !ch.CanSend() {
		return steps // blocked on full or missing channel
	}

	value := p.nextValue
	steps = append(steps, func(w *kripke.World) {
		msg := kripke.Message{
			From: kripke.Address{
				ActorID:     p.id,
				ChannelName: "out",
			},
			To:      p.target,
			Payload: value,
			ReplyTo: nil,
		}
		ok := kripke.SendMessage(w, msg)
		if !ok {
			// Should not happen if CanSend() was true.
			return
		}
		p.nextValue++
	})

	return steps
}

// Consumer receives integers from its owned channel and accumulates them.
type Consumer struct {
	id         string
	channel    kripke.Address
	Total      int // exported so we can print it
	RecvCount  int
	LastRecvAt int // logical time of last receive
}

func NewConsumer(id, chanName string) *Consumer {
	return &Consumer{
		id: id,
		channel: kripke.Address{
			ActorID:     id,
			ChannelName: chanName,
		},
	}
}

func (c *Consumer) ID() string { return c.id }

func (c *Consumer) Ready(w *kripke.World) []kripke.Step {
	var steps []kripke.Step

	ch := w.ChannelByAddress(c.channel)
	if ch == nil || !ch.CanRecv() {
		return steps // no messages waiting
	}

	steps = append(steps, func(w *kripke.World) {
		msg, ok := kripke.RecvAndLog(w, ch)
		if !ok {
			return
		}

		// Try to interpret payload as int.
		if v, ok := msg.Payload.(int); ok {
			c.Total += v
		}
		c.RecvCount++
		c.LastRecvAt = w.Time
	})

	return steps
}

func main() {
	consumerID := "C1"
	chanName := "inbox"

	inbox := kripke.NewChannel(consumerID, chanName, 2) // capacity=2

	consumer := NewConsumer(consumerID, chanName)
	targetAddr := inbox.Address()
	producer := NewProducer("P1", targetAddr, 5) // send 1..5

	procs := []kripke.Process{producer, consumer}
	chans := []*kripke.Channel{inbox}

	w := kripke.NewWorld(procs, chans, 1) // fixed seed for reproducibility

	w.RunSteps(50)

	fmt.Println("=== Kripke engine demo (Producer/Consumer) ===")
	fmt.Printf("Producer sent up to: %d (max %d)\n", producer.nextValue-1, producer.maxValue)
	fmt.Printf("Consumer total: %d (RecvCount=%d, LastRecvAt=%d)\n",
		consumer.Total, consumer.RecvCount, consumer.LastRecvAt)

	fmt.Printf("Total receive events: %d\n", len(w.Events))
	for _, ev := range w.Events {
		fmt.Printf(
			"t=%d MsgID=%d CorrID=%d %s -> %s payload=%v qDelay=%d (enq=%d)\n",
			ev.Time,
			ev.MsgID,
			ev.CorrelationID,
			ev.From.String(),
			ev.To.String(),
			ev.Payload,
			ev.QueueDelay,
			ev.EnqueueTime,
		)
	}
}
