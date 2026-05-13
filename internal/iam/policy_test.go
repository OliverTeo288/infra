package iam_test

import (
	"encoding/json"
	"strings"
	"testing"

	"raid/infra/internal/iam"
)

func TestCreateTrustPolicy_UsesDynamicHost(t *testing.T) {
	const (
		host    = "gitlab.example.com"
		baseURL = "https://gitlab.example.com"
		sub     = "project_path:group/subgroup/*:ref_type:branch:ref:main"
	)

	got, err := iam.CreateTrustPolicy("123456789012", host, baseURL, sub)
	if err != nil {
		t.Fatalf("CreateTrustPolicy: %v", err)
	}

	if !strings.Contains(got, "arn:aws:iam::123456789012:oidc-provider/"+host) {
		t.Errorf("trust policy missing dynamic OIDC provider ARN; got: %s", got)
	}
	if !strings.Contains(got, `"`+host+`:sub"`) {
		t.Errorf("trust policy missing host-prefixed sub key; got: %s", got)
	}
	if !json.Valid([]byte(got)) {
		t.Errorf("trust policy is not valid JSON: %s", got)
	}
}

func TestCreateECRTrustPolicy_UsesDynamicCrossAccountUser(t *testing.T) {
	got, err := iam.CreateECRTrustPolicy("999999999999", "crossacc-ecrreader")
	if err != nil {
		t.Fatalf("CreateECRTrustPolicy: %v", err)
	}
	if !strings.Contains(got, "arn:aws:iam::999999999999:user/crossacc-ecrreader") {
		t.Errorf("ECR trust policy missing principal; got: %s", got)
	}
	if !json.Valid([]byte(got)) {
		t.Errorf("ECR trust policy is not valid JSON: %s", got)
	}
}
