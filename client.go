package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const defaultAPIBase = "https://api.anthropic.com/v1/messages"
const anthropicVersion = "2023-06-01"

// ContentBlock is one block of a message: plain text, a tool call the
// model wants to make, or a tool result being sent back to the model.
type ContentBlock struct {
	Type string `json:"type"`

	// type == "text"
	Text string `json:"text,omitempty"`

	// type == "tool_use" (from the model)
	ID    string          `json:"id,omitempty"`
	Name  string          `json:"name,omitempty"`
	Input json.RawMessage `json:"input,omitempty"`

	// type == "tool_result" (sent back to the model)
	ToolUseID string `json:"tool_use_id,omitempty"`
	Content   string `json:"content,omitempty"`
	IsError   bool   `json:"is_error,omitempty"`
}

// Message is a single turn in a conversation.
type Message struct {
	Role    string         `json:"role"`
	Content []ContentBlock `json:"content"`
}

// ToolDef describes a tool to the model.
type ToolDef struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	InputSchema json.RawMessage `json:"input_schema"`
}

// Client talks to the Anthropic Messages API.
type Client struct {
	APIKey     string
	Model      string
	SystemMsg  string
	BaseURL    string
	HTTPClient *http.Client
}

// NewClient builds a Client with sane defaults.
func NewClient(apiKey, model, systemMsg string) *Client {
	return &Client{
		APIKey:    apiKey,
		Model:     model,
		SystemMsg: systemMsg,
		BaseURL:   defaultAPIBase,
		HTTPClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

type messagesRequest struct {
	Model     string    `json:"model"`
	MaxTokens int       `json:"max_tokens"`
	System    string    `json:"system,omitempty"`
	Messages  []Message `json:"messages"`
	Tools     []ToolDef `json:"tools,omitempty"`
}

// Response is the parsed result of a Messages API call.
type Response struct {
	Content    []ContentBlock `json:"content"`
	StopReason string         `json:"stop_reason"`
	Error      *apiError      `json:"error"`
}

type apiError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

// Send posts the conversation so far, with the given tool definitions,
// and returns the parsed response.
func (c *Client) Send(history []Message, tools []ToolDef) (*Response, error) {
	reqBody := messagesRequest{
		Model:     c.Model,
		MaxTokens: 1024,
		System:    c.SystemMsg,
		Messages:  history,
		Tools:     tools,
	}

	payload, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("encoding request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, c.BaseURL, bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.APIKey)
	req.Header.Set("anthropic-version", anthropicVersion)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("calling anthropic api: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	var parsed Response
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, fmt.Errorf("parsing response (status %d): %w", resp.StatusCode, err)
	}

	if parsed.Error != nil {
		return nil, fmt.Errorf("api error (%s): %s", parsed.Error.Type, parsed.Error.Message)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	return &parsed, nil
}
