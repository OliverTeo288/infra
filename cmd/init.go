package cmd

import (
	"fmt"

	"raid/infra/internal/flows"
	"raid/infra/internal/observability"
	"raid/infra/internal/prompt"

	"github.com/spf13/cobra"
)

var autoApprove bool

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Bootstrap a project: clone template, create S3 state bucket, create GitOps role",
	Long: `Initialise a project end-to-end:
  1. Clone the templated repo
  2. Create the S3 bucket for Terraform remote state
  3. Create the GitOps IAM role + OIDC provider

Use --auto-approve to skip the between-step confirmation prompts.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		log := observability.FromContext(ctx)

		if err := flows.InitialiseProject(); err != nil {
			return err
		}

		if !autoApprove && !prompt.Confirm("Do you want to proceed to creating S3 terraform state bucket? (Y/N)") {
			log.Info("init_step_skipped", "step", "s3_bucket", "reason", "user_declined")
			fmt.Println("Exiting script.")
			return nil
		}

		profile, region, err := login(ctx)
		if err != nil {
			// The cloning step has already completed; surface that so the
			// user knows there are files in CWD they may want to clean up.
			fmt.Fprintln(cmd.ErrOrStderr(),
				"Warning: template repo was cloned into the current directory. "+
					"You may want to remove those files before retrying.")
			return err
		}

		if err := flows.CreateS3StateBucket(ctx, profile, region); err != nil {
			return err
		}

		if !autoApprove && !prompt.Confirm("Do you want to proceed to GitOps role creation? (Y/N)") {
			log.Info("init_step_skipped", "step", "gitops_role", "reason", "user_declined")
			fmt.Println("Exiting script.")
			return nil
		}

		return flows.CreateGitOpsRole(ctx, profile, region)
	},
}

var initRepoCmd = &cobra.Command{
	Use:   "repo",
	Short: "Clone the templated project repository only",
	Long:  "Subcommand: clone the templated DevOps repo without creating any AWS resources.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return flows.InitialiseProject()
	},
}

var initS3Cmd = &cobra.Command{
	Use:   "s3",
	Short: "Create the S3 Terraform state bucket only",
	RunE: func(cmd *cobra.Command, args []string) error {
		profile, region, err := login(cmd.Context())
		if err != nil {
			return err
		}
		return flows.CreateS3StateBucket(cmd.Context(), profile, region)
	},
}

var initRoleCmd = &cobra.Command{
	Use:   "role",
	Short: "Create the GitOps IAM role + OIDC provider only",
	RunE: func(cmd *cobra.Command, args []string) error {
		profile, region, err := login(cmd.Context())
		if err != nil {
			return err
		}
		return flows.CreateGitOpsRole(cmd.Context(), profile, region)
	},
}

func init() {
	initCmd.Flags().BoolVarP(&autoApprove, "auto-approve", "a", false,
		"Skip confirmation prompts between steps")

	initCmd.AddCommand(initRepoCmd)
	initCmd.AddCommand(initS3Cmd)
	initCmd.AddCommand(initRoleCmd)
	rootCmd.AddCommand(initCmd)
}
