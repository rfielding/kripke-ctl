package kripke

// Node IDs for the PURPLE keyspace collapse scenario.
const (
	NodeUnknownKey      NodeID = "UnknownKey"
	NodeUnsolved        NodeID = "Unsolved"
	NodeVowelSolved     NodeID = "VowelSolved"
	NodeConsonantSolved NodeID = "ConsonantSolved"
	NodeUniqueKey       NodeID = "UniqueKey"
)

// BuildPurpleGraph returns a SimpleGraph plus its initial node ID.
// This graph is purely for documentation/diagram purposes.
func BuildPurpleGraph() (*SimpleGraph, NodeID) {
	g := &SimpleGraph{
		States: []NodeID{
			NodeUnknownKey,
			NodeUnsolved,
			NodeVowelSolved,
			NodeConsonantSolved,
			NodeUniqueKey,
		},
		Succ: make(map[NodeID][]NodeID),
	}

	// High-level story:
	// UnknownKey -> Unsolved -> VowelSolved -> ConsonantSolved -> UniqueKey
	// Plus a self-loop on Unsolved to represent repeated traffic analysis.

	g.Succ[NodeUnknownKey] = []NodeID{NodeUnsolved}
	g.Succ[NodeUnsolved] = []NodeID{
		NodeUnsolved,    // observe more traffic, refine keyspace
		NodeVowelSolved, // at some point, vowel channel is solved
	}
	g.Succ[NodeVowelSolved] = []NodeID{NodeConsonantSolved}
	g.Succ[NodeConsonantSolved] = []NodeID{NodeUniqueKey}
	g.Succ[NodeUniqueKey] = nil // terminal

	return g, NodeUnknownKey
}
