// Package cmd wires up the cobra command tree. Each command is a thin
// adapter over a function in the flows package: it parses flags, threads a
// session-tagged context, and lets cobra handle exit codes via RunE.
package cmd

import (
	"context"
	"log/slog"
	"os"

	"raid/infra/internal/observability"

	"github.com/spf13/cobra"
)

var (
	logFormat string
	logLevel  string
)

var rootCmd = &cobra.Command{
	Use:           "infra",
	Short:         "AWS CLI to make infrastructure management easier",
	Long:          "infra is a workflow CLI that wraps common AWS operations: port-forwarding into private RDS via ECS or EC2, ECS exec sessions, Terraform GitOps role + state bucket bootstrap, and cross-account ECR roles.",
	SilenceUsage:  true,
	SilenceErrors: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		observability.Init(logFormat, logLevel)
		// Attach a session ID to the cobra context so every flow under a
		// single CLI invocation shares one correlation key.
		cmd.SetContext(observability.WithSession(cmd.Context()))
		return nil
	},
}

// Execute runs the cobra command tree with a background context attached.
// Errors returned from RunE handlers are logged at error level and the
// process exits non-zero.
func Execute() {
	rootCmd.SetContext(context.Background())
	if err := rootCmd.Execute(); err != nil {
		slog.Error("command failed", "err", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&logFormat, "log-format", "text",
		`Log output format: "text" or "json" (env: INFRA_LOG_FORMAT)`)
	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "info",
		`Log level: debug|info|warn|error (env: INFRA_LOG_LEVEL)`)
}
