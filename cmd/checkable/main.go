package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

// -------------------- Config --------------------

type Config struct {
	Listen        string
	Provider      string
	OpenAIKey     string
	AnthropicKey  string
	OpenAIModel   string
	AnthropicModel string
}

func loadConfig() Config {
	cfg := Config{
		Listen:         getenv("BCTL_LISTEN", "127.0.0.1:8080"),
		Provider:       getenv("LLM_PROVIDER", "openai"),
		OpenAIKey:      os.Getenv("OPENAI_API_KEY"),
		AnthropicKey:   os.Getenv("ANTHROPIC_API_KEY"),
		OpenAIModel:    getenv("OPENAI_MODEL", "gpt-5.2"),
		AnthropicModel: getenv("ANTHROPIC_MODEL", "claude-sonnet-4-5"),
	}
	return cfg
}

func getenv(k, def string) string {
	v := os.Getenv(k)
	if v == "" {
		return def
	}
	return v
}

// -------------------- LLM Interface --------------------

type LLM interface {
	GenerateBoundedCTL(ctx context.Context, req GenerateReq) (GenerateResp, error)
}

type GenerateReq struct {
	// The evolving spec (previous Lisp) and the user's new clarification.
	PrevLisp string `json:"prev_lisp"`
	UserText string `json:"user_text"`

	// Optional: stable requirement list, extra instructions, etc.
	// Keep it simple for now.
}

type GenerateResp struct {
	Lisp      string `json:"lisp"`
	Provider  string `json:"provider"`
	Model     string `json:"model"`
	RequestID string `json:"request_id,omitempty"`
	Raw       string `json:"raw,omitempty"` // for debugging; you can omit in UI
}

// -------------------- OpenAI (Responses API) --------------------

type OpenAIClient struct {
	APIKey string
	Model  string
	HTTP   *http.Client
}

func (c *OpenAIClient) GenerateBoundedCTL(ctx context.Context, req GenerateReq) (GenerateResp, error) {
	if c.APIKey == "" {
		return GenerateResp{}, errors.New("OPENAI_API_KEY is empty")
	}
	system := boundedCTLSystemPrompt()

	// We ask for ONLY the Lisp text. The engine/web UI can treat this as the sole artifact.
	// Using Responses API per OpenAI docs. :contentReference[oaicite:3]{index=3}
	body := map[string]any{
		"model": c.Model,
		"input": []any{
			map[string]any{"role": "system", "content": system},
			map[string]any{"role": "user", "content": openAIUserPrompt(req)},
		},
		// Keep verbosity low; you can tune.
		"text": map[string]any{"verbosity": "low"},
	}

	b, _ := json.Marshal(body)
	httpReq, _ := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/responses", bytes.NewReader(b))
	httpReq.Header.Set("Authorization", "Bearer "+c.APIKey)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTP.Do(httpReq)
	if err != nil {
		return GenerateResp{}, err
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return GenerateResp{}, fmt.Errorf("openai status=%d body=%s", resp.StatusCode, string(raw))
	}

	// ... after reading raw
	var parsed struct {
		ID     string `json:"id"`
		Output []struct {
			Type    string `json:"type"`
			Content []struct {
				Type string `json:"type"`
				Text string `json:"text"`
			} `json:"content"`
		} `json:"output"`
	}

	if err := json.Unmarshal(raw, &parsed); err != nil {
		return GenerateResp{}, fmt.Errorf("openai unmarshal: %w", err)
	}

	var out strings.Builder
	for _, o := range parsed.Output {
		for _, c := range o.Content {
			if c.Type == "output_text" || c.Type == "text" {
				out.WriteString(c.Text)
			}
		}
	}
	text := strings.TrimSpace(out.String())
	text = stripCodeFences(text)

	return GenerateResp{
		Lisp:      text,
		Provider:  "openai",
		Model:     c.Model,
		RequestID: parsed.ID,
		Raw:       string(raw),
	}, nil
}

