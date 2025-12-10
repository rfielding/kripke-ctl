package kripke

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

// GenerateInteractionDiagram generates sequence diagram from actual execution
func (w *World) GenerateInteractionDiagram(maxEvents int) string {
	// Same as GenerateSequenceDiagram
	return w.GenerateSequenceDiagram(maxEvents)
}

// Metric represents an observability counter
type Metric struct {
	Name        string
	Type        string
	Value       float64
	Unit        string
	Description string
}

// MetricsCollector tracks observability metrics
type MetricsCollector struct {
	metrics map[string]*Metric
}

func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		metrics: make(map[string]*Metric),
	}
}

func (mc *MetricsCollector) Counter(name, desc, unit string) *Metric {
	if m, exists := mc.metrics[name]; exists {
		return m
	}
	m := &Metric{
		Name:        name,
		Type:        "counter",
		Value:       0,
		Unit:        unit,
		Description: desc,
	}
	mc.metrics[name] = m
	return m
}

func (m *Metric) Inc() {
	m.Value++
}

func (m *Metric) Add(delta float64) {
	m.Value += delta
}

// GenerateMetricsTable generates markdown table of metrics
func (mc *MetricsCollector) GenerateMetricsTable() string {
	var sb strings.Builder
	sb.WriteString("| Metric | Type | Value | Unit | Description |\n")
	sb.WriteString("|--------|------|-------|------|-------------|\n")
	
	names := make([]string, 0, len(mc.metrics))
	for name := range mc.metrics {
		names = append(names, name)
	}
	sort.Strings(names)
	
	for _, name := range names {
		m := mc.metrics[name]
		sb.WriteString(fmt.Sprintf("| %s | %s | %.2f | %s | %s |\n",
			m.Name, m.Type, m.Value, m.Unit, m.Description))
	}
	
	return sb.String()
}

// GenerateMetricsChart generates Mermaid chart from metrics
func (mc *MetricsCollector) GenerateMetricsChart(metricNames []string) string {
	var sb strings.Builder
	
	sb.WriteString("xychart-beta\n")
	sb.WriteString("    title \"Metrics\"\n")
	
	sb.WriteString("    x-axis [")
	for i, name := range metricNames {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(fmt.Sprintf("\"%s\"", name))
	}
	sb.WriteString("]\n")
	
	sb.WriteString("    bar [")
	for i, name := range metricNames {
		if i > 0 {
			sb.WriteString(", ")
		}
		if m, exists := mc.metrics[name]; exists {
			sb.WriteString(fmt.Sprintf("%.0f", m.Value))
		} else {
			sb.WriteString("0")
		}
	}
	sb.WriteString("]\n")
	
	return sb.String()
}

// Throughput calculates throughput from counters
type Throughput struct {
	TotalBytes int64
	TotalTime  time.Duration
	Rate       float64
}

func CalculateThroughput(totalBytes int64, startTime, endTime time.Time) Throughput {
	duration := endTime.Sub(startTime)
	rate := 0.0
	if duration.Seconds() > 0 {
		rate = float64(totalBytes) / duration.Seconds()
	}
	return Throughput{
		TotalBytes: totalBytes,
		TotalTime:  duration,
		Rate:       rate,
	}
}

func (t Throughput) String() string {
	return fmt.Sprintf("%.2f MB/s (%d bytes in %v)",
		t.Rate/1024/1024, t.TotalBytes, t.TotalTime)
}

// GenerateThroughputTable generates markdown table of throughput metrics
func GenerateThroughputTable(throughputs map[string]Throughput) string {
	var sb strings.Builder
	sb.WriteString("| Operation | Total Bytes | Total Time | Throughput |\n")
	sb.WriteString("|-----------|-------------|------------|------------|\n")
	
	names := make([]string, 0, len(throughputs))
	for name := range throughputs {
		names = append(names, name)
	}
	sort.Strings(names)
	
	for _, name := range names {
		t := throughputs[name]
		sb.WriteString(fmt.Sprintf("| %s | %d bytes | %v | %.2f MB/s |\n",
			name, t.TotalBytes, t.TotalTime, t.Rate/1024/1024))
	}
	
	return sb.String()
}

// ActorState represents the state of an actor (its variables)
type ActorState struct {
	ActorID   string
	Variables map[string]interface{}
}

func (as ActorState) String() string {
	var parts []string
	for k, v := range as.Variables {
		parts = append(parts, fmt.Sprintf("%s=%v", k, v))
	}
	sort.Strings(parts)
	return strings.Join(parts, ", ")
}

// GenerateActorStateTable generates table showing actor states
func GenerateActorStateTable(states []ActorState) string {
	var sb strings.Builder
	sb.WriteString("| Actor | State Variables |\n")
	sb.WriteString("|-------|----------------|\n")
	
	for _, state := range states {
		sb.WriteString(fmt.Sprintf("| %s | %s |\n",
			state.ActorID, state.String()))
	}
	
	return sb.String()
}
