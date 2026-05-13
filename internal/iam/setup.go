package iam

import (
	"context"
	"fmt"

	"raid/infra/internal/awscfg"
	"raid/infra/internal/observability"

	"github.com/aws/aws-sdk-go-v2/service/iam"
)

// SetupGitOpsRole creates the GitOps IAM role, attaches the admin inline policy,
// and is safe to call repeatedly (CreateRole skips if the role exists).
func SetupGitOpsRole(ctx context.Context, audit observability.AuditHook, profile, region, roleName, gitlabHost, baseURL, subjectPattern string) error {
	cfg, err := awscfg.LoadConfig(profile, region)
	if err != nil {
		return err
	}

	accountID, err := GetAccountID(ctx, cfg)
	if err != nil {
		return err
	}

	// Bind account_id + region to the logger for the rest of this flow so
	// every audit line carries them without per-call duplication.
	ctx = observability.WithAttrs(ctx, "account_id", accountID, "region", region)

	trustPolicy, err := CreateTrustPolicy(accountID, gitlabHost, baseURL, subjectPattern)
	if err != nil {
		return err
	}

	client := iam.NewFromConfig(cfg)
	if err := CreateRole(ctx, client, audit, roleName, trustPolicy); err != nil {
		return err
	}
	return AttachAdminInlinePolicy(ctx, client, audit, roleName)
}

// SetupECRRole creates a cross-account-assumable role with the supplied
// ECR action set scoped to the calling account's repositories.
func SetupECRRole(ctx context.Context, audit observability.AuditHook, profile, region, roleName, commonAccountID, crossAccountUser string, actions []string) error {
	cfg, err := awscfg.LoadConfig(profile, region)
	if err != nil {
		return err
	}

	accountID, err := GetAccountID(ctx, cfg)
	if err != nil {
		return err
	}
	ctx = observability.WithAttrs(ctx, "account_id", accountID, "region", region)

	trustPolicy, err := CreateECRTrustPolicy(commonAccountID, crossAccountUser)
	if err != nil {
		return err
	}

	client := iam.NewFromConfig(cfg)
	if err := CreateRole(ctx, client, audit, roleName, trustPolicy); err != nil {
		return err
	}
	return AttachECRPolicy(ctx, client, audit, roleName, fmt.Sprintf("%sPolicy", roleName), accountID, actions)
}

// RegisterOIDCProvider is a thin wrapper around CreateOIDCProvider that loads
// the AWS SDK config and constructs the client.
func RegisterOIDCProvider(ctx context.Context, audit observability.AuditHook, profile, region, gitURL string) error {
	cfg, err := awscfg.LoadConfig(profile, region)
	if err != nil {
		return err
	}
	ctx = observability.WithAttrs(ctx, "region", region)
	return CreateOIDCProvider(ctx, iam.NewFromConfig(cfg), audit, gitURL)
}
