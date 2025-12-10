package main

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

//go:embed templates/*
var templates embed.FS

//go:embed static/*
var static embed.FS

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
	// Serve static files (CSS, JS)
	http.Handle("/static/", http.FileServer(http.FS(static)))
	
	// Serve documentation from ./docs
	docsFS := http.Dir("../../docs")
	http.Handle("/docs/", http.StripPrefix("/docs/", http.FileServer(docsFS)))
	
	// Main page
	http.HandleFunc("/", handleIndex)
	
	// API endpoints
	http.HandleFunc("/api/generate", handleGenerate)
	http.HandleFunc("/api/projects", handleListProjects)
	http.HandleFunc("/api/project/", handleGetProject)
	
	port := ":8080"
	log.Printf("🚀 Server starting on http://localhost%s", port)
	log.Printf("📚 Documentation: http://localhost%s/docs/", port)
	log.Printf("🤖 LLM Workflow: http://localhost%s", port)
	
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
	goCode, err := generateGoModel(req.English)
	if err != nil {
		log.Printf("❌ Error generating model: %v", err)
		json.NewEncoder(w).Encode(GenerateResponse{
			Error: fmt.Sprintf("Failed to generate model: %v", err),
		})
		return
	}
	
	log.Printf("✅ Generated Go model (%d bytes)", len(goCode))
	
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
		markdown, err := executeModel(projectID, goCode)
		if err != nil {
			log.Printf("❌ Error executing model: %v", err)
			project.Markdown = fmt.Sprintf("# Error\n\nFailed to execute model:\n```\n%v\n```", err)
		} else {
			log.Printf("✅ Model executed successfully")
			project.Markdown = markdown
		}
	}()
}

func generateGoModel(english string) (string, error) {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("ANTHROPIC_API_KEY environment variable not set")
	}
	
	prompt := fmt.Sprintf(`Generate a complete kripke-ctl model in Go based on this English description:

"%s"

Requirements:
1. Import "github.com/rfielding/kripke-ctl/kripke"
2. Define actors as structs with state variables
3. Implement Ready() method with guards and actions
4. Include main() that:
   - Creates actors and channels
   - Runs the system with World
   - Generates markdown output with diagrams
   - Uses library methods: GenerateSequenceDiagram(), GenerateActorStateTable(), etc.
5. Write output to a markdown file
6. Include proper error handling

Generate ONLY the Go code, no explanation. Make it complete and runnable.`, english)
	
	requestBody := map[string]interface{}{
		"model": "claude-sonnet-4-20250514",
		"max_tokens": 4000,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
	}
	
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", err
	}
	
	req, err := http.NewRequest("POST", "https://api.anthropic.com/v1/messages", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")
	
	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}
	
	var result struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	
	if len(result.Content) == 0 {
		return "", fmt.Errorf("no content in response")
	}
	
	goCode := result.Content[0].Text
	
	// Extract code from markdown if present
	if strings.Contains(goCode, "```go") {
		parts := strings.Split(goCode, "```go")
		if len(parts) > 1 {
			parts = strings.Split(parts[1], "```")
			goCode = strings.TrimSpace(parts[0])
		}
	} else if strings.Contains(goCode, "```") {
		parts := strings.Split(goCode, "```")
		if len(parts) > 2 {
			goCode = strings.TrimSpace(parts[1])
		}
	}
	
	return goCode, nil
}

func executeModel(projectID, goCode string) (string, error) {
	// Create projects directory
	projectsDir := "../../projects"
	if err := os.MkdirAll(projectsDir, 0755); err != nil {
		return "", err
	}
	
	// Save Go file
	goFile := filepath.Join(projectsDir, projectID+".go")
	if err := os.WriteFile(goFile, []byte(goCode), 0644); err != nil {
		return "", err
	}
	
	log.Printf("📁 Saved Go file: %s", goFile)
	
	// Execute
	cmd := exec.Command("go", "run", goFile)
	cmd.Dir = projectsDir
	
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	
	log.Printf("🏃 Executing: go run %s", goFile)
	
	if err := cmd.Run(); err != nil {
		log.Printf("❌ Execution error: %v", err)
		log.Printf("📤 Stdout: %s", stdout.String())
		log.Printf("📤 Stderr: %s", stderr.String())
		return "", fmt.Errorf("execution failed: %v\nStderr: %s", err, stderr.String())
	}
	
	log.Printf("✅ Execution complete")
	log.Printf("📤 Output: %s", stdout.String())
	
	// Look for generated markdown file
	mdFile := filepath.Join(projectsDir, projectID+"-output.md")
	
	// Try common output file names
	possibleNames := []string{
		projectID + "-output.md",
		"output.md",
		"producer-consumer-output.md",
		"upload-system-output.md",
		"workflow-output.md",
	}
	
	var markdown string
	for _, name := range possibleNames {
		testFile := filepath.Join(projectsDir, name)
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
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(GenerateResponse{
		ModelID:  project.ID,
		GoCode:   project.GoCode,
		Markdown: project.Markdown,
	})
}
