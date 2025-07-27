package cloud

import (
	"k8s.io/kubernetes/test/e2e/framework"
)

// Config is an alias for framework.CloudConfig
type Config = framework.CloudConfig

// LoadConfig loads cloud configuration
func LoadConfig() (string, *Config, error) {
	// For OTP compatibility - load cloud config
	// Returns provider, config, error
	return "aws", &Config{}, nil
}
