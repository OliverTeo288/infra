//go:build integration

package iam_test

import (
	"context"
	"net"
	"strings"
	"testing"

	"raid/infra/internal/awscfg"
	"raid/infra/internal/iam"

	awssdkiam "github.com/aws/aws-sdk-go-v2/service/iam"
)

const (
	testGitlabHost    = "gitlab.example.com"
	testGitlabBaseURL = "https://gitlab.example.com"
	testOIDCSubject   = "project_path:group/subgroup/*:ref_type:branch:ref:main"
)

func TestSetupGitOpsRole_CreatesRoleAndPolicy(t *testing.T) {
	requireFloci(t)

	ctx := context.Background()
	roleName := "test-gitops-" + randomSuffix(t)

	err := iam.SetupGitOpsRole(ctx, nil,
		"floci-test", "us-east-1",
		roleName, testGitlabHost, testGitlabBaseURL, testOIDCSubject,
	)
	if err != nil {
		t.Fatalf("SetupGitOpsRole failed: %v", err)
	}

	cfg, err := awscfg.LoadConfig("floci-test", "us-east-1")
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}
	client := awssdkiam.NewFromConfig(cfg)

	got, err := client.GetRole(ctx, &awssdkiam.GetRoleInput{RoleName: &roleName})
	if err != nil {
		t.Fatalf("GetRole(%s) failed: %v", roleName, err)
	}
	if got.Role == nil || got.Role.Arn == nil {
		t.Fatalf("role response missing ARN")
	}
	if !strings.Contains(*got.Role.Arn, roleName) {
		t.Errorf("role ARN %q does not contain expected name %q", *got.Role.Arn, roleName)
	}
}

func TestSetupGitOpsRole_Idempotent(t *testing.T) {
	requireFloci(t)

	ctx := context.Background()
	roleName := "test-idempotent-" + randomSuffix(t)

	for i := 0; i < 2; i++ {
		err := iam.SetupGitOpsRole(ctx, nil,
			"floci-test", "us-east-1",
			roleName, testGitlabHost, testGitlabBaseURL, testOIDCSubject,
		)
		if err != nil {
			t.Fatalf("attempt %d: SetupGitOpsRole failed: %v", i+1, err)
		}
	}
}

// TestRegisterOIDCProvider_Idempotent verifies that the OIDC provider creation
// path can fetch a real TLS thumbprint AND register the provider in (floci's)
// IAM, and that a second call against the same URL is a no-op.
//
// GitHub Actions' OIDC issuer is used because it's publicly reachable with a
// valid cert chain — this is the only integration test that hits the public
// internet.
//
// Floci's IAM is "partial" — CreateOpenIDConnectProvider isn't implemented.
// The skipIfUnsupported guard skips gracefully on the emulator while still
// validating the full path against real AWS.
func TestRegisterOIDCProvider_Idempotent(t *testing.T) {
	requireFloci(t)

	const githubOIDC = "https://token.actions.githubusercontent.com"

	// FetchThumbprint does a real HTTPS handshake; skip if we have no
	// DNS resolution for the host (air-gapped CI runners).
	if _, err := net.LookupHost("token.actions.githubusercontent.com"); err != nil {
		t.Skipf("no DNS resolution for GitHub OIDC endpoint; skipping: %v", err)
	}

	ctx := context.Background()
	for i := 0; i < 2; i++ {
		err := iam.RegisterOIDCProvider(ctx, nil, "floci-test", "us-east-1", githubOIDC)
		if i == 0 {
			skipIfUnsupported(t, err)
		}
		if err != nil {
			t.Fatalf("attempt %d: RegisterOIDCProvider failed: %v", i+1, err)
		}
	}
}
