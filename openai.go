package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

// OpenAIClient handles interactions with OpenAI API
type OpenAIClient struct {
	APIKey string
	Model  string
}

// NewOpenAIClient creates a new OpenAI client
func NewOpenAIClient() *OpenAIClient {
	apiKey := os.Getenv("OPENAI_API_KEY")
	return &OpenAIClient{
		APIKey: apiKey,
		Model:  "gpt-4",
	}
}

// ChatMessage represents a message in the chat
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatCompletionRequest represents the request to OpenAI
type ChatCompletionRequest struct {
	Model    string        `json:"model"`
	Messages []ChatMessage `json:"messages"`
}

// ChatCompletionResponse represents the response from OpenAI
type ChatCompletionResponse struct {
	Choices []struct {
		Message ChatMessage `json:"message"`
	} `json:"choices"`
}

// GenerateModel uses OpenAI to convert English description to a Kripke structure
func (c *OpenAIClient) GenerateModel(description string) (*KripkeStructure, error) {
	if c.APIKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY not set")
	}

	systemPrompt := `You are an expert in formal verification and CTL model checking. 
Convert natural language descriptions into Kripke structure definitions.

Output format (JSON):
{
  "initial_state": "state_name",
  "states": ["state1", "state2", ...],
  "transitions": [
    {"from": "state1", "to": "state2"},
    ...
  ],
  "labels": [
    {"state": "state1", "proposition": "prop1"},
    ...
  ]
}

Only output valid JSON, no additional text.`

	userPrompt := fmt.Sprintf("Create a Kripke structure for: %s", description)

	request := ChatCompletionRequest{
		Model: c.Model,
		Messages: []ChatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.APIKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("OpenAI API error: %s", string(body))
	}

	var chatResp ChatCompletionResponse
	if err := json.Unmarshal(body, &chatResp); err != nil {
		return nil, err
	}

	if len(chatResp.Choices) == 0 {
		return nil, fmt.Errorf("no response from OpenAI")
	}

	modelJSON := chatResp.Choices[0].Message.Content
	return c.ParseModelJSON(modelJSON)
}

// ModelJSON represents the JSON structure for a Kripke model
type ModelJSON struct {
	InitialState string `json:"initial_state"`
	States       []string `json:"states"`
	Transitions  []struct {
		From string `json:"from"`
		To   string `json:"to"`
	} `json:"transitions"`
	Labels []struct {
		State       string `json:"state"`
		Proposition string `json:"proposition"`
	} `json:"labels"`
}

// ParseModelJSON parses JSON into a Kripke structure
func (c *OpenAIClient) ParseModelJSON(jsonStr string) (*KripkeStructure, error) {
	// Extract JSON from markdown code blocks if present
	jsonStr = strings.TrimSpace(jsonStr)
	if strings.HasPrefix(jsonStr, "```") {
		lines := strings.Split(jsonStr, "\n")
		jsonStr = strings.Join(lines[1:len(lines)-1], "\n")
	}

	var model ModelJSON
	if err := json.Unmarshal([]byte(jsonStr), &model); err != nil {
		return nil, fmt.Errorf("failed to parse model JSON: %w", err)
	}

	k := NewKripkeStructure(State(model.InitialState))

	for _, state := range model.States {
		k.AddState(State(state))
	}

	for _, trans := range model.Transitions {
		k.AddTransition(State(trans.From), State(trans.To))
	}

	for _, label := range model.Labels {
		k.AddLabel(State(label.State), Proposition(label.Proposition))
	}

	return k, nil
}

// GenerateCTLFormula uses OpenAI to convert English to CTL formula
func (c *OpenAIClient) GenerateCTLFormula(description string) (string, error) {
	if c.APIKey == "" {
		return "", fmt.Errorf("OPENAI_API_KEY not set")
	}

	systemPrompt := `You are an expert in CTL (Computational Tree Logic).
Convert natural language queries into CTL formulas.

CTL operators:
- EX φ: there exists a next state where φ
- AX φ: in all next states φ
- EF φ: there exists a path where eventually φ
- AF φ: on all paths eventually φ
- EG φ: there exists a path where always φ
- AG φ: on all paths always φ
- E[φ U ψ]: there exists a path where φ until ψ
- A[φ U ψ]: on all paths φ until ψ

Boolean operators: ∧ (and), ∨ (or), ¬ (not), → (implies)

Output only the CTL formula, no additional text.`

	userPrompt := fmt.Sprintf("Convert to CTL: %s", description)

	request := ChatCompletionRequest{
		Model: c.Model,
		Messages: []ChatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.APIKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("OpenAI API error: %s", string(body))
	}

	var chatResp ChatCompletionResponse
	if err := json.Unmarshal(body, &chatResp); err != nil {
		return "", err
	}

	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("no response from OpenAI")
	}

	return strings.TrimSpace(chatResp.Choices[0].Message.Content), nil
}
