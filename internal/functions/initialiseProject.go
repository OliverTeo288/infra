
package functions

import (
	"fmt"
	
	"raid/infra/internal/utils"
)

var GitlabHttpsDomain string
var GitlabSshDomain string

func InitialiseProject() error {
	// Step 1: Confirm access
	if !utils.ConfirmPrompt("Do you have access to SHIPHATS GitLab? (Y/N)") {
		return fmt.Errorf("please ensure you have access to SHIPHATS GitLab before running this command")
	}

	// Step 2: Prompt user to select cloning method
	options := []string{"Clone with SSH", "Clone with HTTPS"}
	choice, err := utils.PromptSelection(options)
	if err != nil {
		return fmt.Errorf("failed to prompt for cloning method: %w", err)
	}

	// Step 3: Determine the domain based on user choice
	var selectedDomain string
	switch choice {
	case "Clone with SSH":
		selectedDomain = GitlabSshDomain
	case "Clone with HTTPS":
		selectedDomain = GitlabHttpsDomain
	default:
		return fmt.Errorf("invalid cloning method selected")
	}

	// Step 4: Clone the repository using the selected domain
	if err := utils.CloneRepo(selectedDomain); err != nil {
		return fmt.Errorf("failed to clone repository: %w", err)
	}

	fmt.Println("Repository successfully cloned.")
	return nil
}