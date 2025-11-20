package shell

import (
	"fmt"
	"sync"
)

// ExecutorProvider is the interface that plugins implement to provide custom executors
type ExecutorProvider interface {
	// Name returns the unique identifier for this executor
	// Example: "docker", "npm", "custom-tool"
	Name() string

	// CanHandle returns true if this executor can handle the given command
	CanHandle(cmd *Command) bool

	// CreateExecutor creates a new executor instance for the command
	CreateExecutor(options Options) CommandExecutor
}

// CommandExecutor is the interface that custom executors must implement
type CommandExecutor interface {
	// Execute runs the command and returns the result
	Execute(cmd *Command) (*Result, error)
}

// ExecutorRegistry manages registered executor providers
type ExecutorRegistry struct {
	providers map[string]ExecutorProvider
	mu        sync.RWMutex
}

// NewExecutorRegistry creates a new executor registry
func NewExecutorRegistry() *ExecutorRegistry {
	return &ExecutorRegistry{
		providers: make(map[string]ExecutorProvider),
	}
}

// Register adds an executor provider to the registry
func (r *ExecutorRegistry) Register(provider ExecutorProvider) error {
	if provider == nil {
		return nil
	}

	name := provider.Name()
	if name == "" {
		return fmt.Errorf("executor provider name cannot be empty")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.providers[name]; exists {
		return fmt.Errorf("executor provider %q already registered", name)
	}

	r.providers[name] = provider
	return nil
}

// Get retrieves an executor provider by name
func (r *ExecutorRegistry) Get(name string) (ExecutorProvider, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	provider, ok := r.providers[name]
	return provider, ok
}

// FindProvider finds the first provider that can handle the given command
func (r *ExecutorRegistry) FindProvider(cmd *Command) ExecutorProvider {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, provider := range r.providers {
		if provider.CanHandle(cmd) {
			return provider
		}
	}
	return nil
}

// All returns all registered executor providers
func (r *ExecutorRegistry) All() map[string]ExecutorProvider {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Return a copy to prevent external modification
	result := make(map[string]ExecutorProvider, len(r.providers))
	for k, v := range r.providers {
		result[k] = v
	}
	return result
}

// Unregister removes an executor provider from the registry
func (r *ExecutorRegistry) Unregister(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.providers, name)
}

// Clear removes all executor providers
func (r *ExecutorRegistry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.providers = make(map[string]ExecutorProvider)
}

// globalRegistry is the default global executor registry
var globalRegistry = NewExecutorRegistry()

// RegisterGlobal registers an executor provider in the global registry
func RegisterGlobal(provider ExecutorProvider) error {
	return globalRegistry.Register(provider)
}

// GetGlobalRegistry returns the global executor registry
func GetGlobalRegistry() *ExecutorRegistry {
	return globalRegistry
}
