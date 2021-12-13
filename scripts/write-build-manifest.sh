#!/bin/bash

set -eo pipefail

if [[ $# -ne 1 ]]; then
  echo "Usage: $0 path/to/dist" >&2
  exit 1
fi

DIST_DIR="$1"
TIMESTAMP=$( date +%Y-%m-%dT%H:%M:%S%z )

cat <<MANIFEST > "${DIST_DIR}/build.json"
{
  "version": "${VERSION}",
  "gitRef": "${GIT_REF}",
  "os": "${OS}",
  "arch": "${ARCH}",
  "timestamp": "${TIMESTAMP}"
}
MANIFEST