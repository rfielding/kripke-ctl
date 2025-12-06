package main

import (
    "bytes"
    "flag"
    "fmt"
    "os"
    "path/filepath"
    "strings"

    "github.com/rfielding/kripke-ctl/kripke"
    "github.com/rfielding/kripke-ctl/models/mm1"
    "github.com/rfielding/kripke-ctl/models/purple"
)

var nameFlag = flag.String("name", "purple", "model name (purple or mm1)")

func main() {
    flag.Parse()
    modelName := *nameFlag

    spec, err := loadModelSpec(modelName)
    if err != nil {
        fmt.Fprintln(os.Stderr, "load model:", err)
        os.Exit(1)
    }

    graph, initial := spec.BuildGraph()

    var buf bytes.Buffer
    buf.WriteString("```mermaid\n")

    if err := kripke.WriteMermaidStateDiagram(graph, initial, &buf); err != nil {
        fmt.Fprintln(os.Stderr, "write mermaid:", err)
        os.Exit(1)
    }

    buf.WriteString("```\n")

    if err := injectDiagramMarkdown(spec.Name(), buf.String()); err != nil {
        fmt.Fprintln(os.Stderr, "update markdown:", err)
        os.Exit(1)
    }

    fmt.Println("Updated diagrams in docs/models/" + spec.Name() + ".md")
}

func loadModelSpec(name string) (kripke.ModelSpec, error) {
    switch name {
    case "purple":
        return purple.PurpleModel{}, nil
    case "mm1":
        return mm1.Model{}, nil
    }
    return nil, fmt.Errorf("unknown model name: %s", name)
}

func injectDiagramMarkdown(name, mermaid string) error {
    path := filepath.Join("docs", "models", name+".md")
    b, err := os.ReadFile(path)
    if err != nil {
        return err
    }
    text := string(b)

    needle := "## Diagrams"
    idx := strings.Index(text, needle)
    if idx == -1 {
        return fmt.Errorf("no '## Diagrams' section in %s", path)
    }

    head := text[:idx+len(needle)]
    newText := head + "\n\n" + mermaid + "\n"

    return os.WriteFile(path, []byte(newText), 0o644)
}
