
package functions

import (
	"regexp"
	"errors"

	"net/url"
	"raid/infra/internal/utils"
	"raid/infra/internal/aws"
)


func CreateGitopsRole(profile, region string) error {

		validate := func(input string) error {
			// Regex to allow only alphanumeric characters and dashes, no spaces
			re := regexp.MustCompile(`^[a-zA-Z0-9\-]+$`)
			if !re.MatchString(input) {
				return errors.New("role name must only contain alphanumeric characters or dashes, and no spaces")
			}
			return nil
		}
		
		roleName, err := utils.PromptInput("Enter the name of IAM Role for Terraform Gitops", validate, "TerraformGitopsRole")
		if err != nil {
			return err
		}

		err = aws.SetupRole(profile, region, roleName)
		if err != nil {
			return err
		}


	  // Extract the base URL from GitlabHttpsDomain
		parsedURL, err := url.Parse(GitlabHttpsDomain)
		if err != nil {
			return errors.New("invalid GitLab HTTPS domain provided")
		}
		baseURL := parsedURL.Scheme + "://" + parsedURL.Host
		// Create the OIDC provider
		if err := aws.CreateOIDCProvider(profile, region, baseURL); err != nil {
			return err
		}

		return nil
}