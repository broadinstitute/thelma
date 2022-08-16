#!/bin/bash

set -eo pipefail

if [[ $# -ne 1 ]]; then
  echo "Usage: $0 ./path/to/release/dir" >&2
  exit 1
fi

VERSION=${VERSION:-development}
RELEASE_DIR=$1
OUTFILE="thelma_${VERSION}_SHA256SUMS"

SCRIPT_DIR="$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
OS="$( ${SCRIPT_DIR}/get-os.sh )"

mkdir -p $RELEASE_DIR
cd $RELEASE_DIR

if [[ "${OS}" == "darwin" ]]; then
  # Use native OSX shasum utility
  shasum -a 256 *.tar.gz > ${OUTFILE}
else
  # Use Linux sha256sum utility
  sha256sum *.tar.gz > ${OUTFILE}
fi
