// Package tokencmd provides token command utilities for tests
package tokencmd

import "k8s.io/client-go/rest"

// RequestTokenOptions represents options for requesting a token
type RequestTokenOptions struct {
	Server     string
	User       string
	Password   string
	ClientID   string
	Scopes     []string
	Insecure   bool
	CACert     string
	BasicAuth  bool
	OsinConfig *rest.Config
	Issuer     string
}

// NewRequestTokenOptions creates new request token options
func NewRequestTokenOptions(config *rest.Config, reader interface{}, server string, clientID string, basicAuth bool) *RequestTokenOptions {
	return &RequestTokenOptions{
		ClientID:   clientID,
		Scopes:     []string{"user:info"},
		Server:     server,
		BasicAuth:  basicAuth,
		OsinConfig: config,
	}
}

// Stub package for OTP compatibility
// TODO: Migrate actual token command functionality from OTP
