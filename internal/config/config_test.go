package config

import (
	"strings"
	"testing"
)

const (
	fakeDomainURL    = "https://gitlab.example.com/group/subgroup/project.git"
	fakeExpectedHost = "gitlab.example.com"
	fakeExpectedBase = "https://gitlab.example.com"
)

func TestLoad_ExtractsHostFromURL(t *testing.T) {
	GitlabHTTPSDomain = fakeDomainURL
	t.Cleanup(func() { GitlabHTTPSDomain = "" })

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.GitlabHost != fakeExpectedHost {
		t.Errorf("GitlabHost = %q; want %q", cfg.GitlabHost, fakeExpectedHost)
	}
	if cfg.GitlabBaseURL != fakeExpectedBase {
		t.Errorf("GitlabBaseURL = %q; want %q", cfg.GitlabBaseURL, fakeExpectedBase)
	}
}

func TestLoad_MissingHTTPSDomain(t *testing.T) {
	GitlabHTTPSDomain = ""
	_, err := Load()
	if err == nil || !strings.Contains(err.Error(), "GitlabHTTPSDomain") {
		t.Errorf("expected error mentioning GitlabHTTPSDomain, got %v", err)
	}
}

func TestLoad_MalformedURL(t *testing.T) {
	GitlabHTTPSDomain = "not-a-url"
	t.Cleanup(func() { GitlabHTTPSDomain = "" })

	_, err := Load()
	if err == nil {
		t.Errorf("expected error for malformed URL; got nil")
	}
}

func TestLoad_OIDCSubject_EnvOverridesBuildTime(t *testing.T) {
	GitlabHTTPSDomain = fakeDomainURL
	OIDCSubjectPattern = "build-time-default"
	t.Setenv("INFRA_OIDC_SUB", "runtime-override")
	t.Cleanup(func() {
		GitlabHTTPSDomain = ""
		OIDCSubjectPattern = ""
	})

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.OIDCSubjectPattern != "runtime-override" {
		t.Errorf("OIDCSubjectPattern = %q; want runtime-override", cfg.OIDCSubjectPattern)
	}
}

func TestLoad_OIDCSubject_UsesBuildTimeWhenEnvUnset(t *testing.T) {
	GitlabHTTPSDomain = fakeDomainURL
	OIDCSubjectPattern = "build-time-default"
	t.Cleanup(func() {
		GitlabHTTPSDomain = ""
		OIDCSubjectPattern = ""
	})

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.OIDCSubjectPattern != "build-time-default" {
		t.Errorf("OIDCSubjectPattern = %q; want build-time-default", cfg.OIDCSubjectPattern)
	}
}

func TestLoad_EnvOverrides(t *testing.T) {
	GitlabHTTPSDomain = fakeDomainURL
	t.Cleanup(func() { GitlabHTTPSDomain = "" })

	t.Setenv("INFRA_GITOPS_ROLE_DEFAULT", "MyCustomGitopsRole")
	t.Setenv("INFRA_ECR_READER_ROLE", "my-reader")
	t.Setenv("INFRA_ECR_WRITER_ROLE", "my-writer")
	t.Setenv("INFRA_CROSSACC_USER", "custom-crossacc")
	t.Setenv("INFRA_DEFAULT_REGION", "ap-southeast-1")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	cases := map[string]string{
		"GitOpsRoleDefault": cfg.GitOpsRoleDefault,
		"ECRReaderRole":     cfg.ECRReaderRole,
		"ECRWriterRole":     cfg.ECRWriterRole,
		"CrossAccountUser":  cfg.CrossAccountUser,
		"DefaultRegion":     cfg.DefaultRegion,
	}
	want := map[string]string{
		"GitOpsRoleDefault": "MyCustomGitopsRole",
		"ECRReaderRole":     "my-reader",
		"ECRWriterRole":     "my-writer",
		"CrossAccountUser":  "custom-crossacc",
		"DefaultRegion":     "ap-southeast-1",
	}
	for field, got := range cases {
		if got != want[field] {
			t.Errorf("%s = %q; want %q", field, got, want[field])
		}
	}
}

func TestLoad_DefaultsWhenEnvUnset(t *testing.T) {
	GitlabHTTPSDomain = fakeDomainURL
	t.Cleanup(func() { GitlabHTTPSDomain = "" })

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.GitOpsRoleDefault != "TerraformGitopsRole" {
		t.Errorf("GitOpsRoleDefault default = %q; want TerraformGitopsRole", cfg.GitOpsRoleDefault)
	}
	if cfg.ECRReaderRole != "ecrreader" {
		t.Errorf("ECRReaderRole default = %q; want ecrreader", cfg.ECRReaderRole)
	}
	if cfg.DefaultRegion != "us-east-1" {
		t.Errorf("DefaultRegion default = %q; want us-east-1", cfg.DefaultRegion)
	}
}
