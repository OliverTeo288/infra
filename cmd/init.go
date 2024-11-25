package cmd

import (
	"fmt"
	"os"

  "raid/infra/internal/functions"
	// "raid/infra/internal/utils"
	"github.com/spf13/cobra"
)


var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize the project by cloning the repository",
	Long: `This command initializes the project by cloning a specified repository.
Ensure you have the required access before running this command.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Execute the initialization logic
		if err := functions.InitializeProject(); err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
	},
}

// var createS3BucketCmd = &cobra.Command{
// 	Use:   "create-s3-bucket",
// 	Short: "Create an S3 bucket for the project",
// 	Long:  "This subcommand allows you to create an S3 bucket for the project with a specified AWS profile and region.",
// 	Run: func(cmd *cobra.Command, args []string) {

// 		// Step 1: Login to AWS
// 		selectedProfile, selectedRegion, err := utils.Login()
// 		if err != nil {
// 			return err
// 		}

// 		fmt.Printf("Login successful!")

// 		// Step 1: Prompt for bucket name
// 		bucketName, err := utils.PromptInput("Enter the name of the S3 bucket:")
// 		if err != nil {
// 			fmt.Println("Error:", err)
// 			os.Exit(1)
// 		}


// 		// Step 4: Call utility to create the S3 bucket
// 		err = utils.CreateS3Bucket(bucketName, selectedProfile, selectedRegion)
// 		if err != nil {
// 			fmt.Println("Error creating S3 bucket:", err)
// 			os.Exit(1)
// 		}

// 		fmt.Printf("S3 bucket %s created successfully in region %s using profile %s.\n", bucketName, selectedRegion, selectedProfile)
// 	},
// }


func init() {
	rootCmd.AddCommand(initCmd)
	// initCmd.AddCommand(createS3BucketCmd)
}


