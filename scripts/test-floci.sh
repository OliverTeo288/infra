#!/usr/bin/env bash
# Run the infra integration tests (and optionally the CLI itself) against a
# local floci AWS emulator.
#
# Usage:
#   scripts/test-floci.sh                 # run integration tests, tear down on exit
#   scripts/test-floci.sh test            # same as above (explicit)
#   scripts/test-floci.sh shell           # start floci, drop you into an env-loaded subshell
#   scripts/test-floci.sh up              # start floci and leave it running
#   scripts/test-floci.sh down            # stop and remove the floci container
#
# Env overrides:
#   FLOCI_IMAGE       (default: floci/floci:latest)
#   FLOCI_PORT        (default: 4566)
#   FLOCI_CONTAINER   (default: floci-infra-test)
#   FLOCI_WAIT_SECS   (default: 60)

set -euo pipefail

FLOCI_IMAGE="${FLOCI_IMAGE:-floci/floci:latest}"
FLOCI_PORT="${FLOCI_PORT:-4566}"
FLOCI_CONTAINER="${FLOCI_CONTAINER:-floci-infra-test}"
FLOCI_WAIT_SECS="${FLOCI_WAIT_SECS:-60}"
FLOCI_ENDPOINT="http://localhost:${FLOCI_PORT}"

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

color() { printf "\033[%sm%s\033[0m\n" "$1" "$2"; }
info() { color "1;34" "==> $*"; }
err()  { color "1;31" "ERR $*" >&2; }

require_docker() {
  if ! command -v docker >/dev/null 2>&1; then
    err "docker not found in PATH"; exit 127
  fi
  if ! docker info >/dev/null 2>&1; then
    err "docker daemon is not running"; exit 1
  fi
}

floci_running() {
  [[ -n "$(docker ps -q --filter "name=^/${FLOCI_CONTAINER}$" 2>/dev/null)" ]]
}

floci_exists() {
  [[ -n "$(docker ps -aq --filter "name=^/${FLOCI_CONTAINER}$" 2>/dev/null)" ]]
}

start_floci() {
  if floci_running; then
    info "floci already running as ${FLOCI_CONTAINER}"
    return 0
  fi
  if floci_exists; then
    info "removing stale ${FLOCI_CONTAINER} container"
    docker rm -f "${FLOCI_CONTAINER}" >/dev/null
  fi

  info "starting floci (${FLOCI_IMAGE}) on :${FLOCI_PORT}"
  docker run -d --rm \
    --name "${FLOCI_CONTAINER}" \
    -p "${FLOCI_PORT}:4566" \
    "${FLOCI_IMAGE}" >/dev/null

  info "waiting up to ${FLOCI_WAIT_SECS}s for floci to become reachable"
  local elapsed=0
  while (( elapsed < FLOCI_WAIT_SECS )); do
    if curl -fsS "${FLOCI_ENDPOINT}/_floci/health" >/dev/null 2>&1 \
       || curl -fsS "${FLOCI_ENDPOINT}" >/dev/null 2>&1; then
      info "floci is up"
      return 0
    fi
    sleep 2
    elapsed=$((elapsed + 2))
  done

  err "floci never became reachable on ${FLOCI_ENDPOINT}"
  docker logs "${FLOCI_CONTAINER}" 2>&1 | tail -40 || true
  exit 1
}

stop_floci() {
  if floci_running; then
    info "stopping ${FLOCI_CONTAINER}"
    docker rm -f "${FLOCI_CONTAINER}" >/dev/null
  else
    info "${FLOCI_CONTAINER} is not running"
  fi
}

export_aws_env() {
  export FLOCI_ENDPOINT
  export AWS_ACCESS_KEY_ID="test"
  export AWS_SECRET_ACCESS_KEY="test"
  export AWS_REGION="us-east-1"
}

run_tests() {
  cd "${REPO_ROOT}"
  export_aws_env
  info "running integration tests"
  go test -tags=integration -count=1 -v ./...
}

cmd="${1:-test}"
case "${cmd}" in
  up)
    require_docker
    start_floci
    info "endpoint: ${FLOCI_ENDPOINT}"
    info "stop with: $0 down"
    ;;
  down)
    require_docker
    stop_floci
    ;;
  shell)
    require_docker
    start_floci
    export_aws_env
    info "spawning subshell with FLOCI_ENDPOINT + dummy AWS creds set"
    info "exit the subshell to tear down floci"
    trap 'stop_floci' EXIT
    "${SHELL:-/bin/bash}"
    ;;
  test|"")
    require_docker
    start_floci
    trap 'stop_floci' EXIT
    run_tests
    ;;
  *)
    err "unknown command: ${cmd}"
    grep -E "^#" "$0" | sed 's/^# //; s/^#//'
    exit 2
    ;;
esac