func openAIUserPrompt(req GenerateReq) string {
	return fmt.Sprintf(
		`You are updating a BoundedCTL-LISP spec.
Rules:
- Output ONLY the BoundedCTL-LISP text. No prose. No markdown.
- Keep requirements cumulative; do not delete existing (require ...) unless user explicitly says so.
- Prefer minimal diffs and stable names.

PREVIOUS_LISP:
%s

NEW_CLARIFICATION:
%s
`, req.PrevLisp, req.UserText)
}

// -------------------- Anthropic (Messages API) --------------------

type AnthropicClient struct {
	APIKey string
	Model  string
	HTTP   *http.Client
}

func (c *AnthropicClient) GenerateBoundedCTL(ctx context.Context, req GenerateReq) (GenerateResp, error) {
	if c.APIKey == "" {
		return GenerateResp{}, errors.New("ANTHROPIC_API_KEY is empty")
	}
	system := boundedCTLSystemPrompt()

	// Messages API per Anthropic docs. :contentReference[oaicite:4]{index=4}
	body := map[string]any{
		"model":      c.Model,
		"max_tokens": 4000,
		"system":     system,
		"messages": []map[string]any{
			{"role": "user", "content": anthropicUserPrompt(req)},
		},
	}

	b, _ := json.Marshal(body)
	httpReq, _ := http.NewRequestWithContext(ctx, "POST", "https://api.anthropic.com/v1/messages", bytes.NewReader(b))
	httpReq.Header.Set("x-api-key", c.APIKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")
	httpReq.Header.Set("content-type", "application/json")

	resp, err := c.HTTP.Do(httpReq)
	if err != nil {
		return GenerateResp{}, err
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return GenerateResp{}, fmt.Errorf("anthropic status=%d body=%s", resp.StatusCode, string(raw))
	}

	// Extract the first text content block.
	var parsed struct {
		ID      string `json:"id"`
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	}
	_ = json.Unmarshal(raw, &parsed)

	var out strings.Builder
	for _, c := range parsed.Content {
		if c.Type == "text" {
			out.WriteString(c.Text)
		}
	}
	text := strings.TrimSpace(out.String())
	text = stripCodeFences(text)

	return GenerateResp{
		Lisp:      text,
		Provider:  "anthropic",
		Model:     c.Model,
		RequestID: parsed.ID,
		Raw:       string(raw),
	}, nil
}

func anthropicUserPrompt(req GenerateReq) string {
	return fmt.Sprintf(
		`Update the BoundedCTL-LISP spec.
Output ONLY the BoundedCTL-LISP text. No prose. No markdown.
Keep requirements cumulative.

PREVIOUS_LISP:
%s

NEW_CLARIFICATION:
%s
`, req.PrevLisp, req.UserText)
}

// -------------------- Prompts --------------------

func boundedCTLSystemPrompt() string {
	// Keep this short and strict. You can expand later with a preamble “stdlib” and sanity checks.
	return strings.TrimSpace(`
You are a compiler assistant that outputs BoundedCTL-LISP only.

Hard rules:
- Return ONLY BoundedCTL-LISP text (no explanations, no markdown, no code fences).
- Do not invent APIs outside the language. Use simple S-expressions.
- Keep requirements cumulative: preserve existing (require ...) entries unless explicitly told to remove/relax.
- Use stable names (actors/chans/preds/pcs) so diagrams remain comparable across edits.
- Assume bounded queues everywhere; overload manifests as blocking/deadlock, not unbounded growth.
`)
}

// -------------------- HTTP server --------------------

type Server struct {
	cfg Config
	llm LLM
}

func main() {
	cfg := loadConfig()
	httpClient := &http.Client{Timeout: 90 * time.Second}

	var llm LLM
	switch strings.ToLower(cfg.Provider) {
	case "openai":
		llm = &OpenAIClient{APIKey: cfg.OpenAIKey, Model: cfg.OpenAIModel, HTTP: httpClient}
	case "anthropic":
		llm = &AnthropicClient{APIKey: cfg.AnthropicKey, Model: cfg.AnthropicModel, HTTP: httpClient}
	default:
		log.Fatalf("unknown LLM_PROVIDER=%q (use openai|anthropic)", cfg.Provider)
	}

	s := &Server{cfg: cfg, llm: llm}

	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleIndex)
	mux.HandleFunc("/healthz", s.handleHealth)
	mux.HandleFunc("/api/llm", s.handleLLM)
	mux.HandleFunc("/api/markdown", s.handleMarkdown)

	log.Printf("listening on http://%s (provider=%s)", cfg.Listen, cfg.Provider)
	log.Fatal(http.ListenAndServe(cfg.Listen, withLogging(mux)))
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "text/plain")
	w.WriteHeader(200)
	_, _ = w.Write([]byte("ok\n"))
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(indexHTML))
}

