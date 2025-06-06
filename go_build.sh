#!/bin/bash
# Build the Go binary with version, commit, and build time
# Usage: ./go_build.sh <target OS: linux/windows> <input folder> <output filename> [version]

if [ $# -lt 3 ]; then
    echo "Usage: $0 <target OS: linux/windows> <input folder> <output filename> [version]"
    exit 1
fi

TARGET_OS=$1
INPUT_FOLDER=$2
FILE_NAME=$3
VERSION=${4:-"dev"}  # Default version is "dev" if not specified

# Ensure the input folder exists
if [ ! -d "$INPUT_FOLDER" ]; then
    echo "Error: Input folder '$INPUT_FOLDER' does not exist."
    exit 1
fi

cd "$INPUT_FOLDER" || exit 1

# Get git commit hash if available, else "manual"
if git rev-parse --is-inside-work-tree > /dev/null 2>&1; then
    COMMIT=$(git rev-parse --short HEAD)
else
    COMMIT="manual"
fi

# Get current UTC time in ISO 8601 format
BUILD_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

LDFLAGS="-X main.Version=$VERSION -X main.Commit=$COMMIT -X main.BuildTime=$BUILD_TIME"

# Set build parameters based on target OS
case $TARGET_OS in
    "linux")
        echo "Building for Linux..."
        GOARCH=amd64 GOOS=linux CGO_ENABLED=0 go build -ldflags "$LDFLAGS" -o "./$FILE_NAME" .
        ;;
    "windows")
        echo "Building for Windows..."
        GOARCH=amd64 GOOS=windows CGO_ENABLED=0 go build -ldflags "$LDFLAGS" -o "./$FILE_NAME.exe" .
        ;;
    *)
        echo "Invalid target OS. Please specify 'linux' or 'windows'"
        exit 1
        ;;
esac

echo "Build complete!"
echo "  Version:   $VERSION"
echo "  Commit:    $COMMIT"
echo "  BuildTime: $BUILD_TIME"
echo "Output saved in '$INPUT_FOLDER'"
