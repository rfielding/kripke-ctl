package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

const (
	defaultAddr    = ":8080"
	openAIEndpoint = "https://api.openai.com/v1/responses"
	openAIModel    = "gpt-4.1" // or gpt-4.1-mini if you prefer
)

// ChatRequest is what the browser sends to /api/chat.
type ChatRequest struct {
	Prompt  string    `json:"prompt"`
	History []Message `json:"history,omitempty"`
}

// Message is an input item for the Responses API.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// openAIRequest is the wire format for the Responses API (simplified).
type openAIRequest struct {
	Model string    `json:"model"`
	Input []Message `json:"input"`
}

func main() {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Println("ERROR: OPENAI_API_KEY is not set.")
		log.Println()
		log.Println("kripke-ctl docs server expects an OpenAI API key in the environment.")
		log.Println("To fix this:")
		log.Println("  1) Obtain an API key from https://platform.openai.com/")
		log.Println("  2) Export it in your shell, e.g.:")
		log.Println("       export OPENAI_API_KEY=sk-****************************")
		log.Println("  3) Re-run:")
		log.Println("       go run ./cmd/docs")
		os.Exit(1)
	}

	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("could not get working directory: %v", err)
	}

	docsDir := filepath.Join(cwd, "docs")
	webDir := filepath.Join(cwd, "web")

	if _, err := os.Stat(docsDir); err != nil {
		log.Fatalf("docs directory %q not found: %v", docsDir, err)
	}
	if _, err := os.Stat(webDir); err != nil {
		log.Fatalf("web directory %q not found: %v", webDir, err)
	}

	mux := http.NewServeMux()

	// Serve the static docs UI.
	mux.Handle("/", http.FileServer(http.Dir(webDir)))
	mux.Handle("/docs/", http.StripPrefix("/docs/", http.FileServer(http.Dir(docsDir))))

	// Health endpoint.
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok\n"))
	})

	// Chat proxy endpoint.
	mux.Handle("/api/chat", chatHandler(apiKey))

	addr := defaultAddr
	if v := os.Getenv("KRIPKE_HTTP_ADDR"); v != "" {
		addr = v
	}

	log.Printf("kripke-ctl docs server on http://%s\n", addr)
	log.Printf("Static docs: %s (mounted at /docs/)\n", docsDir)
	log.Printf("Web UI:      %s (mounted at /)\n", webDir)

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("ListenAndServe failed: %v", err)
	}
}

// chatHandler proxies chat-style requests to the OpenAI Responses API.
func chatHandler(apiKey string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "POST only", http.StatusMethodNotAllowed)
			return
		}

		var req ChatRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "bad json: "+err.Error(), http.StatusBadRequest)
			return
		}
		if req.Prompt == "" {
			http.Error(w, "missing prompt", http.StatusBadRequest)
			return
		}

		systemPrompt := `
You are the assistant for the kripke-ctl project.

- You model systems as communicating state machines and Markov decision processes.
- You understand CTL (EF, EG, AF, AG, EX, AX, EU).
- You can emit Markdown that may include Mermaid diagrams and/or references to generated images.
- Prefer explicit structures and diagrams over vague prose.
- When emitting Mermaid, clearly mark which code blocks are Mermaid so a renderer can detect them.
`

		// Build the Responses API request.
		input := []Message{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: req.Prompt},
		}

		// Optionally include prior turns if we wire History from the UI in future.
		for _, m := range req.History {
			input = append(input, m)
		}

		oaReq := openAIRequest{
			Model: openAIModel,
			Input: input,
		}

		body, err := json.Marshal(oaReq)
		if err != nil {
			http.Error(w, "marshal error: "+err.Error(), http.StatusInternalServerError)
			return
		}

		httpReq, err := http.NewRequest(http.MethodPost, openAIEndpoint, bytes.NewReader(body))
		if err != nil {
			http.Error(w, "request build error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		httpReq.Header.Set("Authorization", "Bearer "+apiKey)
		httpReq.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(httpReq)
		if err != nil {
			http.Error(w, "upstream error: "+err.Error(), http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(resp.StatusCode)
		_, _ = io.Copy(w, resp.Body)
	})
}

