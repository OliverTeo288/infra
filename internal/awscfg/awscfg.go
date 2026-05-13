// Package awscfg builds AWS SDK configuration for the CLI.
//
// LoadConfig is the only entry point CLI code should use. It transparently
// swaps to a floci-routed config (dummy credentials + custom endpoint) when
// the FLOCI_ENDPOINT environment variable is set, so production code never
// branches on "are we testing or not".
package awscfg

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// flociEndpointEnv is the env var consulted to decide whether AWS calls are
// routed at a local floci emulator. Documented so users can grep for it.
const flociEndpointEnv = "FLOCI_ENDPOINT"

// LoadConfig returns an AWS SDK config for the given profile and region.
// When FLOCI_ENDPOINT is set, profile is ignored and the SDK is pointed at
// the local emulator with static dummy credentials.
func LoadConfig(profile, region string) (aws.Config, error) {
	if endpoint := os.Getenv(flociEndpointEnv); endpoint != "" {
		return loadFlociConfig(endpoint, region)
	}

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(region),
		config.WithSharedConfigProfile(profile),
	)
	if err != nil {
		return aws.Config{}, fmt.Errorf("failed to load AWS config: %w", err)
	}
	return cfg, nil
}

// IsFlociRouted reports whether the AWS SDK is currently routed at a local
// emulator. Callers that need to tweak service-client options (e.g. S3 path
// style) consult this rather than reading the env var directly.
func IsFlociRouted() bool {
	return os.Getenv(flociEndpointEnv) != ""
}

// S3Options returns S3 client functional options appropriate for the current
// environment. When floci-routed, forces path-style addressing so the local
// emulator can match the request URL; real AWS uses the SDK default
// (virtual-hosted).
func S3Options() func(*s3.Options) {
	return func(o *s3.Options) {
		if IsFlociRouted() {
			o.UsePathStyle = true
		}
	}
}

func loadFlociConfig(endpoint, region string) (aws.Config, error) {
	if region == "" {
		region = "us-east-1"
	}
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("test", "test", "")),
		config.WithBaseEndpoint(endpoint),
	)
	if err != nil {
		return aws.Config{}, fmt.Errorf("failed to load floci AWS config: %w", err)
	}
	return cfg, nil
}
