# goagent

A small tool-calling (ReAct-style) agent for the Anthropic Messages API,
written in plain Go with no third-party dependencies.

Given a question, it lets the model call tools, executes them locally,
feeds the results back, and repeats until the model gives a final
answer — the core loop behind most "AI agent" frameworks, in under 300
lines.

## Tools

- **calculator** — evaluates `+ - * /` and parenthesized arithmetic
  expressions ([calc.go](calc.go)).
- **read_file** — reads a text file, restricted to the `--workdir`
  directory tree. Paths that try to escape it (`../..`, absolute paths
  outside the root) are rejected, so a confused or adversarially-prompted
  model can't use it to read arbitrary files on the machine.

## Install

```
go install github.com/rafigzadasamir-prog/goagent@latest
```

Or build from source:

```
go build -o goagent .
```

## Usage

```
export ANTHROPIC_API_KEY=sk-ant-...
goagent "what is (14 + 6) * 3, and then read notes.txt and summarize it in one line"
```

```
goagent --workdir ./notes "what does the file todo.md say I should do first?"
```

## Design

- [client.go](client.go) — Messages API client with tool definitions
  and tool_use/tool_result content blocks.
- [agent.go](agent.go) — the loop: send, check `stop_reason`, execute
  any `tool_use` blocks, send `tool_result`s back, repeat (capped at 10
  iterations).
- [tools.go](tools.go) — the `Tool` interface and the two built-in
  tools.

[agent_test.go](agent_test.go) drives the loop against a local
`httptest` server that plays the model, so the tool-calling logic is
tested without needing a real API key.

## License

MIT
