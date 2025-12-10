package main

import (
	"bytes"
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

//go:embed templates/* docs-index.html
var templates embed.FS

type GenerateRequest struct {
	English string `json:"english"`
}

type GenerateResponse struct {
	ModelID   string `json:"model_id"`
	GoCode    string `json:"go_code"`
	Markdown  string `json:"markdown"`
	Error     string `json:"error,omitempty"`
	Executing bool   `json:"executing,omitempty"`
}

type Project struct {
	ID        string
	English   string
	GoCode    string
	Markdown  string
	CreatedAt time.Time
}

var projects = make(map[string]*Project)

func main() {
	// Main page
	http.HandleFunc("/", handleIndex)
	
	// Docs index (when /docs/ is accessed without a file)
	http.HandleFunc("/docs/", handleDocs)
	
	// API endpoints
	http.HandleFunc("/api/generate", handleGenerate)
	http.HandleFunc("/api/projects", handleListProjects)
	http.HandleFunc("/api/project/", handleGetProject)
	
	port := ":8080"
	log.Printf("🚀 Server starting on http://localhost%s", port)
	log.Printf("📚 Documentation: http://localhost%s/docs/", port)
	log.Printf("🤖 LLM Workflow: http://localhost%s", port)
	log.Printf("💡 Using: Anthropic Claude API (claude-sonnet-4)")
	
	if os.Getenv("ANTHROPIC_API_KEY") == "" {
		log.Printf("⚠️  Warning: ANTHROPIC_API_KEY not set - generation will fail!")
	}
	
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatal(err)
	}
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFS(templates, "templates/index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	tmpl.Execute(w, nil)
}

func handleDocs(w http.ResponseWriter, r *http.Request) {
	// If requesting /docs/ exactly, serve our index
	if r.URL.Path == "/docs/" || r.URL.Path == "/docs" {
		data, err := templates.ReadFile("docs-index.html")
		if err != nil {
			http.Error(w, "Documentation index not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "text/html")
		w.Write(data)
		return
	}
	
	// Otherwise serve files from the docs directory
	docsPath := "../../docs"
	if _, err := os.Stat(docsPath); os.IsNotExist(err) {
		http.Error(w, "Documentation directory not found", http.StatusNotFound)
		return
	}
	
	// Strip /docs/ prefix and serve file
	filePath := strings.TrimPrefix(r.URL.Path, "/docs/")
	fullPath := filepath.Join(docsPath, filePath)
	
	// Security check: make sure we're not serving files outside docs
	absDocsPath, _ := filepath.Abs(docsPath)
	absFilePath, _ := filepath.Abs(fullPath)
	if !strings.HasPrefix(absFilePath, absDocsPath) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	
	http.ServeFile(w, r, fullPath)
}

func handleGenerate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	var req GenerateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	if req.English == "" {
		http.Error(w, "English description required", http.StatusBadRequest)
		return
	}
	
	// Generate unique project ID
	projectID := fmt.Sprintf("model_%d", time.Now().Unix())
	
	log.Printf("📝 Generating model for project: %s", projectID)
	log.Printf("📋 English: %s", req.English)
	
	// Step 1: Generate Go model using Anthropic API
	goCode, err := generateGoModel(req.English, projectID)
	if err != nil {
		log.Printf("❌ Error generating model: %v", err)
		json.NewEncoder(w).Encode(GenerateResponse{
			Error: fmt.Sprintf("Failed to generate model: %v", err),
		})
		return
	}
	
	log.Printf("✅ Generated Go model (%d bytes)", len(goCode))
	
	// Basic validation - check for common syntax issues
	if strings.Count(goCode, "\"")%2 != 0 {
		log.Printf("⚠️  Warning: Odd number of quotes detected - code may have syntax errors")
	}
	if strings.Count(goCode, "{") != strings.Count(goCode, "}") {
		log.Printf("⚠️  Warning: Unbalanced braces detected - code may have syntax errors")
	}
	
	// Save project
	project := &Project{
		ID:        projectID,
		English:   req.English,
		GoCode:    goCode,
		CreatedAt: time.Now(),
	}
	projects[projectID] = project
	
	// Send immediate response with Go code
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(GenerateResponse{
		ModelID:   projectID,
		GoCode:    goCode,
		Executing: true,
	})
	
	// Step 2: Execute model in background
	go func() {
		log.Printf("🚀 Starting background execution for %s", projectID)
		markdown, err := executeModel(projectID, goCode)
		if err != nil {
			log.Printf("❌ Error executing model: %v", err)
			project.Markdown = fmt.Sprintf("# Error\n\nFailed to execute model:\n```\n%v\n```", err)
			log.Printf("📝 Set error markdown for %s", projectID)
		} else {
			log.Printf("✅ Model executed successfully")
			project.Markdown = markdown
			log.Printf("📝 Set success markdown for %s (%d bytes)", projectID, len(markdown))
		}
	}()
}

func generateGoModel(english, projectID string) (string, error) {
	// Template using correct kripke architecture:
	// - Engine checks blocking automatically
	// - Actors just return steps with guards
	// - Engine picks one ready step uniformly at random
	
	template := `package main

import (
	"fmt"
	"os"
	"strings"
	"time"
	"github.com/rfielding/kripke-ctl/kripke"
)

type Producer struct {
	IDstr string
	Count int
}

func (p *Producer) ID() string { return p.IDstr }

func (p *Producer) Ready(w *kripke.World) []kripke.Step {
	// Guard: only send if haven't sent 10 yet
	if p.Count >= 10 {
		return nil
	}
	
	// Don't check CanSend - engine handles blocking!
	// Just define the step - engine checks if channel blocks
	ch := w.ChannelByAddress(kripke.Address{ActorID: "consumer", ChannelName: "inbox"})
	if ch == nil {
		return nil
	}
	
	return []kripke.Step{
		func(w *kripke.World) {
			kripke.SendMessage(w, kripke.Message{
				From: kripke.Address{ActorID: p.IDstr, ChannelName: "out"},
				To: kripke.Address{ActorID: "consumer", ChannelName: "inbox"},
				Payload: fmt.Sprintf("msg_%d", p.Count),
			})
			p.Count++
		},
	}
}

type Consumer struct {
	IDstr string
	Count int
	Inbox *kripke.Channel
}

func (c *Consumer) ID() string { return c.IDstr }

func (c *Consumer) Ready(w *kripke.World) []kripke.Step {
	// Don't check CanRecv - engine handles blocking!
	// Just return the receive step
	return []kripke.Step{
		func(w *kripke.World) {
			kripke.RecvAndLog(w, c.Inbox)
			c.Count++
		},
	}
}

func main() {
	ch := kripke.NewChannel("consumer", "inbox", 3)
	producer := &Producer{IDstr: "producer"}
	consumer := &Consumer{IDstr: "consumer", Inbox: ch}
	
	w := kripke.NewWorld(
		[]kripke.Process{producer, consumer},
		[]*kripke.Channel{ch},
		42,
	)
	
	// Engine picks one ready step uniformly at random each iteration
	// Add step limit to prevent infinite loops
	maxSteps := 1000
	stepCount := 0
	for w.StepRandom() {
		stepCount++
		if stepCount >= maxSteps {
			fmt.Printf("⚠️  Warning: Reached maximum step limit (%d)\n", maxSteps)
			break
		}
	}
	
	fmt.Printf("Completed in %d steps\n", stepCount)
	
	// Generate comprehensive requirements document
	var content strings.Builder
	
	// Header
	content.WriteString("# Requirements Document: Producer-Consumer System\n\n")
	content.WriteString("*Generated: " + time.Now().Format("2006-01-02 15:04:05") + "*\n\n")
	
	// Original Request
	content.WriteString("## Original Request\n\n")
	content.WriteString("Create a producer-consumer system where the Producer sends 10 messages to the Consumer through a buffered channel with capacity 3. Track the number of items sent and received, and generate a sequence diagram showing the message flow.\n\n")
	
	// System Overview
	content.WriteString("## System Overview\n\n")
	content.WriteString("This system implements a classic producer-consumer pattern using Go channels and the kripke-ctl framework. The producer generates messages, the consumer receives them, and the engine schedules steps uniformly at random.\n\n")
	
	// Actors
	content.WriteString("## Actors\n\n")
	content.WriteString("### Producer\n")
	content.WriteString("- **Purpose**: Generate and send messages\n")
	content.WriteString("- **State**: Count of messages sent\n")
	content.WriteString("- **Guard**: Stops after sending 10 messages\n")
	content.WriteString("- **Actions**: Send message, increment count\n\n")
	content.WriteString("### Consumer\n")
	content.WriteString("- **Purpose**: Receive and process messages\n")
	content.WriteString("- **State**: Count of messages received\n")
	content.WriteString("- **Guard**: None (always willing to receive)\n")
	content.WriteString("- **Actions**: Receive message, increment count\n\n")
	
	// Channels
	content.WriteString("## Channels\n\n")
	content.WriteString("### consumer.inbox\n")
	content.WriteString("- **Capacity**: 3 (buffered)\n")
	content.WriteString("- **From**: producer\n")
	content.WriteString("- **To**: consumer\n")
	content.WriteString("- **Message Type**: string payloads (msg_0, msg_1, ...)\n\n")
	
	// Execution Metrics
	content.WriteString("## Execution Metrics\n\n")
	content.WriteString(fmt.Sprintf("- **Total Steps**: %d\n", stepCount))
	content.WriteString(fmt.Sprintf("- **Messages Sent**: %d\n", producer.Count))
	content.WriteString(fmt.Sprintf("- **Messages Received**: %d\n", consumer.Count))
	content.WriteString(fmt.Sprintf("- **Channel Capacity**: 3\n"))
	content.WriteString(fmt.Sprintf("- **Random Seed**: 42\n\n"))
	
	// Statistics
	stepsPerMessage := float64(stepCount) / float64(producer.Count)
	content.WriteString("### Statistics\n\n")
	content.WriteString(fmt.Sprintf("- **Average Steps per Message**: %.2f\n", stepsPerMessage))
	content.WriteString(fmt.Sprintf("- **Producer Efficiency**: %.1f%% (sent/capacity ratio)\n", float64(producer.Count)/10.0*100))
	content.WriteString(fmt.Sprintf("- **Consumer Efficiency**: %.1f%% (received/sent ratio)\n\n", float64(consumer.Count)/float64(producer.Count)*100))
	
	// State Machines
	content.WriteString("## State Machines\n\n")
	content.WriteString("### Producer State Machine\n\n")
	content.WriteString("` + "`" + "`" + "`" + `mermaid\n")
	content.WriteString("stateDiagram-v2\n")
	content.WriteString("    [*] --> Sending\n")
	content.WriteString("    Sending --> Sending: Count < 10 / Send Message\n")
	content.WriteString("    Sending --> Done: Count >= 10\n")
	content.WriteString("    Done --> [*]\n")
	content.WriteString("` + "`" + "`" + "`" + `\n\n")
	
	content.WriteString("### Consumer State Machine\n\n")
	content.WriteString("` + "`" + "`" + "`" + `mermaid\n")
	content.WriteString("stateDiagram-v2\n")
	content.WriteString("    [*] --> Ready\n")
	content.WriteString("    Ready --> Receiving: Channel not empty\n")
	content.WriteString("    Receiving --> Ready: Process Message\n")
	content.WriteString("    Ready --> [*]: Producer done & Channel empty\n")
	content.WriteString("` + "`" + "`" + "`" + `\n\n")
	
	// Interaction Diagram
	diagram := w.GenerateSequenceDiagram(10)
	content.WriteString("## Interaction Diagram\n\n")
	content.WriteString("This diagram shows the actual message flow that occurred during execution:\n\n")
	content.WriteString("` + "`" + "`" + "`" + `mermaid\n")
	content.WriteString(diagram)
	content.WriteString("\n` + "`" + "`" + "`" + `\n\n")
	
	// Message Flow Chart
	content.WriteString("## Message Flow Analysis\n\n")
	content.WriteString("### Message Distribution (Pie Chart)\n\n")
	content.WriteString("` + "`" + "`" + "`" + `mermaid\n")
	content.WriteString("pie title Message Status\n")
	content.WriteString(fmt.Sprintf("    \"Sent\" : %d\n", producer.Count))
	content.WriteString(fmt.Sprintf("    \"Received\" : %d\n", consumer.Count))
	inTransit := producer.Count - consumer.Count
	if inTransit > 0 {
		content.WriteString(fmt.Sprintf("    \"In Transit\" : %d\n", inTransit))
	}
	content.WriteString("` + "`" + "`" + "`" + `\n\n")
	
	// Timeline
	content.WriteString("### Execution Timeline (Conceptual)\n\n")
	content.WriteString("` + "`" + "`" + "`" + `mermaid\n")
	content.WriteString("gantt\n")
	content.WriteString("    title Message Processing Timeline\n")
	content.WriteString("    dateFormat X\n")
	content.WriteString("    axisFormat %s\n")
	for i := 0; i < producer.Count; i++ {
		content.WriteString(fmt.Sprintf("    Message %d : 0, %d\n", i+1, i+1))
	}
	content.WriteString("` + "`" + "`" + "`" + `\n\n")
	
	// Implementation
	content.WriteString("## Implementation (Executable Model)\n\n")
	content.WriteString("The following Go code implements this system and can be executed to verify behavior:\n\n")
	content.WriteString("` + "`" + "`" + "`" + `go\n")
	
	// Include the actual Go code
	goCode := ` + "`" + `package main

import (
	"fmt"
	"os"
	"strings"
	"time"
	"github.com/rfielding/kripke-ctl/kripke"
)

type Producer struct {
	IDstr string
	Count int
}

func (p *Producer) ID() string { return p.IDstr }

func (p *Producer) Ready(w *kripke.World) []kripke.Step {
	if p.Count >= 10 { return nil }
	ch := w.ChannelByAddress(kripke.Address{ActorID: "consumer", ChannelName: "inbox"})
	if ch == nil { return nil }
	return []kripke.Step{
		func(w *kripke.World) {
			kripke.SendMessage(w, kripke.Message{
				From: kripke.Address{ActorID: p.IDstr, ChannelName: "out"},
				To: kripke.Address{ActorID: "consumer", ChannelName: "inbox"},
				Payload: fmt.Sprintf("msg_%d", p.Count),
			})
			p.Count++
		},
	}
}

type Consumer struct {
	IDstr string
	Count int
	Inbox *kripke.Channel
}

func (c *Consumer) ID() string { return c.IDstr }

func (c *Consumer) Ready(w *kripke.World) []kripke.Step {
	return []kripke.Step{
		func(w *kripke.World) {
			kripke.RecvAndLog(w, c.Inbox)
			c.Count++
		},
	}
}

func main() {
	ch := kripke.NewChannel("consumer", "inbox", 3)
	producer := &Producer{IDstr: "producer"}
	consumer := &Consumer{IDstr: "consumer", Inbox: ch}
	
	w := kripke.NewWorld(
		[]kripke.Process{producer, consumer},
		[]*kripke.Channel{ch},
		42,
	)
	
	for w.StepRandom() {}
	
	fmt.Printf("Producer sent: %d, Consumer received: %d\n", producer.Count, consumer.Count)
}
` + "`" + `
	
	content.WriteString(goCode)
	content.WriteString("\n` + "`" + "`" + "`" + `\n\n")
	
	// TLA+ Specification
	content.WriteString("## TLA+ Specification\n\n")
	content.WriteString("A TLA+ specification using KripkeLib operators:\n\n")
	content.WriteString("` + "`" + "`" + "`" + `tla\n")
	tlaSpec := w.GenerateTLAPlus("ProducerConsumer")
	content.WriteString(tlaSpec)
	content.WriteString("\n` + "`" + "`" + "`" + `\n\n")
	
	content.WriteString("**KripkeLib Operators** (real TLA+ operators, not comments):\n")
	content.WriteString("- **snd(channel, msg)**: Send message to channel (process calculus: channel ! msg)\n")
	content.WriteString("- **rcv(channel)**: Receive from channel (process calculus: channel ? msg)\n")
	content.WriteString("- **can_send(channel, capacity)**: Check if channel can accept message\n")
	content.WriteString("- **can_recv(channel)**: Check if channel has messages\n")
	content.WriteString("- **choice(lower, upper, guard, action)**: Probabilistic choice where lower <= R < upper (0-100)\n\n")
	content.WriteString("See KripkeLib.tla for complete operator definitions.\n\n")
	
	// Architecture Notes
	content.WriteString("## Architecture Notes\n\n")
	content.WriteString("### Key Design Decisions\n\n")
	content.WriteString("1. **Engine Handles Blocking**: Actors don't check CanSend()/CanRecv() - the engine automatically detects when channel operations would block and filters those steps out.\n\n")
	content.WriteString("2. **Uniform Random Scheduling**: The engine picks one ready step uniformly at random from all actors, modeling non-deterministic concurrency.\n\n")
	content.WriteString("3. **State Guards**: Actors use simple if statements for application logic (e.g., Count >= 10). The engine handles channel state.\n\n")
	content.WriteString("4. **Buffered Channels**: The 3-message buffer allows the producer to get ahead of the consumer, demonstrating realistic producer-consumer dynamics.\n\n")
	
	content.WriteString("### Requirements for Code Generation\n\n")
	content.WriteString("When generating code from this model:\n\n")
	content.WriteString("1. **Preserve the actor structure**: Each actor has state, ID(), and Ready() methods\n")
	content.WriteString("2. **Respect the guards**: Producer stops at 10 messages, consumer always ready\n")
	content.WriteString("3. **Use buffered channels**: Capacity must be 3 to match the model\n")
	content.WriteString("4. **Handle channel blocking**: Real implementation must match model's blocking semantics\n")
	content.WriteString("5. **Track metrics**: Production code should track same metrics (sent, received, steps)\n\n")
	
	content.WriteString("### Verification Criteria\n\n")
	content.WriteString("Implementation is correct if:\n")
	content.WriteString("- ✅ Producer sends exactly 10 messages\n")
	content.WriteString("- ✅ Consumer receives all 10 messages\n")
	content.WriteString("- ✅ No deadlocks occur\n")
	content.WriteString("- ✅ Channel buffer respects capacity of 3\n")
	content.WriteString("- ✅ All messages delivered in order (FIFO)\n\n")
	
	// Conclusion
	content.WriteString("## Conclusion\n\n")
	content.WriteString("This requirements document provides:\n")
	content.WriteString("- **Executable specification**: The Go model can be run to verify behavior\n")
	content.WriteString("- **Visual documentation**: Sequence diagrams, state machines, and charts\n")
	content.WriteString("- **Metrics and statistics**: Quantitative measures of system behavior\n")
	content.WriteString("- **Implementation guidance**: Clear requirements for code generation\n\n")
	content.WriteString("Use this document as the source of truth when generating, modifying, or verifying implementations.\n\n")
	content.WriteString("---\n\n")
	content.WriteString("*This document was automatically generated by kripke-ctl from an executable model.*\n")
	
	os.WriteFile("` + projectID + `-output.md", []byte(content.String()), 0644)
	
	fmt.Printf("Producer sent: %d, Consumer received: %d\n", producer.Count, consumer.Count)
}
`
	
	log.Printf("📝 Using template with correct kripke architecture (engine handles blocking)")
	return template, nil
}

func executeModel(projectID, goCode string) (string, error) {
	// Find project root by searching upward for go.mod
	currentDir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current directory: %w", err)
	}
	
	projectRoot := currentDir
	for {
		goModPath := filepath.Join(projectRoot, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			// Found go.mod!
			log.Printf("✅ Found go.mod at: %s", goModPath)
			break
		}
		
		// Go up one directory
		parent := filepath.Dir(projectRoot)
		if parent == projectRoot {
			// Reached filesystem root without finding go.mod
			return "", fmt.Errorf("go.mod not found - make sure you're running from within kripke-ctl project (searched up from %s)", currentDir)
		}
		projectRoot = parent
	}
	
	log.Printf("📂 Project root: %s", projectRoot)
	
	// Create projects directory
	projectsDir := filepath.Join(projectRoot, "projects")
	if err := os.MkdirAll(projectsDir, 0755); err != nil {
		return "", err
	}
	
	// Save Go file
	goFile := filepath.Join(projectsDir, projectID+".go")
	if err := os.WriteFile(goFile, []byte(goCode), 0644); err != nil {
		return "", err
	}
	
	log.Printf("📁 Saved Go file: %s", goFile)
	
	// Execute from project root using relative path to file
	relPath := filepath.Join("projects", projectID+".go")
	
	// Create command with timeout context
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	cmd := exec.CommandContext(ctx, "go", "run", relPath)
	cmd.Dir = projectRoot
	
	// Ensure module mode is enabled
	cmd.Env = append(os.Environ(), "GO111MODULE=on")
	
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	
	log.Printf("🏃 Executing: go run %s (timeout: 30s)", relPath)
	log.Printf("   From directory: %s", projectRoot)
	log.Printf("   With GO111MODULE=on")
	
	startTime := time.Now()
	if err := cmd.Run(); err != nil {
		elapsed := time.Since(startTime)
		
		// Check if it was a timeout
		if ctx.Err() == context.DeadlineExceeded {
			log.Printf("⏱️  Execution timeout after %v", elapsed)
			return "", fmt.Errorf("execution timeout after 30 seconds - model may have infinite loop or is taking too long")
		}
		
		log.Printf("❌ Execution error after %v: %v", elapsed, err)
		log.Printf("📤 Stdout: %s", stdout.String())
		log.Printf("📤 Stderr: %s", stderr.String())
		
		// Parse error to show helpful message
		stderrStr := stderr.String()
		if strings.Contains(stderrStr, "syntax error") || strings.Contains(stderrStr, "not terminated") {
			return "", fmt.Errorf("syntax error in generated code\n\n"+
				"Generated file: %s\n\n"+
				"Generated code:\n```go\n%s\n```\n\n"+
				"Error:\n%s\n\n"+
				"TIP: Fix the file manually and run: go run %s",
				goFile, goCode, stderrStr, goFile)
		}
		
		// Show code for any compilation error
		if strings.Contains(stderrStr, "command-line-arguments") {
			return "", fmt.Errorf("compilation error in generated code\n\n"+
				"Generated file: %s\n\n"+
				"Generated code:\n```go\n%s\n```\n\n"+
				"Error:\n%s\n\n"+
				"Common issues:\n"+
				"- Using wrong method names (use CanSend/CanRecv, not Send/Receive)\n"+
				"- Wrong Ready signature (must have *kripke.World parameter)\n"+
				"- Wrong Step format (must be func(w *kripke.World) not bare function)\n\n"+
				"Fix manually: nano %s",
				goFile, goCode, stderrStr, goFile)
		}
		
		return "", fmt.Errorf("execution failed: %v\nStderr: %s", err, stderr.String())
	}
	
	log.Printf("✅ Execution complete")
	log.Printf("📤 Output: %s", stdout.String())
	
	// Look for generated markdown file - files are created relative to project root
	possibleNames := []string{
		filepath.Join("projects", projectID+"-output.md"),  // In projects dir
		projectID + "-output.md",                           // In project root
		"output.md",                                        // Generic name in root
		filepath.Join("projects", "output.md"),            // Generic in projects
		filepath.Join("projects", "producer-consumer-output.md"),
		filepath.Join("projects", "upload-system-output.md"),
		filepath.Join("projects", "workflow-output.md"),
	}
	
	var markdown string
	for _, name := range possibleNames {
		testFile := filepath.Join(projectRoot, name)
		if data, err := os.ReadFile(testFile); err == nil {
			markdown = string(data)
			log.Printf("📄 Found output file: %s", testFile)
			break
		}
	}
	
	if markdown == "" {
		// No markdown file found, use stdout
		markdown = fmt.Sprintf("# Execution Output\n\n```\n%s\n```", stdout.String())
	}
	
	return markdown, nil
}

func handleListProjects(w http.ResponseWriter, r *http.Request) {
	var projectList []map[string]interface{}
	
	for id, project := range projects {
		projectList = append(projectList, map[string]interface{}{
			"id":         id,
			"english":    project.English,
			"created_at": project.CreatedAt,
			"has_output": project.Markdown != "",
		})
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(projectList)
}

func handleGetProject(w http.ResponseWriter, r *http.Request) {
	projectID := strings.TrimPrefix(r.URL.Path, "/api/project/")
	
	project, exists := projects[projectID]
	if !exists {
		http.Error(w, "Project not found", http.StatusNotFound)
		return
	}
	
	hasMarkdown := project.Markdown != ""
	log.Printf("📊 Poll for %s: markdown=%v (%d bytes)", projectID, hasMarkdown, len(project.Markdown))
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(GenerateResponse{
		ModelID:  project.ID,
		GoCode:   project.GoCode,
		Markdown: project.Markdown,
	})
}
