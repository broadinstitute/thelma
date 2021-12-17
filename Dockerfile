#
# Compile Go tools and install to /tools/bin
#
ARG GO_IMAGE_VERSION='1.16'
ARG ALPINE_IMAGE_VERSION='3.14'

FROM golang:${GO_IMAGE_VERSION}-bullseye as build

ARG THELMA_VERSION='unknown'

WORKDIR /build
COPY . .

# Compile & install runtime dependencies into output/release
RUN make release VERSION=${THELMA_VERSION}

RUN pwd

#
# Copy dist into runtime image
#
FROM alpine:${ALPINE_IMAGE_VERSION} as runtime

# OS updates for security
RUN apk update
RUN apk upgrade

# Copy Thelma into runtime image
COPY --from=build /build/output/release /thelma

ENV PATH="/thelma/bin:${PATH}"

RUN find /thelma/

# Make sure thelma executes
RUN /thelma/bin/thelma --help
