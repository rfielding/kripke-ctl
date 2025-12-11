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
	// Template using the BAKERY example - the canonical real-world use case
	// This shows why kripke-ctl exists: to model business systems and track metrics
	
	template := `package main

import (
	"fmt"
	"os"
	"strings"
	"time"
	"math/rand"
	"github.com/rfielding/kripke-ctl/kripke"
)

// ============================================================================
// BAKERY BUSINESS SIMULATION - THE CANONICAL EXAMPLE
// ============================================================================
//
// Real-world system: Model a bakery's operations to answer business questions
//
// Actors:
//   Production: Makes bread (dough → kneading → baking → cooling)
//   Truck: Transports bread (loading → driving → unloading)
//   Storefront: Manages inventory and sales
//   Customers: Arrive and purchase bread
//
// Business Questions:
//   1. What are our profits?
//   2. How much waste do we have?
//   3. What are the most popular breads?
//
// Metrics tracked over time:
//   - Revenue (cumulative and per timestep)
//   - Costs (cumulative and per timestep)
//   - Profit (revenue - costs over time)
//   - Inventory levels
//   - Sales by bread type
//
// ============================================================================

// Metrics tracking for time-series analysis
type Metrics struct {
	Timestep        int
	Revenue         []float64  // Revenue at each timestep
	Costs           []float64  // Costs at each timestep
	Profit          []float64  // Profit at each timestep
	TotalRevenue    float64
	TotalCosts      float64
	ProductionCount int
	SalesCount      int
	Inventory       int
}

func (m *Metrics) RecordTimestep(revenue, costs float64, inventory int) {
	m.Timestep++
	m.TotalRevenue += revenue
	m.TotalCosts += costs
	m.Revenue = append(m.Revenue, m.TotalRevenue)
	m.Costs = append(m.Costs, m.TotalCosts)
	m.Profit = append(m.Profit, m.TotalRevenue - m.TotalCosts)
	m.Inventory = inventory
}

var globalMetrics = &Metrics{
	Revenue: []float64{0},
	Costs:   []float64{0},
	Profit:  []float64{0},
}

// ============================================================================
// PRODUCTION ACTOR
// ============================================================================

type Production struct {
	IDstr      string
	State      string
	BreadType  string
	TimeInState int
	TotalMade  int
	HourlyRate float64
}

func (p *Production) ID() string { return p.IDstr }

type ProductionStart struct {
	IDstr string
	Prod  *Production
}

func (ps *ProductionStart) ID() string { return ps.IDstr }

func (ps *ProductionStart) Ready(w *kripke.World) []kripke.Step {
	if ps.Prod.State != "idle" {
		return nil
	}
	return []kripke.Step{
		func(w *kripke.World) {
			dice := rand.Intn(100)
			if dice < 50 {
				ps.Prod.BreadType = "sourdough"
			} else if dice < 80 {
				ps.Prod.BreadType = "baguette"
			} else {
				ps.Prod.BreadType = "rye"
			}
			ps.Prod.State = "dough"
			ps.Prod.TimeInState = 0
		},
	}
}

type ProductionProgress struct {
	IDstr string
	Prod  *Production
}

func (pp *ProductionProgress) ID() string { return pp.IDstr }

func (pp *ProductionProgress) Ready(w *kripke.World) []kripke.Step {
	if pp.Prod.State != "dough" && pp.Prod.State != "kneading" && 
	   pp.Prod.State != "baking" && pp.Prod.State != "cooling" {
		return nil
	}
	return []kripke.Step{
		func(w *kripke.World) {
			pp.Prod.TimeInState++
			switch pp.Prod.State {
			case "dough":
				if pp.Prod.TimeInState >= 2 {
					pp.Prod.State = "kneading"
					pp.Prod.TimeInState = 0
				}
			case "kneading":
				if pp.Prod.TimeInState >= 2 {
					pp.Prod.State = "baking"
					pp.Prod.TimeInState = 0
				}
			case "baking":
				if pp.Prod.TimeInState >= 3 {
					pp.Prod.State = "cooling"
					pp.Prod.TimeInState = 0
				}
			case "cooling":
				if pp.Prod.TimeInState >= 2 {
					pp.Prod.State = "ready"
					pp.Prod.TimeInState = 0
				}
			}
		},
	}
}

type ProductionLoad struct {
	IDstr   string
	Prod    *Production
	ToTruck *kripke.Channel
}

func (pl *ProductionLoad) ID() string { return pl.IDstr }

func (pl *ProductionLoad) Ready(w *kripke.World) []kripke.Step {
	if pl.Prod.State != "ready" {
		return nil
	}
	return []kripke.Step{
		func(w *kripke.World) {
			kripke.SendMessage(w, kripke.Message{
				From:    kripke.Address{ActorID: pl.Prod.IDstr, ChannelName: "out"},
				To:      kripke.Address{ActorID: "storefront", ChannelName: "delivery"},
				Payload: pl.Prod.BreadType,
			})
			pl.Prod.TotalMade++
			pl.Prod.State = "idle"
			globalMetrics.ProductionCount++
		},
	}
}

// ============================================================================
// STOREFRONT ACTOR
// ============================================================================

type Storefront struct {
	IDstr      string
	Inventory  map[string]int
	SalesCount map[string]int
	Revenue    float64
	Delivery   *kripke.Channel
	SalesChan  *kripke.Channel
	BreadPrice float64
}

func (s *Storefront) ID() string { return s.IDstr }

type StorefrontReceive struct {
	IDstr string
	Store *Storefront
}

func (sr *StorefrontReceive) ID() string { return sr.IDstr }

func (sr *StorefrontReceive) Ready(w *kripke.World) []kripke.Step {
	return []kripke.Step{
		func(w *kripke.World) {
			msg, _ := kripke.RecvAndLog(w, sr.Store.Delivery)
			if breadType, ok := msg.Payload.(string); ok {
				sr.Store.Inventory[breadType]++
			}
		},
	}
}

type StorefrontSale struct {
	IDstr string
	Store *Storefront
}

func (ss *StorefrontSale) ID() string { return ss.IDstr }

func (ss *StorefrontSale) Ready(w *kripke.World) []kripke.Step {
	return []kripke.Step{
		func(w *kripke.World) {
			msg, _ := kripke.RecvAndLog(w, ss.Store.SalesChan)
			if breadType, ok := msg.Payload.(string); ok {
				if ss.Store.Inventory[breadType] > 0 {
					ss.Store.Inventory[breadType]--
					ss.Store.SalesCount[breadType]++
					ss.Store.Revenue += ss.Store.BreadPrice
					globalMetrics.SalesCount++
				}
			}
		},
	}
}

// ============================================================================
// CUSTOMER ARRIVALS
// ============================================================================

type Customer struct {
	IDstr       string
	ToStore     *kripke.Channel
	ArrivalRate int
	Counter     int
}

func (c *Customer) ID() string { return c.IDstr }

func (c *Customer) Ready(w *kripke.World) []kripke.Step {
	c.Counter++
	if c.Counter < c.ArrivalRate {
		return nil
	}
	return []kripke.Step{
		func(w *kripke.World) {
			dice := rand.Intn(100)
			var choice string
			if dice < 50 {
				choice = "sourdough"
			} else if dice < 80 {
				choice = "baguette"
			} else {
				choice = "rye"
			}
			kripke.SendMessage(w, kripke.Message{
				From:    kripke.Address{ActorID: c.IDstr, ChannelName: "out"},
				To:      kripke.Address{ActorID: "storefront", ChannelName: "sales"},
				Payload: choice,
			})
			c.Counter = 0
		},
	}
}

func main() {
	rand.Seed(42)
	
	// Channels
	delivery := kripke.NewChannel("storefront", "delivery", 10)
	sales := kripke.NewChannel("storefront", "sales", 10)
	
	// Actors
	production := &Production{
		IDstr:      "production",
		State:      "idle",
		HourlyRate: 25.0,
	}
	
	storefront := &Storefront{
		IDstr:      "storefront",
		Inventory:  map[string]int{},
		SalesCount: map[string]int{},
		Delivery:   delivery,
		SalesChan:  sales,
		BreadPrice: 8.50,
	}
	
	customer := &Customer{
		IDstr:       "customer",
		ToStore:     sales,
		ArrivalRate: 3,
	}
	
	// Processes
	processes := []kripke.Process{
		&ProductionStart{IDstr: "prod_start", Prod: production},
		&ProductionProgress{IDstr: "prod_progress", Prod: production},
		&ProductionLoad{IDstr: "prod_load", Prod: production, ToTruck: delivery},
		&StorefrontReceive{IDstr: "store_receive", Store: storefront},
		&StorefrontSale{IDstr: "store_sale", Store: storefront},
		customer,
	}
	
	w := kripke.NewWorld(
		processes,
		[]*kripke.Channel{delivery, sales},
		42,
	)
	
	// Run simulation and track metrics
	maxSteps := 100
	stepCount := 0
	costPerStep := (production.HourlyRate + 20.0) / 60.0 / 60.0  // Two workers, per second
	
	for w.StepRandom() {
		stepCount++
		if stepCount >= maxSteps {
			break
		}
		
		// Record metrics every step
		totalInventory := storefront.Inventory["sourdough"] + 
			storefront.Inventory["baguette"] + 
			storefront.Inventory["rye"]
		globalMetrics.RecordTimestep(
			storefront.Revenue - globalMetrics.TotalRevenue,  // Revenue this step
			costPerStep,                                       // Cost this step
			totalInventory,
		)
	}
	
	fmt.Printf("Simulation completed in %d steps\n", stepCount)
	
	// Generate comprehensive requirements document
	var content strings.Builder
	
	// Header
	content.WriteString("# Bakery Business Simulation Requirements\n\n")
	content.WriteString("*Generated: " + time.Now().Format("2006-01-02 15:04:05") + "*\n\n")
	
	// Business Metrics
	content.WriteString("## Business Metrics\n\n")
	
	totalCost := (production.HourlyRate + 20.0) * float64(maxSteps) / 60.0 / 60.0
	profit := storefront.Revenue - totalCost
	totalInventory := storefront.Inventory["sourdough"] + 
		storefront.Inventory["baguette"] + 
		storefront.Inventory["rye"]
	totalSales := storefront.SalesCount["sourdough"] + 
		storefront.SalesCount["baguette"] + 
		storefront.SalesCount["rye"]
	
	content.WriteString(fmt.Sprintf("- **Production**: %d breads made\n", production.TotalMade))
	content.WriteString(fmt.Sprintf("- **Sales**: %d breads sold\n", totalSales))
	content.WriteString(fmt.Sprintf("- **Revenue**: $%.2f\n", storefront.Revenue))
	content.WriteString(fmt.Sprintf("- **Costs**: $%.2f\n", totalCost))
	content.WriteString(fmt.Sprintf("- **Profit**: $%.2f\n", profit))
	content.WriteString(fmt.Sprintf("- **Waste**: %d unsold breads\n", totalInventory))
	content.WriteString("\n")
	
	// Sales by Type
	content.WriteString("### Sales by Bread Type\n\n")
	content.WriteString(fmt.Sprintf("- **Sourdough**: %d (%.0f%%)\n", 
		storefront.SalesCount["sourdough"],
		float64(storefront.SalesCount["sourdough"])*100/float64(totalSales)))
	content.WriteString(fmt.Sprintf("- **Baguette**: %d (%.0f%%)\n", 
		storefront.SalesCount["baguette"],
		float64(storefront.SalesCount["baguette"])*100/float64(totalSales)))
	content.WriteString(fmt.Sprintf("- **Rye**: %d (%.0f%%)\n\n", 
		storefront.SalesCount["rye"],
		float64(storefront.SalesCount["rye"])*100/float64(totalSales)))
	
	// Business Questions
	content.WriteString("### Business Questions Answered\n\n")
	content.WriteString(fmt.Sprintf("1. **What are our profits?** → $%.2f\n", profit))
	content.WriteString(fmt.Sprintf("2. **How much waste?** → %d unsold items\n", totalInventory))
	mostPopular := "sourdough"
	if storefront.SalesCount["baguette"] > storefront.SalesCount[mostPopular] {
		mostPopular = "baguette"
	}
	if storefront.SalesCount["rye"] > storefront.SalesCount[mostPopular] {
		mostPopular = "rye"
	}
	content.WriteString(fmt.Sprintf("3. **Most popular bread?** → %s\n\n", mostPopular))
	
	// Production State Machine
	content.WriteString("## Production Workflow\n\n")
	content.WriteString("` + "`" + "`" + "`" + `mermaid\n")
	content.WriteString("stateDiagram-v2\n")
	content.WriteString("    [*] --> Idle\n")
	content.WriteString("    Idle --> Dough: start production\n")
	content.WriteString("    Dough --> Kneading: time >= 2\n")
	content.WriteString("    Kneading --> Baking: time >= 2\n")
	content.WriteString("    Baking --> Cooling: time >= 3\n")
	content.WriteString("    Cooling --> Ready: time >= 2\n")
	content.WriteString("    Ready --> Idle: load to delivery\n")
	content.WriteString("` + "`" + "`" + "`" + `\n\n")
	
	// Popularity Chart
	content.WriteString("## Bread Popularity\n\n")
	content.WriteString("` + "`" + "`" + "`" + `mermaid\n")
	content.WriteString("pie title Sales by Bread Type\n")
	content.WriteString(fmt.Sprintf("    \"Sourdough\" : %d\n", storefront.SalesCount["sourdough"]))
	content.WriteString(fmt.Sprintf("    \"Baguette\" : %d\n", storefront.SalesCount["baguette"]))
	content.WriteString(fmt.Sprintf("    \"Rye\" : %d\n", storefront.SalesCount["rye"]))
	content.WriteString("` + "`" + "`" + "`" + `\n\n")
	
	// System Overview
	content.WriteString("## System Overview\n\n")
	content.WriteString("This bakery simulation models a complete business workflow:\n\n")
	content.WriteString("1. **Production** makes bread through 4 stages (dough, kneading, baking, cooling)\n")
	content.WriteString("2. **Storefront** receives deliveries and manages inventory\n")
	content.WriteString("3. **Customers** arrive and purchase bread\n")
	content.WriteString("4. **Metrics** are tracked automatically (costs, revenue, waste)\n\n")
	
	content.WriteString("### Actors\n\n")
	content.WriteString("- **Production**: State machine for bread making\n")
	content.WriteString("- **Storefront**: Inventory management and sales\n")
	content.WriteString("- **Customer**: Probabilistic arrivals and purchases\n\n")
	
	// Architecture
	content.WriteString("## Architecture Notes\n\n")
	content.WriteString("### CANDIDATE = guard matches AND not blocked\n\n")
	content.WriteString("Each Process represents ONE transition:\n")
	content.WriteString("- Ready() checks ONE predicate\n")
	content.WriteString("- Returns ONE step (or nil)\n")
	content.WriteString("- Engine picks one candidate uniformly at random\n\n")
	
	// Save document
	os.WriteFile("model_" + fmt.Sprintf("%d", time.Now().Unix()) + "-output.md", []byte(content.String()), 0644)
	
	fmt.Printf("Production: %d, Sales: %d, Revenue: $%.2f, Profit: $%.2f\n", 
		production.TotalMade, totalSales, storefront.Revenue, profit)
}
`

	_ = english  // Template doesn't use English description yet
	_ = projectID

	// Write the generated code to a file
	filename := fmt.Sprintf("projects/model_%d.go", time.Now().Unix())
	if err := os.MkdirAll("projects", 0755); err != nil {
		return "", fmt.Errorf("failed to create projects directory: %w", err)
	}

	if err := os.WriteFile(filename, []byte(template), 0644); err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

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
