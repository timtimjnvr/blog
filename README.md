# blog

A static site generator built in Go.

## Prerequisites

* Go 1.25+
* [Task](https://taskfile.dev/)
* [fswatch](https://github.com/emcrisostomo/fswatch) (optional, for `dev:watch` live reload)
* [browser-sync](https://browsersync.io/) (optional, for `dev:watch` live reload)

## Setup

```bash
task setup
```

## Quick Start

```bash
# Generate the site
task generate
```

The generated site will be in the `target/build/` directory.

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
└── target/                      # Generated output
```

## Available Tasks

Run `task --list` to see all available tasks.

## Development

```bash
# Full validation before committing
task validate

# Generate and preview site locally
task dev

# Generate, serve, and watch for changes
task dev:watch
```
