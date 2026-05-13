package cmd

import (
	"context"
	"fmt"

	"raid/infra/internal/awsauth"
)

// login is a thin helper used by every command that needs an authenticated
// AWS profile. Centralised so the error wrapping is consistent and tests
// that wrap cobra commands have a single seam to stub.
//
// loginFn is package-level so tests can replace it.
var loginFn = awsauth.Login

func login(ctx context.Context) (profile, region string, err error) {
	profile, region, err = loginFn(ctx)
	if err != nil {
		return "", "", fmt.Errorf("login failed: %w", err)
	}
	return profile, region, nil
}
