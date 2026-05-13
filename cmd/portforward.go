package cmd

import (
	"raid/infra/internal/flows"

	"github.com/spf13/cobra"
)

var portforwardCmd = &cobra.Command{
	Use:   "portforward",
	Short: "Port-forward into a private RDS instance via ECS or EC2 over SSM",
	Long: `Port-forward into a private RDS instance from your local machine.

Prereqs:
  - AWS profile configured
  - session-manager-plugin installed
  - ECS Fargate task has SSM access enabled (when targeting ECS)`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return flows.PortForward(cmd.Context())
	},
}

func init() {
	rootCmd.AddCommand(portforwardCmd)
}
