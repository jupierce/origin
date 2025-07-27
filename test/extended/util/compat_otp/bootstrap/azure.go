package bootstrap

import (
	"os"

	exutil "github.com/openshift/origin/test/extended/util"
	e2e "k8s.io/kubernetes/test/e2e/framework"
)

const (
	// EnvVarSSHCloudPrivAzureUser stores the environment variable used to configure the Azure ssh user
	EnvVarSSHCloudPrivAzureUser = "SSH_CLOUD_PRIV_AZURE_USER"
)

// AzureBSInfoProvider implements interface BSInfoProvider
type AzureBSInfoProvider struct{}

// GetIPs returns the IPs of the boostrap machine if this machine exists in Azure
func (a AzureBSInfoProvider) GetIPs(oc *exutil.CLI) (*Ips, error) {
	// TODO: Implement this function properly once all dependencies are resolved
	e2e.Logf("Azure bootstrap GetIPs not implemented in compat_otp")
	return nil, &InstanceNotFound{"bootstrap"}
}

// GetSSHUser returns the ssh username of the boostrap machine if this machine exists in Azure
func (a AzureBSInfoProvider) GetSSHUser() string {
	user, exists := os.LookupEnv(EnvVarSSHCloudPrivAzureUser)
	if !exists {
		user = DefaultSSHUser
	}
	return user
}
