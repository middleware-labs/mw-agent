#!/bin/bash
set -e

CLUSTER="app-cluster"
SERVICE="mw-agent-daemon-service"
REGION="us-east-1"
LAUNCH_JSON="$(cd "$(dirname "$0")" && pwd)/.vscode/launch.json"

TASK_ARN=$(aws ecs list-tasks \
  --cluster "$CLUSTER" \
  --service-name "$SERVICE" \
  --desired-status RUNNING \
  --query 'taskArns[0]' \
  --output text \
  --region "$REGION")

if [ "$TASK_ARN" = "None" ] || [ -z "$TASK_ARN" ]; then
  echo "Error: No running tasks found for $SERVICE"
  exit 1
fi

CI_ARN=$(aws ecs describe-tasks \
  --cluster "$CLUSTER" \
  --tasks "$TASK_ARN" \
  --query 'tasks[0].containerInstanceArn' \
  --output text \
  --region "$REGION")

IID=$(aws ecs describe-container-instances \
  --cluster "$CLUSTER" \
  --container-instances "$CI_ARN" \
  --query 'containerInstances[0].ec2InstanceId' \
  --output text \
  --region "$REGION")

HOST_IP=$(aws ec2 describe-instances \
  --instance-ids "$IID" \
  --query 'Reservations[0].Instances[0].PublicIpAddress' \
  --output text \
  --region "$REGION")

if [ "$HOST_IP" = "None" ] || [ -z "$HOST_IP" ]; then
  echo "Error: No public IP found for instance $IID"
  exit 1
fi

OLD_IP=$(grep -o '"host": "[^"]*"' "$LAUNCH_JSON" | head -1 | cut -d'"' -f4)

if [ "$OLD_IP" = "$HOST_IP" ]; then
  echo "Host IP unchanged: $HOST_IP"
else
  sed -i '' "s/\"host\": \"$OLD_IP\"/\"host\": \"$HOST_IP\"/" "$LAUNCH_JSON"
  echo "Updated launch.json: $OLD_IP → $HOST_IP"
fi
