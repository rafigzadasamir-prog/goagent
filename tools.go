package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Tool is something the agent can call by name with a JSON input object,
// getting back a text result to feed back to the model.
type Tool interface {
	Name() string
	Description() string
	InputSchema() json.RawMessage
	Execute(input json.RawMessage) (string, error)
}

// --- calculator ---

type calculatorTool struct{}

func (calculatorTool) Name() string        { return "calculator" }
func (calculatorTool) Description() string { return "Evaluate a basic arithmetic expression (+ - * / and parentheses)." }
func (calculatorTool) InputSchema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"expression": {"type": "string", "description": "e.g. (2 + 3) * 4"}
		},
		"required": ["expression"]
	}`)
}

func (calculatorTool) Execute(input json.RawMessage) (string, error) {
	var args struct {
		Expression string `json:"expression"`
	}
	if err := json.Unmarshal(input, &args); err != nil {
		return "", fmt.Errorf("invalid input: %w", err)
	}

	result, err := EvalExpr(args.Expression)
	if err != nil {
		return "", fmt.Errorf("evaluating %q: %w", args.Expression, err)
	}
	return fmt.Sprintf("%v", result), nil
}

// --- read_file ---

// readFileTool reads text files, restricted to stay within Root, so an
// agent driven by model output can't be steered into reading arbitrary
// files on the machine via path traversal or absolute paths.
type readFileTool struct {
	Root string
}

func newReadFileTool(root string) (*readFileTool, error) {
	abs, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	return &readFileTool{Root: abs}, nil
}

func (readFileTool) Name() string { return "read_file" }
func (readFileTool) Description() string {
	return "Read the contents of a text file within the agent's working directory."
}
func (readFileTool) InputSchema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"path": {"type": "string", "description": "path relative to the working directory"}
		},
		"required": ["path"]
	}`)
}

func (t *readFileTool) Execute(input json.RawMessage) (string, error) {
	var args struct {
		Path string `json:"path"`
	}
	if err := json.Unmarshal(input, &args); err != nil {
		return "", fmt.Errorf("invalid input: %w", err)
	}

	resolved, err := t.resolve(args.Path)
	if err != nil {
		return "", err
	}

	data, err := os.ReadFile(resolved)
	if err != nil {
		return "", fmt.Errorf("reading %s: %w", args.Path, err)
	}
	return string(data), nil
}

// resolve joins path onto Root and rejects any result that escapes Root.
func (t *readFileTool) resolve(path string) (string, error) {
	joined := filepath.Join(t.Root, path)
	abs, err := filepath.Abs(joined)
	if err != nil {
		return "", err
	}

	rel, err := filepath.Rel(t.Root, abs)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("path %q escapes working directory", path)
	}
	return abs, nil
}
