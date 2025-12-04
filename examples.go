package main

// CreateTrafficLightExample creates a simple traffic light Kripke structure
func CreateTrafficLightExample() *KripkeStructure {
	k := NewKripkeStructure("red")
	
	// Add transitions
	k.AddTransition("red", "green")
	k.AddTransition("green", "yellow")
	k.AddTransition("yellow", "red")
	
	// Add labels
	k.AddLabel("red", "stop")
	k.AddLabel("green", "go")
	k.AddLabel("yellow", "caution")
	
	return k
}

// CreateMutualExclusionExample creates a mutual exclusion example
func CreateMutualExclusionExample() *KripkeStructure {
	k := NewKripkeStructure("n1n2")
	
	// States: n1n2 (both non-critical), t1n2, c1n2, n1t2, n1c2, t1t2, c1t2, t1c2
	// n = non-critical, t = trying, c = critical
	
	k.AddTransition("n1n2", "t1n2")
	k.AddTransition("n1n2", "n1t2")
	k.AddTransition("t1n2", "c1n2")
	k.AddTransition("t1n2", "t1t2")
	k.AddTransition("n1t2", "t1t2")
	k.AddTransition("n1t2", "n1c2")
	k.AddTransition("c1n2", "n1n2")
	k.AddTransition("n1c2", "n1n2")
	k.AddTransition("t1t2", "c1t2")
	k.AddTransition("t1t2", "t1c2")
	k.AddTransition("c1t2", "n1t2")
	k.AddTransition("t1c2", "t1n2")
	
	// Labels
	k.AddLabel("c1n2", "critical1")
	k.AddLabel("n1c2", "critical2")
	k.AddLabel("c1t2", "critical1")
	k.AddLabel("t1c2", "critical2")
	
	k.AddLabel("t1n2", "trying1")
	k.AddLabel("t1t2", "trying1")
	k.AddLabel("t1c2", "trying1")
	
	k.AddLabel("n1t2", "trying2")
	k.AddLabel("t1t2", "trying2")
	k.AddLabel("c1t2", "trying2")
	
	return k
}

// CreateSimpleExample creates a very simple example for testing
func CreateSimpleExample() *KripkeStructure {
	k := NewKripkeStructure("s0")
	
	k.AddTransition("s0", "s1")
	k.AddTransition("s1", "s2")
	k.AddTransition("s2", "s1")
	
	k.AddLabel("s0", "p")
	k.AddLabel("s1", "q")
	k.AddLabel("s2", "p")
	k.AddLabel("s2", "q")
	
	return k
}
