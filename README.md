# infra

`infra` is a command-line tool designed to ease AWS pain points, streamlining tasks like AWS resource management, GitOps setup, and infrastructure initialization.

## Features

- **`infra portforward`**: Port forwards into a private RDS instance through ECS Fargate & EC2 using SSM.
- **`infra ecs exec`**: Execute shell commands interactively in ECS containers.
- **`infra init`**: Initializes your repository by:
  1. Creating Terraform GitOps templates.
  2. Creating an S3 state bucket for Terraform.
  3. Creating an IAM role for GitOps integration.
- **`infra ecr read`**: Creates an IAM role named 'ecrreader' with read-only ECR permissions and cross-account trust relationship.
- **`infra ecr write`**: Creates an IAM role named 'ecrwriter' with ECR push permissions and cross-account trust relationship.

## Installation via Homebrew

To install the `infra` CLI via Homebrew, use the following commands:

```bash
brew tap oliverteo288/infra
brew install oliverteo288/infra/infra
```

### Supported Architectures

- **Linux**: amd64, arm64
- **macOS**: amd64 (Intel), arm64 (Apple Silicon)
- **Windows**: amd64, arm64

Usage
-----

### Prerequisites

1. Ensure you have `awscli` installed
2. Ensure you have `session-manager-plugin` installed for portforwarding and ECS exec
3. Ensure you have an **AWS profile** configured with proper permissions.
4. Ensure you have **SHIPHATS access**. Without access, you will not be able to create the Terraform GitOps template.
5. See [IAM_PERMISSIONS.md](IAM_PERMISSIONS.md) for detailed IAM permissions required for each command


### Commands

#### 1\. **`infra portforward`**

This command allows you to port-forward into a private RDS instance via ECS Fargate or EC2 using SSM.

Example usage:

```
infra portforward
```

#### 2\. **`infra ecs exec`**

This command allows you to execute shell commands interactively in ECS containers.

Example usage:

```
infra ecs exec
```

#### 3\. **`infra init`**

The `init` command sets up your repository and AWS resources for Terraform GitOps. It includes:

-   Creating Terraform GitOps templates.
-   Creating an S3 bucket for Terraform state management.
-   Creating an IAM role for Terraform GitOps.

#### 4\. **`infra ecr read`**

Creates an IAM role named 'ecrreader' with read-only ECR permissions (pull, describe, list) and cross-account trust.

```
infra ecr read
```

#### 5\. **`infra ecr write`**

Creates an IAM role named 'ecrwriter' with ECR push permissions (upload, put image) and cross-account trust.

```
infra ecr write
```

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

---

## IAM Permissions

### `infra portforward`

Required permissions for port forwarding into RDS via ECS/EC2:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "rds:DescribeDBInstances",
        "rds:DescribeDBProxies",
        "ec2:DescribeInstances",
        "ec2:DescribeRegions",
        "ecs:ListClusters",
        "ecs:ListServices",
        "ecs:ListTasks",
        "ecs:DescribeTasks"
      ],
      "Resource": "*"
    },
    {
      "Effect": "Allow",
      "Action": "ssm:StartSession",
      "Resource": [
        "arn:aws:ec2:*:*:instance/*",
        "arn:aws:ecs:*:*:task/*",
        "arn:aws:ssm:*:*:document/AWS-StartPortForwardingSessionToRemoteHost"
      ]
    }
  ]
}
```

### `infra ecs exec`

Required permissions for executing commands in ECS containers:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "ecs:ListClusters",
        "ecs:ListServices",
        "ecs:ListTasks",
        "ecs:DescribeTasks",
        "ecs:ExecuteCommand",
        "ec2:DescribeRegions"
      ],
      "Resource": "*"
    },
    {
      "Effect": "Allow",
      "Action": "ssm:StartSession",
      "Resource": [
        "arn:aws:ecs:*:*:task/*",
        "arn:aws:ssm:*:*:document/AmazonECS-ExecuteInteractiveCommand"
      ]
    }
  ]
}
```

### `infra init`

Required permissions for initializing Terraform GitOps setup:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "s3:CreateBucket",
        "s3:PutBucketVersioning",
        "s3:PutBucketPolicy",
        "s3:PutObject"
      ],
      "Resource": [
        "arn:aws:s3:::*-backend-tf-*",
        "arn:aws:s3:::*-backend-tf-*/*"
      ]
    },
    {
      "Effect": "Allow",
      "Action": [
        "iam:CreateRole",
        "iam:PutRolePolicy",
        "iam:CreateOpenIDConnectProvider",
        "iam:GetOpenIDConnectProvider"
      ],
      "Resource": [
        "arn:aws:iam::*:role/TerraformGitopsRole",
        "arn:aws:iam::*:oidc-provider/sgts.gitlab-dedicated.com"
      ]
    },
    {
      "Effect": "Allow",
      "Action": [
        "sts:GetCallerIdentity",
        "ec2:DescribeRegions"
      ],
      "Resource": "*"
    }
  ]
}
```

### `infra ecr read` / `infra ecr write`

Required permissions for creating ECR roles:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "iam:CreateRole",
        "iam:PutRolePolicy"
      ],
      "Resource": [
        "arn:aws:iam::*:role/ecrreader",
        "arn:aws:iam::*:role/ecrwriter"
      ]
    },
    {
      "Effect": "Allow",
      "Action": [
        "sts:GetCallerIdentity",
        "ec2:DescribeRegions"
      ],
      "Resource": "*"
    }
  ]
}
```

For more detailed information about IAM permissions, see [IAM_PERMISSIONS.md](IAM_PERMISSIONS.md).
