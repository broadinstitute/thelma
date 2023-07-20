# thelma

[![codecov](https://codecov.io/gh/broadinstitute/thelma/branch/main/graph/badge.svg?token=QYQHL6UE6Y)](https://codecov.io/gh/broadinstitute/thelma)
[![Go Report Card](https://goreportcard.com/badge/github.com/broadinstitute/thelma)](https://goreportcard.com/report/github.com/broadinstitute/thelma)

**`thelma`** (short for **T**erra **Helm** **A**utomator) is DSP-Devops' self service CLI tool for interacting with DSP infrastructure.

Thelma includes sub commands for a variety of different common use cases

The most common ones are 
1. Rendering kubernetes manifests for helm charts in the terra-helmfile repo
2. Packaging and publishing helm charts
3. Lifecycle management for BEEs (Branch Engineering Environment)
4. Connect securely to Cloudsql instances

And more ... `thelma --help` for more info

### Local Development

The Makefile has targets for facilitating local development, including:

    make help    # Print out all Makefile targets w/ brief description
    make build   # Compile thelma binary into output
    make release # Assemble a thelma release
    make test    # Run unit tests
    make smoke   # Run unit and smoke tests
    make cover   # View code coverage report for tests in browser

The `build` and `release` targets accept useful parameters:

    # Run a cross-platform build if desired
    make build OS=linux ARCH=amd64

    # Assemble a cross-platform build into a release archive
    make release OS=linux ARCH=arm64

#### Environment Setup
1. Ensure you have the go 1.19 toolchain installed on your local machine `brew install go@1.19`
2. Running thelma requires a local copy of the terra-helmfile repo. clone [terra-helmfile](https://github.com/broadinstitute/terra-helmfile)
3. Set the `THELMA_HOME` environment variable as the path to your local clone of terra-helmfile

#### Testing
The provided Makefile has utilities to easily run Thelma's test suites. 
From root of this repo `make test` will run all of thelma's test suites. 
*Note to future Thelma developers* Go's testing too chain runs all tests in parallel by default, be wary of this if writing any tests that rely on shared state.

`make smoke` will run just the smoke tests.

Both will automatically build a new thelma binary incorporating your latest changes before running the tests

#### Building Thelma Locally
`make build` will build a thelma binary incorporating the latest changes in your branch.
The resulting build artifact will be output to `./output/bin/thelma`, this will not interfere with other installs of thelma on your system.
From there you can test new functionality by running thelma commands directly on your machine via `$ ./output/bin/thelma [COMMAND]`
Since thelma is compiled there is no hot building. After each change you make it is necessary to rerun `make build`.

#### Common Use Cases
1. I want to try creating a BEE using my branch of thelma:
   `make build && ./output/bin/thelma bee create --name <UNIQUE_BEE_NAME>`

2. I want to render terra-helmfile k8s manifests for my application in my bee
   `make build && ./output/bin/thelma render -e [BEE_NAME] -a [MY_APP]`
