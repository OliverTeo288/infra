#!/usr/bin/env bash
# Build the infra CLI with build-time variables baked in for local testing.
#
# Usage:
#   scripts/build-local.sh                 # build ./infra
#   scripts/build-local.sh run [args...]   # build then run with the given args
#   scripts/build-local.sh floci [args...] # build then run, routed at the local floci emulator
#
# Env overrides (set before invoking):
#   INFRA_GITLAB_HTTPS_DOMAIN
#   INFRA_GITLAB_SSH_DOMAIN
#   INFRA_COMMON_AWS_ACCOUNT_ID
#   INFRA_OIDC_SUB              (build-time OIDC subject pattern)
#   INFRA_OUTPUT                (default: ./infra)
#   FLOCI_ENDPOINT              (default: http://localhost:4566)

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

# Defaults — change these or override via env if you want different values.
: "${INFRA_GITLAB_HTTPS_DOMAIN:=https://gitlab.example.com/group/subgroup/project.git}"
: "${INFRA_GITLAB_SSH_DOMAIN:=git@gitlab.example.com:group/subgroup/project.git}"
: "${INFRA_COMMON_AWS_ACCOUNT_ID:=123456789012}"
: "${INFRA_OIDC_SUB:=project_path:group/subgroup/*:ref_type:branch:ref:main}"
: "${INFRA_OUTPUT:=${REPO_ROOT}/infra}"
: "${FLOCI_ENDPOINT:=http://localhost:4566}"

color() { printf "\033[%sm%s\033[0m\n" "$1" "$2"; }
info() { color "1;34" "==> $*"; }

build() {
  cd "${REPO_ROOT}"
  info "building -> ${INFRA_OUTPUT}"
  info "  GitlabHTTPSDomain  = ${INFRA_GITLAB_HTTPS_DOMAIN}"
  info "  GitlabSSHDomain    = ${INFRA_GITLAB_SSH_DOMAIN}"
  info "  CommonAWSAccountID = ${INFRA_COMMON_AWS_ACCOUNT_ID}"
  info "  OIDCSubjectPattern = ${INFRA_OIDC_SUB}"

  # -s -w strips debug symbols + DWARF, matching what goreleaser produces;
  # roughly halves the binary size.
  go build \
    -ldflags "-s -w \
      -X 'raid/infra/internal/config.GitlabHTTPSDomain=${INFRA_GITLAB_HTTPS_DOMAIN}' \
      -X 'raid/infra/internal/config.GitlabSSHDomain=${INFRA_GITLAB_SSH_DOMAIN}' \
      -X 'raid/infra/internal/config.CommonAWSAccountID=${INFRA_COMMON_AWS_ACCOUNT_ID}' \
      -X 'raid/infra/internal/config.OIDCSubjectPattern=${INFRA_OIDC_SUB}'" \
    -o "${INFRA_OUTPUT}" .

  info "built ${INFRA_OUTPUT}"
}

cmd="${1:-build}"
case "${cmd}" in
  build|"")
    build
    ;;
  run)
    shift || true
    build
    info "running: ${INFRA_OUTPUT} $*"
    "${INFRA_OUTPUT}" "$@"
    ;;
  floci)
    shift || true
    build
    info "routing at floci: ${FLOCI_ENDPOINT}"
    FLOCI_ENDPOINT="${FLOCI_ENDPOINT}" \
    AWS_ACCESS_KEY_ID="test" \
    AWS_SECRET_ACCESS_KEY="test" \
    AWS_REGION="${AWS_REGION:-us-east-1}" \
      "${INFRA_OUTPUT}" "$@"
    ;;
  *)
    echo "unknown command: ${cmd}" >&2
    grep -E "^#" "$0" | sed 's/^# //; s/^#//'
    exit 2
    ;;
esac
