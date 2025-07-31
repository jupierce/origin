package util

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	crdv1 "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// This file contains functions that extend origin's CLI to be compatible with OTP.

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

// SetGuestKubeconf sets the guest cluster kubeconf file
func (c *CLI) SetGuestKubeconf(guestKubeconf string) *CLI {
	c.guestConfigPath = guestKubeconf
	return c
}

// GetGuestKubeconf gets the guest cluster kubeconf file
func (c *CLI) GetGuestKubeconf() string {
	return c.guestConfigPath
}

// GuestKubeClient provides a Kubernetes client for the guest cluster user.
func (c *CLI) GuestKubeClient() kubernetes.Interface {
	return kubernetes.NewForConfigOrDie(c.GuestConfig())
}

// GuestConfig provides a REST client config for the guest cluster user.
func (c *CLI) GuestConfig() *rest.Config {
	clientConfig, err := GetClientConfig(c.guestConfigPath)
	if err != nil {
		FatalErr(err)
	}
	return turnOffRateLimiting(clientConfig)
}

// CreateNamespaceUDN creates a new namespace with required user defined network label during creation time only
// required for testing networking UDN features on 4.17z+
func (c *CLI) CreateNamespaceUDN() {
	labels := map[string]string{
		"k8s.ovn.org/primary-user-defined-network": "udn-net",
	}
	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:   c.Namespace(),
			Labels: labels,
		},
	}
	_, err := c.AdminKubeClient().CoreV1().Namespaces().Create(context.Background(), namespace, metav1.CreateOptions{})
	if err != nil {
		FatalErr(err)
	}
}

// CreateSpecificNamespaceUDN creates a specific namespace with required user defined network label during creation time only
// required for testing networking UDN features on 4.17z+
// Important Note:  the namespace created by this function will not be automatically deleted, user need to explicitly delete the namespace after test is done
func (c *CLI) CreateSpecificNamespaceUDN(ns string) {
	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: ns,
			Labels: map[string]string{
				"k8s.ovn.org/primary-user-defined-network": "udn-net",
			},
		},
	}
	_, err := c.AdminKubeClient().CoreV1().Namespaces().Create(context.Background(), namespace, metav1.CreateOptions{})
	if err != nil {
		FatalErr(err)
	}
}

// CreateSpecifiedNamespaceAsAdmin creates specified name namespace.
func (c *CLI) CreateSpecifiedNamespaceAsAdmin(namespace string) {
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}
	_, err := c.AdminKubeClient().CoreV1().Namespaces().Create(context.Background(), ns, metav1.CreateOptions{})
	if err != nil {
		FatalErr(err)
	}
}

// DeleteSpecifiedNamespaceAsAdmin deletes specified name namespace.
func (c *CLI) DeleteSpecifiedNamespaceAsAdmin(namespace string) {
	err := c.AdminKubeClient().CoreV1().Namespaces().Delete(context.Background(), namespace, metav1.DeleteOptions{})
	if err != nil {
		// Log but don't fail if namespace doesn't exist
		fmt.Printf("Failed to delete namespace %s: %v\n", namespace, err)
	}
}

// AddPathsToDelete adds paths to be deleted after the test
func (c *CLI) AddPathsToDelete(dir string) {
	// OTP tracks paths to delete in a field, origin doesn't
	// For now, this is a no-op
}

// GetClientConfigForExtOIDCUser gets a client config for an external OIDC cluster
func (c *CLI) GetClientConfigForExtOIDCUser(tokenCacheDir string) *rest.Config {
	// This is a simplified implementation
	// In OTP this does more complex token caching
	return c.UserConfig()
}

// SilentOutput executes the command and returns stdout/stderr combined into one string
func (c *CLI) SilentOutput() (string, error) {
	// Save current verbose state
	wasVerbose := c.verbose
	c.verbose = false
	defer func() { c.verbose = wasVerbose }()

	return c.Output()
}

// AdminAPIExtensionsV1Client returns the API extensions v1 client
func (c *CLI) AdminAPIExtensionsV1Client() crdv1.ApiextensionsV1Interface {
	return crdv1.NewForConfigOrDie(c.AdminConfig())
}

// AsGuestKubeconf returns a CLI configured to use the guest kubeconfig
func (c *CLI) AsGuestKubeconf() *CLI {
	// Create a copy of the CLI with guest kubeconfig enabled
	copy := *c
	// In OTP this sets a flag and uses guestConfigPath
	// We'll use the guestConfigPath as the configPath
	if c.guestConfigPath != "" {
		copy.configPath = c.guestConfigPath
	}
	return &copy
}
