# thelma

[![codecov](https://codecov.io/gh/broadinstitute/thelma/branch/main/graph/badge.svg?token=QYQHL6UE6Y)](https://codecov.io/gh/broadinstitute/thelma)
[![Go Report Card](https://goreportcard.com/badge/github.com/broadinstitute/thelma)](https://goreportcard.com/report/github.com/broadinstitute/thelma)

**`thelma`** (short for **T**erra **Helm** **A**utomator) is a CLI tool for interacting with Terra's Helm charts.

It includes subcommands for publishing charts to the terra-helm repo as well as rendering manifests locally.

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
