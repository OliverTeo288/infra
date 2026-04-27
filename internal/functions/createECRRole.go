package functions

import (
	"fmt"
	"os"
	"regexp"

	"raid/infra/internal/aws"
)

func getCommonAccountID() (string, error) {
	id := os.Getenv("COMMON_AWS_ACCOUNT_ID")
	if id == "" {
		return "", fmt.Errorf("COMMON_AWS_ACCOUNT_ID environment variable is not set")
	}
	if matched, _ := regexp.MatchString(`^\d{12}$`, id); !matched {
		return "", fmt.Errorf("COMMON_AWS_ACCOUNT_ID must be a 12-digit number, got: %s", id)
	}
	return id, nil
}

func CreateECRReadRole(profile, region string) error {
	accountID, err := getCommonAccountID()
	if err != nil {
		return err
	}
	return aws.SetupECRRole(profile, region, "ecrreader", accountID, aws.ECRReadActions)
}

func CreateECRWriteRole(profile, region string) error {
	accountID, err := getCommonAccountID()
	if err != nil {
		return err
	}
	return aws.SetupECRRole(profile, region, "ecrwriter", accountID, aws.ECRWriteActions)
}
