# ADR-013: Dependency Injection Architecture

**Status:** Accepted
**Date:** 2025-11-26
**Deciders:** Engineering Team
**Related:** Phase 1 - Gold Standard Remediation

## Context

The current `pkg/app/application.go` implements a **Service Locator anti-pattern** where a God Object holds references to all major system dependencies. This approach has several critical problems:

### Current State (Application God Object)

```go
type Application struct {
    OutputManager  *output.Manager
    ProjectContext *context.ProjectContext
    Config         *config.Config
    ShellExecutor  *shell.Executor
    ConfigLoader   *config.Loader
    Writer         io.Writer
}
```

### Problems Identified

1. **No Lifecycle Management**
   - Dependencies created in arbitrary order in `main.go`
   - No graceful shutdown mechanism
   - Resource leaks possible (open files, connections, etc.)
   - No startup validation

2. **Hidden Dependencies and Circular Deps**
   - Dependency graph not explicit
   - Circular dependencies hidden until runtime
   - Hard to understand component relationships
   - Makes refactoring dangerous

3. **Testing Complexity**
   - Must bootstrap entire Application for unit tests
   - Hard to mock individual dependencies
   - Forces integration tests instead of unit tests
   - Slower test execution

4. **No Compile-Time Safety**
   - Dependencies can be nil (`ConfigLoader` lazily initialized)
   - Runtime panics instead of compile errors
   - No validation of dependency graph
   - Easy to forget initialization

5. **Maintenance Issues**
   - Adding new dependencies requires updating multiple locations
   - Hard to track what depends on what
   - Unclear initialization order requirements
   - Difficult to reason about system startup

## Decision

We will adopt **uber-fx** as our dependency injection framework and replace the `Application` God Object with a proper DI container pattern.

### Key Architectural Changes

1. **Introduce `pkg/container` Package**
   - Central location for dependency wiring
   - Explicit provider functions for each dependency
   - Lifecycle hook management
   - Testing support with override options

2. **Provider-Based Dependency Injection**
   - Each component gets a provider function
   - Dependencies declared via function parameters
   - fx automatically resolves dependency graph
   - Compile-time validation of dependency graph

3. **Lifecycle Management**
   - Startup hooks for initialization
   - Shutdown hooks for cleanup
   - Proper error handling during startup
   - Graceful shutdown on signals

4. **Backward Compatibility**
   - Keep `Application` type with deprecation warning
   - Implement using container internally
   - 6-month deprecation period
   - Remove in v3.0.0

## Alternatives Considered

### 1. Google Wire

**Pros:**
- Compile-time code generation
- Zero runtime overhead
- Static analysis friendly

**Cons:**
- Requires codegen step in build
- Less flexible than runtime DI
- More boilerplate
- Harder to override for testing

**Decision:** Rejected. The codegen requirement adds build complexity, and the testing story is worse than uber-fx.

### 2. Manual Dependency Injection

**Pros:**
- No external dependencies
- Complete control
- Explicit wiring

**Cons:**
- Massive boilerplate
- Error-prone
- No lifecycle management
- Hard to maintain

**Decision:** Rejected. Too much boilerplate, and we'd be reimplementing features that uber-fx provides.

### 3. Keep Service Locator Pattern

**Pros:**
- No changes required
- Familiar to team

**Cons:**
- All the problems listed in "Context" remain
- Doesn't solve any issues
- Technical debt accumulates

**Decision:** Rejected. This is the problem we're solving.

### 4. Simple Factory Functions

**Pros:**
- Minimal dependencies
- Easy to understand
- Explicit wiring

**Cons:**
- No lifecycle management
- Manual dependency ordering
- Testing still difficult
- No compile-time validation

**Decision:** Rejected. Doesn't address lifecycle and testing concerns.

## Consequences

### Positive

1. **Explicit Dependency Graph**
   - Every dependency declared in provider signature
   - Compile-time validation of graph
   - Easy to visualize system architecture
   - fx.Visualize() can generate diagrams

