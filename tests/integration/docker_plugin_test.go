package integration_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	glidectx "github.com/ivannovak/glide/internal/context"
	"github.com/ivannovak/glide/pkg/plugin"
	"github.com/ivannovak/glide/plugins/docker"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDockerPluginEndToEnd tests the Docker plugin from end to end
func TestDockerPluginEndToEnd(t *testing.T) {
	// Skip if Docker is not available
	if err := exec.Command("docker", "info").Run(); err != nil {
		t.Skip("Docker is not available - skipping Docker plugin E2E tests")
		return
	}

	t.Run("plugin_initialization", func(t *testing.T) {
		p := docker.NewDockerPlugin()

		require.NotNil(t, p)
		assert.Equal(t, "docker", p.Name())
		assert.Equal(t, "1.0.0", p.Version())
		assert.NotEmpty(t, p.Description())
	})

	t.Run("plugin_metadata", func(t *testing.T) {
		p := docker.NewDockerPlugin()
		metadata := p.Metadata()

		assert.Equal(t, "docker", metadata.Name)
		assert.Equal(t, "1.0.0", metadata.Version)
		assert.NotEmpty(t, metadata.Commands)

		// Verify docker command is in metadata
		hasDockerCmd := false
		for _, cmd := range metadata.Commands {
			if cmd.Name == "docker" {
				hasDockerCmd = true
				break
			}
		}
		assert.True(t, hasDockerCmd, "metadata should include docker command")
	})

	t.Run("context_extension_detection", func(t *testing.T) {
		// Create temporary project with compose file
		tmpDir := t.TempDir()
		composeFile := filepath.Join(tmpDir, "docker-compose.yml")
		composeContent := `version: '3.8'
services:
  web:
    image: nginx:latest
  db:
    image: postgres:latest
`
		require.NoError(t, os.WriteFile(composeFile, []byte(composeContent), 0644))

		// Test context detection
		p := docker.NewDockerPlugin()
		extension := p.ProvideContext()

		require.NotNil(t, extension)
		assert.Equal(t, "docker", extension.Name())

		// Run detection
		ctx := context.Background()
		result, err := extension.Detect(ctx, tmpDir)

		assert.NoError(t, err)
		require.NotNil(t, result, "should detect Docker setup")

		// Verify result structure
		resultMap, ok := result.(map[string]interface{})
		require.True(t, ok, "result should be a map")

		// Check expected fields
		assert.Contains(t, resultMap, "docker_running")
		assert.Contains(t, resultMap, "compose_files")
		assert.Contains(t, resultMap, "compose_override")

		// Verify compose files detected
		if composeFiles, ok := resultMap["compose_files"].([]string); ok {
			assert.NotEmpty(t, composeFiles, "should find compose files")
		}
	})

	t.Run("command_registration", func(t *testing.T) {
		p := docker.NewDockerPlugin()
		root := &cobra.Command{
			Use: "glide",
		}

		err := p.Register(root)
		require.NoError(t, err)

		// Verify docker command was registered
		dockerCmd, _, err := root.Find([]string{"docker"})
		require.NoError(t, err)
		assert.NotNil(t, dockerCmd)
		assert.Equal(t, "docker", dockerCmd.Name())
	})

	t.Run("plugin_interface_compliance", func(t *testing.T) {
		p := docker.NewDockerPlugin()

		// Verify it implements Plugin interface
		var _ plugin.Plugin = p

		// Test all Plugin interface methods
		assert.Equal(t, "docker", p.Name())
		assert.NotEmpty(t, p.Version())
		assert.NotEmpty(t, p.Description())

		metadata := p.Metadata()
		assert.NotNil(t, metadata)

		err := p.Configure(nil)
		assert.NoError(t, err)
	})
}

