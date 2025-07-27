// Package logext provides log extension utilities for tests
package logext

import (
	"fmt"
	"os"

	"github.com/onsi/ginkgo/v2"
)

// EnableDebugLog is the environment variable name for enabling debug logging
const EnableDebugLog = "GINKGO_TEST_ENABLE_DEBUG_LOG"

// Infof logs an info message
func Infof(format string, args ...interface{}) {
	fmt.Fprintf(ginkgo.GinkgoWriter, "[INFO] "+format+"\n", args...)
}

// Errorf logs an error message
func Errorf(format string, args ...interface{}) {
	fmt.Fprintf(ginkgo.GinkgoWriter, "[ERROR] "+format+"\n", args...)
}

// Debugf logs a debug message
func Debugf(format string, args ...interface{}) {
	if _, enabled := os.LookupEnv(EnableDebugLog); enabled {
		fmt.Fprintf(ginkgo.GinkgoWriter, "[DEBUG] "+format+"\n", args...)
	}
}

// Warnf logs a warning message
func Warnf(format string, args ...interface{}) {
	fmt.Fprintf(ginkgo.GinkgoWriter, "[WARN] "+format+"\n", args...)
}

// Stub package for OTP compatibility
// TODO: Migrate actual log extension functionality from OTP
