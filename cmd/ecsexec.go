package cmd

import (
	"raid/infra/internal/flows"

	"github.com/spf13/cobra"
)

var (
	ecsExecShell string

	ecsCmd = &cobra.Command{
		Use:   "ecs",
		Short: "ECS-related commands",
	}

	ecsExecCmd = &cobra.Command{
		Use:   "exec",
		Short: "Execute a shell interactively in an ECS container",
		Long:  "Interactively pick an ECS cluster/service/task/container and open a shell session via ECS Exec.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return flows.ECSExec(cmd.Context(), ecsExecShell)
		},
	}
)

func init() {
	ecsExecCmd.Flags().StringVar(&ecsExecShell, "shell", "/bin/sh",
		"Shell to launch in the container (e.g. /bin/bash for images without sh)")
	ecsCmd.AddCommand(ecsExecCmd)
	rootCmd.AddCommand(ecsCmd)
}
