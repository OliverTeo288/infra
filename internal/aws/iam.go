package aws

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	iamtypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

// ECR permission sets following least privilege principle.
var (
	ECRReadActions = []string{
		"ecr:BatchCheckLayerAvailability",
		"ecr:GetDownloadUrlForLayer",
		"ecr:GetRepositoryPolicy",
		"ecr:DescribeRepositories",
		"ecr:ListImages",
		"ecr:DescribeImages",
		"ecr:BatchGetImage",
		"ecr:GetLifecyclePolicy",
		"ecr:GetLifecyclePolicyPreview",
		"ecr:ListTagsForResource",
		"ecr:DescribeImageScanFindings",
	}

	ECRWriteActions = []string{
		// Read
		"ecr:BatchCheckLayerAvailability",
		"ecr:GetDownloadUrlForLayer",
		"ecr:GetRepositoryPolicy",
		"ecr:DescribeRepositories",
		"ecr:ListImages",
		"ecr:DescribeImages",
		"ecr:BatchGetImage",
		"ecr:GetLifecyclePolicy",
		"ecr:GetLifecyclePolicyPreview",
		"ecr:ListTagsForResource",
		"ecr:DescribeImageScanFindings",
		// Write
		"ecr:PutImage",
		"ecr:InitiateLayerUpload",
		"ecr:UploadLayerPart",
		"ecr:CompleteLayerUpload",
	}
)

// LoadAWSConfig loads the AWS configuration for the given profile and region.
func LoadAWSConfig(profile, region string) (aws.Config, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(region),
		config.WithSharedConfigProfile(profile),
	)
	if err != nil {
		return aws.Config{}, fmt.Errorf("failed to load AWS config: %w", err)
	}
	return cfg, nil
}

// GetAWSAccountID retrieves the AWS account ID for the current profile.
func GetAWSAccountID(cfg aws.Config) (string, error) {
	stsClient := sts.NewFromConfig(cfg)
	callerIdentity, err := stsClient.GetCallerIdentity(context.TODO(), &sts.GetCallerIdentityInput{})
	if err != nil {
		return "", fmt.Errorf("failed to get AWS account ID: %w", err)
	}
	return *callerIdentity.Account, nil
}

// CreateTrustPolicy generates a trust policy JSON for the given AWS account ID.
func CreateTrustPolicy(accountID string) (string, error) {
	trustPolicy := map[string]interface{}{
		"Version": "2012-10-17",
		"Statement": []map[string]interface{}{
			{
				"Effect": "Allow",
				"Principal": map[string]string{
					"Federated": fmt.Sprintf("arn:aws:iam::%s:oidc-provider/sgts.gitlab-dedicated.com", accountID),
				},
				"Action": "sts:AssumeRoleWithWebIdentity",
				"Condition": map[string]interface{}{
					"StringEquals": map[string]string{
						"sgts.gitlab-dedicated.com:aud": "https://sgts.gitlab-dedicated.com",
					},
					"StringLike": map[string]string{
						"sgts.gitlab-dedicated.com:sub": "project_path:wog/mod/raidshiphats/*:ref_type:*:ref:*",
					},
				},
			},
		},
	}

	trustPolicyJSON, err := json.Marshal(trustPolicy)
	if err != nil {
		return "", fmt.Errorf("failed to marshal trust policy: %w", err)
	}
	return string(trustPolicyJSON), nil
}

// CreateECRTrustPolicy generates a trust policy for ECR role with cross-account user access.
func CreateECRTrustPolicy(commonAccountID string) (string, error) {
	trustPolicy := map[string]interface{}{
		"Version": "2012-10-17",
		"Statement": []map[string]interface{}{
			{
				"Effect": "Allow",
				"Principal": map[string]string{
					"AWS": fmt.Sprintf("arn:aws:iam::%s:user/crossacc-ecrreader", commonAccountID),
				},
				"Action": "sts:AssumeRole",
			},
		},
	}

	trustPolicyJSON, err := json.Marshal(trustPolicy)
	if err != nil {
		return "", fmt.Errorf("failed to marshal ECR trust policy: %w", err)
	}
	return string(trustPolicyJSON), nil
}

// CreateIAMRole creates an IAM role with the specified role name and trust policy.
func CreateIAMRole(iamClient *iam.Client, roleName, trustPolicy string) error {
	createRoleInput := &iam.CreateRoleInput{
		RoleName:                 aws.String(roleName),
		AssumeRolePolicyDocument: aws.String(trustPolicy),
		Description:              aws.String(fmt.Sprintf("Role %s created using infra cli.", roleName)),
	}

	result, err := iamClient.CreateRole(context.TODO(), createRoleInput)
	if err != nil {
		var alreadyExists *iamtypes.EntityAlreadyExistsException
		if errors.As(err, &alreadyExists) {
			fmt.Printf("Role %q already exists, skipping creation.\n", roleName)
			return nil
		}
		return fmt.Errorf("failed to create role: %w", err)
	}
	fmt.Printf("Role %q created successfully. ARN: %s\n", roleName, *result.Role.Arn)
	return nil
}

