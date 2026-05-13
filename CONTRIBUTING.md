# Contributing to `infra`

Thanks for working on this. The goal here is short: keep the CLI fast,
hermetic.

## Local development

```bash
make help            # list everything below
make build           # build ./infra with default build-time vars baked in
make test            # unit tests (no AWS / no floci)
make test-race       # unit tests + race detector
make test-integration  # boots floci in Docker, runs the -tags=integration suite
make lint            # gofmt + go vet (+ staticcheck if installed)
make fmt             # rewrite all .go files with gofmt
```

Floci is a local AWS emulator. The integration suite is opt-in via the
`integration` build tag and is automatically skipped when `FLOCI_ENDPOINT`
is unset. See [`scripts/test-floci.sh`](scripts/test-floci.sh) for the
docker plumbing.

## Project layout

```
cmd/                      cobra command tree (thin adapters; RunE only)
internal/
  awsauth/                AWS profile discovery + SSO login
  awscfg/                 SDK config loader (handles floci routing)
  config/                 build-time + env config
  ec2/  ecs/  rds/        SDK-backed wrappers, one per service
  flows/                  end-to-end orchestration of CLI commands
  iam/                    IAM role + OIDC + trust-policy primitives
  observability/          slog setup + AuditHook contract
  prompt/                 interactive stdin prompts
  s3/                     S3 bucket bootstrap
  scaffold/               clone-template repo workflow
```

Rules of thumb:

- **`cmd/` is wiring only.** Parse flags, thread `cmd.Context()`, call
  one function in `flows`. No business logic, no AWS imports.
- **`flows/` is orchestration.** Compose primitives from the lower
  packages. One public function per `infra <command>`.
- **AWS service packages (`iam`, `s3`, `ec2`, `ecs`, `rds`)** know
  nothing about `cobra` or `prompt` for their lower-level functions;
  they accept primitives.
- **`internal/observability` is depended on, never the other way.**
  Lower-level packages accept an `observability.AuditHook` rather than
  importing the observability package directly.

## Adding a new command

1. Add a function in `internal/flows/` (e.g. `flows/MyThing(ctx, args...)`)
2. Add a thin `cmd/mything.go` wrapping it in a cobra `RunE`
3. Wire it in via `rootCmd.AddCommand(mythingCmd)` in `init()`
4. Add unit tests in the lowest package that holds testable logic

Don't put cobra commands directly into `init()` blocks of the leaf
packages — keep `cmd/` as the single registration point.

## Build-time secrets

Production builds inject GitLab URLs and the OIDC subject pattern via
`go build -ldflags -X`. See [`scripts/build-local.sh`](scripts/build-local.sh)
for the canonical invocation and [`.goreleaser.yml`](.goreleaser.yml) for
the release pipeline.

Local builds substitute generic placeholders (`gitlab.example.com`,
`group/subgroup/*`) so accidental leaks don't expose internal paths.

## Commit + PR conventions

- Conventional commits (`feat:`, `fix:`, `chore:`, `docs:`, `test:`)
- One concern per PR — bug fixes shouldn't carry refactors
- CI must be green (lint + unit + integration); release jobs only run on
  strict semver tags

## Release

The release workflow only fires when you push a tag matching strict
`MAJOR.MINOR.PATCH` semver (optionally with a pre-release suffix). Any
other tag is ignored — no accidental releases from a `dev` or `wip` tag.

| Example       | Triggers release? |
| ------------- | ----------------- |
| `0.1.0`       | ✅                |
| `5.1.2`       | ✅                |
| `12.0.0`      | ✅                |
| `1.2.3-rc.1`  | ✅                |
| `v1.2.3`      | ❌ (no `v` prefix) |
| `1.2`         | ❌ (missing patch) |
| `1.2.3.4`     | ❌                |
| `dev`, `wip`  | ❌                |

To cut a release:

```bash
git tag -a 0.1.0 -m "Release 0.1.0"
git push --tags
```

The `release` workflow then builds via goreleaser and publishes to the
homebrew tap automatically.

The release pipeline reads four GitHub Actions secrets:

| Secret                  | Purpose                                   |
| ----------------------- | ----------------------------------------- |
| `GITLAB_HTTPS_DOMAIN`   | HTTPS clone URL for the template repo     |
| `GITLAB_SSH_DOMAIN`     | SSH clone URL for the template repo      |
| `COMMON_AWS_ACCOUNT_ID` | Account ID hosting cross-account ECR      |
| `OIDC_SUBJECT_PATTERN`  | GitLab JWT `sub` pattern for trust policy |
