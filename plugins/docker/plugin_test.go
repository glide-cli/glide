package docker

import (
	"testing"

	"github.com/ivannovak/glide/pkg/plugin"
	"github.com/ivannovak/glide/pkg/plugin/sdk"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDockerPlugin(t *testing.T) {
	p := NewDockerPlugin()

	require.NotNil(t, p, "plugin should not be nil")
	assert.NotNil(t, p.detector, "detector should be initialized")
	assert.NotNil(t, p.executor, "executor should be initialized")
}

func TestDockerPlugin_Name(t *testing.T) {
	p := NewDockerPlugin()

	assert.Equal(t, "docker", p.Name(), "plugin name should be 'docker'")
}

func TestDockerPlugin_Version(t *testing.T) {
	p := NewDockerPlugin()

	version := p.Version()
	assert.NotEmpty(t, version, "version should not be empty")
	assert.Equal(t, "1.0.0", version, "version should be 1.0.0")
}

func TestDockerPlugin_Description(t *testing.T) {
	p := NewDockerPlugin()

	desc := p.Description()
	assert.NotEmpty(t, desc, "description should not be empty")
	assert.Contains(t, desc, "Docker", "description should mention Docker")
}

func TestDockerPlugin_Metadata(t *testing.T) {
	p := NewDockerPlugin()

	metadata := p.Metadata()

	assert.Equal(t, "docker", metadata.Name, "metadata name should match plugin name")
	assert.Equal(t, "1.0.0", metadata.Version, "metadata version should match plugin version")
	assert.NotEmpty(t, metadata.Author, "author should not be empty")
	assert.NotEmpty(t, metadata.Description, "description should not be empty")
	assert.NotEmpty(t, metadata.Commands, "commands should be defined")

	// Check that at least the docker command is registered
	foundDockerCmd := false
	for _, cmd := range metadata.Commands {
		if cmd.Name == "docker" {
			foundDockerCmd = true
			assert.Equal(t, "Development", cmd.Category, "docker command should be in Development category")
			assert.Contains(t, cmd.Description, "Docker", "docker command description should mention Docker")
		}
	}
	assert.True(t, foundDockerCmd, "docker command should be in metadata")
}

func TestDockerPlugin_ProvideContext(t *testing.T) {
	p := NewDockerPlugin()

	extension := p.ProvideContext()

	require.NotNil(t, extension, "context extension should not be nil")

	// Verify it implements the ContextExtension interface
	var _ sdk.ContextExtension = extension

	assert.Equal(t, "docker", extension.Name(), "extension name should be 'docker'")
}

func TestDockerPlugin_ProvideCommands(t *testing.T) {
	p := NewDockerPlugin()

	commands := p.ProvideCommands()

	require.NotNil(t, commands, "commands should not be nil")
	assert.NotEmpty(t, commands, "should provide at least one command")

	// Verify at least one command is defined
	foundDockerCmd := false
	for _, cmd := range commands {
		if cmd != nil && cmd.Name == "docker" {
			foundDockerCmd = true
			assert.NotEmpty(t, cmd.Short, "docker command should have a short description")
		}
	}
	assert.True(t, foundDockerCmd, "docker command should be provided")
}

func TestDockerPlugin_ProvideExecutor(t *testing.T) {
	p := NewDockerPlugin()

	executor := p.ProvideExecutor()

	require.NotNil(t, executor, "executor should not be nil")
}

func TestDockerPlugin_Register(t *testing.T) {
	p := NewDockerPlugin()
	root := &cobra.Command{
		Use: "glide",
	}

	err := p.Register(root)

	require.NoError(t, err, "registration should not error")

	// Verify that commands were added to root
	commands := root.Commands()
	assert.NotEmpty(t, commands, "root should have commands after registration")

	// Check that docker command was registered
	foundDockerCmd := false
	for _, cmd := range commands {
		if cmd.Name() == "docker" {
			foundDockerCmd = true
		}
	}
	assert.True(t, foundDockerCmd, "docker command should be registered with root")
}

func TestDockerPlugin_Configure(t *testing.T) {
	p := NewDockerPlugin()

	tests := []struct {
		name   string
		config map[string]interface{}
		want   error
	}{
		{
			name:   "nil config",
			config: nil,
			want:   nil,
		},
		{
			name:   "empty config",
			config: map[string]interface{}{},
			want:   nil,
		},
		{
			name: "with config values",
			config: map[string]interface{}{
				"docker": map[string]interface{}{
					"compose_path": "/usr/local/bin/docker-compose",
				},
			},
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := p.Configure(tt.config)
			assert.Equal(t, tt.want, err)
		})
	}
}

func TestDockerPlugin_ImplementsPluginInterface(t *testing.T) {
	p := NewDockerPlugin()

	// Verify plugin implements the Plugin interface
	var _ plugin.Plugin = p
}

func TestDockerPlugin_ImplementsSDKInterfaces(t *testing.T) {
	p := NewDockerPlugin()

	// Verify plugin implements SDK interfaces
	var _ sdk.ContextProvider = p
	var _ sdk.CommandProvider = p
}
