#!/bin/bash

set -eo pipefail

if [[ $# -ne 1 ]]; then
  echo "Usage: $0 ./path/to/release/dir" >&2
  exit 1
fi

VERSION=${VERSION:-unknown}
RELEASE_DIR=$1
OUTFILE="thelma_${VERSION}_SHA256SUMS"

mkdir -p $RELEASE_DIR
cd $RELEASE_DIR
sha256sum *.tar.gz > ${OUTFILE}
