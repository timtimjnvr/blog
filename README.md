# blog

A static site generator built in Go.

## Prerequisites

- Go 1.21+
- [Task](https://taskfile.dev/) (optional, for task automation)

## Setup

```bash
# Check if required tools are installed
task setup:check

# Install golangci-lint (for linting)
task setup

# Show instructions to install Task itself
task setup:task
```

## Quick Start

```bash
# Generate the site
go run .

# Or using Task
task generate
```

The generated site will be in the `build/` directory.

## Project Structure

```
blog/
├── main.go                      # Entry point
├── internal/
│   ├── context/                 # Page context interface and implementation
│   ├── substitution/            # Template substitution system
│   │   ├── registry.go          # Generic registry for substitutions
│   │   ├── title.go             # {{title}} substitution
│   │   └── content.go           # {{content}} substitution
│   ├── markdown/                # Markdown processing
│   │   ├── converter.go         # Goldmark wrapper
│   │   └── links.go             # Link conversion and path resolution
│   └── generator/               # Site generation
│       └── generator.go         # Main generation logic
├── content/                     # Markdown source files
├── models/
│   └── page.html                # HTML template
└── build/                       # Generated output
```

## Available Tasks

Run `task --list` to see all available tasks.

| Task | Description |
|------|-------------|
| `task setup` | Install required tools (golangci-lint) |
| `task setup:check` | Check if all tools are installed |
| `task setup:task` | Show instructions to install Task |
| `task` | Run all validation (fmt, lint, test, build) |
| `task validate` | Run all validation (fmt, lint, test, build) |
| `task test` | Run all tests |
| `task test:unit` | Run only unit tests (internal packages) |
| `task test:integration` | Run only integration tests |
| `task test:coverage` | Run tests with coverage report |
| `task test:coverage:html` | Generate HTML coverage report |
| `task build` | Build the binary |
| `task generate` | Generate the static site |
| `task dev` | Generate site and serve locally at :8080 |
| `task clean` | Remove build artifacts |
| `task ci` | Run CI pipeline |
| `task fmt` | Format Go files |
| `task lint` | Run golangci-lint (skipped if not installed) |

## Development

```bash
# Full validation before committing
task validate

# Generate and preview site locally
task dev
```
