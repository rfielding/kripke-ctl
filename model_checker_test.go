package main

import "testing"

func TestKripkeStructureBasics(t *testing.T) {
	k := NewKripkeStructure("s0")
	k.AddTransition("s0", "s1")
	k.AddLabel("s0", "p")

	if k.InitialState != "s0" {
		t.Errorf("Expected initial state s0, got %s", k.InitialState)
	}

	if !k.HasLabel("s0", "p") {
		t.Error("Expected s0 to have label p")
	}

	if k.HasLabel("s1", "p") {
		t.Error("Expected s1 not to have label p")
	}

	successors := k.GetSuccessors("s0")
	if len(successors) != 1 || successors[0] != "s1" {
		t.Errorf("Expected s0 to have successor s1, got %v", successors)
	}
}

func TestAtomicProposition(t *testing.T) {
	k := NewKripkeStructure("s0")
	k.AddLabel("s0", "p")
	k.AddLabel("s1", "q")

	mc := NewModelChecker(k)
	
	// Test atomic proposition p
	result := mc.Check(AtomicProp{"p"})
	if !result["s0"] {
		t.Error("Expected p to hold in s0")
	}
	if result["s1"] {
		t.Error("Expected p not to hold in s1")
	}
}

func TestNegation(t *testing.T) {
	k := NewKripkeStructure("s0")
	k.AddState("s1")
	k.AddLabel("s0", "p")

	mc := NewModelChecker(k)
	
	// Test ¬p
	result := mc.Check(Not{AtomicProp{"p"}})
	if result["s0"] {
		t.Error("Expected ¬p not to hold in s0")
	}
	if !result["s1"] {
		t.Error("Expected ¬p to hold in s1")
	}
}

func TestConjunction(t *testing.T) {
	k := NewKripkeStructure("s0")
	k.AddLabel("s0", "p")
	k.AddLabel("s0", "q")
	k.AddState("s1")
	k.AddLabel("s1", "p")

	mc := NewModelChecker(k)
	
	// Test p ∧ q
	result := mc.Check(And{AtomicProp{"p"}, AtomicProp{"q"}})
	if !result["s0"] {
		t.Error("Expected p ∧ q to hold in s0")
	}
	if result["s1"] {
		t.Error("Expected p ∧ q not to hold in s1")
	}
}

func TestDisjunction(t *testing.T) {
	k := NewKripkeStructure("s0")
	k.AddLabel("s0", "p")
	k.AddState("s1")
	k.AddLabel("s1", "q")

	mc := NewModelChecker(k)
	
	// Test p ∨ q
	result := mc.Check(Or{AtomicProp{"p"}, AtomicProp{"q"}})
	if !result["s0"] {
		t.Error("Expected p ∨ q to hold in s0")
	}
	if !result["s1"] {
		t.Error("Expected p ∨ q to hold in s1")
	}
}

func TestEX(t *testing.T) {
	k := NewKripkeStructure("s0")
	k.AddTransition("s0", "s1")
	k.AddTransition("s1", "s2")
	k.AddLabel("s1", "p")

	mc := NewModelChecker(k)
	
	// Test EX p
	result := mc.Check(EX{AtomicProp{"p"}})
	if !result["s0"] {
		t.Error("Expected EX p to hold in s0")
	}
	if result["s1"] {
		t.Error("Expected EX p not to hold in s1")
	}
}

func TestAX(t *testing.T) {
	k := NewKripkeStructure("s0")
	k.AddTransition("s0", "s1")
	k.AddTransition("s0", "s2")
	k.AddLabel("s1", "p")
	k.AddLabel("s2", "p")

	mc := NewModelChecker(k)
	
	// Test AX p
	result := mc.Check(AX{AtomicProp{"p"}})
	if !result["s0"] {
		t.Error("Expected AX p to hold in s0")
	}
}

func TestEF(t *testing.T) {
	k := NewKripkeStructure("s0")
	k.AddTransition("s0", "s1")
	k.AddTransition("s1", "s2")
	k.AddLabel("s2", "p")

	mc := NewModelChecker(k)
	
	// Test EF p
	result := mc.Check(EF{AtomicProp{"p"}})
	if !result["s0"] {
		t.Error("Expected EF p to hold in s0")
	}
	if !result["s1"] {
		t.Error("Expected EF p to hold in s1")
	}
	if !result["s2"] {
		t.Error("Expected EF p to hold in s2")
	}
}

