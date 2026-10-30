package main

import "fmt"

// maxIterations bounds the tool-call loop so a confused model can't
// make the agent spin forever.
const maxIterations = 10

// Agent runs a tool-calling loop against an Anthropic model: send the
// conversation, execute any tools the model asks for, feed the results
// back, and repeat until the model stops asking for tools.
type Agent struct {
	client   *Client
	tools    map[string]Tool
	toolDefs []ToolDef
}

// NewAgent builds an Agent with the given client and tool set.
func NewAgent(client *Client, tools []Tool) *Agent {
	a := &Agent{
		client: client,
		tools:  make(map[string]Tool, len(tools)),
	}
	for _, t := range tools {
		a.tools[t.Name()] = t
		a.toolDefs = append(a.toolDefs, ToolDef{
			Name:        t.Name(),
			Description: t.Description(),
			InputSchema: t.InputSchema(),
		})
	}
	return a
}

// Run sends userInput, resolves any tool calls the model makes, and
// returns the model's final text answer.
func (a *Agent) Run(userInput string) (string, error) {
	history := []Message{
		{Role: "user", Content: []ContentBlock{{Type: "text", Text: userInput}}},
	}

	for i := 0; i < maxIterations; i++ {
		resp, err := a.client.Send(history, a.toolDefs)
		if err != nil {
			return "", err
		}

		history = append(history, Message{Role: "assistant", Content: resp.Content})

		if resp.StopReason != "tool_use" {
			return extractText(resp.Content), nil
		}

		results := a.executeTools(resp.Content)
		history = append(history, Message{Role: "user", Content: results})
	}

	return "", fmt.Errorf("agent: exceeded %d tool-call iterations without a final answer", maxIterations)
}

// executeTools runs every tool_use block in content and returns the
// matching tool_result blocks.
func (a *Agent) executeTools(content []ContentBlock) []ContentBlock {
	var results []ContentBlock

	for _, block := range content {
		if block.Type != "tool_use" {
			continue
		}

		tool, ok := a.tools[block.Name]
		if !ok {
			results = append(results, ContentBlock{
				Type:      "tool_result",
				ToolUseID: block.ID,
				Content:   fmt.Sprintf("unknown tool: %s", block.Name),
				IsError:   true,
			})
			continue
		}

		output, err := tool.Execute(block.Input)
		if err != nil {
			results = append(results, ContentBlock{
				Type:      "tool_result",
				ToolUseID: block.ID,
				Content:   err.Error(),
				IsError:   true,
			})
			continue
		}

		results = append(results, ContentBlock{
			Type:      "tool_result",
			ToolUseID: block.ID,
			Content:   output,
		})
	}

	return results
}

func extractText(content []ContentBlock) string {
	var text string
	for _, block := range content {
		if block.Type == "text" {
			text += block.Text
		}
	}
	return text
}
