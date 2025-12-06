
package main

import (
    "log"
    "net/http"
    "os"
    "path/filepath"
)

func main() {
    cwd, err := os.Getwd()
    if err != nil {
        log.Fatal(err)
    }

    docsDir := filepath.Join(cwd, "docs")
    webDir := filepath.Join(cwd, "web")

    mux := http.NewServeMux()
    mux.Handle("/docs/", http.StripPrefix("/docs/", http.FileServer(http.Dir(docsDir))))
    mux.Handle("/", http.FileServer(http.Dir(webDir)))

    addr := ":8080"
    log.Printf("kripke-ctl docs server on http://%s\n", addr)
    log.Printf("Static docs: %s (mounted at /docs/)\n", docsDir)
    log.Printf("Web UI:      %s (mounted at /)\n", webDir)

    if err := http.ListenAndServe(addr, mux); err != nil {
        log.Fatal(err)
    }
}
