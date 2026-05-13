package flows

import (
	"context"
	"errors"
	"regexp"

	"raid/infra/internal/observability"
	"raid/infra/internal/prompt"
	"raid/infra/internal/s3"
)

// s3BucketRecommendedName is the project's recommended bucket-naming
// convention shown as the default suggestion at the prompt. The user can
// press Enter to accept it or type a custom name.
const s3BucketRecommendedName = "test-dev-backend-tf-0000"

var bucketNameRegex = regexp.MustCompile(`^[a-zA-Z0-9\-]+$`)

func validateBucketName(input string) error {
	if !bucketNameRegex.MatchString(input) {
		return errors.New("bucket name must only contain alphanumeric characters or dashes, and no spaces")
	}
	return nil
}

// CreateS3StateBucket walks the user through naming and provisioning the S3
// bucket used for the Terraform remote state backend.
func CreateS3StateBucket(ctx context.Context, profile, region string) error {
	bucketName, err := prompt.Input(
		"Enter the name of the S3 bucket",
		validateBucketName,
		s3BucketRecommendedName,
	)
	if err != nil {
		return err
	}
	observability.FromContext(ctx).Info("s3_state_bucket_provisioning",
		"bucket", bucketName, "region", region,
	)
	return s3.CreateTerraformStateBucket(ctx, observability.AuditFunc(), profile, region, bucketName)
}
