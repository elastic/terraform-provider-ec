#!/usr/bin/env bash

set -ex

EXAMPLES=$(find examples -maxdepth 1 -type d | grep "/")
BASEPATH=$(pwd)

for example in ${EXAMPLES}; do
    cd ${BASEPATH}/${example}
    terraform init
    terraform validate
done
