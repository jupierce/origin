// Package bootstrap provides bootstrap-related utilities for tests
package bootstrap

import (
	"fmt"

	"github.com/openshift/origin/test/extended/util"
)

// GetBootstrap gets bootstrap machine information
func GetBootstrap(oc *util.CLI) (*Bootstrap, error) {
	// Simplified implementation - returns bootstrap info
	// In real implementation, this would:
	// 1. Find the bootstrap machine
	// 2. Get its IP address
	// 3. Set up SSH connection
	return &Bootstrap{
		SSH: &util.SshClient{
			Host:       "bootstrap.example.com",
			User:       "core",
			PrivateKey: "ssh-key",
			Port:       22,
		},
		IP:   "10.0.0.1",
		Name: "bootstrap",
	}, nil
}

// Stub package for OTP compatibility
// TODO: Migrate actual bootstrap functionality from OTP

// InstanceNotFound is returned when an instance is not found
type InstanceNotFound struct {
	InstanceName string
}

func (e *InstanceNotFound) Error() string {
	return fmt.Sprintf("instance not found: %s", e.InstanceName)
}

// Bootstrap represents bootstrap machine information
type Bootstrap struct {
	// SSH is the SSH client for connecting to bootstrap
	SSH *util.SshClient
	// Other bootstrap fields
	IP   string
	Name string
}
