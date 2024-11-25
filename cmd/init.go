package cmd

import (
	"fmt"
	"os"

  "raid/infra/internal/functions"
	"raid/infra/internal/utils"

	"github.com/spf13/cobra"
)


var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialise the project by cloning the repository",
	Long: `This command initialises the project by cloning a specified repository.
Ensure you have the required access before running this command.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Check if the `--auto-approve` flag is set
		autoApprove, err := cmd.Flags().GetBool("auto-approve")
		if err != nil {
			fmt.Println("Error parsing flags:", err)
			os.Exit(1)
		}

		// Execute the initialization logic
		if err := functions.InitialiseProject(); err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}

		// Prompt to continue if not auto-approved
		if !autoApprove && !utils.ConfirmPrompt("Do you want to proceed to creating S3 terraform state bucket? (Y/N)") {
			fmt.Println("Exiting script.")
			return
		}

		// Step 1: Login to AWS
		selectedProfile, selectedRegion, err := utils.Login()
		if err != nil {
			fmt.Println("Error logging in:", err)
			os.Exit(1)
		}

		fmt.Println("Login successful!")


		// Step 2: Create S3 Bucket
		if err := functions.CreateS3(selectedProfile, selectedRegion); err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}

		// Prompt to continue if not auto-approved
		if !autoApprove && !utils.ConfirmPrompt("Do you want to proceed to GitOps role creation? (Y/N)") {
			fmt.Println("Exiting script.")
			return
		}

		// Step 3: Create GitOps Role
		if err := functions.CreateGitopsRole(selectedProfile, selectedRegion); err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
	},
}

var initiliseProject = &cobra.Command{
	Use:   "repo",
	Short: "Clones RAiD's templated project from SHIPHATS GitLab",
	Long:  "This subcommand allows you to clone templated project which meant for DevOps engineer to work with.",
	Run: func(cmd *cobra.Command, args []string) {
		if err := functions.InitialiseProject(); err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
	},
}

var createS3BucketCmd = &cobra.Command{
	Use:   "s3",
	Short: "Create an S3 bucket for the project",
	Long:  "This subcommand allows you to create an S3 bucket for the project with a specified AWS profile and region.",
	Run: func(cmd *cobra.Command, args []string) {

		// Step 1: Login to AWS
		selectedProfile, selectedRegion, err := utils.Login()
		if err != nil {
			fmt.Println("Error logging in:", err)
			os.Exit(1)
		}

		fmt.Println("Login successful!")

		// Step 2: Create S3 Bucket
		if err := functions.CreateS3(selectedProfile, selectedRegion); err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
	},
}

var createGitopsRole = &cobra.Command{
	Use:   "role",
	Short: "Create an IAM Role for the Gitops in Gitlab",
	Long:  "This subcommand allows you to create an IAM Role for the project with a specified AWS profile and region.",
	Run: func(cmd *cobra.Command, args []string) {

		// Step 1: Login to AWS
		selectedProfile, selectedRegion, err := utils.Login()
		if err != nil {
			fmt.Println("Error logging in:", err)
			os.Exit(1)
		}

		fmt.Println("Login successful!")

		// Step 2: Create S3 Bucket
		if err := functions.CreateGitopsRole(selectedProfile, selectedRegion); err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
	},
}



func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().BoolP("auto-approve", "a" , false, "Skip confirmation prompts and proceed automatically")

	initCmd.AddCommand(createS3BucketCmd)
	initCmd.AddCommand(createGitopsRole)
	initCmd.AddCommand(initiliseProject)
}


