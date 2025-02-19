# infra

`infra` is a command-line tool designed to ease AWS pain points, streamlining tasks like AWS resource management, GitOps setup, and infrastructure initialization.

## Features

- **`infra portforward`**: Port forwards into a private RDS instance through ECS Fargate & EC2 using SSM.
- **`infra init`**: Initializes your repository by:
  1. Creating Terraform GitOps templates.
  2. Creating an S3 state bucket for Terraform.
  3. Creating an IAM role for GitOps integration.

## Installation via Homebrew

To install the `infra` CLI via Homebrew, use the following commands:

```bash
brew tap oliverteo288/infra
brew install oliverteo288/infra/infra
```

Usage
-----

### Prerequisites

1.  Ensure you have an **AWS profile** configured with proper permissions.
2.  Ensure you have **SHIPHATS access**. Without access, you will not be able to create the Terraform GitOps template.

### Commands

#### 1\. **`infra portforward`**

This command allows you to port-forward into a private RDS instance via ECS Fargate using SSM.

Example usage:

```
infra portforward
```

#### 2\. **`infra init`**

The `init` command sets up your repository and AWS resources for Terraform GitOps. It includes:

-   Creating Terraform GitOps templates.
-   Creating an S3 bucket for Terraform state management.
-   Creating an IAM role for Terraform GitOps.

##### Auto-approve Option

By using the `--auto-approve` flag (or `-a` for shorthand), you can skip all confirmation prompts and proceed with the default actions.

Example usage:

```
infra init
infra init -a
infra init --auto-approve
```

##### Subcommands

You can also run `infra init` with specific subcommands to create only individual resources:

-   **Create GitOps Templates, S3 Bucket, and IAM Role (default)**:

    ```
    infra init
    ```

-   **Create GitOps Templates Only**:

    ```
    infra init repo
    ```

-   **Create Only the S3 Bucket**:

    ```
    infra init s3
    ```

-   **Create Only the IAM Role**:

    ```
    infra init role
    ```

### Additional Notes

-   The `infra init` process requires your AWS profile to have the necessary permissions for creating resources such as S3 buckets and IAM roles.
-   The `infra portforward` command requires the appropriate ECS and RDS access through your AWS profile.
-   Ensure that your AWS profile has sufficient permissions to manage the resources created by this tool (IAM, S3, etc.).


Update gitlab domain