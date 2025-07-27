package bootstrap

import (
	"os"

	exutil "github.com/openshift/origin/test/extended/util"
	e2e "k8s.io/kubernetes/test/e2e/framework"
)

const (
	// EnvVarSSHCloudPrivAWSUser stores the environment variable used to configure the AWS ssh user
	EnvVarSSHCloudPrivAWSUser = "SSH_CLOUD_PRIV_AWS_USER"
)

// AWSBSInfoProvider implements interface BSInfoProvider
type AWSBSInfoProvider struct{}

// GetIPs returns the IPs of the boostrap machine if this machine exists in AWS
func (a AWSBSInfoProvider) GetIPs(oc *exutil.CLI) (*Ips, error) {
	// TODO: Implement this function properly once all dependencies are resolved
	e2e.Logf("AWS bootstrap GetIPs not implemented in compat_otp")
	return nil, &InstanceNotFound{"bootstrap"}
}

// GetSSHUser returns the ssh username of the boostrap machine if this machine exists in AWS
func (a AWSBSInfoProvider) GetSSHUser() string {
	userAWS := os.Getenv(EnvVarSSHCloudPrivAWSUser)
	if userAWS == "" {
		return DefaultSSHUser
	}
	return userAWS
}