func TestAF(t *testing.T) {
	k := NewKripkeStructure("s0")
	k.AddTransition("s0", "s1")
	k.AddTransition("s1", "s2")
	k.AddTransition("s2", "s2")
	k.AddLabel("s2", "p")

	mc := NewModelChecker(k)
	
	// Test AF p
	result := mc.Check(AF{AtomicProp{"p"}})
	if !result["s0"] {
		t.Error("Expected AF p to hold in s0")
	}
}

func TestEG(t *testing.T) {
	k := NewKripkeStructure("s0")
	k.AddTransition("s0", "s1")
	k.AddTransition("s1", "s1")
	k.AddLabel("s0", "p")
	k.AddLabel("s1", "p")

	mc := NewModelChecker(k)
	
	// Test EG p
	result := mc.Check(EG{AtomicProp{"p"}})
	if !result["s0"] {
		t.Error("Expected EG p to hold in s0")
	}
	if !result["s1"] {
		t.Error("Expected EG p to hold in s1")
	}
}

func TestAG(t *testing.T) {
	k := NewKripkeStructure("s0")
	k.AddTransition("s0", "s1")
	k.AddTransition("s1", "s0")
	k.AddLabel("s0", "p")
	k.AddLabel("s1", "p")

	mc := NewModelChecker(k)
	
	// Test AG p
	result := mc.Check(AG{AtomicProp{"p"}})
	if !result["s0"] {
		t.Error("Expected AG p to hold in s0")
	}
	if !result["s1"] {
		t.Error("Expected AG p to hold in s1")
	}
}

func TestEU(t *testing.T) {
	k := NewKripkeStructure("s0")
	k.AddTransition("s0", "s1")
	k.AddTransition("s1", "s2")
	k.AddLabel("s0", "p")
	k.AddLabel("s1", "p")
	k.AddLabel("s2", "q")

	mc := NewModelChecker(k)
	
	// Test E[p U q]
	result := mc.Check(EU{AtomicProp{"p"}, AtomicProp{"q"}})
	if !result["s0"] {
		t.Error("Expected E[p U q] to hold in s0")
	}
	if !result["s1"] {
		t.Error("Expected E[p U q] to hold in s1")
	}
	if !result["s2"] {
		t.Error("Expected E[p U q] to hold in s2")
	}
}

func TestAU(t *testing.T) {
	k := NewKripkeStructure("s0")
	k.AddTransition("s0", "s1")
	k.AddTransition("s1", "s2")
	k.AddTransition("s2", "s2")
	k.AddLabel("s0", "p")
	k.AddLabel("s1", "p")
	k.AddLabel("s2", "q")

	mc := NewModelChecker(k)
	
	// Test A[p U q]
	result := mc.Check(AU{AtomicProp{"p"}, AtomicProp{"q"}})
	if !result["s0"] {
		t.Error("Expected A[p U q] to hold in s0")
	}
}

func TestTrafficLightSafety(t *testing.T) {
	k := CreateTrafficLightExample()
	mc := NewModelChecker(k)

	// The traffic light should eventually reach green (go)
	if !mc.Holds(EF{AtomicProp{"go"}}) {
		t.Error("Expected to eventually reach green light")
	}

	// The traffic light should always eventually reach red (stop)
	if !mc.Holds(AF{AtomicProp{"stop"}}) {
		t.Error("Expected to always eventually reach red light")
	}

	// It's not always at caution
	if mc.Holds(AG{AtomicProp{"caution"}}) {
		t.Error("Should not always be at caution")
	}
}

func TestMutualExclusion(t *testing.T) {
	k := CreateMutualExclusionExample()
	mc := NewModelChecker(k)

	// Mutual exclusion: not both in critical section
	// AG ¬(critical1 ∧ critical2)
	notBothCritical := AG{Not{And{AtomicProp{"critical1"}, AtomicProp{"critical2"}}}}
	if !mc.Holds(notBothCritical) {
		t.Error("Mutual exclusion should hold: never both in critical section")
	}

	// If trying, eventually critical: AG(trying1 → AF critical1)
	// Note: This is a liveness property that would need proper fairness assumptions
}

func TestSimpleExample(t *testing.T) {
	k := CreateSimpleExample()
	mc := NewModelChecker(k)

	// s0 has p
	if !mc.Holds(AtomicProp{"p"}) {
		t.Error("Expected p to hold in initial state s0")
	}

	// Eventually q
	if !mc.Holds(EF{AtomicProp{"q"}}) {
		t.Error("Expected to eventually reach q")
	}

	// Not always p (since s1 doesn't have p)
	if mc.Holds(AG{AtomicProp{"p"}}) {
		t.Error("Should not always have p")
	}
}
