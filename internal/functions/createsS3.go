
package functions

import (
	"regexp"
	"errors"
	"raid/infra/internal/utils"
	"raid/infra/internal/aws"
)


func CreateS3(profile, region string) error {

		validate := func(input string) error {
			// Regex to allow only alphanumeric characters and dashes, no spaces
			re := regexp.MustCompile(`^[a-zA-Z0-9\-]+$`)
			if !re.MatchString(input) {
				return errors.New("role name must only contain alphanumeric characters or dashes, and no spaces")
			}
			return nil
		}
		
		bucketName, err := utils.PromptInput("Enter the name of the S3 bucket", validate, "test-dev-backend-tf-0000")
		if err != nil {
			return err
		}

		err = aws.CreateS3Bucket(profile, region, bucketName)
		if err != nil {
			return err
		}

		return nil
}