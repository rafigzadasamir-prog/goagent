// Command goagent runs a small tool-calling agent against the
// Anthropic Messages API: it can evaluate arithmetic and read files
// within its working directory to help answer a single-shot question.
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

func main() {
	model := flag.String("model", "claude-sonnet-5", "Anthropic model name")
	workdir := flag.String("workdir", ".", "directory the read_file tool is restricted to")
	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "usage: goagent [--model ...] [--workdir ...] <question>")
		os.Exit(1)
	}
	question := strings.Join(args, " ")

	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		fmt.Fprintln(os.Stderr, "goagent: ANTHROPIC_API_KEY is not set")
		os.Exit(1)
	}

	readFile, err := newReadFileTool(*workdir)
	if err != nil {
		fmt.Fprintln(os.Stderr, "goagent:", err)
		os.Exit(1)
	}

	client := NewClient(apiKey, *model, "You are a helpful assistant with access to tools. Use them when needed, then give a concise final answer.")
	agent := NewAgent(client, []Tool{calculatorTool{}, readFile})

	answer, err := agent.Run(question)
	if err != nil {
		fmt.Fprintln(os.Stderr, "goagent:", err)
		os.Exit(1)
	}
	fmt.Println(answer)
}
