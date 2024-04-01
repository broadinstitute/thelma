ARG ALPINE_IMAGE_VERSION='3.14'

#
# Copy dist into runtime image
#
FROM alpine:${ALPINE_IMAGE_VERSION}

ARG THELMA_LINUX_RELEASE

COPY output/releases/${THELMA_LINUX_RELEASE} .

# Include ArgoCD plugin.yaml
# https://argo-cd.readthedocs.io/en/stable/operator-manual/config-management-plugins/#place-the-plugin-configuration-file-in-the-sidecar
COPY argocd/plugin.yaml /home/argocd/cmp-server/config/

# OS updates for security
RUN apk update
RUN apk upgrade

# Unpack Thelma into runtime image
RUN mkdir /thelma && tar -xvf ${THELMA_LINUX_RELEASE} -C /thelma

# Remove the copied tarball
RUN rm ${THELMA_LINUX_RELEASE}

ENV PATH="/thelma/bin:/thelma/tools/bin:${PATH}"

# Make sure thelma executes
RUN thelma --help
