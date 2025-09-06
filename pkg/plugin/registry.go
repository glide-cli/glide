package plugin

import (
	"fmt"
	"sync"

	"github.com/spf13/cobra"
)

// Registry manages plugin registration and lifecycle
type Registry struct {
	mu      sync.RWMutex
	plugins map[string]Plugin
	aliases map[string]string // maps plugin alias to plugin name
	config  map[string]interface{}
}

// global registry instance
var globalRegistry = &Registry{
	plugins: make(map[string]Plugin),
	aliases: make(map[string]string),
	config:  make(map[string]interface{}),
}

// NewRegistry creates a new plugin registry
func NewRegistry() *Registry {
	return &Registry{
		plugins: make(map[string]Plugin),
		aliases: make(map[string]string),
		config:  make(map[string]interface{}),
	}
}

// Register adds a plugin to the global registry
func Register(p Plugin) error {
	return globalRegistry.Register(p)
}

// Register adds a plugin to the registry
func (r *Registry) Register(p Plugin) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if p == nil {
		return fmt.Errorf("cannot register nil plugin")
	}

	name := p.Name()
	if name == "" {
		return fmt.Errorf("plugin must have a name")
	}

	if _, exists := r.plugins[name]; exists {
		return fmt.Errorf("plugin %s already registered", name)
	}

	// Check if plugin name conflicts with existing alias
	if _, exists := r.aliases[name]; exists {
		return fmt.Errorf("plugin name %s conflicts with existing alias", name)
	}

	// Get plugin metadata to register aliases
	meta := p.Metadata()
	
	// Check for alias conflicts
	for _, alias := range meta.Aliases {
		if _, exists := r.plugins[alias]; exists {
			return fmt.Errorf("plugin alias %s conflicts with existing plugin", alias)
		}
		if _, exists := r.aliases[alias]; exists {
			return fmt.Errorf("plugin alias %s already registered", alias)
		}
	}

	// Register the plugin
	r.plugins[name] = p
	
	// Register all aliases
	for _, alias := range meta.Aliases {
		r.aliases[alias] = name
	}
	
	return nil
}

// Get returns a plugin by name or alias
func (r *Registry) Get(name string) (Plugin, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Check if it's a direct plugin name
	plugin, exists := r.plugins[name]
	if exists {
		return plugin, true
	}

	// Check if it's an alias
	if canonicalName, isAlias := r.aliases[name]; isAlias {
		return r.plugins[canonicalName], true
	}

	return nil, false
}

// List returns all registered plugins
func (r *Registry) List() []Plugin {
	r.mu.RLock()
	defer r.mu.RUnlock()

	plugins := make([]Plugin, 0, len(r.plugins))
	for _, p := range r.plugins {
		plugins = append(plugins, p)
	}
	return plugins
}

// SetConfig sets the configuration for all plugins
func (r *Registry) SetConfig(config map[string]interface{}) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.config = config
}

// LoadAll configures and registers all plugin commands
func (r *Registry) LoadAll(root *cobra.Command) error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for name, plugin := range r.plugins {
		// Configure the plugin
		if err := plugin.Configure(r.config); err != nil {
			return fmt.Errorf("failed to configure plugin %s: %w", name, err)
		}

		// Register plugin commands
		if err := plugin.Register(root); err != nil {
			return fmt.Errorf("failed to register plugin %s: %w", name, err)
		}
	}

	return nil
}

// Global registry functions

// GetGlobalRegistry returns the global plugin registry
func GetGlobalRegistry() *Registry {
	return globalRegistry
}

// List returns all plugins from the global registry
func List() []Plugin {
	return globalRegistry.List()
}

// Get returns a plugin from the global registry
func Get(name string) (Plugin, bool) {
	return globalRegistry.Get(name)
}

// LoadAll loads all plugins from the global registry
func LoadAll(root *cobra.Command) error {
	return globalRegistry.LoadAll(root)
}

// SetConfig sets configuration for the global registry
func SetConfig(config map[string]interface{}) {
	globalRegistry.SetConfig(config)
}

// ResolveAlias resolves a plugin alias to its canonical name
func (r *Registry) ResolveAlias(alias string) (string, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	canonical, ok := r.aliases[alias]
	return canonical, ok
}

// GetAliases returns all aliases for a given plugin
func (r *Registry) GetAliases(pluginName string) []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	plugin, exists := r.plugins[pluginName]
	if !exists {
		return nil
	}
	
	return plugin.Metadata().Aliases
}

// IsAlias checks if a given name is a plugin alias
func (r *Registry) IsAlias(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	_, ok := r.aliases[name]
	return ok
}
