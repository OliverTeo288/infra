package awscfg_test

import (
	"context"
	"testing"

	"raid/infra/internal/awscfg"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func TestIsFlociRouted_UnsetByDefault(t *testing.T) {
	t.Setenv("FLOCI_ENDPOINT", "")
	if awscfg.IsFlociRouted() {
		t.Error("IsFlociRouted should be false when FLOCI_ENDPOINT is unset")
	}
}

func TestIsFlociRouted_TrueWhenSet(t *testing.T) {
	t.Setenv("FLOCI_ENDPOINT", "http://localhost:4566")
	if !awscfg.IsFlociRouted() {
		t.Error("IsFlociRouted should be true when FLOCI_ENDPOINT is set")
	}
}

func TestLoadConfig_FlociRoutingUsesDummyCreds(t *testing.T) {
	t.Setenv("FLOCI_ENDPOINT", "http://localhost:4566")
	cfg, err := awscfg.LoadConfig("any-profile-name", "")
	if err != nil {
		t.Fatalf("LoadConfig with floci endpoint: %v", err)
	}
	// Region should fall back to us-east-1 when blank.
	if cfg.Region != "us-east-1" {
		t.Errorf("region = %q; want us-east-1 (the default when blank)", cfg.Region)
	}
	// Credentials should resolve to the static dummy "test/test" pair.
	creds, err := cfg.Credentials.Retrieve(context.Background())
	if err != nil {
		t.Fatalf("Credentials.Retrieve: %v", err)
	}
	if creds.AccessKeyID != "test" {
		t.Errorf("AccessKeyID = %q; want test", creds.AccessKeyID)
	}
}

func TestLoadConfig_FlociRoutingHonoursExplicitRegion(t *testing.T) {
	t.Setenv("FLOCI_ENDPOINT", "http://localhost:4566")
	cfg, err := awscfg.LoadConfig("any-profile", "ap-southeast-1")
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	if cfg.Region != "ap-southeast-1" {
		t.Errorf("region = %q; want ap-southeast-1", cfg.Region)
	}
}

func TestS3Options_PathStyleOnlyWhenFlociRouted(t *testing.T) {
	t.Run("default no-op", func(t *testing.T) {
		t.Setenv("FLOCI_ENDPOINT", "")
		var opts s3.Options
		awscfg.S3Options()(&opts)
		if opts.UsePathStyle {
			t.Error("UsePathStyle should remain false in production routing")
		}
	})
	t.Run("path-style under floci", func(t *testing.T) {
		t.Setenv("FLOCI_ENDPOINT", "http://localhost:4566")
		var opts s3.Options
		awscfg.S3Options()(&opts)
		if !opts.UsePathStyle {
			t.Error("UsePathStyle should be true when FLOCI_ENDPOINT is set")
		}
	})
}
