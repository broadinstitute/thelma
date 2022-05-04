#!/bin/bash

set -eo pipefail

# Convert local OS arch into a value supported by GOARCH

output="$( uname -m )"
case "${output}" in
    x86_64*)
      echo amd64
      ;;
    aarch64*)
      echo arm64
      ;;
    arm64*)
      echo arm64
      ;;
    *)
      echo "Unrecognized architecture: ${output}" >& 2
      exit 1
esac
