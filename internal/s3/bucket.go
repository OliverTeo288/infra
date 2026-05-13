// Package s3 implements the S3 bucket bootstrap used for the Terraform
// remote state backend: versioning enabled, SSL-only policy, starter
// backend.tfvars uploaded.
package s3

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"raid/infra/internal/awscfg"
	"raid/infra/internal/observability"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

const apiTimeout = 60 * time.Second

// CreateTerraformStateBucket creates the bucket, enables versioning, applies
// the SSL-only deny policy, and uploads a starter backend.tfvars file.
func CreateTerraformStateBucket(ctx context.Context, audit observability.AuditHook, profile, region, bucketName string) error {
	cfg, err := awscfg.LoadConfig(profile, region)
	if err != nil {
		return err
	}

	client := s3.NewFromConfig(cfg, awscfg.S3Options())
	arn := fmt.Sprintf("arn:aws:s3:::%s", bucketName)

	// Bind region + bucket to every log line emitted from this flow.
	ctx = observability.WithAttrs(ctx, "region", region, "bucket", bucketName)

	cctx, cancel := context.WithTimeout(ctx, apiTimeout)
	defer cancel()

	if err := create(cctx, client, bucketName, region); err != nil {
		return fmt.Errorf("failed to create bucket: %w", err)
	}
	audit.Run(ctx, "s3:CreateBucket", "resource_arn", arn)

	if err := enableVersioning(cctx, client, bucketName); err != nil {
		return fmt.Errorf("failed to enable versioning: %w", err)
	}
	audit.Run(ctx, "s3:PutBucketVersioning", "resource_arn", arn, "status", "Enabled")

	if err := applyDenyNonSSLPolicy(cctx, client, bucketName); err != nil {
		return fmt.Errorf("failed to apply bucket policy: %w", err)
	}
	audit.Run(ctx, "s3:PutBucketPolicy", "resource_arn", arn, "policy", "deny-non-ssl")

	if err := uploadBackendTfvars(cctx, client, bucketName, region); err != nil {
		return fmt.Errorf("failed to upload backend.tfvars: %w", err)
	}
	audit.Run(ctx, "s3:PutObject", "resource_arn", arn+"/backend.tfvars", "key", "backend.tfvars")

	return nil
}

func create(ctx context.Context, client *s3.Client, bucketName, region string) error {
	input := &s3.CreateBucketInput{Bucket: aws.String(bucketName)}
	// us-east-1 is the only region that must omit LocationConstraint.
	if region != "us-east-1" {
		input.CreateBucketConfiguration = &types.CreateBucketConfiguration{
			LocationConstraint: types.BucketLocationConstraint(region),
		}
	}
	_, err := client.CreateBucket(ctx, input)
	return err
}

func enableVersioning(ctx context.Context, client *s3.Client, bucketName string) error {
	_, err := client.PutBucketVersioning(ctx, &s3.PutBucketVersioningInput{
		Bucket: aws.String(bucketName),
		VersioningConfiguration: &types.VersioningConfiguration{
			Status: types.BucketVersioningStatusEnabled,
		},
	})
	return err
}

func applyDenyNonSSLPolicy(ctx context.Context, client *s3.Client, bucketName string) error {
	policy := map[string]any{
		"Version": "2012-10-17",
		"Statement": []map[string]any{{
			"Sid":       "DenyNonSSLRequests",
			"Effect":    "Deny",
			"Principal": "*",
			"Action":    "s3:*",
			"Resource": []string{
				fmt.Sprintf("arn:aws:s3:::%s", bucketName),
				fmt.Sprintf("arn:aws:s3:::%s/*", bucketName),
			},
			"Condition": map[string]any{
				"Bool": map[string]string{"aws:SecureTransport": "false"},
			},
		}},
	}

	body, err := json.Marshal(policy)
	if err != nil {
		return fmt.Errorf("failed to marshal bucket policy: %w", err)
	}

	_, err = client.PutBucketPolicy(ctx, &s3.PutBucketPolicyInput{
		Bucket: aws.String(bucketName),
		Policy: aws.String(string(body)),
	})
	return err
}

func uploadBackendTfvars(ctx context.Context, client *s3.Client, bucketName, region string) error {
	content := fmt.Sprintf(`bucket  = "%s"
key     = "terraform.tfstate"
encrypt      = true
use_lockfile = true
region  = "%s"`, bucketName, region)

	_, err := client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String("backend.tfvars"),
		Body:   bytes.NewReader([]byte(content)),
	})
	return err
}