// TestDockerPluginContextCompatibility tests backward compatibility with old context fields
func TestDockerPluginContextCompatibility(t *testing.T) {
	// Skip if Docker is not available
	if err := exec.Command("docker", "info").Run(); err != nil {
		t.Skip("Docker is not available - skipping Docker plugin compatibility tests")
		return
	}

	t.Run("compatibility_layer_populates_deprecated_fields", func(t *testing.T) {
		// Create temporary project with compose file
		tmpDir := t.TempDir()
		composeFile := filepath.Join(tmpDir, "docker-compose.yml")
		composeContent := `version: '3.8'
services:
  test:
    image: alpine:latest
`
		require.NoError(t, os.WriteFile(composeFile, []byte(composeContent), 0644))

		// Create context
		ctx := &glidectx.ProjectContext{
			ProjectRoot:     tmpDir,
			WorkingDir:      tmpDir,
			DevelopmentMode: glidectx.ModeSingleRepo,
			Location:        glidectx.LocationProject,
			Extensions:      make(map[string]interface{}),
		}

		// Get Docker plugin and run detection
		p := docker.NewDockerPlugin()
		extension := p.ProvideContext()

		goCtx := context.Background()
		result, err := extension.Detect(goCtx, tmpDir)
		assert.NoError(t, err)

		if result != nil {
			// Store in extensions
			ctx.Extensions["docker"] = result

			// Use compatibility layer to populate deprecated fields
			glidectx.PopulateCompatibilityFields(ctx)

			// Verify deprecated fields are populated
			assert.NotEmpty(t, ctx.ComposeFiles, "ComposeFiles should be populated")
			// DockerRunning might be false if daemon isn't running, but field should exist

			// Verify extensions still contain data
			assert.NotNil(t, ctx.Extensions["docker"])
		}
	})

	t.Run("extensions_to_compatibility_fields", func(t *testing.T) {
		// Create context with extension data
		ctx := &glidectx.ProjectContext{
			ProjectRoot:     "/test/project",
			WorkingDir:      "/test/project",
			DevelopmentMode: glidectx.ModeSingleRepo,
			Location:        glidectx.LocationProject,
			Extensions: map[string]interface{}{
				"docker": map[string]interface{}{
					"docker_running":   true,
					"compose_files":    []string{"docker-compose.yml"},
					"compose_override": "docker-compose.override.yml",
				},
			},
		}

		// Populate compatibility fields
		glidectx.PopulateCompatibilityFields(ctx)

		// Verify fields populated correctly
		assert.Equal(t, true, ctx.DockerRunning)
		assert.Equal(t, []string{"docker-compose.yml"}, ctx.ComposeFiles)
		assert.Equal(t, "docker-compose.override.yml", ctx.ComposeOverride)
	})

	t.Run("compatibility_fields_to_extensions", func(t *testing.T) {
		// Create context with deprecated fields
		ctx := &glidectx.ProjectContext{
			ProjectRoot:     "/test/project",
			WorkingDir:      "/test/project",
			DevelopmentMode: glidectx.ModeSingleRepo,
			Location:        glidectx.LocationProject,
			Extensions:      make(map[string]interface{}),
			// Set deprecated fields
			DockerRunning:   true,
			ComposeFiles:    []string{"docker-compose.yml"},
			ComposeOverride: "docker-compose.override.yml",
		}

		// Update extensions from compatibility fields
		glidectx.UpdateExtensionsFromCompatibility(ctx)

		// Verify extensions populated
		dockerExt, ok := ctx.Extensions["docker"]
		require.True(t, ok, "docker extension should exist")

		dockerData, ok := dockerExt.(map[string]interface{})
		require.True(t, ok, "docker extension should be a map")

		assert.Equal(t, true, dockerData["docker_running"])
		assert.Equal(t, []string{"docker-compose.yml"}, dockerData["compose_files"])
		assert.Equal(t, "docker-compose.override.yml", dockerData["compose_override"])
	})
}

