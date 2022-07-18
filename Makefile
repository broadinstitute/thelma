# LOCAL_OS: local os  in GOOS format. Eg. "darwin", "linux"
LOCAL_OS=$(shell ./scripts/get-os.sh)

# LOCAL_ARCH: local arch  in GOARCH format. Eg. "amd64", "arm64"
LOCAL_ARCH=$(shell ./scripts/get-arch.sh)

# OS, ARCH: os & arch to _target_ for build artifacts (in GOOS/GOARCH format)
OS=${LOCAL_OS}
ARCH=${LOCAL_ARCH}

# TARGET_OS, TARGET_ARCH: Aliases for the OS and ARCH parameters that are used within this Makefile for the sake of clarity
TARGET_OS=${OS}
TARGET_ARCH=${ARCH}

# VERSION: semantic version (eg. "1.2.3") to use when versioning thelma build artifacts
VERSION=development

# GIT_SHA: git sha to use when versioning thelma build artifacts
GIT_SHA=$(shell git rev-parse HEAD)

# BUILD_TIMESTAMP: timestamp to use when versioning thelma build artifacts
BUILD_TIMESTAMP=$(shell date "+%Y-%m-%dT%H:%M:%S%z")

# VERSION_IMPORT_PATH: path where version ldflags should be set
VERSION_IMPORT_PATH=github.com/broadinstitute/thelma/internal/thelma/app/version

# CROSSPLATFORM: true if this is a cross-platform build, i.e. we're building for Linux on OSX or vice versa
ifeq ($(LOCAL_OS)-$(LOCAL_ARCH),$(TARGET_OS)-$(TARGET_ARCH))
	CROSSPLATFORM=false
else
	CROSSPLATFORM=true
endif

# RUNTIME_DEPS_TESTEXEC: value TESTEXEC is set to in runtime-deps target
# if we're compiling Linux executables on OSX, we can't execute them for smoke testing
ifeq ($(CROSSPLATFORM),true)
	RUNTIME_DEPS_TESTEXEC=false
else
	RUNTIME_DEPS_TESTEXEC=true
endif

# MACOS_SIGN_AND_NOTARIZE: When false, creates a release tarball as part of the make process,
#                          when true, create it using a separate GHA job using the sign
#                          and notarize script.
#                          When a macOS host and target are used, default to sign and
#                          notarize releases. This value can be manually set to false in
#                          cases where you want to generate a macOS release tarball from
#                          a macOS host without performing any signing and notarizing,
#                          i.e. when locally testing updates to the release process without
#                          having the certs and other creds on your local machine.
ifeq ($(LOCAL_OS)-$(TARGET_OS),darwin-darwin)
	MACOS_SIGN_AND_NOTARIZE=true
else
	MACOS_SIGN_AND_NOTARIZE=false
endif

# OUTPUT_DIR root directory for all build output
OUTPUT_DIR=./output

# BIN_DIR location where compiled binaries are generated
BIN_DIR=${OUTPUT_DIR}/bin

# RUNTIME_DEPS_BIN_DIR location where 3rd-party runtime dependency binaries are downloaded
RUNTIME_DEPS_BIN_DIR=${OUTPUT_DIR}/runtime-deps/${TARGET_OS}/${TARGET_ARCH}/bin

# RELEASE_STAGING_DIR directory where release archive is staged before assembly
RELEASE_STAGING_DIR=${OUTPUT_DIR}/release-assembly

# RELEASE_ARCHIVE_NAME name of generated release archive
RELEASE_ARCHIVE_NAME=thelma_${VERSION}_${TARGET_OS}_${TARGET_ARCH}.tar.gz

# RELEASE_ARCHIVE_DIR directory where release archives are generated
RELEASE_ARCHIVE_DIR=${OUTPUT_DIR}/releases

# COVERAGE_DIR directory where coverage reports are generated
COVERAGE_DIR=${OUTPUT_DIR}/coverage

# Add runtime dependencies to PATH
export PATH := $(shell pwd)/${RUNTIME_DEPS_BIN_DIR}:$(PATH)

# Self-documenting help target copied from https://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
# NOTE:
#  High-level targets that devs are expected to run are documented with two pound signs (##), which makes them appear in help output.
#  Low-level helper targets are documented with a single pound sign to reduce clutter in the help output.
.PHONY: help
help: # Prints list of targets in this Makefile to terminal
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

