package test

import (
	"os"
	"testing"
)

const (
	IntegrationTestEnvVarName = "INTEGRATION_TEST"
	NoTestCleanupEnvVarName   = "TEST_NOCLEANUP"
)

func ShouldRunIntegrationTest(t *testing.T) bool {
	if os.Getenv(IntegrationTestEnvVarName) != "TRUE" {
		t.Skip("Integration tests are skipped. To run them, set the environment variable INTEGRATION_TEST to TRUE.")
		return false
	}
	return true
}

func GetTestUrl() string {
	testUrl := os.Getenv("EVENT_REACTOR_TEST_URL")
	if testUrl == "" {
		return "http://app.event-reactor-127-0-0-1.nip.io"
	}
	return testUrl
}
