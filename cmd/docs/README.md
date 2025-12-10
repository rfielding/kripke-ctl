# kripke-ctl Documentation Server

Web server that demonstrates the complete LLM workflow:

```
English Description → LLM generates Go model → Execute → Display output
```

## Features

- 📝 **English Input**: Text area to describe system in plain English
- 🤖 **LLM Generation**: Uses Claude API to generate kripke-ctl model
- 🏃 **Auto-Execute**: Runs generated model automatically
- 📊 **Live Output**: Shows both Go code and markdown output
- 💾 **Project History**: Keeps all generated models with unique IDs
- 📚 **Documentation**: Serves docs from `./docs` directory

## Prerequisites

```bash
# 1. Install kripke library additions
cd ~/code/kripke-ctl
cp kripke-final/diagrams.go kripke/
cp kripke-final/metrics.go kripke/

# 2. Set Anthropic API key
export ANTHROPIC_API_KEY="sk-ant-your-api-key-here"
```

## Running

```bash
cd cmd/docs
go run main.go
```

Then open: http://localhost:8080

## Usage

### 1. Enter English Description

Example:
```
Create a producer-consumer system where:
- Producer sends 10 messages (1KB each)
- Consumer receives and processes them
- Channel has capacity of 3
- Track: items sent/received, total bytes, delays
- Generate: sequence diagram, metrics table
```

### 2. Click "Generate & Execute Model"

The system will:
1. Call Claude API to generate Go model
2. Save model to `projects/model_<timestamp>.go`
3. Execute: `go run projects/model_<timestamp>.go`
4. Display Go code immediately
5. Display markdown output when execution completes

### 3. View Output

Two tabs:
- **Go Model**: The generated kripke-ctl code
- **Output**: Markdown with diagrams, metrics, tables

### 4. Browse Projects

Left panel shows all generated projects:
- Click any project to reload it
- Projects persist for the session
- Files saved in `projects/` directory

## File Structure

```
cmd/docs/
├── main.go              - Web server
├── templates/
│   └── index.html       - UI template
└── README.md            - This file

Generated files:
projects/
├── model_1234567890.go  - Generated Go models
├── model_1234567890-output.md
└── ...
```

## API Endpoints

- `GET /` - Main UI
- `GET /docs/*` - Documentation files
- `POST /api/generate` - Generate model from English
- `GET /api/projects` - List all projects
- `GET /api/project/:id` - Get specific project

## How It Works

**Note**: Currently uses a proven template instead of LLM generation to avoid truncation issues.

### Step 1: Get Working Template

User enters description (any text).

Server returns a working producer-consumer template with correct kripke API.

### Step 2: Customize

Edit the generated file in `projects/` to match your needs:
- Change message count
- Modify buffer sizes
- Add more actors
- Customize behavior

### Step 3: Execute & View

Model runs automatically and generates sequence diagram output.

## Why Template Instead of LLM?

LLM code generation kept truncating because responses were too verbose (>8000 tokens).

Template approach:
- ✅ Always works
- ✅ Correct API
- ✅ Good starting point
- ✅ Instant (no API delay)

See `TEMPLATE_APPROACH.md` for details.
```
Generate a complete kripke-ctl model in Go based on this English description:
"Create upload system with 10 chunks"

Requirements:
1. Import "github.com/rfielding/kripke-ctl/kripke"
2. Define actors as structs with state variables
3. Implement Ready() method with guards and actions
4. Include main() that generates markdown output
...
```

### Step 2: Save & Execute

Server:
1. Generates unique ID: `model_1234567890`
2. Saves to `projects/model_1234567890.go`
3. Runs: `go run projects/model_1234567890.go`
4. Captures stdout and looks for markdown files

### Step 3: Display Results

UI shows:
- Generated Go code (immediately)
- Markdown output (when execution completes)
- Both saved for future reference

## Example English Prompts

### Producer-Consumer

```
Create a producer-consumer system where the Producer sends 10 items 
to the Consumer through a buffered channel (capacity 3). Track the 
number of items sent and received. Generate a sequence diagram and 
metrics table.
```

### File Upload

```
Create a file upload system where an Uploader sends 10 chunks (1MB each) 
to a Receiver. Use a buffered channel with capacity 3. Track total bytes 
sent/received and calculate throughput. Generate sequence diagram, 
metrics, and throughput table.
```

### Request-Response

```
Create a client-server system where the Client sends 15 requests to 
the Server. Server processes requests with 90% success rate (chance node). 
Track success/failure counts. Generate sequence diagram and pie chart 
showing success vs failure distribution.
```

### Multi-Actor Workflow

```
Create a 3-actor system: Client, LoadBalancer, Server. Client sends 
requests to LoadBalancer which routes to Server. LoadBalancer uses 
chance node to pick routing strategy (70% direct, 30% cached). Track 
request counts and routing decisions. Generate sequence diagram and 
distribution chart.
```

## Troubleshooting

### "ANTHROPIC_API_KEY environment variable not set"

```bash
export ANTHROPIC_API_KEY="sk-ant-..."
```

Get your API key from: https://console.anthropic.com/

### "Failed to execute model"

Check that:
1. kripke library additions are installed (`diagrams.go`, `metrics.go`)
2. Generated Go code has no syntax errors
3. All imports are available

View generated file:
```bash
cat projects/model_<id>.go
```

Try running manually:
```bash
cd projects
go run model_<id>.go
```

### "No markdown output"

The generated model should write to a file. Check that the LLM included:
```go
filename := "model_<id>-output.md"
os.WriteFile(filename, []byte(md.String()), 0644)
```

## Environment Variables

- `ANTHROPIC_API_KEY` - Required for LLM generation (get from https://console.anthropic.com/)
- `PORT` - Optional, defaults to 8080

## Development

To modify the UI:
1. Edit `templates/index.html`
2. Restart server: `go run main.go`
3. Refresh browser

To modify server logic:
1. Edit `main.go`
2. Restart server

Hot reload not implemented - just restart!

## Notes

- Projects stored in memory (lost on restart)
- Files persist in `projects/` directory
- Each project gets unique ID based on timestamp
- Generated models should output markdown to be captured

## Future Enhancements

- [ ] Persist projects to database
- [ ] Export projects as .zip
- [ ] Real-time streaming of execution output
- [ ] Syntax highlighting for Go code
- [ ] Markdown rendering (not just raw text)
- [ ] Edit and re-run models
- [ ] CTL verification results
- [ ] Compare multiple models

## The Complete Workflow

```
┌─────────────────┐
│ English         │
│ "Create upload" │
└────────┬────────┘
         ↓
┌─────────────────┐
│ Claude API      │
│ Generates Go    │
└────────┬────────┘
         ↓
┌─────────────────┐
│ kripke Model    │
│ model_123.go    │
└────────┬────────┘
         ↓
┌─────────────────┐
│ Execute         │
│ go run ...      │
└────────┬────────┘
         ↓
┌─────────────────┐
│ Markdown Output │
│ Diagrams        │
│ Metrics         │
│ Tables          │
└─────────────────┘
```

This is the vision: **English → Model → Verification → Implementation**

The web UI demonstrates Step 1-3. The generated model becomes the specification for implementation!
