#!/usr/bin/env bash

#
# This script takes in a single parameter with the current provider version
# and updates all the files that contain a declaration of the 'ec' provider
# to use the current version.
#

set -e

echo "-> Updating the version field on references to the previous 'ec' provider declaration"

declare -a UPDATE_FILES=("README.md")
for f in $(grep -R 'ec = {' examples | cut -d : -f1); do
    UPDATE_FILES+=("$f")
done

for f in "${UPDATE_FILES[@]}"; do
    # Instead of using the -i flag which its implementation differs on macOS and
    # Linux, save the output in the temporary folder and move the result over to
    # the original file for better OS cross-compatibility.
    FILE_NAME="$(basename $f)"
    sed "s/ version = \".*\"/ version = \"${1}\"/" $f > /tmp/$FILE_NAME
    mv /tmp/$FILE_NAME $f
done

echo "-> Updated all the ec provider declarations"
