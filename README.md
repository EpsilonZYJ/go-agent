# go-agent

A command-line coding agent built with Go and the Anthropic Claude API.

## Quick Start

### Requirements

- Go 1.26+
- A reachable Claude-compatible API endpoint and API key

### Configuration

The agent supports two ways of providing model and auth info, with the
config file taking priority:

#### 1. JSON config file (`gocode.json`, preferred)

Create a `gocode.json` in the project root (the working directory):

```json
{
  "system": {
    "url": "https://api.anthropic.com",
    "api_key": "sk-ant-..."
  },
  "model": {
    "model_name": "claude-3-5-sonnet-20241022",
    "max_tokens": 16000
  }
}
```

| Field                  | Description                                   |
| ---------------------- | --------------------------------------------- |
| `system.url`           | Claude-compatible API endpoint                |
| `system.api_key`       | Your API key                                  |
| `model.model_name`     | Model name                                    |
| `model.max_tokens`     | Maximum tokens per response                   |

If `gocode.json` is present, it is loaded on startup and the environment
variables below are ignored.

#### 2. Environment variables (fallback)

If `gocode.json` is not found, the agent falls back to environment
variables:

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