2. **Lifecycle Management**
   - Proper startup/shutdown hooks
   - Graceful error handling
   - Resource cleanup guaranteed
   - Startup validation

3. **Testing Improved**
   - Easy to mock any dependency
   - Override providers for tests
   - Faster unit tests (no full bootstrap)
   - Better test isolation

4. **Type Safety**
   - Compile-time detection of missing deps
   - No nil pointer panics
   - Better IDE support (autocomplete, refactoring)
   - Safer refactoring

5. **Maintainability**
   - Clear separation of concerns
   - Easy to add new dependencies
   - Dependency changes localized
   - Self-documenting architecture

### Negative

1. **External Dependency**
   - Add uber-fx to dependencies
   - Team needs to learn fx patterns
   - Another framework to maintain
   - **Mitigation:** fx is stable and well-maintained

2. **Migration Effort**
   - Need to refactor main.go
   - Update all dependency consumers
   - Test backward compatibility
   - **Mitigation:** Phased migration with backward compat shim

3. **Learning Curve**
   - Team needs to understand fx concepts
   - Provider pattern is new
   - Lifecycle hooks are new
   - **Mitigation:** Comprehensive ADR, examples, and documentation

4. **Slight Performance Overhead**
   - Reflection-based DI has minimal overhead
   - Container initialization slightly slower
   - **Mitigation:** Negligible for CLI tool (< 10ms)

### Risks and Mitigations

| Risk | Impact | Probability | Mitigation |
|------|--------|-------------|------------|
| Breaking changes | High | Medium | Backward compat shim for 6 months |
| Performance regression | Medium | Low | Benchmark container init (<10ms acceptable) |
| Plugin system conflicts | High | Medium | Plugins as fx providers with group tags |
| Team adoption | Medium | Medium | Training, examples, pairing sessions |
| Testing complexity | Medium | Low | Comprehensive testing helpers in container/options.go |

## Implementation Plan

See [DI Architecture Design Document](../specs/gold-standard-remediation/DI-ARCHITECTURE-DESIGN.md) for detailed implementation plan.

### High-Level Steps

1. **Phase 1: Foundation (Task 1.1 - 20h)**
   - Create `pkg/container` package
   - Implement core providers
   - Add lifecycle management
   - Create backward compatibility shim
   - Write ADR (this document)

2. **Phase 2: Migration (Task 1.2 - 24h)**
   - Update `cmd/glide/main.go` to use container
   - Migrate CLI builder to providers
   - Migrate plugin loading to providers
   - Integration testing

3. **Phase 3: Cleanup (Task 1.3 - 16h)**
   - Remove Application references from codebase
   - Update all tests to use container
   - Remove global state (output.SetGlobalManager)

4. **Phase 4: Deprecation (Task 1.4 - 8h)**
   - Mark Application as deprecated
   - Add migration guide
   - Update documentation

5. **Phase 5: Removal (v3.0.0)**
   - Remove deprecated Application type
   - Remove backward compat shims

## Examples

### Current Approach (Before)

```go
// cmd/glide/main.go
func main() {
    // Manual dependency creation (order matters!)
    cfg, err := config.Load()
    // ...

    ctx := context.DetectWithExtensions(plugins)

    app := app.NewApplication(
        app.WithProjectContext(ctx),
        app.WithConfig(cfg),
        app.WithOutputFormat(output.FormatTable, false, false),
    )

    // app is a God Object with all dependencies
    cli := cliPkg.New(app)
    // ...
}
```

### New Approach (After)

```go
// cmd/glide/main.go
func main() {
    ctx := context.Background()

    // Create DI container
    c, err := container.New()
    if err != nil {
        log.Fatal(err)
    }

    // Run with lifecycle management
    if err := c.Run(ctx, func() error {
        var rootCmd *cobra.Command
        // Extract root command from container
        if err := c.Invoke(func(cmd *cobra.Command) {
            rootCmd = cmd
        }); err != nil {
            return err
        }

        return rootCmd.Execute()
    }); err != nil {
        log.Fatal(err)
    }
}
```

