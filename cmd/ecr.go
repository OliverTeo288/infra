package cmd

import (
	"fmt"
	"os"
	"raid/infra/internal/functions"
	"raid/infra/internal/utils"

	"github.com/spf13/cobra"
)

var ecrCmd = &cobra.Command{
	Use:   "ecr",
	Short: "Manage ECR IAM roles",
	Long:  "Create IAM roles for ECR image pull (read) or push (write) with cross-account trust.",
}

var ecrReadCmd = &cobra.Command{
	Use:   "read",
	Short: "Create an IAM role with read-only ECR permissions",
	Run: func(cmd *cobra.Command, args []string) {
		profile, region, err := utils.Login()
		if err != nil {
			fmt.Println("Error logging in:", err)
			os.Exit(1)
		}
		if err := functions.CreateECRReadRole(profile, region); err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
	},
}

var ecrWriteCmd = &cobra.Command{
	Use:   "write",
	Short: "Create an IAM role with ECR push permissions",
	Run: func(cmd *cobra.Command, args []string) {
		profile, region, err := utils.Login()
		if err != nil {
			fmt.Println("Error logging in:", err)
			os.Exit(1)
		}
		if err := functions.CreateECRWriteRole(profile, region); err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(ecrCmd)
	ecrCmd.AddCommand(ecrReadCmd)
	ecrCmd.AddCommand(ecrWriteCmd)
}
