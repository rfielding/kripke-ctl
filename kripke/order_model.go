package kripke

// OrderGraph builds a small Kripke model for a single order.
//
// States:
//   s0: New
//   s1: Accepted
//   s2: Delivered
//   s3: Cancelled
//
// Atomic propositions:
//   "accepted", "delivered", "cancelled"
//
// Transitions:
//   s0 -> s1
//   s1 -> s2
//   s1 -> s3
//   s2 -> s2 (absorbing)
//   s3 -> s3 (absorbing)
//
// This is a "good" model where accepted orders always resolve.

func OrderGraph() *Graph {
	g := NewGraph()

	// s0: New (no props set)
	g.AddState("s0", map[string]bool{
		"accepted":  false,
		"delivered": false,
		"cancelled": false,
	})

	// s1: Accepted
	g.AddState("s1", map[string]bool{
		"accepted":  true,
		"delivered": false,
		"cancelled": false,
	})

	// s2: Delivered
	g.AddState("s2", map[string]bool{
		"accepted":  true,
		"delivered": true,
		"cancelled": false,
	})

	// s3: Cancelled
	g.AddState("s3", map[string]bool{
		"accepted":  true,
		"delivered": false,
		"cancelled": true,
	})

	// Edges: New -> Accepted -> (Delivered | Cancelled)
	g.AddEdge("s0", "s1")
	g.AddEdge("s1", "s2")
	g.AddEdge("s1", "s3")

	// Make s2 and s3 absorbing
	g.AddEdge("s2", "s2")
	g.AddEdge("s3", "s3")

	g.SetInitial("s0")
	return g
}
