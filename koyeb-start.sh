#!/bin/sh
set -e

# Authenticate to Docker Hub before the in-container compose build to avoid
# anonymous pull rate limits (HTTP 429) from the shared egress IP.
if [ -n "$DOCKERHUB_USERNAME" ] && [ -n "$DOCKERHUB_TOKEN" ]; then
  echo "$DOCKERHUB_TOKEN" | docker login -u "$DOCKERHUB_USERNAME" --password-stdin
fi

exec docker compose -f docker-compose.koyeb.yml up
