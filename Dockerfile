FROM portainer/base:latest AS production

ARG TARGETARCH
ARG COMPOSE_UNPACKER_IMAGE=themodcrafttmc/compose-unpacker:2.39.3.2.3
ENV COMPOSE_UNPACKER_IMAGE=${COMPOSE_UNPACKER_IMAGE}

LABEL org.opencontainers.image.title="Portainer" \
  org.opencontainers.image.description="Docker container management made simple, with the world's most popular GUI-based container management platform." \
  org.opencontainers.image.vendor="Portainer.io" \
  com.docker.desktop.extension.api.version=">= 0.2.2" \
  com.docker.extension.publisher-url="https://www.portainer.io" \
  com.docker.extension.additional-urls="[{\"title\":\"Website\",\"url\":\"https://www.portainer.io?utm_campaign=DockerCon&utm_source=DockerDesktop\"},{\"title\":\"Documentation\",\"url\":\"https://docs.portainer.io\"},{\"title\":\"Support\",\"url\":\"https://join.slack.com/t/portainer/shared_invite/zt-txh3ljab-52QHTyjCqbe5RibC2lcjKA\"}]"

COPY dist/mustache-templates /mustache-templates/
COPY --chmod=0755 dist/portainer-${TARGETARCH} /portainer
COPY dist/public /public/

COPY build/docker-extension /

# storybook exists only in portainerci builds
COPY dist/storybook* /storybook/

VOLUME /data
WORKDIR /

EXPOSE 9000
EXPOSE 9443
EXPOSE 8000

ARG GIT_COMMIT=unspecified
ARG BUILD_DATE=unspecified
LABEL git_commit=$GIT_COMMIT \
  org.opencontainers.image.revision=$GIT_COMMIT \
  org.opencontainers.image.created=$BUILD_DATE \
  org.opencontainers.image.title="Portainer CE" \
  org.opencontainers.image.description="Portainer Community Edition server." \
  org.opencontainers.image.vendor="Portainer.io" \
  org.opencontainers.image.url="https://www.portainer.io" \
  org.opencontainers.image.documentation="https://docs.portainer.io" \
  io.portainer.server="true"

ENTRYPOINT ["/portainer"]
