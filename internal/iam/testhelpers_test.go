//go:build integration

package iam_test

import (
	"crypto/rand"
	"encoding/hex"
	"os"
	"strings"
	"testing"
)

// requireFloci skips the test unless FLOCI_ENDPOINT points at a reachable
// floci emulator.
func requireFloci(t *testing.T) {
	t.Helper()
	if os.Getenv("FLOCI_ENDPOINT") == "" {
		t.Skip("FLOCI_ENDPOINT not set; skipping integration test")
	}
}

// randomSuffix returns 8 hex chars suitable for namespacing per-test AWS
// resources so parallel runs and reruns don't collide.
func randomSuffix(t *testing.T) string {
	t.Helper()
	b := make([]byte, 4)
	if _, err := rand.Read(b); err != nil {
		t.Fatalf("rand.Read: %v", err)
	}
	return hex.EncodeToString(b)
}

// skipIfUnsupported skips the test when the error indicates the emulator
// doesn't implement the operation.
func skipIfUnsupported(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		return
	}
	msg := err.Error()
	if strings.Contains(msg, "UnsupportedOperation") ||
		strings.Contains(msg, "is not supported") ||
		strings.Contains(msg, "NotImplemented") {
		t.Skipf("operation not supported by emulator; skipping: %v", err)
	}
}
