package functions

import (
	"fmt"

	"raid/infra/internal/aws"
)

var CommonAwsAccountId string

func CreateECRReadRole(profile, region string) error {
	if CommonAwsAccountId == "" {
		return fmt.Errorf("common AWS account ID is not configured (missing build-time variable)")
	}
	return aws.SetupECRRole(profile, region, "ecrreader", CommonAwsAccountId, aws.ECRReadActions)
}

func CreateECRWriteRole(profile, region string) error {
	if CommonAwsAccountId == "" {
		return fmt.Errorf("common AWS account ID is not configured (missing build-time variable)")
	}
	return aws.SetupECRRole(profile, region, "ecrwriter", CommonAwsAccountId, aws.ECRWriteActions)
}
