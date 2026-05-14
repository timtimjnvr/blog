# blog

A static site generator built in Go.

## Prerequisites

* Go 1.25+
* [Task](https://taskfile.dev/)
* [browser-sync](https://browsersync.io/) (optional, for `dev` live reload)

## Development

```bash
# Install needed tools
task setup

# Validate the Go generator
task validate

# Generate and validate the site
task generate

# Generate the site and serve it locally
task serve &

task dev
```
