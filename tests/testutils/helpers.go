package testutils

import (
	"os"
	"testing"
)

// MustGetEnvForTest fails the test if the environment variable is not set.
func MustGetEnvForTest(t *testing.T, key string) string {
	t.Helper()
	v := os.Getenv(key)
	if v == "" {
		t.Fatalf("%s not set", key)
	}
	return v
}
