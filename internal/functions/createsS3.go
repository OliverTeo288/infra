
package functions

import (
	
	"raid/infra/internal/utils"
	"raid/infra/internal/aws"
)


func CreateS3(profile, region string) error {
		bucketName, err := utils.PromptInput("Enter the name of the S3 bucket:", nil, "test-dev-backend-tf-0000")
		if err != nil {
			return err
		}

		err = aws.CreateS3Bucket(profile, region, bucketName)
		if err != nil {
			return err
		}

		return nil
}