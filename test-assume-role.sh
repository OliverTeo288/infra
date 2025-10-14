#!/bin/bash

# Test script to assume role with profile selection

# Get AWS profiles and sort alphabetically
echo "Fetching AWS profiles..."
PROFILES=$(grep '\[profile' ~/.aws/config | sed 's/\[profile \(.*\)\]/\1/' | sort)

if [ -z "$PROFILES" ]; then
    echo "No AWS profiles found. Please configure profiles first."
    exit 1
fi

# Display profiles for selection
echo "Available AWS profiles:"
i=1
declare -a profile_array
while IFS= read -r profile; do
    echo "[$i] $profile"
    profile_array[$i]="$profile"
    ((i++))
done <<< "$PROFILES"

# Prompt user to select profile
read -p "Enter the number of your choice: " choice

if [[ ! "$choice" =~ ^[0-9]+$ ]] || [ "$choice" -lt 1 ] || [ "$choice" -ge "$i" ]; then
    echo "Invalid choice. Exiting."
    exit 1
fi

SELECTED_PROFILE="${profile_array[$choice]}"
echo "Selected profile: $SELECTED_PROFILE"

# Test assume role
ROLE_ARN="arn:aws:iam::637423367961:role/crossacc-ecrreader"
echo "Testing assume role to: $ROLE_ARN"

# Attempt to assume role
set +e  # Disable exit on error for this command
OUTPUT=$(aws sts assume-role \
    --role-arn "$ROLE_ARN" \
    --role-session-name "test-session" \
    --profile "$SELECTED_PROFILE" \
    --output json 2>&1)
EXIT_CODE=$?
set -e  # Re-enable exit on error

if [ $EXIT_CODE -eq 0 ]; then
    echo "✅ SUCCESS: Successfully assumed role $ROLE_ARN"
    echo "Response:"
    echo "$OUTPUT"
else
    echo "❌ FAILED: Could not assume role $ROLE_ARN"
    echo "Error details:"
    echo "$OUTPUT"
fi