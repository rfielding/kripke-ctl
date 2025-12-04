package main

import (
	"strings"
	"testing"
)

func TestGraphvizGeneration(t *testing.T) {
	k := CreateSimpleExample()
	dot := k.GenerateGraphviz()

	// Check that DOT output contains expected elements
	if !strings.Contains(dot, "digraph KripkeStructure") {
		t.Error("Expected digraph declaration")
	}

	if !strings.Contains(dot, "start") {
		t.Error("Expected start node")
	}

	if !strings.Contains(dot, `"s0"`) {
		t.Error("Expected s0 state")
	}

	if !strings.Contains(dot, `"s1"`) {
		t.Error("Expected s1 state")
	}

	if !strings.Contains(dot, `"s2"`) {
		t.Error("Expected s2 state")
	}

	// Check for transitions
	if !strings.Contains(dot, `"s0" -> "s1"`) {
		t.Error("Expected s0 -> s1 transition")
	}

	if !strings.Contains(dot, `"s1" -> "s2"`) {
		t.Error("Expected s1 -> s2 transition")
	}
}

func TestGraphvizLabels(t *testing.T) {
	k := NewKripkeStructure("s0")
	k.AddTransition("s0", "s1")
	k.AddLabel("s0", "p")
	k.AddLabel("s0", "q")

	dot := k.GenerateGraphviz()

	// Check that labels are included
	if !strings.Contains(dot, "{p, q}") && !strings.Contains(dot, "{q, p}") {
		t.Error("Expected labels to be included in DOT output")
	}
}

func TestTrafficLightVisualization(t *testing.T) {
	k := CreateTrafficLightExample()
	dot := k.GenerateGraphviz()

	// Verify all states are present
	states := []string{"red", "green", "yellow"}
	for _, state := range states {
		if !strings.Contains(dot, `"`+string(state)+`"`) {
			t.Errorf("Expected state %s in visualization", state)
		}
	}

	// Verify labels are present
	labels := []string{"stop", "go", "caution"}
	for _, label := range labels {
		if !strings.Contains(dot, label) {
			t.Errorf("Expected label %s in visualization", label)
		}
	}
}
