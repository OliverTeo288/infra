// Package flows orchestrates the higher-level user-facing operations of the
// CLI: creating IAM roles, S3 buckets, scaffolding a project, port-forwarding,
// and ECS Exec. Each public function is exactly one `infra <command>`
// workflow and runs end-to-end on a single context with a single session ID.
package flows

import (
	"context"
	"errors"
	"fmt"
	"regexp"

	"raid/infra/internal/config"
	"raid/infra/internal/iam"
	"raid/infra/internal/observability"
	"raid/infra/internal/prompt"
)

// roleNameRegex enforces the IAM-naming subset we accept from users.
var roleNameRegex = regexp.MustCompile(`^[a-zA-Z0-9\-]+$`)

func validateRoleName(input string) error {
	if !roleNameRegex.MatchString(input) {
		return errors.New("role name must only contain alphanumeric characters or dashes, and no spaces")
	}
	return nil
}

// CreateGitOpsRole walks the user through naming and provisioning the GitOps
// IAM role plus its OIDC provider for GitLab CI.
func CreateGitOpsRole(ctx context.Context, profile, region string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	if cfg.OIDCSubjectPattern == "" {
		return fmt.Errorf("OIDCSubjectPattern is not configured (missing build-time variable)")
	}

	roleName, err := prompt.Input("Enter the name of IAM Role for Terraform Gitops", validateRoleName, cfg.GitOpsRoleDefault)
	if err != nil {
		return err
	}

	// Bind role-level attributes once so they appear on every log line.
	ctx = observability.WithAttrs(ctx,
		"role_name", roleName,
		"gitlab_host", cfg.GitlabHost,
	)
	observability.FromContext(ctx).Info("gitops_role_provisioning")

	audit := observability.AuditFunc()
	if err := iam.SetupGitOpsRole(ctx, audit, profile, region, roleName, cfg.GitlabHost, cfg.GitlabBaseURL, cfg.OIDCSubjectPattern); err != nil {
		return err
	}
	return iam.RegisterOIDCProvider(ctx, audit, profile, region, cfg.GitlabBaseURL)
}
