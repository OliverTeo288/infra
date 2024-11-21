package cmd

import (
	"fmt"
	"os"

  "raid/infra/internal/utils"
	"github.com/spf13/cobra"
)
var GitlabHttpsDomain string
var GitlabSshDomain string

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize the project by cloning the repository",
	Long: `This command initializes the project by cloning a specified repository.
Ensure you have the required access before running this command.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Step 1: Confirm access
		if !utils.ConfirmAccess() {
			fmt.Println("Please ensure you have access to SHIPHATS GitLab before running this command.")
			os.Exit(1)
		}

		// Step 2: Prompt user to select cloning method
		options := []string{"Clone with SSH", "Clone with HTTPS"}
		choice, err := utils.PromptSelection(options)
		if err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}

		// Step 3: Determine the domain based on user choice
		var selectedDomain string
		switch choice {
		case "Clone with SSH":
			selectedDomain = GitlabSshDomain
		case "Clone with HTTPS":
			selectedDomain = GitlabHttpsDomain
		default:
			fmt.Println("Invalid choice. Exiting.")
			os.Exit(1)
		}

		// Step 4: Clone the repository using the selected domain
		err = utils.CloneRepo(selectedDomain)
		if err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}

		fmt.Println("Repository successfully cloned.")
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}


