# go-agent

A command-line coding agent built with Go and the Anthropic Claude API.

## Quick Start

### Requirements

- Go 1.26+
- A reachable Claude-compatible API endpoint and API key

### Configure Environment Variables

Provide model and auth info via environment variables:

```bash
export URL="https://api.anthropic.com"          # or a compatible endpoint
export API_KEY="sk-ant-..."                      # your API key
export MODEL="claude-3-5-sonnet-20241022"        # model name
export LOG_LEVEL=debug                           # optional, enable debug logging
```

Alternatively, copy `.env.example` to `.env` and fill in the values. The
`task` commands below load `.env` automatically.

### Build & Run

With [Task](https://taskfile.dev) (recommended):

```bash
# build
task build

# build and start the interactive agent (loads .env)
task run

# list all available tasks (vet, fmt, tidy, clean, ...)
task
```

Or with `go` directly:

```bash
# build
go build -o build/go-agent .

# run
./build/go-agent
```

Or run directly:

```bash
go run .
```

### Usage

After launch, you enter an interactive REPL:

```
Welcome to Go Agent! Type `/exit` to quit.
User >> list the Go files in the current directory
Agent:
 ...
User >> /exit
Bye!
```

Type `/exit` to quit. The model decides for itself which tool to call to accomplish the task.

## Roadmap

See [`docs/TODO.md`](docs/TODO.md)

## License

This project is licensed under the [MIT License](LICENSE), copyright © 2026 Yujie Zhou.
