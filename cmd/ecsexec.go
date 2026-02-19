package cmd

import (
	"fmt"
	"os"

	"raid/infra/internal/functions"

	"github.com/spf13/cobra"
)

var ecsExecCmd = &cobra.Command{
	Use:   "ecs exec",
	Short: "Execute shell commands in ECS containers",
	Long:  `Interactively select your ECS cluster, service, task, and container to exec into with a shell session.`,
	Run: func(cmd *cobra.Command, args []string) {
		err := functions.ExecuteECSExec()
		if err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(ecsExecCmd)
}
