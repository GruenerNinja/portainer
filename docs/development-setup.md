# Development Setup Guide

This guide covers the shortest path to running Portainer CE locally for
development.

## Prerequisites

Install:

- Docker
- Go 1.26.4
- Node.js 22.22.1 or newer in the 22.x line
- pnpm 10.26.2

Docker must be running because the development server runs Portainer in a local
container and mounts the Docker socket.

## Install dependencies

```sh
make deps
```

## Start the development environment

```sh
make dev
```

This builds the server, starts the Portainer container, installs frontend
dependencies if needed, and starts the frontend development server.

Default local URLs:

- Portainer server: <https://localhost:9443>
- Frontend dev server: <http://localhost:8999>

## Run server and client separately

Use separate targets when you want independent terminal sessions:

```sh
make dev-server
make dev-client
```

`make dev-server` rebuilds the server binary and starts a container named
`portainer`. `make dev-client` runs the frontend development server.

## Useful development settings

The server container script supports these environment variables:

- `PORTAINER_DATA`: host directory or volume used for Portainer data. Defaults
  to `/tmp/portainer-ce`.
- `PORTAINER_PROJECT`: repository root mounted into the container. Defaults to
  the current directory.
- `PORTAINER_FLAGS`: extra flags passed to the Portainer binary.

Example:

```sh
PORTAINER_DATA=$HOME/.portainer-dev PORTAINER_FLAGS="--log-level=DEBUG" make dev-server
```

Run `make help` for the full list of development, test, lint, and build targets.
