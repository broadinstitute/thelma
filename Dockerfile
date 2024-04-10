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

# ArgoCD plugins always run as UID 999. This is set by Kubernetes, so no need to make it the default here,
# but Thelma does attempt to use the home directory *a lot*, and we need to have the 999 user not have /
# as its home directory (because 999 isn't root, so it can't modify /). Normally we'd just create a new
# user with `adduser -u 999 argocd-thelma`, but `adduser` refuses because it'll say that 999 is taken.
# 999 isn't defined in any of the /etc files, though, so to set its home directory we manually add lines
# to those files.
RUN mkdir /home/argocd-thelma
RUN echo 'argocd-thelma:!::0:::::' >> /etc/shadow
RUN echo 'argocd-thelma:x:999:' >> /etc/group
RUN echo 'argocd-thelma:x:999:999:argocd-thelma:/home/argocd-thelma:/bin/nologin' >> /etc/passwd
RUN chown 999:999 /home/argocd-thelma

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
