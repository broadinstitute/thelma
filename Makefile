# LOCAL_OS: local os  in GOOS format. Eg. "darwin", "linux"
LOCAL_OS := $(shell ./scripts/get-os.sh)

# LOCAL_ARCH: local arch  in GOARCH format. Eg. "amd64", "arm64"
LOCAL_ARCH := $(shell ./scripts/get-arch.sh)

# OS: set to target OS for build artifacts, in GOOS format
OS ?= ${LOCAL_OS}
ARCH ?= ${LOCAL_ARCH}

# TARGET_OS/
TARGET_OS := ${OS}
TARGET_ARCH := ${ARCH}

# VERSION: semantic version (eg. "1.2.3") to use when generating thelma build artifacts
VERSION ?= unknown

# GIT_REF: git ref to use when generating thelma build artifacts
GIT_REF ?= $(shell git rev-parse HEAD)

# CROSSPLATFORM: true if this is a cross-platform build, eg. we're building for Linux on OSX or vice versa
ifeq ($(LOCAL_OS)-$(LOCAL_ARCH),$(TARGET_OS)-$(TARGET_ARCH))
	CROSSPLATFORM = false
else
	CROSSPLATFORM = true
endif

# RUNTIME_DEPS_TESTEXEC: value TESTEXEC is set to in runtime-deps target
ifeq ($(CROSSPLATFORM),true)
	RUNTIME_DEPS_TESTEXEC=false
else
	RUNTIME_DEPS_TESTEXEC=true
endif

# OUTPUT root directory for all build output
OUTPUT=./output

# BIN location where compiled binaries are generated
BIN=${OUTPUT}/bin

# RUNTIME_DEPS_BIN location where 3rd-party runtime dependency binaries are downloaded
RUNTIME_DEPS_BIN=${OUTPUT}/runtime-deps/${TARGET_OS}/${TARGET_ARCH}/bin

# DIST_DIR directory where tarball distribution is staged
DIST_DIR=${OUTPUT}/dist

# DIST_ARCHIVE name of generated dist archive
DIST_ARCHIVE=${OUTPUT}/thelma_${VERSION}_${TARGET_OS}_${TARGET_ARCH}.tar.gz

# COVERAGE directory where coverage reports are generated
COVERAGE=${OUTPUT}/coverage

# echo-vars: Echo makefile variables for debugging purposes
echo-vars:
	@echo VERSION: ${VERSION}
	@echo GIT_REF: ${GIT_REF}
	@echo TARGET_OS: ${TARGET_OS}
	@echo TARGET_ARCH: ${TARGET_ARCH}
	@echo LOCAL_OS: ${LOCAL_OS}
	@echo LOCAL_ARCH: ${LOCAL_ARCH}
	@echo CROSSPLATFORM: ${CROSSPLATFORM}
	@echo
	@echo OUTPUT: ${OUTPUT}
	@echo BIN: ${BIN}
	@echo RUNTIME_DEPS_BIN: ${RUNTIME_DEPS_BIN}
	@echo DIST_DIR: ${DIST_DIR}
	@echo DIST_ARCHIVE: ${DIST_ARCHIVE}

# init: Initialization steps for build & other targets
init: echo-vars
	mkdir -p ${OUTPUT}

# runtime-deps: Download runtime binary dependencies, such as helm, helmfile, and so on, to output/bin
runtime-deps: init
	env OS=${TARGET_OS} ARCH=${TARGET_ARCH} SCRATCH_DIR=./output/downloads TESTEXEC=${RUNTIME_DEPS_TESTEXEC} ./scripts/install-runtime-deps.sh ${RUNTIME_DEPS_BIN}

# build: Compile thelma into output/bin
build: init
	CGO_ENABLED=0 GO111MODULE=on GOBIN=./output/bin/ GOOS=${TARGET_OS} GOARCH=${TARGET_ARCH} go build -o ./output/bin/ ./...

# dist: Package thelma binary + runtime dependencies into a tarball distribution
dist: runtime-deps build
	mkdir -p ${DIST_DIR}
	cp -R ${RUNTIME_DEPS_BIN}/ ${DIST_DIR}/bin
	cp -R ${BIN}/ ${DIST_DIR}/bin
	VERSION=${VERSION} GIT_REF=${GIT_REF} OS=${TARGET_OS} ARCH=${TARGET_ARCH} ./scripts/write-build-manifest.sh ${DIST_DIR}
	tar -C ${DIST_DIR} -czf ${DIST_ARCHIVE} .

# test: Run unit tests
test: init
	go test -covermode=atomic -race -coverprofile=${COVERAGE} ./...

# smoke: Run unit and smoke tests
smoke: runtime-deps
	PATH=${PATH}:${RUNTIME_DEPS_BIN} go test -tags smoke -covermode=atomic -race -coverprofile=${COVERAGE} ./...

# lint: Run golangci linter
lint:
	golangci-lint run ./...

# fmt: Fmt go source code
fmt:
	go fmt ./...

# cover: Open coverage report from test run in browser
cover:
	go tool cover -html=${COVERAGE}

# clean: Clean up all generated files
clean:
	rm -rf ${OUTPUT}
