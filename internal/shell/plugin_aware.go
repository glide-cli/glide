package shell

import "context"

// PluginAwareExecutor wraps the standard executor with plugin-provided executor support
type PluginAwareExecutor struct {
	standardExecutor *Executor
	registry         *ExecutorRegistry
}

// NewPluginAwareExecutor creates a new plugin-aware executor
func NewPluginAwareExecutor(options Options) *PluginAwareExecutor {
	return &PluginAwareExecutor{
		standardExecutor: NewExecutor(options),
		registry:         GetGlobalRegistry(),
	}
}

// NewPluginAwareExecutorWithRegistry creates a plugin-aware executor with a custom registry
func NewPluginAwareExecutorWithRegistry(options Options, registry *ExecutorRegistry) *PluginAwareExecutor {
	return &PluginAwareExecutor{
		standardExecutor: NewExecutor(options),
		registry:         registry,
	}
}

// Execute runs a command, delegating to plugin executors when available
func (e *PluginAwareExecutor) Execute(cmd *Command) (*Result, error) {
	// Check if a plugin can handle this command
	provider := e.registry.FindProvider(cmd)
	if provider != nil {
		// Use plugin-provided executor
		executor := provider.CreateExecutor(e.standardExecutor.options)
		return executor.Execute(cmd)
	}

	// Fall back to standard executor
	return e.standardExecutor.Execute(cmd)
}

// ExecuteWithContext runs a command with context, delegating to plugin executors when available
func (e *PluginAwareExecutor) ExecuteWithContext(ctx context.Context, cmd *Command) (*Result, error) {
	// Check if a plugin can handle this command
	provider := e.registry.FindProvider(cmd)
	if provider != nil {
		// Use plugin-provided executor
		executor := provider.CreateExecutor(e.standardExecutor.options)

		// Check if executor supports context
		if ctxExecutor, ok := executor.(ContextAwareExecutor); ok {
			return ctxExecutor.ExecuteWithContext(ctx, cmd)
		}

		// Fall back to Execute without context
		return executor.Execute(cmd)
	}

	// Fall back to standard executor with context
	return e.standardExecutor.ExecuteWithContext(ctx, cmd)
}

// ContextAwareExecutor is an optional interface for executors that support context
type ContextAwareExecutor interface {
	CommandExecutor
	ExecuteWithContext(ctx context.Context, cmd *Command) (*Result, error)
}

// Run is a convenience method for simple command execution
func (e *PluginAwareExecutor) Run(name string, args ...string) error {
	cmd := NewPassthroughCommand(name, args...)
	result, err := e.Execute(cmd)
	if err != nil {
		return err
	}
	if result.Error != nil {
		return result.Error
	}
	if result.ExitCode != 0 {
		return result.Error
	}
	return nil
}

// RunCapture runs a command and returns captured output
func (e *PluginAwareExecutor) RunCapture(name string, args ...string) (string, error) {
	cmd := NewCommand(name, args...)
	result, err := e.Execute(cmd)
	if err != nil {
		return "", err
	}
	if result.Error != nil {
		return "", result.Error
	}
	if result.ExitCode != 0 {
		return string(result.Stderr), result.Error
	}
	return string(result.Stdout), nil
}

// SetRegistry allows changing the executor registry
func (e *PluginAwareExecutor) SetRegistry(registry *ExecutorRegistry) {
	e.registry = registry
}

// GetRegistry returns the current executor registry
func (e *PluginAwareExecutor) GetRegistry() *ExecutorRegistry {
	return e.registry
}
