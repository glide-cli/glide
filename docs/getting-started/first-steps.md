# First Steps with Glide

Welcome! This guide will walk you through your first 5 minutes with Glide.

## Understanding Context

Glide's superpower is understanding your project context. Let's see what it detects:

```bash
cd your-project
glid context
```

Example output:
```
Project Context:
  Type: Docker Compose Project
  Language: Go
  Git Branch: main
  Working Directory: /Users/you/project
  Available Plugins: docker, git-tools
```

Glide detected your project type and loaded relevant plugins automatically.

## Essential Commands

### Getting Help

See all available commands:

```bash
glid help
```

Get help for a specific command:

```bash
glid help worktree
```

### Working with Your Project

If Glide detected a Docker project, you'll have commands like:

```bash
# Start your project
glid up

# Check status
glid status

# View logs
glid logs

# Stop everything
glid down
```

## The Two Modes

### Standard Mode
Most commands run quickly and return:

```bash
glid status
# Output appears, then back to your shell
```

### Interactive Mode
Some commands need a full terminal session:

```bash
glid shell
# You're now inside your container
# Type 'exit' to return
```

Interactive commands are perfect for:
- Debugging inside containers
- Running database clients
- Interactive testing sessions

## Working with Plugins

### See What's Available

```bash
# List installed plugins
glid plugins list

# See commands from a specific plugin
glid plugins info docker
```

### Plugin Commands Feel Native

Once a plugin is installed, its commands work like built-in ones:

```bash
# These might come from different plugins
glid db migrate        # Database plugin
glid test --watch      # Testing plugin
glid deploy staging    # Deployment plugin
```

## Your First Worktree

Glide makes working on multiple features easy with Git worktrees:

```bash
# Create a worktree for a new feature
glid worktree feature/awesome-feature

# The worktree is created in: worktrees/feature-awesome-feature
cd worktrees/feature-awesome-feature

# Start this feature's isolated environment
glid up
```

Now you can switch between features without stopping and restarting services!

## Configuration

Glide looks for configuration in this order:

1. Project-specific: `.glide.yml` in your project
2. Global: `~/.glide/config.yml`

Example `.glide.yml`:
```yaml
# Project-specific configuration
plugins:
  docker:
    compose_file: docker-compose.dev.yml
  
# Custom aliases
aliases:
  build: docker build --no-cache
  clean: docker system prune -af
```

## Tips for Success

### 1. Let Context Guide You
Don't memorize commands. Use `glid help` in different projects to see what's available.

### 2. Use Tab Completion
If you set up shell completion, double-tab shows available options:
```bash
glid <TAB><TAB>
```

### 3. Explore Gradually
Start with basic commands, then explore plugins as you need them.

### 4. Check Plugin Ecosystem
Many common workflows already have plugins:
```bash
glid plugins search database
```

## Common Workflows

### Morning Routine
```bash
# Update Glide
glid self-update

# Pull latest code
git pull

# Start your environment
glid up

# Check everything's running
glid status
```

### Switching Features
```bash
# Save current work
git commit -am "WIP"

# Switch to another feature
cd ~/project/worktrees/feature-other
glid up
```

### Debugging
```bash
# Check logs
glid logs --tail 50

# Jump into container
glid shell

# Run tests
glid test
```

## Next Steps

- Learn about [Core Concepts](../core-concepts/README.md)
- Set up Glide for [Your Project](project-setup.md)
- Explore [Common Workflows](../guides/README.md)

## Getting Help

- Run `glid help` for command reference
- Check the [guides](../guides/) for specific scenarios
- Visit [GitHub Issues](https://github.com/ivannovak/glide/issues) for support