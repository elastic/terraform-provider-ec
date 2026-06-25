#!/bin/bash
set -euo pipefail

echo "--- Importing GPG key"
echo -n "$GPG_PRIVATE_SECRET" | base64 --decode | gpg --import --batch --yes --passphrase "$GPG_PASSPHRASE_SECRET"

echo "--- Caching GPG passphrase"
echo "$GPG_PASSPHRASE_SECRET" | gpg --armor --detach-sign --passphrase-fd 0 --pinentry-mode loopback

echo "--- Installing jq"
apt-get update
apt-get install jq -y

echo "--- Release the binaries"
make release
