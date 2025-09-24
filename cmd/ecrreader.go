package cmd

import (
	"fmt"
	"os"
	"raid/infra/internal/functions"
	"raid/infra/internal/utils"
	"github.com/spf13/cobra"
)

var ecrReaderRole = &cobra.Command{
	Use:   "ecrrole",
	Short: "Creates an IAM role with permissions to pull ECR images",
	Long: `Create an IAM role with permissions to pull ECR images.

	Additionally, giving access to trust relationship to our management account
	`,
	Run: func(cmd *cobra.Command, args []string) {
		// Step 1: Login to AWS
		selectedProfile, selectedRegion, err := utils.Login()
		if err != nil {
			fmt.Println("Error logging in:", err)
			os.Exit(1)
		}

		fmt.Println("Login successful!")

		// Step 2: Create ECR Role
		err = functions.CreateECRRole(selectedProfile, selectedRegion)
		if err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(ecrReaderRole)
}

