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

# OUTPUT_DIR root directory for all build output
OUTPUT_DIR=./output

# BIN_DIR location where compiled binaries are generated
BIN_DIR=${OUTPUT_DIR}/bin

# RUNTIME_DEPS_BIN_DIR location where 3rd-party runtime dependency binaries are downloaded
RUNTIME_DEPS_BIN_DIR=${OUTPUT_DIR}/runtime-deps/${TARGET_OS}/${TARGET_ARCH}/bin

# RELEASE_STAGING_DIR directory where release archive is staged before assembly
RELEASE_STAGING_DIR=${OUTPUT_DIR}/release

# RELEASE_ARCHIVE_NAME name of generated release archive
RELEASE_ARCHIVE_NAME=thelma_${VERSION}_${TARGET_OS}_${TARGET_ARCH}.tar.gz

# RELEASE_ARCHIVE_DIR directory where release archives are generated
RELEASE_ARCHIVE_DIR=${OUTPUT_DIR}/releases

# COVERAGE_DIR directory where coverage reports are generated
COVERAGE_DIR=${OUTPUT_DIR}/coverage

# echo-vars: Echo makefile variables for debugging purposes
echo-vars:
	@echo TARGET_OS: ${TARGET_OS}
	@echo TARGET_ARCH: ${TARGET_ARCH}
	@echo LOCAL_OS: ${LOCAL_OS}
	@echo LOCAL_ARCH: ${LOCAL_ARCH}
	@echo VERSION: ${VERSION}
	@echo GIT_SHA: ${GIT_SHA}
	@echo BUILD_TIMESTAMP: ${BUILD_TIMESTAMP}
	@echo VERSION_IMPORT_PATH: ${VERSION_IMPORT_PATH}
	@echo CROSSPLATFORM: ${CROSSPLATFORM}
	@echo
	@echo OUTPUT_DIR: ${OUTPUT_DIR}
	@echo BIN_DIR: ${BIN_DIR}
	@echo RUNTIME_DEPS_BIN_DIR: ${RUNTIME_DEPS_BIN_DIR}
	@echo RELEASE_STAGING_DIR: ${RELEASE_STAGING_DIR}
	@echo RELEASE_ARCHIVE_NAME: ${RELEASE_ARCHIVE_NAME}
	@echo RELEASE_ARCHIVE_DIR: ${RELEASE_ARCHIVE_DIR}
	@echo COVERAGE_DIR: ${COVERAGE_DIR}

# init: Initialization steps for build & other targets
init: echo-vars
	mkdir -p ${OUTPUT_DIR}

# runtime-deps: Download runtime binary dependencies, such as helm, helmfile, and so on, to output directory
runtime-deps: init
	env OS=${TARGET_OS} ARCH=${TARGET_ARCH} SCRATCH_DIR=${OUTPUT_DIR}/downloads TESTEXEC=${RUNTIME_DEPS_TESTEXEC} ./scripts/install-runtime-deps.sh ${RUNTIME_DEPS_BIN_DIR}

# build: Compile thelma into output directory
build: init
	CGO_ENABLED=0 \
	GO111MODULE=on \
	GOBIN=${BIN_DIR} \
	GOOS=${TARGET_OS} \
	GOARCH=${TARGET_ARCH} \
	go build \
	-ldflags="-X '${VERSION_IMPORT_PATH}.Version=${VERSION}' -X '${VERSION_IMPORT_PATH}.GitSha=${GIT_SHA}' -X '${VERSION_IMPORT_PATH}.BuildTimestamp=${BUILD_TIMESTAMP}'" \
	-o ${BIN_DIR}/ ./...

# release: Assemble thelma binary + runtime dependencies into a tarball distribution
release: runtime-deps build
	# Always clean release staging dir before assembly, just so we don't end up with pollution
	rm -rf ${RELEASE_STAGING_DIR}
	mkdir -p ${RELEASE_STAGING_DIR}
	mkdir -p ${RELEASE_ARCHIVE_DIR}

	cp -r ${RUNTIME_DEPS_BIN_DIR}/. ${RELEASE_STAGING_DIR}/bin
	cp -r ${BIN_DIR}/. ${RELEASE_STAGING_DIR}/bin
	VERSION=${VERSION} GIT_SHA=${GIT_SHA} BUILD_TIMESTAMP=${BUILD_TIMESTAMP} OS=${TARGET_OS} ARCH=${TARGET_ARCH} ./scripts/write-build-manifest.sh ${RELEASE_STAGING_DIR}/build.json
	tar -C ${RELEASE_STAGING_DIR} -czf ${RELEASE_ARCHIVE_DIR}/${RELEASE_ARCHIVE_NAME} .

# checksum: Generate sha256sum file for tarball archives in the release archive directory
checksum:
	env VERSION=${VERSION} ./scripts/checksum.sh ${RELEASE_ARCHIVE_DIR}

# test: Run unit tests
test: init
	go test -covermode=atomic -race -coverprofile=${COVERAGE_DIR} ./...

# smoke: Run unit and smoke tests
smoke: runtime-deps
	PATH=${PATH}:${RUNTIME_DEPS_BIN_DIR} go test -tags smoke -covermode=atomic -race -coverprofile=${COVERAGE_DIR} ./...

# lint: Run golangci linter
lint:
	golangci-lint run ./...

# fmt: Fmt go source code
fmt:
	go fmt ./...

# coverage: Open coverage report from test run in browser
coverage:
	go tool cover -html=${COVERAGE_DIR}

# clean: Clean up all generated files
clean:
	rm -rf ${OUTPUT_DIR}
