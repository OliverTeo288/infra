package cmd

import (
	"raid/infra/internal/flows"

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
	RunE: func(cmd *cobra.Command, args []string) error {
		profile, region, err := login(cmd.Context())
		if err != nil {
			return err
		}
		return flows.CreateECRReadRole(cmd.Context(), profile, region)
	},
}

var ecrWriteCmd = &cobra.Command{
	Use:   "write",
	Short: "Create an IAM role with ECR push permissions",
	RunE: func(cmd *cobra.Command, args []string) error {
		profile, region, err := login(cmd.Context())
		if err != nil {
			return err
		}
		return flows.CreateECRWriteRole(cmd.Context(), profile, region)
	},
}

func init() {
	ecrCmd.AddCommand(ecrReadCmd)
	ecrCmd.AddCommand(ecrWriteCmd)
	rootCmd.AddCommand(ecrCmd)
}
