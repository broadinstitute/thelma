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

# TOOLS_BIN_DIR symlink to 3rd-party tools binaries (helm, kubectl, etc)
TOOLS_BIN_DIR=${OUTPUT_DIR}/tools/bin

# RUNTIME_DEPS_BIN_DIR location where 3rd-party tools dependency binaries are downloaded
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
export PATH := $(shell pwd)/${TOOLS_BIN_DIR}:$(PATH)

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
	@echo "Target OS:      ${TARGET_OS}"
	@echo "Target Arch:    ${TARGET_ARCH}"
	@echo "Local OS:       ${LOCAL_OS}"
	@echo "Local Arch:     ${LOCAL_ARCH}"
	@echo "Cross-platform? ${CROSSPLATFORM}"
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

EMPTY :=
SPACE := $(EMPTY) $(EMPTY)
RELATIVE_PREFIX := $(subst $(SPACE),/,$(patsubst %,..,$(subst /, ,$(subst ${OUTPUT_DIR},,${TOOLS_BIN_DIR}))))

runtime-deps: init # Download runtime binary dependencies, such as helm, helmfile, and so on, to output directory
	env OS=${TARGET_OS} ARCH=${TARGET_ARCH} SCRATCH_DIR=${OUTPUT_DIR}/downloads TESTEXEC=${RUNTIME_DEPS_TESTEXEC} ./scripts/install-runtime-deps.sh ${RUNTIME_DEPS_BIN_DIR}

EMPTY :=
SPACE := $(EMPTY) $(EMPTY)
runtime-deps-symlink: # Create symlink from ${TOOLS_BIN_DIR} to ${RUNTIME_DEPS_BIN_DIR} so compiled thelma can find its tools
	mkdir -p $(dir ${TOOLS_BIN_DIR})
	# "fun" hack to compute relative path --
	# all the gnarly pattern substitution does is add the appropriate number of
	# "../"'s to the runtime deps directory for symlinking
	ln -sfn $(subst $(SPACE),/,$(patsubst %,..,$(subst /,$(SPACE),$(subst ${OUTPUT_DIR},,${TOOLS_BIN_DIR}))))/${RUNTIME_DEPS_BIN_DIR} ${TOOLS_BIN_DIR}

build: init runtime-deps runtime-deps-symlink ## Compile thelma into output directory
	CGO_ENABLED=0 \
	GO111MODULE=on \
	GOBIN=${BIN_DIR} \
	GOOS=${TARGET_OS} \
	GOARCH=${TARGET_ARCH} \
	go build \
	-ldflags="-X '${VERSION_IMPORT_PATH}.Version=${VERSION}' -X '${VERSION_IMPORT_PATH}.GitSha=${GIT_SHA}' -X '${VERSION_IMPORT_PATH}.BuildTimestamp=${BUILD_TIMESTAMP}'" \
	-o ${BIN_DIR}/ ./cmd/thelma

release: runtime-deps build ## Assemble thelma binary + runtime dependencies into a tarball distribution. Set OS and ARCH to desired platform.
	# Clean staging dir
	rm -rf ${RELEASE_STAGING_DIR}
	mkdir -p ${RELEASE_STAGING_DIR}
	mkdir -p ${RELEASE_ARCHIVE_DIR}

    # Copy runtime dependencies into staging dir
	mkdir -p ${RELEASE_STAGING_DIR}/tools/bin
	cp -r ${RUNTIME_DEPS_BIN_DIR}/. ${RELEASE_STAGING_DIR}/tools/bin

    # Copy compiled thelma binary into staging dir
	cp -r ${BIN_DIR}/. ${RELEASE_STAGING_DIR}/bin

    # Generate build.json manifest in staging dir
	VERSION=${VERSION} GIT_SHA=${GIT_SHA} BUILD_TIMESTAMP=${BUILD_TIMESTAMP} OS=${TARGET_OS} ARCH=${TARGET_ARCH} ./scripts/write-build-manifest.sh ${RELEASE_STAGING_DIR}/build.json

    # Package all files into tar.gz archive
	tar -C ${RELEASE_STAGING_DIR} -czf ${RELEASE_ARCHIVE_DIR}/${RELEASE_ARCHIVE_NAME} .

checksum: # Generate sha256sum file for tarball archives in the release archive directory
	env VERSION=${VERSION} ./scripts/checksum.sh ${RELEASE_ARCHIVE_DIR}

test: init ## Run unit tests
	go test -covermode=atomic -race -coverprofile=${COVERAGE_DIR} ./...

smoke: runtime-deps ## Run unit and smoke tests
	go test -tags smoke -covermode=atomic -race -coverprofile=${COVERAGE_DIR} ./...

lint: ## Run golangci linter
	golangci-lint run ./...

