#!/bin/bash

set -eo pipefail

if [[ $# -ne 1 ]]; then
  echo "Usage: $0 path/to/file" >&2
  exit 1
fi

# It would be nice if we could run `thelma version --output-format=json` to do this,
# but unfortunately that doesn't work for cross-platform buils

OUTPUT_FILE="$1"

VERSION=${VERSION:-unset}
GIT_SHA=${GIT_SHA:-unset}
BUILD_TIMESTAMP=${BUILD_TIMESTAMP:-unset}
OS=${OS:-unset}
ARCH=${ARCH:-unset}

cat <<MANIFEST > "${OUTPUT_FILE}"
{
  "version": "${VERSION}",
  "gitSha": "${GIT_SHA}",
  "os": "${OS}",
  "arch": "${ARCH}",
  "buildTimestamp": "${BUILD_TIMESTAMP}"
}
MANIFEST 
