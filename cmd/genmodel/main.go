package main

import (
    "flag"
    "fmt"
    "os"
    "path/filepath"
    "strings"
)

// genmodel -name X
// Produces:
//   models/X/model.go
//   docs/models/X.md

func main() {
    name := flag.String("name", "", "model name (required)")
    flag.Parse()

    if *name == "" {
        fmt.Println("Usage: genmodel -name mm1")
        os.Exit(1)
    }

    modelName := *name
    dir := filepath.Join("models", modelName)
    docDir := filepath.Join("docs", "models")

    if err := os.MkdirAll(dir, 0o755); err != nil {
        panic(err)
    }
    if err := os.MkdirAll(docDir, 0o755); err != nil {
        panic(err)
    }

    // --------------------------------------------------------------------
    // Write Go model file
    // --------------------------------------------------------------------

    goFile := filepath.Join(dir, "model.go")

    goContents := fmt.Sprintf(
`package %s

import "github.com/rfielding/kripke-ctl/kripke"

// BuildModel constructs the Kripke state graph.
// Fill in states and transitions manually.
func BuildModel() *kripke.SimpleGraph {
    g := &kripke.SimpleGraph{
        States: []kripke.NodeID{},
        Succ:   map[kripke.NodeID][]kripke.NodeID{},
    }

    // Example:
    // g.States = []kripke.NodeID{"Start", "End"}
    // g.Succ["Start"] = []kripke.NodeID{"End"}

    return g
}

func Spec() *kripke.ModelSpec {
    return &kripke.ModelSpec{
        Name: "%s",
        CTL: []kripke.CTLSpec{
            // Example:
            // { Name: "NoDeadlock", Description: "No state is deadlocked", Formula: "AG EF Start" },
        },
        Counters: []kripke.CounterSpec{},
    }
}
`, modelName, modelName)

    if err := os.WriteFile(goFile, []byte(goContents), 0o644); err != nil {
        panic(err)
    }

    fmt.Println("Created:", goFile)

    // --------------------------------------------------------------------
    // Write Markdown documentation stub
    // --------------------------------------------------------------------

    mdFile := filepath.Join(docDir, modelName+".md")

    mdContents := fmt.Sprintf(
`# %s Model

This file documents the model generated from English.

## Editing

You should:

- describe the scenario
- list states + transitions
- add CTL formulas
- include charts/diagrams

## Diagrams

~~~mermaid
%% placeholder — cmd/gendiagrams will regenerate this diagram
stateDiagram-v2
    [*] --> Start
    Start --> End
~~~

`, strings.Title(modelName))

    if err := os.WriteFile(mdFile, []byte(mdContents), 0o644); err != nil {
        panic(err)
    }

    fmt.Println("Created:", mdFile)
}
