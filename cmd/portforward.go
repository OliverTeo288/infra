package cmd

import (
	"fmt"
	"os"

	"raid/infra/internal/functions"

	"github.com/spf13/cobra"
)

// portforwardCmd represents the portforward command
var portforwardCmd = &cobra.Command{
	Use:   "portforward",
	Short: "Making it easier for you to portfoward into your Private RDS from your ECS",
	Long: `Automatically discovers your ECS Tasks and RDS hostname to start an SSM session for DB management.
	
	Ensure that you have the following things in place (1) AWS profile conifgured (2) ECS Fargate has SSM access enabled
	`,
	Run: func(cmd *cobra.Command, args []string) {
		err := functions.ExecutePortForwarding()
		if err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(portforwardCmd)
}

