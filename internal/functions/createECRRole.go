package functions

import (
	"raid/infra/internal/aws"
)

var CommonAwsAccountId string

func CreateECRRole(profile, region string) error {
	roleName := "ecrreader"
	
	err := aws.SetupECRRole(profile, region, roleName, CommonAwsAccountId)
	if err != nil {
		return err
	}

	return nil
}