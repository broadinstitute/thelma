ARG ALPINE_IMAGE_VERSION='3.14'

#
# Copy dist into runtime image
#
FROM alpine:${ALPINE_IMAGE_VERSION}

# OS updates for security
RUN apk update
RUN apk upgrade

# Copy Thelma into runtime image
COPY output/release-assembly /thelma

ENV PATH="/thelma/bin:${PATH}"

# Make sure thelma executes
RUN /thelma/bin/thelma --help