echo-vars: # Echo makefile variables for debugging purposes
	@echo "-----------------------------------------------------------"
	@echo "Build Parameters"
	@echo "-----------------------------------------------------------"
	@echo "Target OS:         ${TARGET_OS}"
	@echo "Target Arch:       ${TARGET_ARCH}"
	@echo "Local OS:          ${LOCAL_OS}"
	@echo "Local Arch:        ${LOCAL_ARCH}"
	@echo "Cross-platform?    ${CROSSPLATFORM}"
	@echo "Sign and notarize? ${MACOS_SIGN_AND_NOTARIZE}"
	@echo
	@echo "Version:     ${VERSION}"
	@echo "Git Sha:     ${GIT_SHA}"
	@echo "Timestamp:   ${BUILD_TIMESTAMP}"
	@echo "Import Path: ${VERSION_IMPORT_PATH}"
	@echo
	@echo "Paths:"
	@echo "Compiled binaries:    ${BIN_DIR}"
	@echo "Runtime dependencies: ${RUNTIME_DEPS_BIN_DIR}"
	@echo "Release staging dir:  ${RELEASE_STAGING_DIR}"
	@echo "Release archive file: ${RELEASE_ARCHIVE_DIR}/${RELEASE_ARCHIVE_NAME}"
	@echo

init: echo-vars # Initialization steps for build & other targets
	mkdir -p ${OUTPUT_DIR}

runtime-deps: init # Download runtime binary dependencies, such as helm, helmfile, and so on, to output directory
	env OS=${TARGET_OS} ARCH=${TARGET_ARCH} SCRATCH_DIR=${OUTPUT_DIR}/downloads TESTEXEC=${RUNTIME_DEPS_TESTEXEC} ./scripts/install-runtime-deps.sh ${RUNTIME_DEPS_BIN_DIR}

build: init ## Compile thelma into output directory
	CGO_ENABLED=0 \
	GO111MODULE=on \
	GOBIN=${BIN_DIR} \
	GOOS=${TARGET_OS} \
	GOARCH=${TARGET_ARCH} \
	go build \
	-ldflags="-X '${VERSION_IMPORT_PATH}.Version=${VERSION}' -X '${VERSION_IMPORT_PATH}.GitSha=${GIT_SHA}' -X '${VERSION_IMPORT_PATH}.BuildTimestamp=${BUILD_TIMESTAMP}'" \
	-o ${BIN_DIR}/ ./...

release: runtime-deps build ## Assemble thelma binary + runtime dependencies into a tarball distribution. Set OS and ARCH to desired platform.
	# Clean staging dir
	rm -rf ${RELEASE_STAGING_DIR}
	mkdir -p ${RELEASE_STAGING_DIR}
	mkdir -p ${RELEASE_ARCHIVE_DIR}

    # Copy runtime dependencies into staging dir
	cp -r ${RUNTIME_DEPS_BIN_DIR}/. ${RELEASE_STAGING_DIR}/bin

    # Copy compiled thelma binary into staging dir
	cp -r ${BIN_DIR}/. ${RELEASE_STAGING_DIR}/bin

    # Generate build.json manifest in staging dir
	VERSION=${VERSION} GIT_SHA=${GIT_SHA} BUILD_TIMESTAMP=${BUILD_TIMESTAMP} OS=${TARGET_OS} ARCH=${TARGET_ARCH} ./scripts/write-build-manifest.sh ${RELEASE_STAGING_DIR}/build.json

	# Create a release tarball unless signing and notarizing, which take place in a separate GHA job
	if [ ${MACOS_SIGN_AND_NOTARIZE} = false ]; then \
		tar -C ${RELEASE_STAGING_DIR} -czf ${RELEASE_ARCHIVE_DIR}/${RELEASE_ARCHIVE_NAME} .; \
	fi;

checksum: # Generate sha256sum file for tarball archives in the release archive directory
	env VERSION=${VERSION} ./scripts/checksum.sh ${RELEASE_ARCHIVE_DIR}

test: init ## Run unit tests
	go test -covermode=atomic -race -coverpkg=./... -coverprofile=${COVERAGE_DIR} ./...

smoke: runtime-deps ## Run unit and smoke tests
	go test -tags smoke -covermode=atomic -race -coverpkg=./... -coverprofile=${COVERAGE_DIR} ./...

lint: ## Run golangci linter
	golangci-lint run ./...

fmt: ## Format source code
	go fmt ./...

coverage: ## Open coverage report from test run in browser. Run "make test" first!
	go tool cover -html=${COVERAGE_DIR}

clean: ## Clean up all generated files
	rm -rf ${OUTPUT_DIR}

mocks: ## Generate testify mocks with Mockery
	mockery --dir ./internal/thelma/clients/google/bucket --name Locker --output=./internal/thelma/clients/google/bucket/testing/mocks --outpkg mocks --filename locker.go
	mockery --dir ./internal/thelma/clients/google/bucket --name Bucket --output=./internal/thelma/clients/google/bucket/testing/mocks --outpkg mocks --filename bucket.go
	mockery --dir ./internal/thelma/state/api/terra --name Release --output=./internal/thelma/state/api/terra/mocks --outpkg mocks --filename release.go