fmt: ## Format source code
	go fmt ./...

coverage: ## Open coverage report from test run in browser. Run "make test" first!
	go tool cover -html=${COVERAGE_DIR}

clean: ## Clean up all generated files
	rm -rf ${OUTPUT_DIR}

export MOCKERY_WITH_EXPECTER=true

mocks: ## Generate testify mocks with Mockery
	mockery --dir ./internal/thelma/app/autoupdate/bootstrap --name Bootstrapper --output=./internal/thelma/app/autoupdate/bootstrap/mocks --outpkg mocks --filename bootstrapper.go
	mockery --dir ./internal/thelma/app/autoupdate/installer --name Installer --output=./internal/thelma/app/autoupdate/installer/mocks --outpkg mocks --filename installer.go
	mockery --dir ./internal/thelma/app/autoupdate/releasebucket --name ReleaseBucket --output=./internal/thelma/app/autoupdate/releasebucket/mocks --outpkg mocks --filename release_bucket.go
	mockery --dir ./internal/thelma/app/autoupdate/releases --name Dir --output=./internal/thelma/app/autoupdate/releases/mocks --outpkg mocks --filename dir.go
	mockery --dir ./internal/thelma/app/autoupdate/spawn --name Spawn --output=./internal/thelma/app/autoupdate/spawn/mocks --outpkg mocks --filename spawn.go
	mockery --dir ./internal/thelma/app/scratch --name Scratch --output=./internal/thelma/app/scratch/mocks --outpkg mocks --filename scratch.go
	mockery --dir ./internal/thelma/cli --name RunContext --output=./internal/thelma/cli/mocks --outpkg mocks --filename run_context.go
	mockery --dir ./internal/thelma/clients/google --name Clients --output=./internal/thelma/clients/google/mocks --outpkg mocks --filename clients.go
	mockery --dir ./internal/thelma/clients/google/bucket --name Bucket --output=./internal/thelma/clients/google/bucket/testing/mocks --outpkg mocks --filename bucket.go
	mockery --dir ./internal/thelma/clients/google/bucket --name Locker --output=./internal/thelma/clients/google/bucket/testing/mocks --outpkg mocks --filename locker.go
	mockery --dir ./internal/thelma/clients/google/sqladmin --name Client --output=./internal/thelma/clients/google/sqladmin/mocks --outpkg mocks --filename sqladmin.go
	mockery --dir ./internal/thelma/clients/google/testing/aliases --name ClusterManagerServer --output=./internal/thelma/clients/google/testing/mocks --outpkg mocks --filename cluster_manager_server.go
	mockery --dir ./internal/thelma/clients/google/testing/aliases --name PublisherServer --output=./internal/thelma/clients/google/testing/mocks --outpkg mocks --filename publisher_server.go
	mockery --dir ./internal/thelma/clients/google/testing/aliases --name SubscriberServer --output=./internal/thelma/clients/google/testing/mocks --outpkg mocks --filename subscriber_server.go
	mockery --dir ./internal/thelma/clients/kubernetes --name Clients --output=./internal/thelma/clients/kubernetes/mocks --outpkg mocks --filename clients.go
	mockery --dir ./internal/thelma/clients/kubernetes/kubecfg --name Kubeconfig --output=./internal/thelma/clients/kubernetes/kubecfg/mocks --outpkg mocks --filename kubecfg.go
	mockery --dir ./internal/thelma/clients/kubernetes/kubecfg --name Kubectx --output=./internal/thelma/clients/kubernetes/kubecfg/mocks --outpkg mocks --filename kubectx.go
	mockery --dir ./internal/thelma/clients/kubernetes/testing/aliases --name AppsV1 --output=./internal/thelma/clients/kubernetes/testing/mocks --outpkg mocks --filename appsv1.go
	mockery --dir ./internal/thelma/clients/kubernetes/testing/aliases --name ConfigMaps --output=./internal/thelma/clients/kubernetes/testing/mocks --outpkg mocks --filename configmaps.go
	mockery --dir ./internal/thelma/clients/kubernetes/testing/aliases --name CoreV1 --output=./internal/thelma/clients/kubernetes/testing/mocks --outpkg mocks --filename corev1.go
	mockery --dir ./internal/thelma/clients/kubernetes/testing/aliases --name Deployments --output=./internal/thelma/clients/kubernetes/testing/mocks --outpkg mocks --filename deployments.go
	mockery --dir ./internal/thelma/clients/kubernetes/testing/aliases --name KubeClient --output=./internal/thelma/clients/kubernetes/testing/mocks --outpkg mocks --filename kube_client.go
	mockery --dir ./internal/thelma/clients/kubernetes/testing/aliases --name Pods --output=./internal/thelma/clients/kubernetes/testing/mocks --outpkg mocks --filename pods.go
	mockery --dir ./internal/thelma/clients/kubernetes/testing/aliases --name Secrets --output=./internal/thelma/clients/kubernetes/testing/mocks --outpkg mocks --filename secrets.go
	mockery --dir ./internal/thelma/clients/kubernetes/testing/aliases --name Services --output=./internal/thelma/clients/kubernetes/testing/mocks --outpkg mocks --filename services.go
	mockery --dir ./internal/thelma/clients/kubernetes/testing/aliases --name StatefulSets --output=./internal/thelma/clients/kubernetes/testing/mocks --outpkg mocks --filename statefulsets.go
	mockery --dir ./internal/thelma/clients/kubernetes/testing/aliases --name Watch --output=./internal/thelma/clients/kubernetes/testing/mocks --outpkg mocks --filename watch.go
	mockery --dir ./internal/thelma/clients/sherlock --name ChartVersionUpdater --output=./internal/thelma/clients/sherlock/mocks --outpkg mocks --filename chart_version_updater.go
	mockery --dir ./internal/thelma/clients/sherlock --name StateReadWriter --output=./internal/thelma/clients/sherlock/mocks --outpkg mocks --filename state_read_writer.go
	mockery --dir ./internal/thelma/clients/sherlock --name StateReadWriter --output=./internal/thelma/clients/sherlock/mocks --outpkg mocks --filename state_read_writer.go
	mockery --dir ./internal/thelma/ops/sql/dbms --name DBMS --output=./internal/thelma/ops/sql/dbms/mocks --outpkg mocks --filename dbms.go
	mockery --dir ./internal/thelma/ops/sql/podrun --name Pod --output=./internal/thelma/ops/sql/podrun/mocks --outpkg mocks --filename pod.go
	mockery --dir ./internal/thelma/ops/sql/podrun --name Runner --output=./internal/thelma/ops/sql/podrun/mocks --outpkg mocks --filename runner.go
	mockery --dir ./internal/thelma/ops/sql/provider --name Provider --output=./internal/thelma/ops/sql/provider/mocks --outpkg mocks --filename provider.go
	mockery --dir ./internal/thelma/state/api/terra --name AppRelease --output=./internal/thelma/state/api/terra/mocks --outpkg mocks --filename app_release.go
	mockery --dir ./internal/thelma/state/api/terra --name AutoDelete --output=./internal/thelma/state/api/terra/mocks --outpkg mocks --filename auto_delete.go
	mockery --dir ./internal/thelma/state/api/terra --name Cluster --output=./internal/thelma/state/api/terra/mocks --outpkg mocks --filename cluster.go
	mockery --dir ./internal/thelma/state/api/terra --name ClusterRelease --output=./internal/thelma/state/api/terra/mocks --outpkg mocks --filename cluster_release.go
	mockery --dir ./internal/thelma/state/api/terra --name Clusters --output=./internal/thelma/state/api/terra/mocks --outpkg mocks --filename clusters.go
	mockery --dir ./internal/thelma/state/api/terra --name Destination --output=./internal/thelma/state/api/terra/mocks --outpkg mocks --filename destination.go
	mockery --dir ./internal/thelma/state/api/terra --name Destinations --output=./internal/thelma/state/api/terra/mocks --outpkg mocks --filename destinations.go
	mockery --dir ./internal/thelma/state/api/terra --name Environment --output=./internal/thelma/state/api/terra/mocks --outpkg mocks --filename environment.go
	mockery --dir ./internal/thelma/state/api/terra --name Environments --output=./internal/thelma/state/api/terra/mocks --outpkg mocks --filename environments.go
	mockery --dir ./internal/thelma/state/api/terra --name Release --output=./internal/thelma/state/api/terra/mocks --outpkg mocks --filename release.go
	mockery --dir ./internal/thelma/state/api/terra --name Releases --output=./internal/thelma/state/api/terra/mocks --outpkg mocks --filename releases.go
	mockery --dir ./internal/thelma/state/api/terra --name State --output=./internal/thelma/state/api/terra/mocks --outpkg mocks --filename state.go
	mockery --dir ./internal/thelma/state/api/terra --name StateLoader --output=./internal/thelma/state/api/terra/mocks --outpkg mocks --filename state_loader.go
	mockery --dir ./internal/thelma/state/api/terra --name StateWriter --output=./internal/thelma/state/api/terra/mocks --outpkg mocks --filename state_writer.go
	mockery --dir ./internal/thelma/toolbox/kubectl --name Kubectl --output=./internal/thelma/toolbox/kubectl/mocks --outpkg mocks --filename kubectl.go
	mockery --dir ./internal/thelma/utils/prompt --name Prompt --output=./internal/thelma/utils/prompt/mocks --outpkg mocks --filename prompt.go
