//go:build integration

package integration_test

import (
	"fmt"
	"os"
	"testing"
)

const testPostgresDSNEnv = "TEST_POSTGRES_DSN"

func TestMain(m *testing.M) {
	if os.Getenv(testPostgresDSNEnv) == "" {
		fmt.Fprintf(os.Stderr, "%s is not configured\n", testPostgresDSNEnv)
		os.Exit(1)
	}
	os.Exit(m.Run())
}
