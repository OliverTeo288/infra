package config

import (
	"errors"
	"fmt"
	"net/url"
	"os"
)

// Build-time variables. Set via:
//
//	go build -ldflags "-X raid/infra/internal/config.GitlabHTTPSDomain=https://... \
//	                   -X raid/infra/internal/config.GitlabSSHDomain=git@... \
//	                   -X raid/infra/internal/config.CommonAWSAccountID=123456789012 \
//	                   -X raid/infra/internal/config.OIDCSubjectPattern=project_path:..."
//
// OIDCSubjectPattern is sensitive — it embeds the GitLab project path used in
// the role's trust policy. Inject from a GitHub Actions secret at release
// build time, not committed to the repo. Local dev can override at runtime
// via the INFRA_OIDC_SUB env var.
var (
	GitlabHTTPSDomain  string
	GitlabSSHDomain    string
	CommonAWSAccountID string
	OIDCSubjectPattern string
)

type Config struct {
	GitlabHTTPSDomain  string
	GitlabSSHDomain    string
	CommonAWSAccountID string

	// GitlabHost is the host component derived from GitlabHTTPSDomain
	// e.g. "gitlab.example.com"
	GitlabHost string

	// GitlabBaseURL is "<scheme>://<host>" — used for the OIDC provider URL
	GitlabBaseURL string

	OIDCSubjectPattern string
	GitOpsRoleDefault  string
	ECRReaderRole      string
	ECRWriterRole      string
	CrossAccountUser   string
	DefaultRegion      string
}

// Load builds runtime config from build-time vars and env overrides.
// Returns an error when required build-time vars are missing or malformed.
func Load() (*Config, error) {
	if GitlabHTTPSDomain == "" {
		return nil, errors.New("GitlabHTTPSDomain not configured (missing -ldflags -X at build time)")
	}

	u, err := url.Parse(GitlabHTTPSDomain)
	if err != nil || u.Host == "" || u.Scheme == "" {
		return nil, fmt.Errorf("invalid GitlabHTTPSDomain %q: must be a full URL", GitlabHTTPSDomain)
	}

	return &Config{
		GitlabHTTPSDomain:  GitlabHTTPSDomain,
		GitlabSSHDomain:    GitlabSSHDomain,
		CommonAWSAccountID: CommonAWSAccountID,
		GitlabHost:         u.Host,
		GitlabBaseURL:      u.Scheme + "://" + u.Host,
		OIDCSubjectPattern: resolveOIDCSubject(),
		GitOpsRoleDefault:  envOr("INFRA_GITOPS_ROLE_DEFAULT", "TerraformGitopsRole"),
		ECRReaderRole:      envOr("INFRA_ECR_READER_ROLE", "ecrreader"),
		ECRWriterRole:      envOr("INFRA_ECR_WRITER_ROLE", "ecrwriter"),
		CrossAccountUser:   envOr("INFRA_CROSSACC_USER", "crossacc-ecrreader"),
		DefaultRegion:      envOr("INFRA_DEFAULT_REGION", "us-east-1"),
	}, nil
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// resolveOIDCSubject prefers the runtime env override (for local dev / testing),
// then falls back to the build-time-injected OIDCSubjectPattern var.
// Returns an empty string when neither is set; callers that require a non-empty
// value should validate accordingly.
func resolveOIDCSubject() string {
	if v := os.Getenv("INFRA_OIDC_SUB"); v != "" {
		return v
	}
	return OIDCSubjectPattern
}
