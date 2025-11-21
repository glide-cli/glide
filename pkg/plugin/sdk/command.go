package sdk

import (
	"github.com/spf13/cobra"
)

// PluginCommandDefinition defines a command that a plugin provides
type PluginCommandDefinition struct {
	// Name is the command name (e.g., "docker", "status")
	Name string

	// Use is the full usage string for the command
	Use string

	// Short is a brief description of the command
	Short string

	// Long is a detailed description of the command
	Long string

	// Example provides usage examples
	Example string

	// Aliases are alternative names for this command
	Aliases []string

	// Hidden indicates if the command should be hidden from help
	Hidden bool

	// Args specifies argument validation (optional)
	Args cobra.PositionalArgs

	// Flags defines the flags for this command
	Flags []FlagDefinition

	// RunE is the function to execute when the command is run
	// It receives the cobra.Command and args
	RunE func(cmd *cobra.Command, args []string) error

	// Subcommands are nested commands under this command
	Subcommands []*PluginCommandDefinition

	// PreRunE is executed before RunE (optional)
	PreRunE func(cmd *cobra.Command, args []string) error

	// PostRunE is executed after RunE (optional)
	PostRunE func(cmd *cobra.Command, args []string) error

	// Category is the command category for grouping in help
	Category string
}

// FlagDefinition defines a command flag
type FlagDefinition struct {
	// Name is the flag name (without dashes)
	Name string

	// Shorthand is the single-letter shorthand (optional)
	Shorthand string

	// Usage is the help text for this flag
	Usage string

	// Type is the flag data type (string, bool, int, etc.)
	Type string

	// Default is the default value for this flag
	Default interface{}

	// Required indicates if this flag must be provided
	Required bool

	// Hidden indicates if this flag should be hidden from help
	Hidden bool

	// Deprecated provides a deprecation message
	Deprecated string
}

// CommandProvider is the interface plugins implement to provide commands
type CommandProvider interface {
	// ProvideCommands returns the commands provided by this plugin
	ProvideCommands() []*PluginCommandDefinition
}

// ToCobraCommand converts a PluginCommandDefinition to a cobra.Command
func (d *PluginCommandDefinition) ToCobraCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:      d.Use,
		Short:    d.Short,
		Long:     d.Long,
		Example:  d.Example,
		Aliases:  d.Aliases,
		Hidden:   d.Hidden,
		Args:     d.Args,
		RunE:     d.RunE,
		PreRunE:  d.PreRunE,
		PostRunE: d.PostRunE,
	}

	// Set category if provided
	if d.Category != "" {
		cmd.Annotations = map[string]string{
			"category": d.Category,
		}
	}

	// Add flags
	for _, flag := range d.Flags {
		addFlagToCommand(cmd, flag)
	}

	// Add subcommands
	for _, subCmd := range d.Subcommands {
		cmd.AddCommand(subCmd.ToCobraCommand())
	}

	return cmd
}

// addFlagToCommand adds a flag to a cobra command based on its type
func addFlagToCommand(cmd *cobra.Command, flag FlagDefinition) {
	switch flag.Type {
	case "string":
		defaultVal, _ := flag.Default.(string)
		if flag.Shorthand != "" {
			cmd.Flags().StringP(flag.Name, flag.Shorthand, defaultVal, flag.Usage)
		} else {
			cmd.Flags().String(flag.Name, defaultVal, flag.Usage)
		}

	case "bool":
		defaultVal, _ := flag.Default.(bool)
		if flag.Shorthand != "" {
			cmd.Flags().BoolP(flag.Name, flag.Shorthand, defaultVal, flag.Usage)
		} else {
			cmd.Flags().Bool(flag.Name, defaultVal, flag.Usage)
		}

	case "int":
		defaultVal, _ := flag.Default.(int)
		if flag.Shorthand != "" {
			cmd.Flags().IntP(flag.Name, flag.Shorthand, defaultVal, flag.Usage)
		} else {
			cmd.Flags().Int(flag.Name, defaultVal, flag.Usage)
		}

	case "[]string":
		defaultVal, _ := flag.Default.([]string)
		if flag.Shorthand != "" {
			cmd.Flags().StringSliceP(flag.Name, flag.Shorthand, defaultVal, flag.Usage)
		} else {
			cmd.Flags().StringSlice(flag.Name, defaultVal, flag.Usage)
		}

	default:
		// Default to string type
		defaultVal, _ := flag.Default.(string)
		if flag.Shorthand != "" {
			cmd.Flags().StringP(flag.Name, flag.Shorthand, defaultVal, flag.Usage)
		} else {
			cmd.Flags().String(flag.Name, defaultVal, flag.Usage)
		}
	}

	// Mark as required if needed
	if flag.Required {
		cmd.MarkFlagRequired(flag.Name)
	}

	// Hide if needed
	if flag.Hidden {
		cmd.Flags().MarkHidden(flag.Name)
	}

	// Mark as deprecated if needed
	if flag.Deprecated != "" {
		cmd.Flags().MarkDeprecated(flag.Name, flag.Deprecated)
	}
}

// CommandRegistry manages registered commands from plugins
type CommandRegistry struct {
	commands map[string]*PluginCommandDefinition
}

// NewCommandRegistry creates a new command registry
func NewCommandRegistry() *CommandRegistry {
	return &CommandRegistry{
		commands: make(map[string]*PluginCommandDefinition),
	}
}

// Register adds a command to the registry
func (r *CommandRegistry) Register(cmd *PluginCommandDefinition) error {
	if cmd == nil {
		return nil
	}

	if cmd.Name == "" {
		return ErrInvalidCommandName
	}

	r.commands[cmd.Name] = cmd
	return nil
}

// Get retrieves a command by name
func (r *CommandRegistry) Get(name string) (*PluginCommandDefinition, bool) {
	cmd, ok := r.commands[name]
	return cmd, ok
}

// All returns all registered commands
func (r *CommandRegistry) All() map[string]*PluginCommandDefinition {
	// Return a copy to prevent external modification
	result := make(map[string]*PluginCommandDefinition, len(r.commands))
	for k, v := range r.commands {
		result[k] = v
	}
	return result
}

// AddToCobraCommand adds all registered commands to a cobra command
func (r *CommandRegistry) AddToCobraCommand(rootCmd *cobra.Command) {
	for _, cmdDef := range r.commands {
		cobraCmd := cmdDef.ToCobraCommand()
		rootCmd.AddCommand(cobraCmd)
	}
}
