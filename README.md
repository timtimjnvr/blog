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
├── tailwind.config.js           # Tailwind CSS configuration
├── styles/
│   └── input.css                # Tailwind CSS input file
├── styles.json                  # Optional custom styling configuration
├── internal/
│   ├── context/                 # Page context interface and implementation
│   ├── generator/               # Site generation
│   │   ├── generator.go         # Main generation logic
│   │   └── page.html            # Embedded HTML template
│   ├── markdown/                # Markdown processing
│   │   ├── converter.go         # Goldmark wrapper with styling support
│   │   └── links.go             # Link conversion and path resolution
│   ├── styling/                 # CSS styling system
│   │   ├── config.go            # Style configuration (element → CSS classes)
│   │   └── transformer.go       # Goldmark AST transformer for class injection
│   ├── substitution/            # Template substitution system
│   │   ├── registry.go          # Generic registry for substitutions
│   │   ├── title.go             # {{title}} substitution
│   │   └── content.go           # {{content}} substitution
│   └── validator/               # Post-generation validation
│       └── image.go             # Image reference validator
├── content/                     # Markdown source files
│   ├── home.md                  # Homepage (becomes index.html)
│   ├── posts/                   # Blog posts
│   └── assets/                  # Static assets (images, etc.)
└── target/build/                # Generated output
```

## Architecture

### Generation Pipeline

```
Markdown Files (content/)
    │
    ▼
┌─────────────────────────────────────────────────────────┐
│  Generator                                              │
│  ├─ Markdown Converter (Goldmark + GFM)                 │
│  │   └─ Style Transformer (optional CSS class injection)│
│  ├─ Substitution Registry                               │
│  │   ├─ {{title}} → First H1 from markdown              │
│  │   └─ {{content}} → Converted HTML with fixed links   │
│  └─ Validators                                          │
│      └─ Image Validator (checks local/remote images)    │
└─────────────────────────────────────────────────────────┘
    │
    ▼
HTML Files (target/build/)
```

### Styling System

The generator uses [Tailwind CSS](https://tailwindcss.com/) with the [Typography plugin](https://tailwindcss.com/docs/typography-plugin) for automatic prose styling. 
CSS is built using the [Tailwind Standalone CLI](https://tailwindcss.com/blog/standalone-cli).

* **Default behavior**: All Markdown content is wrapped in `<article class="prose prose-lg">`, which applies consistent typography styles to headings, paragraphs, links, code blocks, etc.

  * Tailwind configuration: Customize the typography styles in `tailwind.config.js`. See the [Typography plugin documentation](https://tailwindcss.com/docs/typography-plugin) for all available options.

* **Custom styling**: Create a `styles.json` file at the project root to add CSS classes to specific elements:

```json
{
  "elements": {
    "heading1": "text-4xl font-bold",
    "image": "rounded-lg shadow-md",
    "link": "text-blue-600 hover:underline",
    "blockquote": "border-l-4 border-gray-300 italic"
  },
  "contexts": {
    "post": {
      "heading1": "text-blue-900"
    }
  }
}
```

  * Supported element keys: `heading1`, `heading2`, `heading3`, `heading4`, `heading5`, `heading6`, `paragraph`, `link`, `image`, `codeblock`, `code`, `blockquote`, `list`, `listitem`.

  * Contexts: Files in `posts/` automatically get the `post` context, allowing context-specific styling.

  * Validation: Invalid keys cause the generator to exit with an error listing valid options.

* **Inline attributes**: For precise control on specific elements, use the inline attribute syntax directly in Markdown (take precedence over `styles.json`) :

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
