
package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// CreateS3Bucket creates an S3 bucket using the specified profile, region, and bucket name.
func CreateS3Bucket(profile, region, bucketName string) error {
	// Load AWS config with the specified profile
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(region),
		config.WithSharedConfigProfile(profile),
	)
	if err != nil {
		return fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Create an S3 client
	s3Client := s3.NewFromConfig(cfg)

	// Prepare the CreateBucket input
	input := &s3.CreateBucketInput{
		Bucket: aws.String(bucketName),
	}

	// Add a location constraint if the region is not us-east-1
	if region != "us-east-1" {
		input.CreateBucketConfiguration = &types.CreateBucketConfiguration{
			LocationConstraint: types.BucketLocationConstraint(region),
		}
	}

	// Create the bucket
	_, err = s3Client.CreateBucket(context.TODO(), input)
	if err != nil {
		return fmt.Errorf("failed to create bucket: %w", err)
	}

	// Enable versioning for the bucket
	versioningInput := &s3.PutBucketVersioningInput{
		Bucket: aws.String(bucketName),
		VersioningConfiguration: &types.VersioningConfiguration{
			Status: types.BucketVersioningStatusEnabled,
		},
	}

	_, err = s3Client.PutBucketVersioning(context.TODO(), versioningInput)
	if err != nil {
		return fmt.Errorf("failed to enable versioning: %w", err)
	}

	fmt.Printf("Bucket %q successfully created in region %q using profile %q.\n", bucketName, region, profile)
	return nil
}