#
# Compile Go tools and install to /tools/bin
#
ARG GO_IMAGE_VERSION='1.16'
ARG ALPINE_IMAGE_VERSION='3.14'

FROM golang:${GO_IMAGE_VERSION}-bullseye as build

ARG THELMA_VERSION='development'

WORKDIR /build
COPY . .

# Compile & install runtime dependencies into output/release-assembly
RUN make release VERSION=${THELMA_VERSION}

#
# Copy dist into runtime image
#
FROM alpine:${ALPINE_IMAGE_VERSION} as runtime

# OS updates for security
RUN apk update
RUN apk upgrade

# Copy Thelma into runtime image
COPY --from=build /build/output/release-assembly /thelma

ENV PATH="/thelma/bin:${PATH}"

# Make sure thelma executes
RUN /thelma/bin/thelma --help
