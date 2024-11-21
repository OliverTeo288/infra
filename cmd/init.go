package cmd

import (
	"fmt"
	"os"

  "raid/infra/internal/utils"
	"github.com/spf13/cobra"
)
var GitlabHttpsDomain string

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize the project by cloning the repository",
	Long: `This command initializes the project by cloning a specified repository.
Ensure you have the required access before running this command.`,
	Run: func(cmd *cobra.Command, args []string) {
		if !utils.ConfirmAccess() {
			fmt.Println("Please ensure you have access to SHIPHATS GitLab before running this command.")
			os.Exit(1)
		}

		err := utils.CloneRepo(GitlabHttpsDomain)
		if err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}


