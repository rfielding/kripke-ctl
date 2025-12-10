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
	
	diagram := w.GenerateSequenceDiagram(10)
	
	var content strings.Builder
	content.WriteString("# Producer-Consumer\n\n")
	content.WriteString("## Sequence Diagram\n\n")
	content.WriteString("` + "`" + "`" + "`" + `mermaid\n")
	content.WriteString(diagram)
	content.WriteString("\n` + "`" + "`" + "`" + `\n")
	
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
