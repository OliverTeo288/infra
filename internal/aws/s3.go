
package aws

import (
	"context"
	"bytes"
	"fmt"
	"encoding/json"

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

	fmt.Println("Creating non-SSL deny policy....")
	// Policy to enforce SSL
	err = applyBucketPolicy(s3Client, bucketName)
	if err != nil {
    return fmt.Errorf("failed to apply bucket policy: %w", err)
}

	fmt.Println("Creating backend.tfvars....")
	// Upload the backend.tfvars file
	err = uploadBackendTfvars(s3Client, bucketName, region)
	if err != nil {
		return fmt.Errorf("failed to upload backend.tfvars: %w", err)
	}

	fmt.Printf("Bucket %q successfully created in region %q using profile %q.\n", bucketName, region, profile)
	return nil
}


// UploadBackendTfvars uploads a backend.tfvars file with specified content to the S3 bucket.
func uploadBackendTfvars(s3Client *s3.Client, bucketName, region string) error {
	// File content
	fileContent := fmt.Sprintf(`bucket  = "%s"
key     = "terraform.tfstate"
encrypt      = true  
use_lockfile = true
region  = "%s"`, bucketName, region)

	// Upload the file
	input := &s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String("backend.tfvars"),
		Body:   bytes.NewReader([]byte(fileContent)),
	}

	_, err := s3Client.PutObject(context.TODO(), input)
	if err != nil {
		return fmt.Errorf("failed to upload backend.tfvars: %w", err)
	}

	fmt.Println("backend.tfvars file successfully uploaded.")
	return nil
}

// applyBucketPolicy adds a bucket policy to deny non-SSL access to the bucket.
func applyBucketPolicy(s3Client *s3.Client, bucketName string) error {
	// Define the bucket policy
	policy := map[string]interface{}{
		"Version": "2012-10-17",
		"Statement": []map[string]interface{}{
			{
				"Sid":    "DenyNonSSLRequests",
				"Effect": "Deny",
				"Principal": "*",
				"Action": "s3:*",
				"Resource": []string{
					fmt.Sprintf("arn:aws:s3:::%s", bucketName),
					fmt.Sprintf("arn:aws:s3:::%s/*", bucketName),
				},
				"Condition": map[string]interface{}{
					"Bool": map[string]string{
						"aws:SecureTransport": "false",
					},
				},
			},
		},
	}

	// Marshal the policy into JSON
	policyJSON, err := json.Marshal(policy)
	if err != nil {
		return fmt.Errorf("failed to marshal bucket policy: %w", err)
	}

	// Apply the bucket policy
	input := &s3.PutBucketPolicyInput{
		Bucket: aws.String(bucketName),
		Policy: aws.String(string(policyJSON)),
	}

	_, err = s3Client.PutBucketPolicy(context.TODO(), input)
	if err != nil {
		return fmt.Errorf("failed to apply bucket policy: %w", err)
	}

	fmt.Println("Bucket policy to deny non-SSL access successfully applied.")
	return nil
}
