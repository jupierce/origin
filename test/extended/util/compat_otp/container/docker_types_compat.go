package container

import (
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
)

// Type aliases for Docker SDK compatibility
type (
	// Image types
	ImageListOptions   = image.ListOptions
	ImageRemoveOptions = image.RemoveOptions
	ImageSummary       = image.Summary

	// Container types
	ContainerRemoveOptions = container.RemoveOptions
	ContainerStartOptions  = container.StartOptions
	// ExecStartCheck is no longer in newer Docker SDK
)

// Helper functions - these are no longer needed with the new SDK
// The types package no longer has ImageListOptions, etc.
