// Package iam builds IAM trust policies, creates roles, and registers OIDC
// providers used by the infra CLI's gitops + ECR flows.
package iam

import (
	"encoding/json"
	"fmt"
)

// PolicyDocument is the typed shape of an IAM policy JSON document.
// Using a struct instead of a map[string]any catches field typos at compile
// time and lets tests assert structural correctness.
type PolicyDocument struct {
	Version   string            `json:"Version"`
	Statement []PolicyStatement `json:"Statement"`
}

// PolicyStatement is a single statement within a PolicyDocument. Fields use
// `any` so callers can pass either a string or a slice as appropriate for
// the action/resource/principal type they need.
type PolicyStatement struct {
	Sid       string `json:"Sid,omitempty"`
	Effect    string `json:"Effect"`
	Principal any    `json:"Principal,omitempty"`
	Action    any    `json:"Action,omitempty"`
	Resource  any    `json:"Resource,omitempty"`
	Condition any    `json:"Condition,omitempty"`
}

// policyVersion is the only IAM policy schema version AWS currently supports.
const policyVersion = "2012-10-17"

// CreateTrustPolicy builds a GitLab-OIDC trust policy JSON for the given
// account.
//
//   - gitlabHost is the OIDC provider host (e.g. "gitlab.example.com")
//   - baseURL is the OIDC audience (e.g. "https://gitlab.example.com")
//   - subjectPattern is the StringLike match for the GitLab JWT sub claim
func CreateTrustPolicy(accountID, gitlabHost, baseURL, subjectPattern string) (string, error) {
	doc := PolicyDocument{
		Version: policyVersion,
		Statement: []PolicyStatement{{
			Effect: "Allow",
			Principal: map[string]string{
				"Federated": fmt.Sprintf("arn:aws:iam::%s:oidc-provider/%s", accountID, gitlabHost),
			},
			Action: "sts:AssumeRoleWithWebIdentity",
			Condition: map[string]map[string]string{
				"StringEquals": {gitlabHost + ":aud": baseURL},
				"StringLike":   {gitlabHost + ":sub": subjectPattern},
			},
		}},
	}
	return marshal(doc, "trust policy")
}

// CreateECRTrustPolicy builds a cross-account trust policy granting an IAM
// user in another account permission to assume this role.
func CreateECRTrustPolicy(commonAccountID, crossAccountUser string) (string, error) {
	doc := PolicyDocument{
		Version: policyVersion,
		Statement: []PolicyStatement{{
			Effect: "Allow",
			Principal: map[string]string{
				"AWS": fmt.Sprintf("arn:aws:iam::%s:user/%s", commonAccountID, crossAccountUser),
			},
			Action: "sts:AssumeRole",
		}},
	}
	return marshal(doc, "ECR trust policy")
}

// adminInlinePolicy is the inline policy attached to the GitOps role.
// Intentionally broad — the GitOps role needs to manage arbitrary AWS
// resources via Terraform. Scoping this is tracked as future work.
func adminInlinePolicy() PolicyDocument {
	return PolicyDocument{
		Version: policyVersion,
		Statement: []PolicyStatement{{
			Effect:   "Allow",
			Action:   []string{"*"},
			Resource: []string{"*"},
		}},
	}
}

// ecrInlinePolicy builds the ECR-scoped inline policy. ecr:GetAuthorizationToken
// requires Resource:* per the AWS docs; all other actions are scoped to
// repositories in the calling account.
func ecrInlinePolicy(accountID string, actions []string) PolicyDocument {
	return PolicyDocument{
		Version: policyVersion,
		Statement: []PolicyStatement{
			{
				Effect:   "Allow",
				Action:   []string{"ecr:GetAuthorizationToken"},
				Resource: "*",
			},
			{
				Effect:   "Allow",
				Action:   actions,
				Resource: fmt.Sprintf("arn:aws:ecr:*:%s:repository/*", accountID),
			},
		},
	}
}

func marshal(doc PolicyDocument, label string) (string, error) {
	b, err := json.Marshal(doc)
	if err != nil {
		return "", fmt.Errorf("failed to marshal %s: %w", label, err)
	}
	return string(b), nil
}
