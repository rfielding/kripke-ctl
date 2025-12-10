# kripke-ctl: Add These 2 Files

## What to do

```bash
# Extract this tarball
tar xzf kripke-additions.tar.gz

# Copy the 2 files to your kripke package
cd ~/code/kripke-ctl
cp kripke-additions/diagrams.go kripke/
cp kripke-additions/metrics.go kripke/

# Test
go run ./cmd/demo/
```

That's it! These are the only 2 files you need to add.

## What these files provide

### diagrams.go
- `g.GenerateStateDiagram()` - Generate Mermaid state diagrams
- `w.GenerateSequenceDiagram(n)` - Generate sequence diagrams from execution
- `g.GenerateCTLTable(reqs)` - Generate CTL verification tables
- Helper functions for custom edge labels, state descriptions

### metrics.go  
- `NewMetricsCollector()` - Create metrics collector
- `CalculateThroughput()` - Calculate throughput from byte counts
- `GenerateActorStateTable()` - Generate tables of actor states
- Helper functions for metrics tables and charts

## Example usage

```go
package main

import "github.com/rfielding/kripke-ctl/kripke"

func main() {
    // Build your graph as usual
    g := kripke.NewGraph()
    // ... add states and edges ...
    
    // NEW: Generate diagram
    diagram := g.GenerateStateDiagram()
    println(diagram)
}
```

## API compatibility

These files are corrected to match your actual kripke API:
- ✅ CTL formulas are functions: `req.Formula(g)` not `req.Formula.Check(g)`
- ✅ Uses public methods: `g.States()` not `g.states`  
- ✅ Doesn't access private fields: removed `w.Processes` access

Should compile without errors.
