package kripke

import (
	"fmt"
	"sort"
	"strings"
)

// GenerateStateDiagram generates a Mermaid state diagram from the graph
func (g *Graph) GenerateStateDiagram(options ...DiagramOption) string {
	opts := &diagramOptions{
		showLabels:     true,
		stateDescriber: nil,
		edgeLabeler:    nil,
	}
	for _, opt := range options {
		opt(opts)
	}

	var sb strings.Builder
	sb.WriteString("stateDiagram-v2\n")

	// Initial states
	for _, sid := range g.InitialStates() {
		name := g.NameOf(sid)
		sb.WriteString(fmt.Sprintf("    [*] --> %s\n", name))
	}

	// Transitions
	for _, sid := range g.States() {
		fromName := g.NameOf(sid)
		for _, tid := range g.Succ(sid) {
			toName := g.NameOf(tid)
			
			if opts.edgeLabeler != nil {
				label := opts.edgeLabeler(sid, tid, g)
				if label != "" {
					sb.WriteString(fmt.Sprintf("    %s --> %s: %s\n", fromName, toName, label))
				} else {
					sb.WriteString(fmt.Sprintf("    %s --> %s\n", fromName, toName))
				}
			} else {
				sb.WriteString(fmt.Sprintf("    %s --> %s\n", fromName, toName))
			}
		}
	}

	// State descriptions
	if opts.stateDescriber != nil {
		sb.WriteString("\n")
		for _, sid := range g.States() {
			name := g.NameOf(sid)
			desc := opts.stateDescriber(sid, g)
			if desc != "" {
				sb.WriteString(fmt.Sprintf("    %s: %s\n", name, desc))
			}
		}
	}

	return sb.String()
}

// GenerateSequenceDiagram generates a Mermaid sequence diagram from events
func (w *World) GenerateSequenceDiagram(maxEvents int) string {
	var sb strings.Builder
	sb.WriteString("sequenceDiagram\n")
	
	if len(w.Events) == 0 {
		return sb.String()
	}

	// Get unique actors
	actors := make(map[string]bool)
	for _, ev := range w.Events {
		actors[ev.From.ActorID] = true
		actors[ev.To.ActorID] = true
	}

	// Add participants
	sortedActors := make([]string, 0, len(actors))
	for actor := range actors {
		sortedActors = append(sortedActors, actor)
	}
	sort.Strings(sortedActors)
	
	for _, actor := range sortedActors {
		sb.WriteString(fmt.Sprintf("    participant %s\n", actor))
	}
	sb.WriteString("\n")

	// Add messages
	limit := len(w.Events)
	if maxEvents > 0 && maxEvents < limit {
		limit = maxEvents
	}

	for i, ev := range w.Events[:limit] {
		sb.WriteString(fmt.Sprintf("    %s->>%s: msg_%d (delay=%d)\n",
			ev.From.ActorID, ev.To.ActorID, i+1, ev.QueueDelay))
	}

	if limit < len(w.Events) {
		sb.WriteString(fmt.Sprintf("    Note over %s: ... (%d more events)\n",
			sortedActors[0], len(w.Events)-limit))
	}

	return sb.String()
}

// GenerateCTLTable generates a markdown table of CTL verification results
func (g *Graph) GenerateCTLTable(requirements []Requirement) string {
	var sb strings.Builder
	sb.WriteString("| ID | Requirement | CTL Formula | Result |\n")
	sb.WriteString("|----|-------------|-------------|--------|\n")
	
	for _, req := range requirements {
		result := "❓ UNKNOWN"
		if req.Formula != nil {
			// Evaluate formula - it returns a StateSet
			satisfying := req.Formula(g)
			totalStates := len(g.States())
			satisfyingCount := len(satisfying)  // Use len() instead of Cardinality()
			
			if satisfyingCount == totalStates {
				result = "✅ PASS"
			} else if satisfyingCount > 0 {
				result = fmt.Sprintf("⚠️ PARTIAL (%d/%d)", satisfyingCount, totalStates)
			} else {
				result = "❌ FAIL"
			}
		}
		
		sb.WriteString(fmt.Sprintf("| %s | %s | `%s` | %s |\n",
			req.ID, req.Description, req.FormulaString, result))
	}
	
	return sb.String()
}

// DiagramOption configures diagram generation
type DiagramOption func(*diagramOptions)

type diagramOptions struct {
	showLabels     bool
	stateDescriber func(StateID, *Graph) string
	edgeLabeler    func(StateID, StateID, *Graph) string
}

// WithStateDescriber sets a custom state description function
func WithStateDescriber(f func(StateID, *Graph) string) DiagramOption {
	return func(opts *diagramOptions) {
		opts.stateDescriber = f
	}
}

// WithEdgeLabeler sets a custom edge label function
func WithEdgeLabeler(f func(StateID, StateID, *Graph) string) DiagramOption {
	return func(opts *diagramOptions) {
		opts.edgeLabeler = f
	}
}

// Requirement represents a formal requirement with CTL formula
type Requirement struct {
	ID            string
	Category      string
	Description   string
	FormulaString string
	Formula       func(*Graph) StateSet  // CTL formula as function
	EnglishRef    string
	Rationale     string
}

// GenerateRequirementsTable generates a markdown table of requirements
func GenerateRequirementsTable(requirements []Requirement) string {
	var sb strings.Builder
	sb.WriteString("| ID | Category | Requirement | CTL Formula | Traces to |\n")
	sb.WriteString("|----|----------|-------------|-------------|----------|\n")
	
	for _, req := range requirements {
		sb.WriteString(fmt.Sprintf("| %s | %s | %s | `%s` | %s |\n",
			req.ID, req.Category, req.Description, req.FormulaString, req.EnglishRef))
	}
	
	return sb.String()
}

// StateTransition represents a transition in the actor state machine
type StateTransition struct {
	FromActor     string
	ToActor       string
	Guard         string
	Action        string
	VariableEdits []string
}

// GenerateTransitionTable generates table of state transitions
func GenerateTransitionTable(transitions []StateTransition) string {
	var sb strings.Builder
	sb.WriteString("| From | To | Guard | Action | Variable Changes |\n")
	sb.WriteString("|------|----|---------|----|------------------|\n")
	
	for _, t := range transitions {
		edits := strings.Join(t.VariableEdits, ", ")
		sb.WriteString(fmt.Sprintf("| %s | %s | `%s` | %s | %s |\n",
			t.FromActor, t.ToActor, t.Guard, t.Action, edits))
	}
	
	return sb.String()
}
