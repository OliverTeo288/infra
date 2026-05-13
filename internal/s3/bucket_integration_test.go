//go:build integration

package s3_test

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"os"
	"strings"
	"testing"

	"raid/infra/internal/awscfg"
	infras3 "raid/infra/internal/s3"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func requireFloci(t *testing.T) {
	t.Helper()
	if os.Getenv("FLOCI_ENDPOINT") == "" {
		t.Skip("FLOCI_ENDPOINT not set; skipping integration test")
	}
}

func randomSuffix(t *testing.T) string {
	t.Helper()
	b := make([]byte, 4)
	if _, err := rand.Read(b); err != nil {
		t.Fatalf("rand.Read: %v", err)
	}
	return hex.EncodeToString(b)
}

func ptr(s string) *string { return &s }

func TestCreateTerraformStateBucket_AppliesVersioningAndPolicy(t *testing.T) {
	requireFloci(t)

	ctx := context.Background()
	bucket := "test-bucket-" + randomSuffix(t)

	if err := infras3.CreateTerraformStateBucket(ctx, nil, "floci-test", "us-east-1", bucket); err != nil {
		t.Fatalf("CreateTerraformStateBucket failed: %v", err)
	}

	cfg, err := awscfg.LoadConfig("floci-test", "us-east-1")
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}
	client := s3.NewFromConfig(cfg, awscfg.S3Options())

	ver, err := client.GetBucketVersioning(ctx, &s3.GetBucketVersioningInput{Bucket: &bucket})
	if err != nil {
		t.Fatalf("GetBucketVersioning failed: %v", err)
	}
	if string(ver.Status) != "Enabled" {
		t.Errorf("expected versioning Enabled, got %q", ver.Status)
	}

	pol, err := client.GetBucketPolicy(ctx, &s3.GetBucketPolicyInput{Bucket: &bucket})
	if err != nil {
		t.Fatalf("GetBucketPolicy failed: %v", err)
	}
	if pol.Policy == nil || !strings.Contains(*pol.Policy, "DenyNonSSLRequests") {
		t.Errorf("bucket policy missing DenyNonSSLRequests statement: %v", pol.Policy)
	}

	obj, err := client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: &bucket,
		Key:    ptr("backend.tfvars"),
	})
	if err != nil {
		t.Fatalf("GetObject(backend.tfvars) failed: %v", err)
	}
	defer obj.Body.Close()
}
