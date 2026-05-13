package flows

import (
	"context"
	"fmt"

	"raid/infra/internal/config"
	"raid/infra/internal/iam"
	"raid/infra/internal/observability"
)

// CreateECRReadRole provisions a cross-account-assumable role with read-only
// ECR permissions.
func CreateECRReadRole(ctx context.Context, profile, region string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	if cfg.CommonAWSAccountID == "" {
		return fmt.Errorf("CommonAWSAccountID is not configured (missing build-time variable)")
	}
	ctx = observability.WithAttrs(ctx, "role_name", cfg.ECRReaderRole, "permission_set", "ecr_read")
	observability.FromContext(ctx).Info("ecr_role_provisioning")
	return iam.SetupECRRole(ctx, observability.AuditFunc(),
		profile, region,
		cfg.ECRReaderRole, cfg.CommonAWSAccountID, cfg.CrossAccountUser,
		iam.ECRReadActions,
	)
}

// CreateECRWriteRole provisions a cross-account-assumable role with read+push
// ECR permissions.
func CreateECRWriteRole(ctx context.Context, profile, region string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	if cfg.CommonAWSAccountID == "" {
		return fmt.Errorf("CommonAWSAccountID is not configured (missing build-time variable)")
	}
	ctx = observability.WithAttrs(ctx, "role_name", cfg.ECRWriterRole, "permission_set", "ecr_write")
	observability.FromContext(ctx).Info("ecr_role_provisioning")
	return iam.SetupECRRole(ctx, observability.AuditFunc(),
		profile, region,
		cfg.ECRWriterRole, cfg.CommonAWSAccountID, cfg.CrossAccountUser,
		iam.ECRWriteActions,
	)
}
