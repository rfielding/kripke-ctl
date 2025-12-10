package kripke

import (
	"fmt"
	"strings"
)

// GenerateTLAPlus generates a TLA+ specification using KripkeLib operators
// This is for documentation purposes - to show what TLA+ would look like
// It's not a complete translation, just a readable example
func (w *World) GenerateTLAPlus(moduleName string) string {
	var tla strings.Builder
	
	// Module header
	tla.WriteString(fmt.Sprintf("---- MODULE %s ----\n", moduleName))
	tla.WriteString("EXTENDS Naturals, Sequences, KripkeLib\n\n")
	
	// Constants
	tla.WriteString("CONSTANTS\n")
	tla.WriteString("    MaxMessages,       \\* Maximum messages to send\n")
	tla.WriteString("    ChannelCapacity    \\* Channel buffer size\n\n")
	
	// Variables
	tla.WriteString("VARIABLES\n")
	tla.WriteString("    producer_count,    \\* Messages sent by producer\n")
	tla.WriteString("    consumer_count,    \\* Messages received by consumer\n")
	tla.WriteString("    channel            \\* Communication channel\n\n")
	
	tla.WriteString("vars == <<producer_count, consumer_count, channel>>\n\n")
	
	// Type invariant
	tla.WriteString("TypeOK ==\n")
	tla.WriteString("    /\\ producer_count \\in Nat\n")
	tla.WriteString("    /\\ consumer_count \\in Nat\n")
	tla.WriteString("    /\\ channel \\in Seq(Nat)\n\n")
	
	// Initial state
	tla.WriteString("Init ==\n")
	tla.WriteString("    /\\ producer_count = 0\n")
	tla.WriteString("    /\\ consumer_count = 0\n")
	tla.WriteString("    /\\ channel = <<>>\n\n")
	
	// Producer action
	tla.WriteString("\\* Producer: sends messages to channel\n")
	tla.WriteString("Producer ==\n")
	tla.WriteString("    /\\ producer_count < MaxMessages\n")
	tla.WriteString("    /\\ can_send(channel, ChannelCapacity)\n")
	tla.WriteString("    /\\ channel' = snd(channel, producer_count)    \\* Process calculus: channel ! msg\n")
	tla.WriteString("    /\\ producer_count' = producer_count + 1\n")
	tla.WriteString("    /\\ UNCHANGED consumer_count\n\n")
	
	// Consumer action
	tla.WriteString("\\* Consumer: receives messages from channel\n")
	tla.WriteString("Consumer ==\n")
	tla.WriteString("    LET result == rcv(channel) IN                 \\* Process calculus: channel ? msg\n")
	tla.WriteString("    /\\ can_recv(channel)\n")
	tla.WriteString("    /\\ channel' = result.channel\n")
	tla.WriteString("    /\\ consumer_count' = consumer_count + 1\n")
	tla.WriteString("    /\\ UNCHANGED producer_count\n\n")
	
	// Example: Chance node (optional - show the pattern)
	tla.WriteString("\\* Example: Action with probabilistic failure\n")
	tla.WriteString("\\* Dice roll R1 on every scheduler attempt (0-100)\n")
	tla.WriteString("ProducerWithFailure ==\n")
	tla.WriteString("    /\\ producer_count < MaxMessages\n")
	tla.WriteString("    /\\ can_send(channel, ChannelCapacity)\n")
	tla.WriteString("    /\\ (\n")
	tla.WriteString("        \\* choice(0 <= R1 < 95, success) - 95% success\n")
	tla.WriteString("        \\/ choice(0, 95, TRUE,\n")
	tla.WriteString("                  /\\ channel' = snd(channel, producer_count)\n")
	tla.WriteString("                  /\\ producer_count' = producer_count + 1)\n")
	tla.WriteString("        \n")
	tla.WriteString("        \\* choice(95 <= R1 < 100, failure) - 5% failure\n")
	tla.WriteString("        \\/ choice(95, 100, TRUE,\n")
	tla.WriteString("                  UNCHANGED vars)\n")
	tla.WriteString("       )\n")
	tla.WriteString("    /\\ UNCHANGED consumer_count\n\n")
	
	// Next state relation
	tla.WriteString("Next ==\n")
	tla.WriteString("    \\/ Producer\n")
	tla.WriteString("    \\/ Consumer\n\n")
	
	// Specification
	tla.WriteString("\\* Temporal specification\n")
	tla.WriteString("Spec == Init /\\ [][Next]_vars\n\n")
	
	// Safety properties
	tla.WriteString("\\* Safety: system never violates these conditions\n")
	tla.WriteString("Safety ==\n")
	tla.WriteString("    /\\ TypeOK\n")
	tla.WriteString("    /\\ producer_count <= MaxMessages\n")
	tla.WriteString("    /\\ Len(channel) <= ChannelCapacity\n")
	tla.WriteString("    /\\ consumer_count <= producer_count\n\n")
	
	// Liveness properties
	tla.WriteString("\\* Liveness: system eventually satisfies these conditions\n")
	tla.WriteString("EventuallyComplete ==\n")
	tla.WriteString("    <>(producer_count = MaxMessages /\\ Len(channel) = 0)\n\n")
	
	// Fairness
	tla.WriteString("\\* Fairness: ensure all actors make progress\n")
	tla.WriteString("Fairness ==\n")
	tla.WriteString("    /\\ WF_vars(Producer)\n")
	tla.WriteString("    /\\ WF_vars(Consumer)\n\n")
	
	// Module footer
	tla.WriteString("====\n")
	
	return tla.String()
}
