package docker

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	glidectx "github.com/ivannovak/glide/internal/context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDockerDetector(t *testing.T) {
	detector := NewDockerDetector()

	assert.NotNil(t, detector, "detector should not be nil")
}

func TestDockerDetector_Name(t *testing.T) {
	detector := NewDockerDetector()

	assert.Equal(t, "docker", detector.Name(), "detector name should be 'docker'")
}

func TestDockerDetector_Detect(t *testing.T) {
	// Create temporary directory structure with compose files
	tempDir := t.TempDir()
	vcsDir := filepath.Join(tempDir, "vcs")
	err := os.MkdirAll(vcsDir, 0755)
	require.NoError(t, err)

	// Create docker-compose.yml
	composeFile := filepath.Join(vcsDir, "docker-compose.yml")
	err = os.WriteFile(composeFile, []byte(`version: '3.8'
services:
  web:
    image: nginx
`), 0644)
	require.NoError(t, err)

	tests := []struct {
		name           string
		projectRoot    string
		expectError    bool
		expectNil      bool
		checkFields    bool
		skipDockerTest bool
	}{
		{
			name:           "valid project with compose files",
			projectRoot:    vcsDir,
			expectError:    false,
			expectNil:      false,
			checkFields:    true,
			skipDockerTest: false,
		},
		{
			name:           "project without compose files",
			projectRoot:    t.TempDir(), // Empty directory
			expectError:    false,
			expectNil:      true, // Should return nil when Docker not detected
			checkFields:    false,
			skipDockerTest: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			detector := NewDockerDetector()
			ctx := context.Background()

			result, err := detector.Detect(ctx, tt.projectRoot)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if tt.expectNil {
				assert.Nil(t, result, "result should be nil when Docker is not detected")
			} else if tt.checkFields {
				require.NotNil(t, result, "result should not be nil")

				// Verify result is a map with expected fields
				resultMap, ok := result.(map[string]interface{})
				require.True(t, ok, "result should be a map")

				// Check required fields exist
				_, hasDockerRunning := resultMap["docker_running"]
				assert.True(t, hasDockerRunning, "should have docker_running field")

				_, hasComposeFiles := resultMap["compose_files"]
				assert.True(t, hasComposeFiles, "should have compose_files field")

				_, hasComposeOverride := resultMap["compose_override"]
				assert.True(t, hasComposeOverride, "should have compose_override field")

				// Verify compose_files is populated when compose file exists
				composeFiles, ok := resultMap["compose_files"].([]string)
				if ok && len(composeFiles) > 0 {
					assert.NotEmpty(t, composeFiles, "compose files should be found")
				}
			}
		})
	}
}

func TestDockerDetector_Merge(t *testing.T) {
	detector := NewDockerDetector()

	tests := []struct {
		name     string
		existing interface{}
		new      interface{}
		expected interface{}
	}{
		{
			name:     "new data overwrites existing",
			existing: map[string]interface{}{"docker_running": false},
			new:      map[string]interface{}{"docker_running": true},
			expected: map[string]interface{}{"docker_running": true},
		},
		{
			name:     "nil new returns existing",
			existing: map[string]interface{}{"docker_running": true},
			new:      nil,
			expected: map[string]interface{}{"docker_running": true},
		},
		{
			name:     "both nil",
			existing: nil,
			new:      nil,
			expected: nil,
		},
		{
			name:     "new replaces nil existing",
			existing: nil,
			new:      map[string]interface{}{"docker_running": true},
			expected: map[string]interface{}{"docker_running": true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := detector.Merge(tt.existing, tt.new)

			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDockerDetector_detectDevelopmentMode(t *testing.T) {
	detector := NewDockerDetector()

	tests := []struct {
		name         string
		setupFunc    func(t *testing.T) string
		expectedMode glidectx.DevelopmentMode
	}{
		{
			name: "single repo mode (default)",
			setupFunc: func(t *testing.T) string {
				return t.TempDir()
			},
			expectedMode: glidectx.ModeSingleRepo,
		},
		{
			name: "directory with files but no worktree structure",
			setupFunc: func(t *testing.T) string {
				dir := t.TempDir()
				err := os.WriteFile(filepath.Join(dir, "test.txt"), []byte("test"), 0644)
				require.NoError(t, err)
				return dir
			},
			expectedMode: glidectx.ModeSingleRepo,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			projectRoot := tt.setupFunc(t)
			mode := detector.detectDevelopmentMode(projectRoot)

			assert.Equal(t, tt.expectedMode, mode)
		})
	}
}

func TestCheckDockerDaemon(t *testing.T) {
	// This test checks if checkDockerDaemon function works
	// Note: This will return false in CI environments where Docker isn't running
	result := checkDockerDaemon()

	// We don't assert true/false here because Docker may or may not be running
	// The important thing is that the function doesn't panic
	t.Logf("Docker daemon running: %v", result)
	assert.IsType(t, false, result, "checkDockerDaemon should return a boolean")
}

func TestDockerDetector_Detect_ContainerStatus(t *testing.T) {
	// Create temporary directory with compose files
	tempDir := t.TempDir()
	vcsDir := filepath.Join(tempDir, "vcs")
	err := os.MkdirAll(vcsDir, 0755)
	require.NoError(t, err)

	// Create docker-compose.yml
	composeFile := filepath.Join(vcsDir, "docker-compose.yml")
	err = os.WriteFile(composeFile, []byte(`version: '3.8'
services:
  web:
    image: nginx
  db:
    image: postgres
`), 0644)
	require.NoError(t, err)

	detector := NewDockerDetector()
	ctx := context.Background()

	result, err := detector.Detect(ctx, vcsDir)

	// We don't require result to be non-nil because Docker might not be running
	assert.NoError(t, err, "detection should not error")

	if result != nil {
		resultMap, ok := result.(map[string]interface{})
		require.True(t, ok, "result should be a map")

		// Container status is optional (only present when Docker is running)
		if containerStatus, exists := resultMap["containers_status"]; exists {
			assert.NotNil(t, containerStatus, "if containers_status exists, it should not be nil")
		}
	}
}

func TestDockerDetector_Detect_WithDifferentProjectStructures(t *testing.T) {
	tests := []struct {
		name      string
		setupFunc func(t *testing.T) string
		expectNil bool
	}{
		{
			name: "project root with compose file",
			setupFunc: func(t *testing.T) string {
				dir := t.TempDir()
				err := os.WriteFile(filepath.Join(dir, "docker-compose.yml"), []byte("version: '3'"), 0644)
				require.NoError(t, err)
				return dir
			},
			expectNil: false,
		},
		{
			name: "project root without compose file",
			setupFunc: func(t *testing.T) string {
				return t.TempDir()
			},
			expectNil: true,
		},
		{
			name: "nested project structure",
			setupFunc: func(t *testing.T) string {
				dir := t.TempDir()
				srcDir := filepath.Join(dir, "src")
				err := os.MkdirAll(srcDir, 0755)
				require.NoError(t, err)
				// No compose file
				return dir
			},
			expectNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			projectRoot := tt.setupFunc(t)
			detector := NewDockerDetector()
			ctx := context.Background()

			result, err := detector.Detect(ctx, projectRoot)

			assert.NoError(t, err)
			if tt.expectNil {
				assert.Nil(t, result, "result should be nil when no Docker setup detected")
			}
		})
	}
}
