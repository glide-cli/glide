package sdk

import "context"

// ContextExtension represents additional context data provided by a plugin
// Plugins can contribute custom data to the project context that will be
// available to all commands and other plugins
type ContextExtension interface {
	// Name returns the unique identifier for this extension
	// Example: "docker", "kubernetes", "terraform"
	Name() string

	// Detect analyzes the project environment and returns extension data
	// The projectRoot is the detected root directory of the project
	// Returns nil if the extension is not applicable to this project
	Detect(ctx context.Context, projectRoot string) (interface{}, error)

	// Merge combines this extension's data with existing extension data
	// This is called when multiple plugins provide overlapping extensions
	// The existing parameter contains the current data for this extension
	// Returns the merged result
	Merge(existing interface{}, new interface{}) (interface{}, error)
}

// ContextProvider is the interface plugins implement to contribute context extensions
type ContextProvider interface {
	// ProvideContext returns the context extension provided by this plugin
	// Returns nil if the plugin does not provide context extensions
	ProvideContext() ContextExtension
}

// ExtensionRegistry manages registered context extensions
type ExtensionRegistry struct {
	extensions map[string]ContextExtension
}

// NewExtensionRegistry creates a new extension registry
func NewExtensionRegistry() *ExtensionRegistry {
	return &ExtensionRegistry{
		extensions: make(map[string]ContextExtension),
	}
}

// Register adds a context extension to the registry
func (r *ExtensionRegistry) Register(ext ContextExtension) error {
	if ext == nil {
		return nil
	}

	name := ext.Name()
	if name == "" {
		return ErrInvalidExtensionName
	}

	r.extensions[name] = ext
	return nil
}

// Get retrieves an extension by name
func (r *ExtensionRegistry) Get(name string) (ContextExtension, bool) {
	ext, ok := r.extensions[name]
	return ext, ok
}

// All returns all registered extensions
func (r *ExtensionRegistry) All() map[string]ContextExtension {
	// Return a copy to prevent external modification
	result := make(map[string]ContextExtension, len(r.extensions))
	for k, v := range r.extensions {
		result[k] = v
	}
	return result
}

// DetectAll runs detection for all registered extensions
func (r *ExtensionRegistry) DetectAll(ctx context.Context, projectRoot string) (map[string]interface{}, error) {
	results := make(map[string]interface{})

	for name, ext := range r.extensions {
		data, err := ext.Detect(ctx, projectRoot)
		if err != nil {
			// Continue with other extensions if one fails
			continue
		}
		if data != nil {
			results[name] = data
		}
	}

	return results, nil
}

// MergeExtensionData merges extension data from multiple sources
func MergeExtensionData(extensions []ContextExtension, dataMap map[string]interface{}) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	for _, ext := range extensions {
		name := ext.Name()
		if existing, ok := result[name]; ok {
			// Merge with existing data
			if newData, ok := dataMap[name]; ok {
				merged, err := ext.Merge(existing, newData)
				if err != nil {
					return nil, err
				}
				result[name] = merged
			}
		} else {
			// No existing data, just set it
			if data, ok := dataMap[name]; ok {
				result[name] = data
			}
		}
	}

	return result, nil
}
