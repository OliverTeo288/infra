// Package awsauth wraps AWS profile discovery, SSO login, and region
// selection. The Login flow is interactive: it lists configured profiles,
// prompts the user to pick one, refreshes credentials via `aws sso login`
// when needed, then prompts for an AWS region.
package awsauth

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"raid/infra/internal/awscfg"
	"raid/infra/internal/prompt"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/aws/smithy-go"
)

const stsTimeout = 30 * time.Second

// defaultBootstrapRegion is the region used for SDK calls that must happen
// before the user has chosen a region (DescribeRegions, GetCallerIdentity).
const defaultBootstrapRegion = "us-east-1"

// expiredCredsErrorCodes are the AWS-side error codes that indicate the
// user's credentials need refreshing (and so warrant a `aws sso login`).
// Any other STS failure is a real problem we should surface.
var expiredCredsErrorCodes = map[string]struct{}{
	"ExpiredToken":                {},
	"ExpiredTokenException":       {},
	"InvalidClientTokenId":        {},
	"UnrecognizedClientException": {},
	"AccessDenied":                {},
}

// Login walks the user through profile selection, credential refresh, and
// region selection. Returns the selected profile and region. When no
// profiles are configured, runs `aws configure sso` and then recurses so
// the freshly-created profile can be picked.
func Login(ctx context.Context) (string, string, error) {
	profiles, err := listProfiles()
	if err != nil {
		return "", "", fmt.Errorf("error fetching AWS profiles: %w", err)
	}

	if len(profiles) == 0 {
		fmt.Println("No AWS profiles found. Running `aws configure sso`...")
		cfg := exec.Command("aws", "configure", "sso")
		cfg.Stdin = os.Stdin
		cfg.Stdout = os.Stdout
		cfg.Stderr = os.Stderr
		if err := cfg.Run(); err != nil {
			return "", "", fmt.Errorf("failed to configure AWS SSO: %w", err)
		}
		// Re-list now that the user has created a profile.
		profiles, err = listProfiles()
		if err != nil {
			return "", "", fmt.Errorf("error re-reading AWS profiles after SSO setup: %w", err)
		}
		if len(profiles) == 0 {
			return "", "", errors.New("no AWS profiles configured after `aws configure sso`; aborting")
		}
	}

	profile, err := prompt.Selection(profiles, "AWS Profile")
	if err != nil {
		return "", "", fmt.Errorf("error selecting AWS profile: %w", err)
	}

	if err := refreshIfExpired(ctx, profile); err != nil {
		return "", "", fmt.Errorf("error handling credentials for profile %q: %w", profile, err)
	}

	region, err := promptRegion(ctx, profile)
	if err != nil {
		return "", "", fmt.Errorf("error selecting AWS region: %w", err)
	}
	return profile, region, nil
}

// listProfiles parses ~/.aws/config and returns profile names sorted
// alphabetically. The [default] profile is intentionally ignored — this tool
// only operates on named profiles managed via `aws configure sso`.
func listProfiles() ([]string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	f, err := os.Open(filepath.Join(home, ".aws", "config"))
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, fmt.Errorf("failed to read AWS config: %w", err)
	}
	defer f.Close()

	var profiles []string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if strings.HasPrefix(line, "[profile ") && strings.HasSuffix(line, "]") {
			name := strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(line, "[profile "), "]"))
			if name != "" {
				profiles = append(profiles, name)
			}
		}
	}
	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("failed to parse AWS config: %w", err)
	}

	sort.Strings(profiles)
	return profiles, nil
}

// refreshIfExpired makes a lightweight STS call to validate the profile's
// credentials. If the call fails with a credentials-related AWS error code,
// it shells out to `aws sso login` to refresh them. Other failure modes
// (network, throttling, unexpected errors) are propagated as-is rather
// than triggering a spurious SSO login prompt.
func refreshIfExpired(ctx context.Context, profile string) error {
	cfg, err := awscfg.LoadConfig(profile, defaultBootstrapRegion)
	if err != nil {
		return err
	}

	cctx, cancel := context.WithTimeout(ctx, stsTimeout)
	defer cancel()

	_, err = sts.NewFromConfig(cfg).GetCallerIdentity(cctx, &sts.GetCallerIdentityInput{})
	if err == nil {
		return nil
	}

	if !isExpiredCredsError(err) {
		// Real failure: don't pop an SSO login flow, surface the underlying error.
		return fmt.Errorf("sts:GetCallerIdentity failed: %w", err)
	}

	fmt.Printf("Credentials for profile %q have expired or are invalid. Logging in...\n", profile)
	login := exec.Command("aws", "sso", "login", "--profile", profile)
	login.Stdin = os.Stdin
	login.Stdout = os.Stdout
	login.Stderr = os.Stderr
	if err := login.Run(); err != nil {
		return fmt.Errorf("failed to log in to AWS SSO: %w", err)
	}
	fmt.Println("AWS SSO login successful.")
	return nil
}

// isExpiredCredsError reports whether err looks like an AWS API error
// indicating the caller needs to re-authenticate.
func isExpiredCredsError(err error) bool {
	var apiErr smithy.APIError
	if !errors.As(err, &apiErr) {
		return false
	}
	_, ok := expiredCredsErrorCodes[apiErr.ErrorCode()]
	return ok
}

// promptRegion fetches the list of AWS regions via the SDK (using profile
// against defaultBootstrapRegion to make the API call itself) and lets the
// user pick one.
func promptRegion(ctx context.Context, profile string) (string, error) {
	cfg, err := awscfg.LoadConfig(profile, defaultBootstrapRegion)
	if err != nil {
		return "", err
	}

	cctx, cancel := context.WithTimeout(ctx, stsTimeout)
	defer cancel()

	out, err := ec2.NewFromConfig(cfg).DescribeRegions(cctx, &ec2.DescribeRegionsInput{})
	if err != nil {
		return "", fmt.Errorf("failed to fetch regions: %w", err)
	}

	names := make([]string, 0, len(out.Regions))
	for _, r := range out.Regions {
		if r.RegionName != nil {
			names = append(names, *r.RegionName)
		}
	}

	return prompt.Selection(names, "AWS Region")
}