// AttachInlinePolicy attaches an inline policy to the specified role.
func AttachInlinePolicy(iamClient *iam.Client, roleName string) error {
	inlinePolicy := map[string]interface{}{
		"Version": "2012-10-17",
		"Statement": []map[string]interface{}{
			{
				"Action":   []string{"*"},
				"Effect":   "Allow",
				"Resource": []string{"*"},
			},
		},
	}

	inlinePolicyJSON, err := json.Marshal(inlinePolicy)
	if err != nil {
		return fmt.Errorf("failed to marshal inline policy: %w", err)
	}

	putRolePolicyInput := &iam.PutRolePolicyInput{
		RoleName:       aws.String(roleName),
		PolicyName:     aws.String(fmt.Sprintf("%sPolicy", roleName)),
		PolicyDocument: aws.String(string(inlinePolicyJSON)),
	}

	_, err = iamClient.PutRolePolicy(context.TODO(), putRolePolicyInput)
	if err != nil {
		return fmt.Errorf("failed to attach inline policy to role: %w", err)
	}
	return nil
}

// AttachECRPolicy attaches an ECR inline policy with the given actions to the specified role.
// Actions are scoped to the account's repositories following least privilege.
func AttachECRPolicy(iamClient *iam.Client, roleName, policyName, accountID string, actions []string) error {
	inlinePolicy := map[string]interface{}{
		"Version": "2012-10-17",
		"Statement": []map[string]interface{}{
			{
				"Effect":   "Allow",
				"Action":   []string{"ecr:GetAuthorizationToken"},
				"Resource": "*",
			},
			{
				"Effect":   "Allow",
				"Action":   actions,
				"Resource": fmt.Sprintf("arn:aws:ecr:*:%s:repository/*", accountID),
			},
		},
	}

	inlinePolicyJSON, err := json.Marshal(inlinePolicy)
	if err != nil {
		return fmt.Errorf("failed to marshal ECR inline policy: %w", err)
	}

	_, err = iamClient.PutRolePolicy(context.TODO(), &iam.PutRolePolicyInput{
		RoleName:       aws.String(roleName),
		PolicyName:     aws.String(policyName),
		PolicyDocument: aws.String(string(inlinePolicyJSON)),
	})
	if err != nil {
		return fmt.Errorf("failed to attach ECR inline policy to role: %w", err)
	}
	return nil
}

// SetupRole creates a GitOps IAM role with trust policy and admin inline policy.
func SetupRole(profile, region, roleName string) error {
	cfg, err := LoadAWSConfig(profile, region)
	if err != nil {
		return err
	}

	accountID, err := GetAWSAccountID(cfg)
	if err != nil {
		return err
	}

	trustPolicy, err := CreateTrustPolicy(accountID)
	if err != nil {
		return err
	}

	iamClient := iam.NewFromConfig(cfg)

	if err := CreateIAMRole(iamClient, roleName, trustPolicy); err != nil {
		return err
	}

	if err := AttachInlinePolicy(iamClient, roleName); err != nil {
		return err
	}

	return nil
}

// SetupECRRole creates an ECR role with cross-account trust and the specified ECR permissions.
func SetupECRRole(profile, region, roleName, commonAccountID string, actions []string) error {
	cfg, err := LoadAWSConfig(profile, region)
	if err != nil {
		return err
	}

	accountID, err := GetAWSAccountID(cfg)
	if err != nil {
		return err
	}

	trustPolicy, err := CreateECRTrustPolicy(commonAccountID)
	if err != nil {
		return err
	}

	iamClient := iam.NewFromConfig(cfg)

	if err := CreateIAMRole(iamClient, roleName, trustPolicy); err != nil {
		return err
	}

	if err := AttachECRPolicy(iamClient, roleName, fmt.Sprintf("%sPolicy", roleName), accountID, actions); err != nil {
		return err
	}

	return nil
}

// FetchThumbprint retrieves the SHA-1 thumbprint of the SSL certificate for a given URL.
func FetchThumbprint(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to fetch the certificate from URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.TLS == nil || len(resp.TLS.PeerCertificates) == 0 {
		return "", fmt.Errorf("no TLS certificates found for URL: %s", url)
	}

	certs := resp.TLS.PeerCertificates
	cert := certs[len(certs)-1]
	thumbprint := sha1.Sum(cert.Raw)
	return hex.EncodeToString(thumbprint[:]), nil
}

// CreateOIDCProvider creates an AWS OpenID Connect (OIDC) Provider for GitLab.
func CreateOIDCProvider(profile, region, gitURL string) error {
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(region),
		config.WithSharedConfigProfile(profile),
	)
	if err != nil {
		return fmt.Errorf("failed to load AWS config: %w", err)
	}

	iamClient := iam.NewFromConfig(cfg)

	thumbprint, err := FetchThumbprint(gitURL)
	if err != nil {
		return fmt.Errorf("failed to fetch thumbprint: %w", err)
	}

	oidcInput := &iam.CreateOpenIDConnectProviderInput{
		Url:            aws.String(gitURL),
		ClientIDList:   []string{gitURL},
		ThumbprintList: []string{thumbprint},
	}

	if _, err := iamClient.CreateOpenIDConnectProvider(context.TODO(), oidcInput); err != nil {
		var alreadyExists *iamtypes.EntityAlreadyExistsException
		if errors.As(err, &alreadyExists) {
			fmt.Println("OIDC provider already exists, skipping creation.")
			return nil
		}
		return fmt.Errorf("failed to create OIDC provider: %w", err)
	}

	fmt.Println("OIDC Provider created successfully")
	return nil
}
