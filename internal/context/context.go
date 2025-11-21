package context

// Detect is a convenience function to detect the current project context
func Detect() *ProjectContext {
	detector, err := NewDetector()
	if err != nil {
		return &ProjectContext{
			WorkingDir: "", // We don't know the working directory
			Error:      err,
		}
	}

	// Extension registry will be set by the caller if plugins are available
	// This avoids import cycles with pkg/plugin

	ctx, err := detector.Detect()
	if err != nil {
		// Even if detection fails, return the context with basic info
		ctx.Error = err
	}
	return ctx
}

// DetectWithExtensions detects context with plugin-provided extensions
func DetectWithExtensions(extensionProviders []interface{}) *ProjectContext {
	detector, err := NewDetector()
	if err != nil {
		return &ProjectContext{
			WorkingDir: "", // We don't know the working directory
			Error:      err,
		}
	}

	// Set up extension registry from provided plugins
	if len(extensionProviders) > 0 {
		detector.SetExtensionRegistry(newPluginExtensionRegistry(extensionProviders))
	}

	ctx, err := detector.Detect()
	if err != nil {
		// Even if detection fails, return the context with basic info
		ctx.Error = err
	}
	return ctx
}

// newPluginExtensionRegistry creates an extension registry from provided plugins
func newPluginExtensionRegistry(providers []interface{}) ExtensionRegistry {
	return &pluginExtensionAdapter{
		providers: providers,
	}
}
