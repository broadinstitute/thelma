#!/bin/bash

set -eo pipefail

# Download and install runtime dependencies such as Helm, Helmfile, etc
# This script implements basic caching -- if the binaries already exist in
# the target directory, it won't download them again.

HELM_VERSION=3.2.4
HELMFILE_VERSION=0.114.0
YQ_VERSION=4.11.2
HELM_DOCS_VERSION=1.5.0

SCRIPT_DIR="$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"

# OS: target OS where these binaries should run. One of "linux" or "darwin".
OS="${OS:-$( ${SCRIPT_DIR}/get-os.sh )}"

# ARCH: target architecture where these binaries should run. Currently only "amd64" is supported.
ARCH="${ARCH:-$( ${SCRIPT_DIR}/get-arch.sh )}"

# SCRATCH_DIR: directory where temporary files should be downloaded
SCRATCH_DIR="${SCRATCH_DIR:-/tmp/thelma-downloads-$$}"

# TESTEXEC: set to false to skip test execution of downloaded binaries (eg. "helm --version").
# Useful when building a dist for a different target OS/ARCH than where this script is running
TESTEXEC="${TESTEXEC:-true}"

# CLEANUP: set to false to skip cleanup of download directory. Useful for debugging installation issues.
CLEANUP="${CLEANUP:-true}"


# WGET_TIMEOUT_SECONDS: set to false to skip cleanup of download directory. Useful for debugging installation issues.
WGET_TIMEOUT_SECONDS="${WGET_TIMEOUT_SECONDS:-15}"


if [[ $# -ne 1 ]]; then
  echo "Usage: $0 install/path" >&2
  exit 1
fi

INSTALL_DIR="$1"

cat <<EOF
-----------------------------------------------------------
Installing runtime deps for ${OS} ${ARCH} to ${INSTALL_DIR}

Scratch directory: ${SCRATCH_DIR}
Test executables after download? ${TESTEXEC}
-----------------------------------------------------------
EOF

cleanup() {
  if [[ "${CLEANUP}" != "true" ]]; then
    return 0
  fi
  echo "Cleaning up scratch directory: ${SCRATCH_DIR}"
  rm -rf "${SCRATCH_DIR}"
}
trap "cleanup" EXIT

testexec() {
  if [[ "${TESTEXEC}" != "true" ]]; then
    return 0
  fi
  echo "Testing binary with: $@"
  "$@"
}

install_helm() {
  URL="https://get.helm.sh/helm-v${HELM_VERSION}-${OS}-${ARCH}.tar.gz"
  echo "Downloading helm from ${URL}"
  wget --timeout="${WGET_TIMEOUT_SECONDS}" -q "${URL}" -O - | \
    tar -xz --strip-components=1 "${OS}-${ARCH}/helm" && \
    chmod +x helm && \
    testexec ./helm version && \
    mv helm "${INSTALL_DIR}/helm"
}

install_helmfile() {
  URL="https://github.com/roboll/helmfile/releases/download/v${HELMFILE_VERSION}/helmfile_${OS}_${ARCH}"
  echo "Downloading helmfile from ${URL}"
  wget --timeout="${WGET_TIMEOUT_SECONDS}" -q "${URL}" -O helmfile && \
    chmod +x helmfile && \
    testexec ./helmfile --version && \
    mv ./helmfile "${INSTALL_DIR}/helmfile"
}

install_yq() {
  URL="https://github.com/mikefarah/yq/releases/download/v${YQ_VERSION}/yq_${OS}_${ARCH}.tar.gz"
  echo "Downloading yq from ${URL}"
  wget --timeout="${WGET_TIMEOUT_SECONDS}" -q "${URL}" -O - |\
    tar xz && \
    mv yq_${OS}_${ARCH} yq && \
    testexec ./yq --version && \
    mv ./yq "${INSTALL_DIR}/yq"
}

install_helm_docs() {
  if [[ "${OS}" == "linux" ]]; then
    # linux artifacts are only made available as rpm or deb packages and not as tarballs :'(
    URL="https://github.com/norwoodj/helm-docs/releases/download/v${HELM_DOCS_VERSION}/helm-docs_${HELM_DOCS_VERSION}_${OS}_${ARCH}.deb"
    echo "Downloading helm-docs from ${URL}"
    wget --timeout="${WGET_TIMEOUT_SECONDS}" -q "${URL}" -O helm-docs.deb && \
      ar x helm-docs.deb && \
      tar -xzvf data.tar.gz && \
      testexec ./usr/local/bin/helm-docs --version && \
      mv ./usr/local/bin/helm-docs "${INSTALL_DIR}/helm-docs"
    return $?
  fi

  if [[ "${OS}" == "darwin" ]]; then
    os="Darwin"
    arch="x86_64"
    URL="https://github.com/norwoodj/helm-docs/releases/download/v${HELM_DOCS_VERSION}/helm-docs_${HELM_DOCS_VERSION}_${os}_${arch}.tar.gz"
    echo "Downloading helm-docs from ${URL}"
    wget --timeout="${WGET_TIMEOUT_SECONDS}" -q "${URL}" -O - |\
      tar -xz && \
      testexec ./helm-docs --version && \
      mv ./helm-docs "${INSTALL_DIR}/helm-docs"
    return $?
  fi

  echo "Unsupported OS / ARCH combo ${OS} ${ARCH}, don't know how to install helm-docs" >&2
  return 1
}

mkdir -p "${INSTALL_DIR}"
mkdir -p "${SCRATCH_DIR}"

# Get fully expanded paths to directories so we can change directories
INSTALL_DIR="$( cd "${INSTALL_DIR}" && pwd )"
SCRATCH_DIR="$( cd "${SCRATCH_DIR}" && pwd )"

cd "${SCRATCH_DIR}"

# Install Helm
if [[ ! -f "${INSTALL_DIR}/helm" ]]; then
  if ! install_helm; then
    echo "helm install failed!" >&2
    exit 1
  fi
fi

# Install Helmfile
if [[ ! -f "${INSTALL_DIR}/helmfile" ]]; then
  if ! install_helmfile; then
    echo "helmfile install failed!" >&2
    exit 1
  fi
fi

# Install yq
if [[ ! -f "${INSTALL_DIR}/yq" ]]; then
  if ! install_yq; then
    echo "yq install failed!" >&2
    exit 1
  fi
fi

# Install helm_docs
if [[ ! -f "${INSTALL_DIR}/helm-docs" ]]; then
  if ! install_helm_docs; then
    echo "helm-docs install failed!" >&2
    exit 1
  fi
fi
