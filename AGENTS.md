# AGENTS.md

## Project overview

`go-agent` is an interactive command-line coding agent written in Go. It uses
the Anthropic Go SDK with a Claude-compatible API, registers local tools, loads
repository skills, and runs a tool-calling loop with context compaction.

## Repository map

- `main.go`: configuration, client setup, tool registration, and the REPL.
- `internal/agent`: the main model/tool execution loop.
- `internal/llm`: Anthropic client requests and retry behavior.
- `internal/tool`: tool schemas, registration, permissions, and execution.
- `internal/tool/builtin`: built-in bash, filesystem, todo, skill, subagent, and
  compaction tools.
- `internal/compact`: context-size estimation and compaction strategies.
- `internal/config`: configuration loading and runtime directory setup.
- `internal/hooks`: lifecycle observers.
- `internal/memory`: persistent agent memory.
- `internal/session`: terminal input handling.
- `internal/skill` and `skills/`: skill discovery plus bundled skill content.
- `internal/consts`: shared tool names, limits, exit codes, and defaults.

## Development commands

The project requires Go 1.26 or newer. Prefer the existing Task targets when
Task is installed:

```sh
task build        # build build/go-agent
task run          # build and start the interactive REPL
task vet          # run go vet ./...
task fmt          # run gofmt -s -w .
task tidy         # run go mod tidy
task build-cross  # cross-compile; accepts GOOS and GOARCH
```

Equivalent direct checks are:

```sh
go test ./...
go vet ./...
go build ./...
```

Run `gofmt` on every changed Go file. Before handing off a code change, run at
least `go test ./...` and `go vet ./...`; also run `go build ./...` when startup,
configuration, or dependency wiring changes.

## Code conventions

- Keep packages focused and follow the current `internal/` boundaries.
- Use standard Go naming and error handling. Wrap errors with useful operation
  context, and preserve the underlying error with `%w` when callers may inspect
  it.
- Use the shared structured logger in `internal/logs` for diagnostics. Reserve
  `fmt.Print*` for deliberate user-facing REPL output.
- Keep tool names and shared limits in `internal/consts`; do not duplicate string
  literals across registration, permission, and execution code.
- Define tool inputs as typed structs with JSON and `jsonschema` tags. Register
  tools through `tool.RegisterTool`, add them to the appropriate registry in
  `internal/tool/builtin`, and consider separately whether subagents should
  receive the tool.
- Treat request-message mutation carefully: Anthropic tool-use blocks must be
  followed by matching tool-result blocks, and compaction must preserve that
  ordering.
- Protect package-level mutable state used across goroutines with synchronization
  consistent with the existing mutex-based patterns.
- Keep comments concise. Existing Chinese comments are acceptable; use the
  language that makes the surrounding code clearest and avoid translating
  unrelated comments.

## Testing guidance

- Place tests beside the implementation in `*_test.go` files.
- Prefer table-driven tests for parsing, permission rules, path validation,
  compaction thresholds, and tool input validation.
- Avoid live model/API calls in unit tests. Isolate network behavior behind
  replaceable functions or clients and use deterministic fakes.
- For filesystem tests, use `t.TempDir()` and do not write into the repository.
- Include failure cases, especially malformed configuration, denied paths,
  mismatched tool results, and context-limit boundaries.

## Configuration and generated data

- `gocode.json` takes precedence over environment variables. The fallback
  variables are `URL`, `API_KEY`, and `MODEL`; `LOG_LEVEL` controls optional
  logging behavior.
- Never commit API keys, `.env`, or `gocode.json`.
- Runtime state lives under `.gocode/`; build output lives under `build/`. Treat
  both as generated data and do not rely on their contents in tests.
- When adding a user-facing configuration field, update the config structs,
  validation/defaulting behavior, `gocode.json.example`, and `README.md` together.

## Change checklist

1. Keep the change scoped and avoid modifying unrelated user work.
2. Add or update tests for behavioral changes.
3. Run formatting and the relevant checks listed above.
4. Update `README.md`, examples, or `docs/TODO.md` when behavior or roadmap status
   changes.
5. Summarize what changed and report any checks that could not be run.