// TestDockerPluginWorktreeModes tests plugin in different worktree configurations
func TestDockerPluginWorktreeModes(t *testing.T) {
	// Skip if Docker is not available
	if err := exec.Command("docker", "info").Run(); err != nil {
		t.Skip("Docker is not available - skipping Docker plugin worktree tests")
		return
	}

	t.Run("single_worktree_mode", func(t *testing.T) {
		// Create single repo structure
		tmpDir := t.TempDir()
		composeFile := filepath.Join(tmpDir, "docker-compose.yml")
		require.NoError(t, os.WriteFile(composeFile, []byte("version: '3'"), 0644))

		p := docker.NewDockerPlugin()
		extension := p.ProvideContext()

		ctx := context.Background()
		result, err := extension.Detect(ctx, tmpDir)

		assert.NoError(t, err)
		if result != nil {
			resultMap := result.(map[string]interface{})
			assert.Contains(t, resultMap, "compose_files")
		}
	})

	t.Run("multi_worktree_mode", func(t *testing.T) {
		// Create multi-worktree structure
		tmpDir := t.TempDir()
		vcsDir := filepath.Join(tmpDir, "vcs")
		worktreesDir := filepath.Join(tmpDir, "worktrees", "feature")

		require.NoError(t, os.MkdirAll(vcsDir, 0755))
		require.NoError(t, os.MkdirAll(worktreesDir, 0755))

		// Create compose files in both locations
		vcsCompose := filepath.Join(vcsDir, "docker-compose.yml")
		require.NoError(t, os.WriteFile(vcsCompose, []byte("version: '3'"), 0644))

		worktreeCompose := filepath.Join(worktreesDir, "docker-compose.yml")
		require.NoError(t, os.WriteFile(worktreeCompose, []byte("version: '3'"), 0644))

		p := docker.NewDockerPlugin()
		extension := p.ProvideContext()

		// Test detection from vcs directory
		ctx := context.Background()
		result, err := extension.Detect(ctx, vcsDir)
		assert.NoError(t, err)

		if result != nil {
			resultMap := result.(map[string]interface{})
			assert.Contains(t, resultMap, "compose_files")
		}

		// Test detection from worktree directory
		result2, err := extension.Detect(ctx, worktreesDir)
		assert.NoError(t, err)

		if result2 != nil {
			resultMap := result2.(map[string]interface{})
			assert.Contains(t, resultMap, "compose_files")
		}
	})
}

// TestDockerPluginWithoutDocker tests plugin behavior when Docker is not available
func TestDockerPluginWithoutDocker(t *testing.T) {
	t.Run("plugin_works_without_docker", func(t *testing.T) {
		// Plugin should initialize even if Docker isn't running
		p := docker.NewDockerPlugin()

		assert.NotNil(t, p)
		assert.Equal(t, "docker", p.Name())

		// Metadata should still be available
		metadata := p.Metadata()
		assert.NotEmpty(t, metadata.Commands)
	})

	t.Run("detection_gracefully_handles_missing_docker", func(t *testing.T) {
		// Create directory without compose files
		tmpDir := t.TempDir()

		p := docker.NewDockerPlugin()
		extension := p.ProvideContext()

		ctx := context.Background()
		result, err := extension.Detect(ctx, tmpDir)

		// Should not error, but may return nil
		assert.NoError(t, err)

		// Result may be nil when Docker setup not detected
		if result != nil {
			t.Logf("Result: %+v", result)
		}
	})
}

// TestDockerPluginExecutorProvider tests the executor provider functionality
func TestDockerPluginExecutorProvider(t *testing.T) {
	// Skip if Docker is not available
	if err := exec.Command("docker", "info").Run(); err != nil {
		t.Skip("Docker is not available - skipping executor provider tests")
		return
	}

	t.Run("executor_provider_exists", func(t *testing.T) {
		p := docker.NewDockerPlugin()
		executor := p.ProvideExecutor()

		require.NotNil(t, executor, "executor provider should not be nil")
	})
}
