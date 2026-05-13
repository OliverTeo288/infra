# infra

`infra` is a command-line tool designed to ease AWS pain points, streamlining tasks like AWS resource management, GitOps setup, and infrastructure initialization.

## Features

- **`infra portforward`** — Port-forward into a private RDS instance through ECS Fargate or EC2 over SSM.
- **`infra ecs exec`** — Execute shell commands interactively in ECS containers (`--shell` overrides the default `/bin/sh`).
- **`infra init`** — End-to-end project bootstrap: clone template repo, create S3 state bucket, create GitOps IAM role + OIDC provider. Use `--auto-approve`/`-a` to skip between-step prompts.
- **`infra ecr read`** — Creates an IAM role named `ecrreader` with read-only ECR permissions and cross-account trust.
- **`infra ecr write`** — Creates an IAM role named `ecrwriter` with ECR push permissions and cross-account trust.

All commands support `--log-format json|text` and `--log-level debug|info|warn|error` (also via `INFRA_LOG_FORMAT` / `INFRA_LOG_LEVEL`).

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

## Project layout

```
cmd/                      cobra command tree (thin RunE adapters)
internal/
  awsauth/                AWS profile discovery + SSO login
  awscfg/                 SDK config loader (transparently handles floci routing)
  config/                 build-time + env config
  ec2/  ecs/  rds/        SDK-backed wrappers, one per AWS service
  flows/                  end-to-end orchestration (one func per CLI command)
  iam/                    IAM role + OIDC + trust-policy primitives
  observability/          slog setup + AuditHook contract
  prompt/                 interactive stdin prompts
  s3/                     S3 bucket bootstrap
  scaffold/               clone-template repo workflow
```

See [CONTRIBUTING.md](CONTRIBUTING.md) for the rules of thumb on where new
code belongs.

### Additional Notes

-   The `infra init` process requires your AWS profile to have the necessary permissions for creating resources such as S3 buckets and IAM roles.
-   The `infra portforward` command requires the appropriate ECS and RDS access through your AWS profile.
-   Ensure that your AWS profile has sufficient permissions to manage the resources created by this tool (IAM, S3, etc.).

---

## Testing

The codebase has two tiers of tests:

| Tier | Scope | How to run | Requires |
|---|---|---|---|
| **Unit** | Pure-Go logic: trust-policy generators, config URL parsing, CWD-empty detection, etc. | `go test ./...` | Go toolchain only |
| **Integration** | Real AWS SDK calls against a local [floci](https://github.com/floci-io/floci) emulator: IAM role creation, S3 bucket setup, OIDC provider idempotency | `scripts/test-floci.sh` | Docker + floci image |

Integration tests are guarded by `//go:build integration` so they never run in plain `go test ./...`.

### Running unit tests

```bash
go test -race ./...
```

### Running integration tests against floci

The repo ships a helper script that boots floci, runs the integration suite, and tears the container down on exit:

```bash
# Full cycle (recommended): start floci → run tests → tear down
scripts/test-floci.sh

# Just bring up floci and leave it running
scripts/test-floci.sh up

# Drop into a subshell with FLOCI_ENDPOINT + dummy AWS creds exported,
# so you can exercise the CLI itself against floci
scripts/test-floci.sh shell
./infra portforward      # hits floci instead of real AWS

# Stop the container
scripts/test-floci.sh down
```

Environment overrides (set before invoking the script):

| Variable | Default | Purpose |
|---|---|---|
| `FLOCI_IMAGE` | `floci/floci:latest` | Docker image tag |
| `FLOCI_PORT` | `4566` | Host port to bind |
| `FLOCI_CONTAINER` | `floci-infra-test` | Docker container name |
| `FLOCI_WAIT_SECS` | `60` | Healthcheck timeout |

### How floci routing works

When the `FLOCI_ENDPOINT` env var is set, `internal/awscfg.LoadConfig` automatically:

- Points the AWS SDK at that endpoint
- Uses static dummy credentials (`test`/`test`) — bypasses SSO / shared config

This is the same code path used by the integration tests and by `scripts/test-floci.sh shell`. Production builds are unaffected — when `FLOCI_ENDPOINT` is empty, the SDK resolves credentials normally.

### CI

`.github/workflows/ci.yml` runs four jobs on every push/PR and daily at 02:00 SGT (GMT+8):

- **lint** — `gofmt`, `go vet`, `staticcheck`
- **unit** — `go build` + `go test -race`, with a pass/fail/skip breakdown
- **security** — Trivy filesystem scan (Go dependencies, secrets, misconfig); findings uploaded to the GitHub Security tab as SARIF
- **integration** — boots a floci service container and runs `go test -tags=integration`

Every job writes a markdown report to `$GITHUB_STEP_SUMMARY`, so a single click on a workflow run shows lint results, test counts, security findings, and failed-test names without drilling into logs.

The daily schedule catches AWS SDK minor releases, floci API drift, and dependency vulnerability advisories before they land in a release.

### Logging

Both the CLI and the integration tests emit structured logs via `log/slog`. Useful flags / env vars:

```bash
./infra --log-format json --log-level debug portforward
INFRA_LOG_FORMAT=json INFRA_LOG_LEVEL=debug ./infra ...
```

Every AWS-mutating action (`iam:CreateRole`, `s3:CreateBucket`, etc.) emits one audit-style log line tagged with a session ID so a multi-step flow (e.g. `infra init`) can be correlated end-to-end. JSON output is ready to ship to Datadog.

---

## IAM Permissions

Per-command IAM policies live in [IAM_PERMISSIONS.md](IAM_PERMISSIONS.md).
Attach the relevant policy to the AWS profile / role you use when invoking
the CLI.
