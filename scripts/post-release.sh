#!/bin/bash
set -e

echo "Starting post-release processing..."

# Get the version from environment or tags
VERSION="${GORELEASER_CURRENT_TAG#v}"
if [ -z "$VERSION" ]; then
    VERSION=$(git describe --tags --abbrev=0 | sed 's/^v//')
fi

echo "Processing version: $VERSION"

# GitHub repository details
OWNER="SolomonHD"
REPO="packer-plugin-ansible-navigator"
PLUGIN_NAME="packer-plugin-ansible-navigator"

# Create temporary working directory
WORK_DIR=$(mktemp -d)
cd "$WORK_DIR"

echo "Working in: $WORK_DIR"

# Download all ZIP files from the release
echo "Downloading release assets..."
gh release download "v${VERSION}" --repo "${OWNER}/${REPO}" --pattern "*.zip" || {
    echo "Failed to download release assets"
    exit 1
}

# Process each zip file
for zipfile in *.zip; do
    if [[ "$zipfile" == *"SHA256SUMS"* ]]; then
        continue
    fi
    
    echo "Processing $zipfile..."
    
    # Extract OS and architecture from filename
    # Format: packer-plugin-ansible-navigator_v0.4.17_x5_darwin_amd64.zip
    if [[ "$zipfile" =~ ${PLUGIN_NAME}_v${VERSION}_x[0-9]+\.[0-9]+_([^_]+)_([^.]+)\.zip ]]; then
        OS="${BASH_REMATCH[1]}"
        ARCH="${BASH_REMATCH[2]}"
        
        # Create temp directory for this archive
        TEMP_DIR=$(mktemp -d)
        
        # Extract the current archive
        unzip -q "$zipfile" -d "$TEMP_DIR"
        
        # Find the binary (it should be the only file)
        BINARY=$(find "$TEMP_DIR" -type f -maxdepth 1 | head -n 1)
        
        if [ -n "$BINARY" ]; then
            # Get the proper binary name
            BINARY_NAME=$(basename "$BINARY")
            
            # Remove the old zip
            rm "$zipfile"
            
            # Create new zip with just the renamed binary
            # The binary inside should be named: packer-plugin-ansible-navigator_vX.X.X_x5.0_os_arch[.exe]
            cd "$TEMP_DIR"
            zip -q "../$zipfile" "$BINARY_NAME"
            cd "$WORK_DIR"
            
            echo "  Repackaged: $zipfile"
        else
            echo "  WARNING: No binary found in $zipfile"
        fi
        
        # Clean up temp directory
        rm -rf "$TEMP_DIR"
    else
        echo "  WARNING: Could not parse filename: $zipfile"
    fi
done

# Download the SHA256SUMS file
echo "Downloading SHA256SUMS file..."
gh release download "v${VERSION}" --repo "${OWNER}/${REPO}" --pattern "*SHA256SUMS" || {
    echo "No SHA256SUMS file found"
}

# Create the _SHA256SUMS version (with underscore)
if [ -f "${PLUGIN_NAME}_v${VERSION}_SHA256SUMS" ]; then
    echo "Creating _SHA256SUMS file..."
    cp "${PLUGIN_NAME}_v${VERSION}_SHA256SUMS" "${PLUGIN_NAME}_v${VERSION}_x5.0_SHA256SUMS"
    
    # Upload the new _SHA256SUMS file
    echo "Uploading _SHA256SUMS file..."
    gh release upload "v${VERSION}" \
        "${PLUGIN_NAME}_v${VERSION}_x5.0_SHA256SUMS" \
        --repo "${OWNER}/${REPO}" \
        --clobber || {
        echo "Failed to upload _SHA256SUMS file"
    }
fi

# Upload all modified zip files back to the release
echo "Uploading modified archives..."
for zipfile in *.zip; do
    if [[ "$zipfile" == *"SHA256SUMS"* ]]; then
        continue
    fi
    
    echo "  Uploading: $zipfile"
    gh release upload "v${VERSION}" \
        "$zipfile" \
        --repo "${OWNER}/${REPO}" \
        --clobber || {
        echo "Failed to upload $zipfile"
    }
done

# Clean up
cd /
rm -rf "$WORK_DIR"

echo "Post-release processing complete!"