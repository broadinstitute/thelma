#
# Compile Go tools and install to /tools/bin
#
ARG THELMA_VERSION=unknown
ARG GO_VERSION='1.16'
ARG ALPINE_VERSION='3.14'

FROM golang:${GO_VERSION}-bullseye as build

WORKDIR /build
COPY . .

# Compile & install runtime dependencies into output/release
RUN make release RELEASE_DIR=/thelma VERSION=${THELMA_VERSION}

#
# Copy dist into runtime image
#
FROM alpine:${ALPINE_VERSION} as runtime

# OS updates for security
RUN apk update
RUN apk upgrade

# Copy Thelma into runtime image
COPY --from=build /thelma /thelma
ENV PATH="/thelma/bin:${PATH}"
