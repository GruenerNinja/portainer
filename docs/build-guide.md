# Build Guide

This guide covers the common local build commands for Portainer CE.

## Prerequisites

Install the project toolchain versions declared in the repo:

- Docker
- Go 1.26.4
- Node.js 22.22.1 or newer in the 22.x line
- pnpm 10.26.2

## Install dependencies

```sh
make deps
```

This installs the frontend dependencies and prepares the server build output
directory.

## Build everything

```sh
make build-all
```

`make build-all` is an alias for the full local build. It runs Go module tidy,
installs dependencies, builds the server binary, and builds the frontend.

The build artifacts are written under `dist/`.

## Build only one side

```sh
make build-server
make build-client
```

Use these targets when you only need the Go server or the frontend bundle.

## Build a local image

```sh
make build-image
```

This builds the app and then creates a local Docker image tagged
`portainerci/portainer-ce:local`. Override the tag with `TAG=<name>` when
needed:

```sh
make build-image TAG=my-local-build
```

## Clean build output

```sh
make clean
```

Run `make help` to see every available target.
