# Examples Overview

This directory contains examples demonstrating kripke-ctl. Some print to stdout, others generate Markdown files with diagrams.

## 📄 Examples That Generate Markdown Reports

### ⭐ example6_complete_report.go (RECOMMENDED)
**Output**: `producer-consumer-complete-report.md`

The most comprehensive example showing the **complete workflow**:
1. Runs the actor engine for 20 steps
2. Logs all events with queue delays
3. Extracts the state space
4. Verifies CTL properties
5. Generates a Markdown report with:
   - Execution trace table
   - Mermaid sequence diagram
   - Mermaid state diagram
   - CTL verification results
   - Performance metrics

```bash
cd kripke-ctl
go run ./outputs/example6_complete_report.go
```

### example5_markdown_report.go
**Output**: `producer-consumer-report.md`

A simpler report focusing on CTL verification:
- System description
- State space
- Mermaid state diagram
- CTL verification table
- Detailed results

```bash
cd kripke-ctl
go run ./outputs/example5_markdown_report.go
```

## 🖥️ Examples That Print to Stdout

### example3_with_import.go
Basic CTL verification printed to terminal:
```bash
go run ./outputs/example3_with_import.go
```

### example4_engine_integration.go
Engine execution + CTL verification printed to terminal:
```bash
go run ./outputs/example4_engine_integration.go
```

## 📁 Self-Contained Examples (Don't Import Package)

These were for quick testing but don't import your kripke package:
- `example1_client_server.go`
- `example2_producer_consumer.go`

**For your own code, use examples 3-6 which import `github.com/rfielding/kripke-ctl/kripke`**

## 🎯 Which Example Should I Use?

**Want a Markdown report with diagrams?**
→ Use **example6_complete_report.go** (most comprehensive)
→ Or **example5_markdown_report.go** (simpler)

**Just want to see how CTL verification works?**
→ Use **example3_with_import.go**

**Want to see engine + CTL integration?**
→ Use **example4_engine_integration.go** or **example6_complete_report.go**

## 📊 Viewing Generated Reports

The `.md` files contain embedded Mermaid diagrams that render automatically on:
- GitHub
- GitLab
- VS Code (with Markdown preview)
- [Mermaid Live Editor](https://mermaid.live)

Open the generated `.md` file to see:
- Beautiful state diagrams
- Sequence diagrams showing message flow
- Tables of CTL verification results
- Complete system analysis

## 🚀 Quick Start

```bash
cd /path/to/kripke-ctl

# Generate a complete report with diagrams
go run ./outputs/example6_complete_report.go

# Open the generated file
cat producer-consumer-complete-report.md
# or open in VS Code, push to GitHub, etc.
```

That's it! You now have a complete formal verification report.
