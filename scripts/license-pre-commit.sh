#!/bin/bash
set -eo pipefail

# Install addlicense if not found
if ! command -v addlicense &> /dev/null; then
    echo "Installing addlicense..."
    go install github.com/google/addlicense@latest
    export PATH=$HOME/go/bin:$PATH
fi

echo "Checking license headers..."

# Find all Go files that are staged for commit
FILES=$(git diff --cached --name-only --diff-filter=ACM | grep '\.go$' || true)

if [[ -z "$FILES" ]]; then
    echo "No Go files to check."
    exit 0
fi

REPO_PATH=$(git rev-parse --show-toplevel)
echo "Checking files in $REPO_PATH"
if addlicense -check -l mit -f $REPO_PATH/LICENSE -c "FLYR, Inc" $(find $REPO_PATH -type f -name "*.go"); then
    echo "License headers are present."
    exit 0
fi

echo "License headers are missing. Please run 'make add-license' to apply."
exit 1
