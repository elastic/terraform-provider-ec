#!/bin/bash
set -euo pipefail

DOCKER_IMAGE="golang:1.21"
APP_PATH="/go/src/github.com/elastic/terraform-provider-ec"

echo "--- Run acceptance tests"
docker run \
  -u "root:root" \
  --env "EC_API_KEY=${TERRAFORM_PROVIDER_API_KEY_SECRET}" \
  -v "$PWD:${APP_PATH}" \
  -w ${APP_PATH} \
  --rm \
  $DOCKER_IMAGE \
  TEST_NAME=TestAccDeploymentExtension_basic make vendor && make testacc
