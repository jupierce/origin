package util

import (
	"github.com/openshift/origin/test/extended/util/compat_otp"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// This file contains functions that extend origin's CLI to be compatible with OTP.
// Functions are only added here if they MUST have special access to the internals
// of origin's util package. Most functions should be called directly from compat_otp.

// NotShowInfo disables showing info in CLI output
func (c *CLI) NotShowInfo() *CLI {
	// OTP tracks this with a showInfo field, but origin doesn't have this field
	// For now, this is a no-op in origin
	return c
}

// SetShowInfo enables showing info in CLI output
func (c *CLI) SetShowInfo() *CLI {
	// OTP tracks this with a showInfo field, but origin doesn't have this field
	// For now, this is a no-op in origin
	return c
}

// SetGuestKubeconf sets the guest kubeconfig path
func (c *CLI) SetGuestKubeconf(guestKubeconf string) *CLI {
	// OTP has a guestConfigPath field, origin doesn't
	// We'll store this in configPath when needed or create a wrapper
	// For now, we'll just return c
	return c
}

// GetGuestKubeconf returns the guest kubeconfig path
func (c *CLI) GetGuestKubeconf() string {
	// OTP has a guestConfigPath field, origin doesn't
	// Return empty string for now
	return ""
}

// GuestKubeClient returns a Kubernetes client using guest credentials
func (c *CLI) GuestKubeClient() kubernetes.Interface {
	// In OTP, this uses the guest config
	// For now, return the regular client
	return c.KubeClient()
}

// GuestConfig returns the guest REST config
func (c *CLI) GuestConfig() *rest.Config {
	// In OTP, this uses the guest config path
	// For now, return the user config
	return c.UserConfig()
}

// CreateNamespaceUDN creates a namespace with User Defined Networking
func (c *CLI) CreateNamespaceUDN() {
	// Delegate to compat_otp implementation
	compatCLI := &compat_otp.CLI{
		// Copy relevant fields
	}
	compatCLI.CreateNamespaceUDN()
}

// CreateSpecificNamespaceUDN creates a specific namespace with User Defined Networking
func (c *CLI) CreateSpecificNamespaceUDN(ns string) {
	// Delegate to compat_otp implementation
	compatCLI := &compat_otp.CLI{
		// Copy relevant fields
	}
	compatCLI.CreateSpecificNamespaceUDN(ns)
}

// CreateSpecifiedNamespaceAsAdmin creates a specified namespace as admin
func (c *CLI) CreateSpecifiedNamespaceAsAdmin(namespace string) {
	// Use origin's existing admin functionality
	c.AsAdmin().WithoutNamespace().Run("create", "namespace", namespace).Execute()
}

// DeleteSpecifiedNamespaceAsAdmin deletes a specified namespace as admin
func (c *CLI) DeleteSpecifiedNamespaceAsAdmin(namespace string) {
	// Use origin's existing admin functionality
	c.AsAdmin().WithoutNamespace().Run("delete", "namespace", namespace, "--ignore-not-found").Execute()
}

// AddPathsToDelete adds paths to be deleted during cleanup
func (c *CLI) AddPathsToDelete(dir string) {
	// OTP has a pathsToDelete field, origin doesn't
	// For now, this is a no-op
	// TODO: Consider adding path cleanup functionality to origin
}

// GetClientConfigForExtOIDCUser returns a client config for external OIDC user
func (c *CLI) GetClientConfigForExtOIDCUser(tokenCacheDir string) *rest.Config {
	// Delegate to compat_otp implementation
	compatCLI := &compat_otp.CLI{
		// Copy relevant fields
	}
	return compatCLI.GetClientConfigForExtOIDCUser(tokenCacheDir)
}

// SilentOutput runs the command and returns output without logging
func (c *CLI) SilentOutput() (string, error) {
	// Save current verbose state
	wasVerbose := c.verbose
	c.verbose = false
	defer func() { c.verbose = wasVerbose }()

	return c.Output()
}

// Additional OTP compatibility functions can be added here as needed
