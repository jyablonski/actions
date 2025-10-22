#!/bin/bash
set -e

# Inputs
STATUS="$1"
WEBHOOK_URL="$2"
CUSTOM_MESSAGE="$3"
MENTION_ON_FAILURE="$4"
REPO_NAME="$5"
WORKFLOW_NAME="$6"
BRANCH="$7"
ACTOR="$8"
RUN_URL="$9"

# Set color based on status
if [ "$STATUS" == "success" ]; then
  COLOR="good"
  EMOJI="✅"
  STATUS_TEXT="succeeded"
elif [ "$STATUS" == "failure" ]; then
  COLOR="danger"
  EMOJI="❌"
  STATUS_TEXT="failed"
elif [ "$STATUS" == "cancelled" ]; then
  COLOR="warning"
  EMOJI="⚠️"
  STATUS_TEXT="was cancelled"
else
  COLOR="#808080"
  EMOJI="ℹ️"
  STATUS_TEXT="$STATUS"
fi

# Build title
TITLE="${EMOJI} ${REPO_NAME} - ${WORKFLOW_NAME}"

# Build message
MESSAGE="*Status:* ${STATUS_TEXT}"
MESSAGE="${MESSAGE}\n*Branch:* \`${BRANCH}\`"
MESSAGE="${MESSAGE}\n*Triggered by:* ${ACTOR}"

# Add mention on failure
if [ "$STATUS" == "failure" ] && [ -n "$MENTION_ON_FAILURE" ]; then
  MESSAGE="${MENTION_ON_FAILURE} ${MESSAGE}"
fi

# Add custom message if provided
if [ -n "$CUSTOM_MESSAGE" ]; then
  MESSAGE="${MESSAGE}\n\n${CUSTOM_MESSAGE}"
fi

MESSAGE="${MESSAGE}\n\n<${RUN_URL}|View Workflow Run>"

# Export to GitHub outputs
echo "color=${COLOR}" >> $GITHUB_OUTPUT
echo "title=${TITLE}" >> $GITHUB_OUTPUT
echo "message<<EOF" >> $GITHUB_OUTPUT
echo -e "${MESSAGE}" >> $GITHUB_OUTPUT
echo "EOF" >> $GITHUB_OUTPUT
