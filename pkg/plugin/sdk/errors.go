package sdk

import "errors"

var (
	// ErrInvalidExtensionName is returned when an extension has an empty name
	ErrInvalidExtensionName = errors.New("extension name cannot be empty")

	// ErrInvalidExecutorName is returned when an executor has an empty name
	ErrInvalidExecutorName = errors.New("executor name cannot be empty")

	// ErrInvalidCommandName is returned when a command has an empty name
	ErrInvalidCommandName = errors.New("command name cannot be empty")

	// ErrInvalidCompletionProvider is returned when a completion provider is invalid
	ErrInvalidCompletionProvider = errors.New("invalid completion provider")
)
