# update

The `update` package implements Thelma's self-update logic.

## Release Publishing

Thelma releases are semantically versioned, with automatic bumping via the [bumper GitHub action](https://github.com/DataBiosphere/github-actions/tree/master/actions/bumper). Releases are published as an [Alpine Docker image](https://console.cloud.google.com/artifacts/docker/dsp-artifact-registry/us-central1/thelma/thelma?project=dsp-artifact-registry) and as cross-platform binaries to a GCS bucket, called [thelma-releases](https://console.cloud.google.com/storage/browser/thelma-releases;tab=objects?forceOnBucketsSortingFiltering=false&project=dsp-artifact-registry&prefix=&forceOnObjectsSortingFiltering=false).

## Binary Releases

Thelma subcommands have a number of runtime dependencies that are also shipped as cross-platform binaries, such as Helm, Helmfile, `yq`, and the ArgoCD CLI client. [^1] Therefore, Thelma's binary releases are published as tar.gz packages that contain thelma as well as these dependencies. Release tarballs have the following structure:

```
    # The root contains a build manifest file with version info
    ./build.json  

    # Executables live in the ./bin/ subdirectory
    ./bin/thelma
    ./bin/helm
    ./bin/helmfile
    ... # more binaries omitted
```

Currently, releases for both linux and darwin are published to the GCS bucket, under the releases/ subdirectory:

```
     releases/v0.0.21/thelma_v0.0.21_SHA256SUMS
     releases/v0.0.21/thelma_v0.0.21_darwin_amd64.tar.gz
     releases/v0.0.21/thelma_v0.0.21_linux_amd64.tar.gz
```

[^1]: Note that this is relatively unusual for a Go program and as a result, we are unable to use the popular [go-releaser](https://goreleaser.com/) project to package Thelma. Instead, packaging logic is contained in Thelma's Makefile. `make release` can be used to assemble a Thelma binary package locally.

## Release Tags

Thelma's Docker images use the `latest` tag as well as semantic version tags like `v0.0.21`. Similar functionality is implemented for binary releases using a `tags.json` file at the root of the GCS bucket. This file is a simple map, with tag names as the keys and semantic release versions as values. For example:

```
{
    "latest": "v0.0.21"
}
```

Thelma's CI/CD pipelines update the `tags.json` file at the same time as they update the `latest` Docker tag. At the time of this writing, `latest` is the only supported tag.
