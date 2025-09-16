package cli

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/ivannovak/glide/internal/config"
)

// ExecuteYAMLCommand runs a YAML-defined command
func ExecuteYAMLCommand(cmdStr string, args []string) error {
	// Expand parameters
	expanded := config.ExpandCommand(cmdStr, args)

	// Check for multi-line commands (contains newlines or &&)
	if strings.Contains(expanded, "\n") || strings.Contains(expanded, "&&") {
		return executeMultiCommand(expanded)
	}

	// Single command execution
	return executeSingleCommand(expanded)
}

// executeSingleCommand runs a single command line
func executeSingleCommand(cmdStr string) error {
	// Check if it's a glid command (recursive call)
	if strings.HasPrefix(cmdStr, "glid ") || strings.HasPrefix(cmdStr, "glide ") {
		// Extract command and args
		parts := strings.Fields(cmdStr)
		if len(parts) > 1 {
			// For now, execute as shell command to avoid circular dependency
			// In the future, this could be integrated more deeply
			return executeShellCommand(cmdStr)
		}
	}

	// Execute as shell command
	return executeShellCommand(cmdStr)
}

// executeShellCommand runs a command through the shell
func executeShellCommand(cmdStr string) error {
	// Use sh -c to handle pipes, redirects, and other shell features
	cmd := exec.Command("sh", "-c", cmdStr)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	// Set environment to include current environment
	cmd.Env = os.Environ()

	return cmd.Run()
}

// executeMultiCommand handles multi-line or chained commands
func executeMultiCommand(cmdStr string) error {
	// Split by newlines and &&
	var commands []string

	// Handle newline-separated commands
	lines := strings.Split(cmdStr, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Further split by &&
		if strings.Contains(line, "&&") {
			parts := strings.Split(line, "&&")
			for _, part := range parts {
				part = strings.TrimSpace(part)
				if part != "" {
					commands = append(commands, part)
				}
			}
		} else {
			commands = append(commands, line)
		}
	}

	// Execute commands in sequence
	for i, cmd := range commands {
		// Show progress for multi-command sequences
		if len(commands) > 1 {
			fmt.Printf("→ [%d/%d] %s\n", i+1, len(commands), cmd)
		}

		if err := executeSingleCommand(cmd); err != nil {
			return fmt.Errorf("command failed: %s: %w", cmd, err)
		}
	}

	return nil
}