### Provider Function Example

```go
// pkg/container/providers.go

// OutputManager depends on Writer and Logger
func provideOutputManager(writer io.Writer, logger *logging.Logger) *output.Manager {
    return output.NewManager(
        output.FormatTable,
        false,
        false,
        writer,
    )
}

// fx automatically injects writer and logger based on their types
```

### Testing Example

```go
// Before (requires full Application)
func TestFeature(t *testing.T) {
    app := app.NewApplication(
        app.WithConfig(testConfig),
        app.WithProjectContext(testContext),
        // ... must provide ALL dependencies
    )
    // ...
}

// After (override only what's needed)
func TestFeature(t *testing.T) {
    buf := bytes.NewBuffer(nil)

    c, err := container.New(
        container.WithWriter(buf), // Override only writer
        container.WithoutLifecycle(), // Skip lifecycle for speed
    )
    require.NoError(t, err)
    // ...
}
```

## Validation

### Acceptance Criteria

- [ ] Container package compiles and passes all tests
- [ ] All existing tests pass with backward compat shim
- [ ] Container initialization <10ms
- [ ] Coverage >90% on new code
- [ ] Zero linter warnings
- [ ] ADR approved by team
- [ ] Migration guide written

### Validation Commands

```bash
# Build and test
go build ./pkg/container/...
go test ./pkg/container/... -v -cover

# Test backward compatibility
go test ./pkg/app/... -v
go test ./... -v

# Benchmark container init
go test -bench=BenchmarkContainerInit ./pkg/container/

# Run linters
golangci-lint run ./pkg/container/...

# Check coverage
go test -coverprofile=coverage.out ./pkg/container/...
go tool cover -func=coverage.out | grep total
```

## References

- [uber-fx Documentation](https://uber-go.github.io/fx/)
- [Dependency Injection in Go](https://blog.drewolson.org/dependency-injection-in-go)
- [Google Wire Documentation](https://github.com/google/wire)
- [Gold Standard Remediation - DI Design](../specs/gold-standard-remediation/DI-ARCHITECTURE-DESIGN.md)
- [ADR-002: Plugin System Design](./ADR-002-plugin-system-design.md)
- [ADR-009: Command Builder Pattern](./ADR-009-command-builder-pattern.md)

## Appendix: Dependency Graph Comparison

### Before (Implicit Dependencies)

```
main.go
  ├─> config.Load() [manually called]
  ├─> context.Detect() [manually called]
  ├─> app.NewApplication() [manually called]
  │     ├─> output.NewManager() [internally created]
  │     ├─> shell.NewExecutor() [internally created]
  │     └─> config.NewLoader() [lazily created]
  └─> plugin.LoadAll() [manually called]
```

**Problems:** Hidden deps, manual ordering, lazy init, no validation

### After (Explicit Dependencies)

```
container.New()
  ├─> provideLogger() -> Logger [no deps]
  │
  ├─> provideWriter() -> Writer [no deps]
  │
  ├─> provideConfigLoader(Logger) -> Loader
  ├─> provideConfig(Loader, Logger) -> Config
  │
  ├─> provideContextDetector(Logger) -> Detector
  ├─> provideProjectContext(Detector, Plugins, Logger) -> ProjectContext
  │
  ├─> provideOutputManager(Writer, Logger) -> Manager
  │
  ├─> provideShellExecutor(Logger) -> Executor
  │
  └─> providePluginRegistry(Logger) -> Registry
```

**Benefits:** Explicit deps, automatic ordering, compile-time validation, proper lifecycle

---

**Status:** Accepted
**Version:** 1.0
**Last Updated:** 2025-11-26
