<p align="center">
  <img src="docs/assets/glide-logotype.png" alt="Glide" width="400">
</p>

<h3 align="center">Streamline your development workflow with context-aware command orchestration</h3>

<p align="center">
  <a href="https://github.com/ivannovak/glide/releases"><img src="https://img.shields.io/github/v/release/ivannovak/glide?style=flat-square" alt="Release"></a>
  <a href="https://github.com/ivannovak/glide/actions"><img src="https://img.shields.io/github/actions/workflow/status/ivannovak/glide/ci.yml?branch=main&style=flat-square" alt="Build Status"></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/license-MIT-blue?style=flat-square" alt="License"></a>
</p>

---

## What is Glide?

Glide is a context-aware command orchestrator that adapts to your project environment, streamlining complex development workflows through an extensible plugin system. It detects what you're working on and provides the right tools at the right time.

### Why Glide?

- **🎯 Context-Aware**: Automatically detects your project type and provides relevant commands
- **🔌 Extensible**: Add custom commands through a powerful plugin system  
- **🌳 Worktree-Optimized**: First-class support for Git worktrees to work on multiple features simultaneously
- **⚡ Fast**: Written in Go for instant command execution
- **🛠️ Developer-First**: Built by developers, for developers who value efficient workflows

## Quick Start

### Install Glide

```bash
# macOS/Linux
curl -sSL https://raw.githubusercontent.com/ivannovak/glide/main/install.sh | bash

# Or download directly from releases
# https://github.com/ivannovak/glide/releases
```

### Your First Commands

```bash
# See what Glide detected about your project
glid context

# List all available commands
glid help

# Run a command (example with docker)
glid up        # Start your project
glid status    # Check project status
glid down      # Stop your project
```

## Core Concepts

### 🎭 Two Development Modes

Glide operates in two modes to match your workflow:

1. **Standard Mode** - Quick commands for immediate tasks
2. **Interactive Mode** - Full terminal sessions for complex operations

```bash
# Standard mode - quick command
glid status

# Interactive mode - when you need a full session
glid shell  # Opens interactive shell in your container
```

### 🔌 Plugin System

Extend Glide with custom commands specific to your team or project:

```bash
# List installed plugins
glid plugins list

# Plugins provide seamless commands
glid deploy staging    # From a deployment plugin
glid db backup        # From a database plugin
```

### 🌳 Worktree Support

Work on multiple features without context switching:

```bash
# Create a new worktree for a feature
glid worktree feature/new-feature

# Each worktree maintains its own environment
cd worktrees/feature-new-feature
glid up  # Isolated environment for this feature
```

## Documentation

### 🚀 Getting Started
- [**Installation Guide**](docs/getting-started/installation.md) - Get Glide running in 2 minutes
- [**First Steps**](docs/getting-started/first-steps.md) - Essential commands and concepts
- [**Project Setup**](docs/getting-started/project-setup.md) - Configure Glide for your project

### 📚 Learn More
- [**Core Concepts**](docs/core-concepts/README.md) - Understand how Glide works
- [**Common Workflows**](docs/guides/README.md) - Real-world usage patterns
- [**Plugin Development**](docs/plugin-development/README.md) - Create your own plugins

## Built-in Commands

Glide includes essential commands out of the box:

| Command | Description |
|---------|------------|
| `context` | Show detected project information |
| `help` | Display available commands |
| `version` | Show Glide version |
| `plugins` | Manage plugins |
| `worktree` | Manage Git worktrees |
| `self-update` | Update Glide to the latest version |

*Additional commands are provided by plugins based on your project context.*

## Philosophy

Glide follows these principles:

1. **Context is King** - Understand the environment and provide relevant tools
2. **Progressive Disclosure** - Show simple options first, reveal complexity as needed
3. **Extensible by Default** - Teams know their needs best
4. **Speed Matters** - Every millisecond counts in development workflows
5. **Respect Existing Tools** - Enhance, don't replace

## Contributing

We welcome contributions! See our [Contributing Guide](CONTRIBUTING.md) for details.

## Support

- **Issues**: [GitHub Issues](https://github.com/ivannovak/glide/issues)
- **Discussions**: [GitHub Discussions](https://github.com/ivannovak/glide/discussions)

## License

MIT License - see [LICENSE](LICENSE) for details.

---

<p align="center">
  <sub>Built with ❤️ by developers who were tired of typing the same commands over and over.</sub>
</p>