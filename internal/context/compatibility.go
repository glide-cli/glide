package context

// PopulateCompatibilityFields populates the deprecated Docker fields from the extensions map
// This ensures backward compatibility with code that still uses the old Docker fields directly
func PopulateCompatibilityFields(ctx *ProjectContext) {
	if ctx.Extensions == nil {
		return
	}

	dockerData, ok := ctx.Extensions["docker"]
	if !ok {
		return
	}

	// Type assert to expected Docker context structure
	dockerCtx, ok := dockerData.(map[string]interface{})
	if !ok {
		return
	}

	// Populate ComposeFiles
	if composeFiles, ok := dockerCtx["compose_files"].([]string); ok {
		ctx.ComposeFiles = composeFiles
	}

	// Populate ComposeOverride
	if composeOverride, ok := dockerCtx["compose_override"].(string); ok {
		ctx.ComposeOverride = composeOverride
	}

	// Populate DockerRunning
	if dockerRunning, ok := dockerCtx["docker_running"].(bool); ok {
		ctx.DockerRunning = dockerRunning
	}

	// Populate ContainersStatus
	if containersStatus, ok := dockerCtx["containers_status"].(map[string]ContainerStatus); ok {
		ctx.ContainersStatus = containersStatus
	}
}

// UpdateExtensionsFromCompatibility updates the extensions map from the deprecated Docker fields
// This allows plugins to access Docker data through the extensions system while maintaining
// backward compatibility with code that sets the old fields
func UpdateExtensionsFromCompatibility(ctx *ProjectContext) {
	if ctx.Extensions == nil {
		ctx.Extensions = make(map[string]interface{})
	}

	// Only update if any Docker fields are set
	hasDockerData := len(ctx.ComposeFiles) > 0 ||
		ctx.ComposeOverride != "" ||
		ctx.DockerRunning ||
		len(ctx.ContainersStatus) > 0

	if !hasDockerData {
		return
	}

	dockerCtx := make(map[string]interface{})

	if len(ctx.ComposeFiles) > 0 {
		dockerCtx["compose_files"] = ctx.ComposeFiles
	}

	if ctx.ComposeOverride != "" {
		dockerCtx["compose_override"] = ctx.ComposeOverride
	}

	dockerCtx["docker_running"] = ctx.DockerRunning

	if len(ctx.ContainersStatus) > 0 {
		dockerCtx["containers_status"] = ctx.ContainersStatus
	}

	ctx.Extensions["docker"] = dockerCtx
}
