package iam

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"time"

	"raid/infra/internal/observability"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	iamtypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
)

// thumbprintHTTPTimeout bounds the TLS handshake performed by FetchThumbprint.
// Without a timeout, a hung remote could deadlock provisioning indefinitely.
const thumbprintHTTPTimeout = 10 * time.Second

// thumbprintClient is a dedicated HTTP client for FetchThumbprint with a
// timeout so we never hang on a slow / unreachable IdP host.
var thumbprintClient = &http.Client{Timeout: thumbprintHTTPTimeout}

// FetchThumbprint returns the SHA-1 thumbprint of the leaf-most cert in the
// chain served by url. AWS uses this thumbprint as the root of trust for OIDC
// identity provider verification.
func FetchThumbprint(url string) (string, error) {
	resp, err := thumbprintClient.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to fetch the certificate from URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.TLS == nil || len(resp.TLS.PeerCertificates) == 0 {
		return "", fmt.Errorf("no TLS certificates found for URL: %s", url)
	}

	cert := resp.TLS.PeerCertificates[len(resp.TLS.PeerCertificates)-1]
	sum := sha1.Sum(cert.Raw)
	return hex.EncodeToString(sum[:]), nil
}

// CreateOIDCProvider registers an OIDC provider with AWS IAM for the supplied
// URL. Idempotent: if a provider for that URL already exists, returns nil and
// emits an info-level log.
//
// Region (and any other contextual attributes) should be attached to ctx
// via observability.WithAttrs by the caller — see iam/setup.go.
func CreateOIDCProvider(ctx context.Context, client *iam.Client, audit observability.AuditHook, gitURL string) error {
	thumbprint, err := FetchThumbprint(gitURL)
	if err != nil {
		return fmt.Errorf("failed to fetch thumbprint: %w", err)
	}

	cctx, cancel := context.WithTimeout(ctx, apiTimeout)
	defer cancel()

	if _, err := client.CreateOpenIDConnectProvider(cctx, &iam.CreateOpenIDConnectProviderInput{
		Url:            aws.String(gitURL),
		ClientIDList:   []string{gitURL},
		ThumbprintList: []string{thumbprint},
	}); err != nil {
		var exists *iamtypes.EntityAlreadyExistsException
		if errors.As(err, &exists) {
			observability.FromContext(ctx).Info("oidc_provider_exists_skip", "url", gitURL)
			return nil
		}
		return fmt.Errorf("failed to create OIDC provider: %w", err)
	}

	audit.Run(ctx, "iam:CreateOpenIDConnectProvider", "url", gitURL)
	return nil
}
