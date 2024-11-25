package aws

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/sts"
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

// CreateIAMRole creates an IAM role with the specified role name and trust policy.
func CreateIAMRole(iamClient *iam.Client, roleName, trustPolicy string) error {
	createRoleInput := &iam.CreateRoleInput{
		RoleName:                 aws.String(roleName),
		AssumeRolePolicyDocument: aws.String(trustPolicy),
		Description:              aws.String(fmt.Sprintf("Role %s for Terraform GitOps with full access policy.", roleName)),
	}

	_, err := iamClient.CreateRole(context.TODO(), createRoleInput)
	if err != nil {
		return fmt.Errorf("failed to create role: %w", err)
	}
	fmt.Printf("Role %q created successfully.\n", roleName)
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

// Main function to create the role and attach the policy.
func SetupRole(profile, region, roleName string) error {
	// Load AWS configuration
	cfg, err := LoadAWSConfig(profile, region)
	if err != nil {
		return err
	}

	// Get AWS account ID
	accountID, err := GetAWSAccountID(cfg)
	if err != nil {
		return err
	}

	// Generate trust policy
	trustPolicy, err := CreateTrustPolicy(accountID)
	if err != nil {
		return err
	}

	// Create an IAM client
	iamClient := iam.NewFromConfig(cfg)

	// Create IAM role
	if err := CreateIAMRole(iamClient, roleName, trustPolicy); err != nil {
		return err
	}

	// Attach inline policy
	if err := AttachInlinePolicy(iamClient, roleName); err != nil {
		return err
	}

	return nil
}
