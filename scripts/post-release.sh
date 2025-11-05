#!/bin/bash
set -e

VERSION=$1
PROJECT_NAME="packer-plugin-ansible-navigator"

cd dist

# Create a properly formatted x5 checksum file
# Remove .zip extensions from the checksum entries
sed 's/\.zip$//' ${PROJECT_NAME}_v${VERSION}_SHA256SUMS > ${PROJECT_NAME}_v${VERSION}_x5_SHA256SUMS

# Upload to GitHub release
gh release upload v${VERSION} ${PROJECT_NAME}_v${VERSION}_x5_SHA256SUMS --clobber

echo "✅ Uploaded x5 checksum file (with .zip extensions removed)"