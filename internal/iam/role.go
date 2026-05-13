package iam

import (
	"context"
	"errors"
	"fmt"
	"time"

	"raid/infra/internal/observability"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	iamtypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

const apiTimeout = 30 * time.Second

// GetAccountID resolves the AWS account ID for the supplied SDK config.
// The call inherits cancellation from ctx and additionally caps each
// attempt at apiTimeout.
func GetAccountID(ctx context.Context, cfg aws.Config) (string, error) {
	cctx, cancel := context.WithTimeout(ctx, apiTimeout)
	defer cancel()

	out, err := sts.NewFromConfig(cfg).GetCallerIdentity(cctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return "", fmt.Errorf("failed to get AWS account ID: %w", err)
	}
	return aws.ToString(out.Account), nil
}

// CreateRole creates an IAM role with the given trust policy. Idempotent:
// if the role already exists, returns nil and emits an info-level log line.
func CreateRole(ctx context.Context, client *iam.Client, audit observability.AuditHook, roleName, trustPolicy string) error {
	cctx, cancel := context.WithTimeout(ctx, apiTimeout)
	defer cancel()

	out, err := client.CreateRole(cctx, &iam.CreateRoleInput{
		RoleName:                 aws.String(roleName),
		AssumeRolePolicyDocument: aws.String(trustPolicy),
		Description:              aws.String(fmt.Sprintf("Role %s created using infra cli.", roleName)),
	})
	if err != nil {
		var exists *iamtypes.EntityAlreadyExistsException
		if errors.As(err, &exists) {
			observability.FromContext(ctx).Info("iam_role_exists_skip", "role_name", roleName)
			return nil
		}
		return fmt.Errorf("failed to create role: %w", err)
	}

	audit.Run(ctx, "iam:CreateRole",
		"resource_arn", aws.ToString(out.Role.Arn),
		"role_name", roleName,
	)
	return nil
}

// AttachAdminInlinePolicy attaches the broad GitOps admin inline policy to roleName.
func AttachAdminInlinePolicy(ctx context.Context, client *iam.Client, audit observability.AuditHook, roleName string) error {
	body, err := marshal(adminInlinePolicy(), "inline admin policy")
	if err != nil {
		return err
	}
	return putRolePolicy(ctx, client, audit, roleName, fmt.Sprintf("%sPolicy", roleName), body, "admin")
}

// AttachECRPolicy attaches an ECR-scoped inline policy with the given actions
// to roleName, scoped to the supplied account ID.
func AttachECRPolicy(ctx context.Context, client *iam.Client, audit observability.AuditHook, roleName, policyName, accountID string, actions []string) error {
	body, err := marshal(ecrInlinePolicy(accountID, actions), "ECR inline policy")
	if err != nil {
		return err
	}
	return putRolePolicy(ctx, client, audit, roleName, policyName, body, "ecr")
}

func putRolePolicy(ctx context.Context, client *iam.Client, audit observability.AuditHook, roleName, policyName, body, policyType string) error {
	cctx, cancel := context.WithTimeout(ctx, apiTimeout)
	defer cancel()

	if _, err := client.PutRolePolicy(cctx, &iam.PutRolePolicyInput{
		RoleName:       aws.String(roleName),
		PolicyName:     aws.String(policyName),
		PolicyDocument: aws.String(body),
	}); err != nil {
		return fmt.Errorf("failed to attach inline policy: %w", err)
	}
	audit.Run(ctx, "iam:PutRolePolicy",
		"role_name", roleName,
		"policy_name", policyName,
		"policy_type", policyType,
	)
	return nil
}
