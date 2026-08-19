FROM portainer/base:latest

ARG COMPOSE_UNPACKER_IMAGE=themodcrafttmc/compose-unpacker:2.39.3.2.3
ARG GIT_COMMIT=unspecified
ARG BUILD_DATE=unspecified
ENV COMPOSE_UNPACKER_IMAGE=${COMPOSE_UNPACKER_IMAGE}

LABEL org.opencontainers.image.title="Portainer CE" \
  org.opencontainers.image.description="Portainer Community Edition server." \
  org.opencontainers.image.vendor="Portainer.io" \
  org.opencontainers.image.revision=$GIT_COMMIT \
  org.opencontainers.image.created=$BUILD_DATE \
  io.portainer.server="true"

COPY --chmod=0755 portainer /portainer
COPY public /public/
COPY mustache-templates /mustache-templates/

VOLUME /data
WORKDIR /
EXPOSE 8000 9000 9443
ENTRYPOINT ["/portainer"]
