#!/bin/bash

# Test script to assume ECR role and pull/push image
set -e

read -p "Enter AWS Access Key ID: " AWS_ACCESS_KEY_ID
read -s -p "Enter AWS Secret Access Key: " AWS_SECRET_ACCESS_KEY
echo
read -p "Enter Account ID to assume role into: " ACCOUNT_ID
read -p "Enter role type (read/write): " ROLE_TYPE

case "$ROLE_TYPE" in
  read)  ROLE_NAME="ecrreader" ;;
  write) ROLE_NAME="ecrwriter" ;;
  *)     echo "Invalid role type. Use 'read' or 'write'."; exit 1 ;;
esac

ROLE_ARN="arn:aws:iam::$ACCOUNT_ID:role/$ROLE_NAME"
read -p "Enter ECR Image (<aws_account_id>.dkr.ecr.ap-southeast-1.amazonaws.com/<ecr_repo>:<image_tag>): " ECR_IMAGE
REGION="ap-southeast-1"

export AWS_ACCESS_KEY_ID
export AWS_SECRET_ACCESS_KEY

echo "1. Assuming role: $ROLE_ARN"

ASSUME_ROLE_OUTPUT=$(aws sts assume-role \
    --role-arn "$ROLE_ARN" \
    --role-session-name "ecr-test-session" \
    --region "$REGION" \
    --output json)

export AWS_ACCESS_KEY_ID=$(echo "$ASSUME_ROLE_OUTPUT" | jq -r '.Credentials.AccessKeyId')
export AWS_SECRET_ACCESS_KEY=$(echo "$ASSUME_ROLE_OUTPUT" | jq -r '.Credentials.SecretAccessKey')
export AWS_SESSION_TOKEN=$(echo "$ASSUME_ROLE_OUTPUT" | jq -r '.Credentials.SessionToken')

ASSUMED_ACCOUNT_ID=$(echo "$ASSUME_ROLE_OUTPUT" | jq -r '.AssumedRoleUser.Arn' | cut -d':' -f5)

echo "2. Successfully assumed role into AWS Account: $ASSUMED_ACCOUNT_ID"

echo "3. Listing all ECR repositories:"
aws ecr describe-repositories --region "$REGION" --output table --query 'repositories[*].[repositoryName,repositoryUri]' || echo "No repositories found or insufficient permissions"
echo

if [[ ! "$ECR_IMAGE" == *".amazonaws.com"* ]]; then
    ECR_IMAGE="$ASSUMED_ACCOUNT_ID.dkr.ecr.ap-southeast-1.amazonaws.com/$ECR_IMAGE"
    echo "Using full ECR URL: $ECR_IMAGE"
fi

echo "4. Getting ECR login token"
aws ecr get-login-password --region "$REGION" | docker login --username AWS --password-stdin "$ASSUMED_ACCOUNT_ID.dkr.ecr.ap-southeast-1.amazonaws.com"

if [ "$ROLE_TYPE" = "read" ]; then
    echo "5. Pulling ECR image: $ECR_IMAGE"
    docker pull "$ECR_IMAGE"
else
    echo "5. Pushing ECR image: $ECR_IMAGE"
    docker push "$ECR_IMAGE"
fi

echo "6. Test completed successfully!"
