package main

import (
	"fmt"
	"strings"
)

// GenerateGraphviz generates a Graphviz DOT representation of the Kripke structure
func (k *KripkeStructure) GenerateGraphviz() string {
	var sb strings.Builder
	
	sb.WriteString("digraph KripkeStructure {\n")
	sb.WriteString("  rankdir=LR;\n")
	sb.WriteString("  node [shape=circle];\n")
	sb.WriteString("\n")
	
	// Add invisible start node pointing to initial state
	sb.WriteString("  start [shape=point];\n")
	sb.WriteString(fmt.Sprintf("  start -> \"%s\" [label=\"start\"];\n", k.InitialState))
	sb.WriteString("\n")
	
	// Add nodes with labels
	for _, state := range k.States {
		labels := k.Labeling[state]
		if len(labels) > 0 {
			labelStr := strings.Join(toStringSlice(labels), ", ")
			sb.WriteString(fmt.Sprintf("  \"%s\" [label=\"%s\\n{%s}\"];\n", state, state, labelStr))
		} else {
			sb.WriteString(fmt.Sprintf("  \"%s\" [label=\"%s\"];\n", state, state))
		}
	}
	sb.WriteString("\n")
	
	// Add edges
	for from, tos := range k.Transitions {
		for _, to := range tos {
			sb.WriteString(fmt.Sprintf("  \"%s\" -> \"%s\";\n", from, to))
		}
	}
	
	sb.WriteString("}\n")
	return sb.String()
}

func toStringSlice(props []Proposition) []string {
	result := make([]string, len(props))
	for i, p := range props {
		result[i] = string(p)
	}
	return result
}

// SaveGraphviz saves the Graphviz representation to a file
func (k *KripkeStructure) SaveGraphviz(filename string) error {
	dot := k.GenerateGraphviz()
	return writeFile(filename, dot)
}

func writeFile(filename, content string) error {
	// In a real implementation, this would write to a file
	// For now, we'll just print it
	fmt.Printf("Generated Graphviz DOT for %s:\n%s\n", filename, content)
	return nil
}