func (s *Server) handleLLM(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "POST only", 405)
		return
	}
	var in GenerateReq
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, "bad json: "+err.Error(), 400)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 120*time.Second)
	defer cancel()

	out, err := s.llm.GenerateBoundedCTL(ctx, in)
	if err != nil {
		http.Error(w, err.Error(), 502)
		return
	}

	// IMPORTANT: The only durable artifact is out.Lisp. The Raw response can be disabled later.
	w.Header().Set("content-type", "application/json")
	json.NewEncoder(w).Encode(out)
}

type MarkdownReq struct {
	Lisp string `json:"lisp"`
}

type MarkdownResp struct {
	Markdown string `json:"markdown"`
	Errors   []string `json:"errors,omitempty"`
}

func (s *Server) handleMarkdown(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "POST only", 405)
		return
	}

	var lisp string
	ct := r.Header.Get("Content-Type")

	if strings.HasPrefix(ct, "application/json") {
		var in MarkdownReq
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			http.Error(w, "bad json: "+err.Error(), 400)
			return
		}
		lisp = in.Lisp
	} else {
		// treat body as raw Lisp
		b, _ := io.ReadAll(r.Body)
		lisp = string(b)
	}

	md, errs := RenderMarkdownMinimal(lisp)

	w.Header().Set("content-type", "application/json")
	_ = json.NewEncoder(w).Encode(MarkdownResp{
		Markdown: md,
		Errors:   errs,
	})
}

// --- Minimal extraction (temporary!) ---
// This is intentionally small and dumb: it pulls :name/:english, (prop ... :english ...),
// and (require ... :ctl ... :english ...).
// Replace later with your real AST parser.
func RenderMarkdownMinimal(lisp string) (string, []string) {
	type Prop struct{ Name, English string }
	type Req struct{ ID, CTL, English string }

	var errs []string
	lisp = strings.TrimSpace(lisp)
	if lisp == "" {
		return "", []string{"empty lisp"}
	}

	sysName := findKeywordString(lisp, "(system", ":name")
	sysEng := findKeywordString(lisp, "(system", ":english")

	props := findForms(lisp, "(prop")
	var propOut []Prop
	for _, form := range props {
		name := firstSymbolAfter(form, "(prop")
		eng := findKeywordString(form, "(prop", ":english")
		if name == "" {
			continue
		}
		if eng == "" {
			errs = append(errs, "prop "+name+" missing :english")
		}
		propOut = append(propOut, Prop{Name: name, English: eng})
	}

	reqs := findForms(lisp, "(require")
	var reqOut []Req
	for _, form := range reqs {
		id := firstSymbolAfter(form, "(require")
		ctl := findKeywordForm(form, ":ctl")
		eng := findKeywordString(form, "(require", ":english")
		if id == "" {
			continue
		}
		if eng == "" {
			errs = append(errs, "require "+id+" missing :english")
		}
		if ctl == "" {
			errs = append(errs, "require "+id+" missing :ctl")
		}
		reqOut = append(reqOut, Req{ID: id, CTL: ctl, English: eng})
	}

	var b strings.Builder
	if sysName == "" {
		sysName = "BoundedCTL Spec"
		errs = append(errs, "system missing :name")
	}
	fmt.Fprintf(&b, "# %s\n\n", sysName)
	if sysEng != "" {
		fmt.Fprintf(&b, "%s\n\n", sysEng)
	} else {
		errs = append(errs, "system missing :english")
	}

	if len(propOut) > 0 {
		b.WriteString("## Propositions\n\n")
		for _, p := range propOut {
			fmt.Fprintf(&b, "### %s\n\n", p.Name)
			if p.English != "" {
				fmt.Fprintf(&b, "**Meaning:** %s\n\n", p.English)
			} else {
				b.WriteString("**Meaning:** *(missing :english)*\n\n")
			}
		}
	}

	if len(reqOut) > 0 {
		b.WriteString("## Requirements\n\n")
		for _, q := range reqOut {
			fmt.Fprintf(&b, "### %s\n\n", q.ID)
			if q.English != "" {
				fmt.Fprintf(&b, "**English:** %s\n\n", q.English)
			} else {
				b.WriteString("**English:** *(missing :english)*\n\n")
			}
			if q.CTL != "" {
				b.WriteString("**Formal (CTL):**\n\n```lisp\n")
				b.WriteString(strings.TrimSpace(q.CTL))
				b.WriteString("\n```\n\n")
			}
			b.WriteString("**Status:** ⬜ Unknown (not yet checked)\n\n")
		}
	}

	return b.String(), errs
}

