package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestCalculatorToolExecute(t *testing.T) {
	tool := calculatorTool{}
	result, err := tool.Execute(json.RawMessage(`{"expression": "2 + 3 * 4"}`))
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if result != "14" {
		t.Errorf("Execute() = %q, want %q", result, "14")
	}
}

func TestCalculatorToolInvalidExpression(t *testing.T) {
	tool := calculatorTool{}
	if _, err := tool.Execute(json.RawMessage(`{"expression": "2 + "}`)); err == nil {
		t.Error("expected an error for an invalid expression")
	}
}

func TestReadFileToolReadsWithinRoot(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "note.txt"), []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}

	tool, err := newReadFileTool(dir)
	if err != nil {
		t.Fatal(err)
	}

	result, err := tool.Execute(json.RawMessage(`{"path": "note.txt"}`))
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if result != "hello" {
		t.Errorf("Execute() = %q, want %q", result, "hello")
	}
}

func TestReadFileToolRejectsPathEscape(t *testing.T) {
	dir := t.TempDir()
	tool, err := newReadFileTool(dir)
	if err != nil {
		t.Fatal(err)
	}

	_, err = tool.Execute(json.RawMessage(`{"path": "../../../../Windows/System32/drivers/etc/hosts"}`))
	if err == nil {
		t.Error("expected an error for a path that escapes the working directory")
	}
}

func TestReadFileToolMissingFile(t *testing.T) {
	dir := t.TempDir()
	tool, err := newReadFileTool(dir)
	if err != nil {
		t.Fatal(err)
	}

	_, err = tool.Execute(json.RawMessage(`{"path": "does-not-exist.txt"}`))
	if err == nil {
		t.Error("expected an error for a missing file")
	}
}
