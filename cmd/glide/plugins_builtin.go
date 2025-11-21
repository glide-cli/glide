package main

import (
	"github.com/ivannovak/glide/pkg/plugin"
	"github.com/ivannovak/glide/plugins/docker"
)

// init registers all built-in plugins
// Built-in plugins are compiled into the glide binary
func init() {
	// Register Docker plugin
	if err := plugin.Register(docker.NewDockerPlugin()); err != nil {
		// Panic during init as this indicates a critical configuration error
		panic(err)
	}
}
