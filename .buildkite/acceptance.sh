#!/bin/bash
set -euo pipefail

echo "--- Download dependencies"
make vendor

echo "--- Run acceptance tests"
EC_API_KEY=$TERRAFORM_PROVIDER_API_KEY_SECRET make testacc
