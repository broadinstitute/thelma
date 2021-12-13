#!/bin/bash

set -eo pipefail

# Convert local OS name into a value supported by GOOS

output="$( uname -s )"
case "${output}" in
    Linux*)
      echo linux
      ;;
    Darwin*)
      echo darwin
      ;;
    *)
      echo "Unrecognized OS: ${output}" >& 2
      exit 1
esac
