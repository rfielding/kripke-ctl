
package main

import (
    "fmt"
    "os"

    "github.com/rfielding/kripke-ctl/kripke"
)

func main() {
    g, initial := kripke.BuildPurpleGraph()

    fmt.Println("PURPLE model (diagram graph) built with", len(g.States), "states.")
    fmt.Println("Initial state:", initial)

    fmt.Println()
    fmt.Println("Mermaid stateDiagram-v2:")
    fmt.Println("```mermaid")
    if err := kripke.WriteMermaidStateDiagram(g, initial, os.Stdout); err != nil {
        fmt.Fprintln(os.Stderr, "error writing Mermaid:", err)
    }
    fmt.Println("```")
}
