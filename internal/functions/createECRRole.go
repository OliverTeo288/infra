package functions

import (
	"raid/infra/internal/aws"
)

var CommonAwsAccountId string

func CreateECRReadRole(profile, region string) error {
	return aws.SetupECRRole(profile, region, "ecrreader", CommonAwsAccountId, aws.ECRReadActions)
}

func CreateECRWriteRole(profile, region string) error {
	return aws.SetupECRRole(profile, region, "ecrwriter", CommonAwsAccountId, aws.ECRWriteActions)
}
