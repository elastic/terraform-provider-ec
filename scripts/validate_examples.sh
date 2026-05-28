#!/usr/bin/env bash

set -ex

# Configure Terraform to use the local filesystem mirror for elastic/ec
# so that validation works even when the version isn't published yet.
CLI_CONFIG=$(mktemp)
cat > "${CLI_CONFIG}" <<EOF
provider_installation {
  filesystem_mirror {
    path = "${HOME}/.terraform.d/plugins"
  }
  direct {
    exclude = ["registry.terraform.io/elastic/ec"]
  }
}
EOF
export TF_CLI_CONFIG_FILE="${CLI_CONFIG}"

EXAMPLES=$(find examples -maxdepth 1 -type d | grep "/")
BASEPATH=$(pwd)

for example in ${EXAMPLES}; do
    cd ${BASEPATH}/${example}
    terraform init
    terraform validate
done

rm -f "${CLI_CONFIG}"
