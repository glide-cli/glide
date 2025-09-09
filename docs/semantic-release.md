# Semantic Release Integration

## Overview

This project uses [semantic-release](https://semantic-release.gitbook.io/) to automate the version management and package publishing process. Semantic-release analyzes commit messages to determine the type of version bump (major, minor, or patch) and automatically creates releases.

## How It Works

1. **Commit Analysis**: When code is pushed to `main`, semantic-release analyzes commit messages
2. **Version Determination**: Based on conventional commits, it determines the next version:
   - `fix:` commits trigger a patch release (1.0.0 → 1.0.1)
   - `feat:` commits trigger a minor release (1.0.0 → 1.1.0)
   - `BREAKING CHANGE:` triggers a major release (1.0.0 → 2.0.0)
3. **Automated Release**: If a release is needed, it will:
   - Update version in `pkg/version/version.go`
   - Generate/update `CHANGELOG.md`
   - Create a git tag
   - Create a GitHub release
   - Trigger the build workflow

## Commit Message Format

Follow the [Conventional Commits](https://www.conventionalcommits.org/) specification:

```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

### Common Types

- `feat`: A new feature (triggers minor release)
- `fix`: A bug fix (triggers patch release)
- `docs`: Documentation changes (no release)
- `style`: Code style changes (no release)
- `refactor`: Code refactoring (no release)
- `perf`: Performance improvements (triggers patch release)
- `test`: Test changes (no release)
- `build`: Build system changes (no release)
- `ci`: CI configuration changes (no release)
- `chore`: Other changes (no release)
- `revert`: Revert a previous commit (triggers patch release)

### Breaking Changes

To trigger a major version bump, include `BREAKING CHANGE:` in the commit body or footer:

```
feat: remove support for Go 1.20

BREAKING CHANGE: Minimum Go version is now 1.21
```

Or use `!` after the type:

```
feat!: remove deprecated API endpoints
```

## Configuration

### `.releaserc.json`

The semantic-release configuration includes:
- **Commit analyzer**: Analyzes commits to determine version bump
- **Release notes generator**: Creates formatted release notes
- **Changelog**: Updates `CHANGELOG.md`
- **Version updater**: Updates version in `pkg/version/version.go`
- **Git**: Commits version changes back to repository
- **GitHub**: Creates GitHub releases

### GitHub Workflow

The `.github/workflows/semantic-release.yml` workflow:
- Runs on pushes to `main` branch
- Can be manually triggered via workflow_dispatch
- Requires `SEMANTIC_RELEASE_TOKEN` secret or uses `GITHUB_TOKEN`

## Local Testing

To test semantic-release locally without making actual releases:

```bash
# Install dependencies
npm ci

# Dry run (no actual release)
npx semantic-release --dry-run --no-ci
```

## Workflow Integration

The semantic-release workflow integrates with the existing release workflow:

1. **Semantic Release** (`semantic-release.yml`):
   - Analyzes commits on push to main
   - Creates version tag if release needed
   - Updates CHANGELOG.md and version.go

2. **Build & Release** (`release.yml`):
   - Triggered by version tags created by semantic-release
   - Builds cross-platform binaries
   - Creates Docker images
   - Publishes GitHub release with artifacts

## Troubleshooting

### No Release Created

If commits aren't triggering releases:
- Ensure commit messages follow conventional format
- Check that commits contain `feat:` or `fix:` prefixes
- Verify the workflow has proper permissions

### Version Not Updated

If `pkg/version/version.go` isn't updated:
- Check the replace plugin configuration in `.releaserc.json`
- Ensure the regex pattern matches the version variable

### Workflow Failures

Common issues:
- Missing `npm ci` before running semantic-release
- Insufficient GitHub token permissions
- Network issues accessing npm registry

## GitHub Token Requirements

The `SEMANTIC_RELEASE_TOKEN` or `GITHUB_TOKEN` needs:
- `contents: write` - To create tags and releases
- `issues: write` - To comment on issues
- `pull-requests: write` - To comment on PRs

## Manual Release

To manually trigger a release:
1. Go to Actions → Semantic Release
2. Click "Run workflow"
3. Select the `main` branch
4. Click "Run workflow"

This will analyze commits since the last release and create a new release if needed.