// --- Helpers: quick-and-dirty S-expression scanning ---

func findForms(src, head string) []string {
	var out []string
	i := 0
	for {
		j := strings.Index(src[i:], head)
		if j < 0 {
			break
		}
		j += i
		form, end := sliceBalanced(src, j)
		if form != "" {
			out = append(out, form)
			i = end
		} else {
			i = j + len(head)
		}
	}
	return out
}

func sliceBalanced(s string, start int) (string, int) {
	// start points at '(' of a form. Return balanced substring.
	if start < 0 || start >= len(s) || s[start] != '(' {
		return "", start
	}
	depth := 0
	inStr := false
	esc := false
	for i := start; i < len(s); i++ {
		ch := s[i]
		if inStr {
			if esc {
				esc = false
				continue
			}
			if ch == '\\' {
				esc = true
				continue
			}
			if ch == '"' {
				inStr = false
			}
			continue
		}
		if ch == '"' {
			inStr = true
			continue
		}
		if ch == '(' {
			depth++
		} else if ch == ')' {
			depth--
			if depth == 0 {
				return s[start : i+1], i + 1
			}
		}
	}
	return "", start
}

func firstSymbolAfter(form, head string) string {
	// e.g. "(prop Foo :english ...)" -> "Foo"
	k := strings.Index(form, head)
	if k < 0 {
		return ""
	}
	rest := strings.TrimSpace(form[k+len(head):])
	// symbol ends at whitespace or ')' or ':'
	for i := 0; i < len(rest); i++ {
		if rest[i] == ' ' || rest[i] == '\n' || rest[i] == '\t' || rest[i] == ')' || rest[i] == ':' {
			return rest[:i]
		}
	}
	return rest
}

func findKeywordString(src, scopeHead, key string) string {
	// finds key followed by a string literal "...", within the first balanced scope of scopeHead
	scope := src
	if scopeHead != "" {
		j := strings.Index(src, scopeHead)
		if j >= 0 {
			if form, _ := sliceBalanced(src, j); form != "" {
				scope = form
			}
		}
	}
	k := strings.Index(scope, key)
	if k < 0 {
		return ""
	}
	rest := scope[k+len(key):]
	q := strings.Index(rest, `"`)
	if q < 0 {
		return ""
	}
	rest = rest[q+1:]
	// read until next unescaped "
	var b strings.Builder
	esc := false
	for i := 0; i < len(rest); i++ {
		ch := rest[i]
		if esc {
			b.WriteByte(ch)
			esc = false
			continue
		}
		if ch == '\\' {
			esc = true
			continue
		}
		if ch == '"' {
			return b.String()
		}
		b.WriteByte(ch)
	}
	return ""
}

