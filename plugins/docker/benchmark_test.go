package docker

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	glidectx "github.com/ivannovak/glide/internal/context"
	"github.com/spf13/cobra"
)

// BenchmarkDockerPluginInitialization benchmarks plugin initialization time
// Target: < 10ms
func BenchmarkDockerPluginInitialization(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = NewDockerPlugin()
	}
}

// BenchmarkDockerDetection benchmarks Docker detection
// Target: < 50ms per detection
func BenchmarkDockerDetection(b *testing.B) {
	// Create temporary project with compose file
	tmpDir := b.TempDir()
	composeFile := filepath.Join(tmpDir, "docker-compose.yml")
	composeContent := `version: '3.8'
services:
  web:
    image: nginx:latest
`
	if err := os.WriteFile(composeFile, []byte(composeContent), 0644); err != nil {
		b.Fatalf("Failed to create compose file: %v", err)
	}

	p := NewDockerPlugin()
	extension := p.ProvideContext()
	ctx := context.Background()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := extension.Detect(ctx, tmpDir)
		if err != nil {
			b.Fatalf("Detection failed: %v", err)
		}
	}
}

// BenchmarkDockerDetectionWithoutCompose benchmarks detection when no compose files present
func BenchmarkDockerDetectionWithoutCompose(b *testing.B) {
	tmpDir := b.TempDir()

	p := NewDockerPlugin()
	extension := p.ProvideContext()
	ctx := context.Background()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _ = extension.Detect(ctx, tmpDir)
	}
}

// BenchmarkCommandRegistration benchmarks command registration
func BenchmarkCommandRegistration(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		p := NewDockerPlugin()
		root := &cobra.Command{Use: "glide"}
		err := p.Register(root)
		if err != nil {
			b.Fatalf("Registration failed: %v", err)
		}
	}
}

// BenchmarkContextExtensionMerge benchmarks extension data merging
func BenchmarkContextExtensionMerge(b *testing.B) {
	detector := NewDockerDetector()

	existing := map[string]interface{}{
		"docker_running": false,
		"compose_files":  []string{"old-compose.yml"},
	}

	new := map[string]interface{}{
		"docker_running": true,
		"compose_files":  []string{"docker-compose.yml"},
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := detector.Merge(existing, new)
		if err != nil {
			b.Fatalf("Merge failed: %v", err)
		}
	}
}

// BenchmarkCompatibilityLayerPopulate benchmarks populating compatibility fields
func BenchmarkCompatibilityLayerPopulate(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		ctx := &glidectx.ProjectContext{
			Extensions: map[string]interface{}{
				"docker": map[string]interface{}{
					"docker_running":   true,
					"compose_files":    []string{"docker-compose.yml"},
					"compose_override": "docker-compose.override.yml",
				},
			},
		}

		glidectx.PopulateCompatibilityFields(ctx)
	}
}

// BenchmarkCompatibilityLayerUpdate benchmarks updating extensions from compatibility fields
func BenchmarkCompatibilityLayerUpdate(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		ctx := &glidectx.ProjectContext{
			Extensions:      make(map[string]interface{}),
			DockerRunning:   true,
			ComposeFiles:    []string{"docker-compose.yml"},
			ComposeOverride: "docker-compose.override.yml",
		}

		glidectx.UpdateExtensionsFromCompatibility(ctx)
	}
}

// TestDetectionPerformance is a performance test (not a benchmark) that ensures
// Docker detection completes within acceptable time
func TestDetectionPerformance(t *testing.T) {
	// Create temporary project with compose file
	tmpDir := t.TempDir()
	composeFile := filepath.Join(tmpDir, "docker-compose.yml")
	composeContent := `version: '3.8'
services:
  web:
    image: nginx:latest
`
	if err := os.WriteFile(composeFile, []byte(composeContent), 0644); err != nil {
		t.Fatalf("Failed to create compose file: %v", err)
	}

	p := NewDockerPlugin()
	extension := p.ProvideContext()
	ctx := context.Background()

	// Warm up
	extension.Detect(ctx, tmpDir)

	// Measure actual detection time
	start := time.Now()
	_, err := extension.Detect(ctx, tmpDir)
	elapsed := time.Since(start)

	if err != nil {
		t.Errorf("Detection failed: %v", err)
	}

	// Target: < 200ms per detection (includes Docker daemon checks and docker-compose operations)
	// Note: This is slower than simple file detection because it involves actual Docker operations
	if elapsed > 200*time.Millisecond {
		t.Errorf("Detection too slow: %v (target: <200ms)", elapsed)
	} else {
		t.Logf("Detection time: %v (target: <200ms) ✓", elapsed)
	}
}

// TestPluginLoadingPerformance tests that plugin loading is fast
func TestPluginLoadingPerformance(t *testing.T) {
	measurements := make([]time.Duration, 10)

	for i := 0; i < 10; i++ {
		start := time.Now()
		_ = NewDockerPlugin()
		measurements[i] = time.Since(start)
	}

	// Calculate average
	var total time.Duration
	for _, d := range measurements {
		total += d
	}
	avg := total / time.Duration(len(measurements))

	// Target: < 10ms
	if avg > 10*time.Millisecond {
		t.Errorf("Plugin loading too slow: %v average (target: <10ms)", avg)
	} else {
		t.Logf("Plugin loading time: %v average (target: <10ms) ✓", avg)
	}
}

// TestCommandRegistrationPerformance tests that command registration is fast
func TestCommandRegistrationPerformance(t *testing.T) {
	p := NewDockerPlugin()

	start := time.Now()
	root := &cobra.Command{Use: "glide"}
	err := p.Register(root)
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("Registration failed: %v", err)
	}

	// Target: < 5ms
	if elapsed > 5*time.Millisecond {
		t.Errorf("Command registration too slow: %v (target: <5ms)", elapsed)
	} else {
		t.Logf("Command registration time: %v (target: <5ms) ✓", elapsed)
	}
}
