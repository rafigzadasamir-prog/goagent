package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAgentRunExecutesToolThenReturnsFinalAnswer(t *testing.T) {
	callCount := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		var req messagesRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decoding request: %v", err)
		}

		switch callCount {
		case 1:
			if len(req.Messages) != 1 {
				t.Errorf("call 1: expected 1 message, got %d", len(req.Messages))
			}
			json.NewEncoder(w).Encode(Response{
				StopReason: "tool_use",
				Content: []ContentBlock{
					{Type: "tool_use", ID: "tool-1", Name: "calculator", Input: json.RawMessage(`{"expression":"2+2"}`)},
				},
			})
		case 2:
			if len(req.Messages) != 3 {
				t.Fatalf("call 2: expected 3 messages (user, assistant, tool_result), got %d", len(req.Messages))
			}
			toolResultMsg := req.Messages[2]
			if len(toolResultMsg.Content) != 1 || toolResultMsg.Content[0].Type != "tool_result" {
				t.Fatalf("call 2: expected a tool_result block, got %+v", toolResultMsg.Content)
			}
			if toolResultMsg.Content[0].Content != "4" {
				t.Errorf("call 2: tool_result content = %q, want %q", toolResultMsg.Content[0].Content, "4")
			}
			json.NewEncoder(w).Encode(Response{
				StopReason: "end_turn",
				Content:    []ContentBlock{{Type: "text", Text: "The answer is 4."}},
			})
		default:
			t.Fatalf("unexpected call count %d", callCount)
		}
	}))
	defer server.Close()

	client := NewClient("test-key", "claude-sonnet-5", "")
	client.BaseURL = server.URL

	agent := NewAgent(client, []Tool{calculatorTool{}})

	answer, err := agent.Run("what is 2+2?")
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if answer != "The answer is 4." {
		t.Errorf("Run() = %q, want %q", answer, "The answer is 4.")
	}
	if callCount != 2 {
		t.Errorf("expected 2 API calls, got %d", callCount)
	}
}

func TestAgentRunReturnsImmediateTextWithoutToolUse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(Response{
			StopReason: "end_turn",
			Content:    []ContentBlock{{Type: "text", Text: "hi"}},
		})
	}))
	defer server.Close()

	client := NewClient("test-key", "claude-sonnet-5", "")
	client.BaseURL = server.URL

	agent := NewAgent(client, nil)

	answer, err := agent.Run("hello")
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if answer != "hi" {
		t.Errorf("Run() = %q, want %q", answer, "hi")
	}
}

func TestAgentRunUnknownToolReportsErrorToModel(t *testing.T) {
	callCount := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount == 1 {
			json.NewEncoder(w).Encode(Response{
				StopReason: "tool_use",
				Content: []ContentBlock{
					{Type: "tool_use", ID: "tool-1", Name: "nonexistent_tool", Input: json.RawMessage(`{}`)},
				},
			})
			return
		}

		var req messagesRequest
		json.NewDecoder(r.Body).Decode(&req)
		toolResultMsg := req.Messages[2]
		if !toolResultMsg.Content[0].IsError {
			t.Error("expected IsError to be true for an unknown tool")
		}
		json.NewEncoder(w).Encode(Response{
			StopReason: "end_turn",
			Content:    []ContentBlock{{Type: "text", Text: "done"}},
		})
	}))
	defer server.Close()

	client := NewClient("test-key", "claude-sonnet-5", "")
	client.BaseURL = server.URL

	agent := NewAgent(client, nil)
	if _, err := agent.Run("use a tool that doesn't exist"); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
}