func findKeywordForm(form string, key string) string {
	// finds key followed by one balanced S-expression, e.g. ":ctl (AG ...)" -> "(AG ...)"
	k := strings.Index(form, key)
	if k < 0 {
		return ""
	}
	rest := strings.TrimSpace(form[k+len(key):])
	// find first '('
	p := strings.Index(rest, "(")
	if p < 0 {
		return ""
	}
	sub, _ := sliceBalanced(rest, p)
	return sub
}

// -------------------- Helpers --------------------

func stripCodeFences(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "```") {
		// Remove first fence line
		nl := strings.Index(s, "\n")
		if nl >= 0 {
			s = s[nl+1:]
		}
		// Remove trailing fence
		if i := strings.LastIndex(s, "```"); i >= 0 {
			s = s[:i]
		}
	}
	return strings.TrimSpace(s)
}

func withLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rid := shortID()
		w.Header().Set("x-request-id", rid)
		next.ServeHTTP(w, r)
		log.Printf("%s %s rid=%s dur=%s", r.Method, r.URL.Path, rid, time.Since(start))
	})
}

func shortID() string {
	var b [6]byte
	_, _ = rand.Read(b[:])
	return hex.EncodeToString(b[:])
}

// -------------------- Minimal UI --------------------

const indexHTML = `<!doctype html>
<html>
<head>
  <meta charset="utf-8" />
  <title>kripke-ctl — BoundedCTL Notebook</title>
    <style>
      :root { color-scheme: dark; }
      body { background:#000; color:#ddd; font-family:sans-serif; margin:0;
             display:grid; grid-template-columns: 1fr 1fr; height:100vh; }
      textarea, input, pre {
        background:#000; color:#ddd; border:1px solid #333;
        font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
      }
      textarea { width:100%; height: calc(100vh - 160px); font-size: 14px; line-height: 1.35; }
      pre { white-space: pre-wrap; padding: 10px; height: calc(100vh - 160px); overflow:auto; }
      button { background:#111; color:#ddd; border:1px solid #333; padding:6px 10px; }
</style>

</head>
<body>
  <div class="pane">
    <div class="bar">
      <button onclick="callLLM()">Ask LLM → Lisp</button>
      <button onclick="clearOut()">Clear output</button>
    </div>
    <div><strong>Prev Lisp (source of truth)</strong></div>
    <textarea id="prev"></textarea>
    <div style="margin-top:8px"><strong>Clarification</strong></div>
    <input id="clar" placeholder="Describe what to change / clarify..." />
  </div>

  <div class="pane">
    <div class="bar">
      <button onclick="renderMarkdown()">Render Markdown</button>
      <button onclick="copyLisp()">Copy Lisp to left</button>
    </div>
    <div><strong>LLM Output (must be Lisp only)</strong></div>
    <pre id="out"></pre>
  </div>
<script>
async function renderMarkdown() {
  const lisp = document.getElementById('prev').value.trim();
  const out = document.getElementById('out');
  if (!lisp) { out.textContent = "No Lisp to render."; return; }

  out.textContent = "Rendering Markdown...";
  const resp = await fetch('/api/markdown', {
    method: 'POST',
    headers: { 'content-type': 'application/json' },
    body: JSON.stringify({ lisp })
  });
  const txt = await resp.text();
  if (!resp.ok) { out.textContent = txt; return; }

  const j = JSON.parse(txt);
  let md = j.markdown || "";
  if (j.errors && j.errors.length) {
    md += "\n\n---\n\n## Spec issues\n" + j.errors.map(e => "- " + e).join("\n") + "\n";
  }
  out.textContent = md;
}
window.addEventListener('DOMContentLoaded', () => {
  loadState();
  hookAutosave();
  renderMarkdown(); // auto-regenerate docs view
});
</script>

</body>
</html>`

