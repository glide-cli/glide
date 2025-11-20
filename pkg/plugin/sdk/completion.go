package sdk

import (
	"github.com/spf13/cobra"
)

// CompletionProvider is the interface plugins implement to provide shell completions
type CompletionProvider interface {
	// ProvideCompletions returns the completions provided by this plugin
	// The method returns a map of command name to completion function
	ProvideCompletions() map[string]CompletionFunc
}

// CompletionFunc is a function that provides completions for a command
// It receives the cobra.Command, args, and the current word being completed (toComplete)
// It should return completion suggestions and a ShellCompDirective
type CompletionFunc func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective)

// CompletionRegistry manages registered completion providers
type CompletionRegistry struct {
	completions map[string]CompletionFunc
}

// NewCompletionRegistry creates a new completion registry
func NewCompletionRegistry() *CompletionRegistry {
	return &CompletionRegistry{
		completions: make(map[string]CompletionFunc),
	}
}

// Register adds a completion function for a command
func (r *CompletionRegistry) Register(commandName string, fn CompletionFunc) error {
	if commandName == "" {
		return ErrInvalidCompletionProvider
	}

	if fn == nil {
		return ErrInvalidCompletionProvider
	}

	r.completions[commandName] = fn
	return nil
}

// Get retrieves a completion function for a command
func (r *CompletionRegistry) Get(commandName string) (CompletionFunc, bool) {
	fn, ok := r.completions[commandName]
	return fn, ok
}

// All returns all registered completion functions
func (r *CompletionRegistry) All() map[string]CompletionFunc {
	// Return a copy to prevent external modification
	result := make(map[string]CompletionFunc, len(r.completions))
	for k, v := range r.completions {
		result[k] = v
	}
	return result
}

// ApplyToCommand applies registered completions to a cobra command tree
func (r *CompletionRegistry) ApplyToCommand(rootCmd *cobra.Command) {
	// Walk through all commands and apply completions
	for cmdName, completionFn := range r.completions {
		if cmd, _, err := rootCmd.Find([]string{cmdName}); err == nil && cmd != nil {
			cmd.ValidArgsFunction = completionFn
		}
	}
}

// Helper functions for common completion patterns

// NoFileCompletion returns a completion directive that disables file completion
func NoFileCompletion() cobra.ShellCompDirective {
	return cobra.ShellCompDirectiveNoFileComp
}

// FileCompletion returns a completion directive that enables file completion
func FileCompletion() cobra.ShellCompDirective {
	return cobra.ShellCompDirectiveDefault
}

// DirectoryCompletion returns a completion directive that enables directory completion only
func DirectoryCompletion() cobra.ShellCompDirective {
	return cobra.ShellCompDirectiveFilterDirs
}

// StaticCompletion creates a completion function that returns a static list of options
func StaticCompletion(options []string) CompletionFunc {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return options, NoFileCompletion()
	}
}

// DynamicCompletion creates a completion function from a provider function
func DynamicCompletion(provider func() []string) CompletionFunc {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return provider(), NoFileCompletion()
	}
}

// FilePathCompletion creates a completion function that completes file paths
func FilePathCompletion() CompletionFunc {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return nil, FileCompletion()
	}
}

// DirectoryPathCompletion creates a completion function that completes directory paths
func DirectoryPathCompletion() CompletionFunc {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return nil, DirectoryCompletion()
	}
}

// ConditionalCompletion creates a completion function that uses different completions based on arg position
func ConditionalCompletion(completions map[int]CompletionFunc, defaultCompletion CompletionFunc) CompletionFunc {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		// Use arg position to determine which completion to use
		position := len(args)
		if fn, ok := completions[position]; ok {
			return fn(cmd, args, toComplete)
		}

		// Fall back to default
		if defaultCompletion != nil {
			return defaultCompletion(cmd, args, toComplete)
		}

		return nil, NoFileCompletion()
	}
}
