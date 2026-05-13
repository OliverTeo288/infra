#!/usr/bin/env bash
# Pick an AWS profile interactively, then attempt to assume a target role
# to verify cross-account trust + permissions.
#
# Usage:
#   scripts/test-assume-role.sh                       # prompt for role ARN
#   scripts/test-assume-role.sh <role-arn>            # use given role ARN
#
# Env overrides:
#   ROLE_ARN     full IAM role ARN to assume (skips the prompt)

set -euo pipefail

# 1. Pick a profile.
echo "Fetching AWS profiles..."
mapfile -t PROFILES < <(grep '\[profile' ~/.aws/config 2>/dev/null | sed 's/\[profile \(.*\)\]/\1/' | sort)

if [ ${#PROFILES[@]} -eq 0 ]; then
  echo "No AWS profiles found in ~/.aws/config. Run 'aws configure sso' first."
  exit 1
fi

echo "Available AWS profiles:"
for i in "${!PROFILES[@]}"; do
  echo "[$((i + 1))] ${PROFILES[$i]}"
done

read -r -p "Enter the number of your choice: " choice
if [[ ! "$choice" =~ ^[0-9]+$ ]] || [ "$choice" -lt 1 ] || [ "$choice" -gt ${#PROFILES[@]} ]; then
  echo "Invalid choice. Exiting."
  exit 1
fi
SELECTED_PROFILE="${PROFILES[$((choice - 1))]}"
echo "Selected profile: $SELECTED_PROFILE"

# 2. Resolve the target role ARN: CLI arg → env var → interactive prompt.
ROLE_ARN="${1:-${ROLE_ARN:-}}"
if [ -z "$ROLE_ARN" ]; then
  read -r -p "Enter the role ARN to assume (arn:aws:iam::ACCOUNT_ID:role/NAME): " ROLE_ARN
fi
if [ -z "$ROLE_ARN" ]; then
  echo "ROLE_ARN is required."
  exit 1
fi

echo "Testing assume role to: $ROLE_ARN"

set +e
OUTPUT=$(aws sts assume-role \
  --role-arn "$ROLE_ARN" \
  --role-session-name "test-session" \
  --profile "$SELECTED_PROFILE" \
  --output json 2>&1)
EXIT_CODE=$?
set -e

if [ $EXIT_CODE -eq 0 ]; then
  echo "SUCCESS: Successfully assumed role $ROLE_ARN"
  echo "$OUTPUT"
else
  echo "FAILED: Could not assume role $ROLE_ARN"
  echo "$OUTPUT"
  exit $EXIT_CODE
fi
