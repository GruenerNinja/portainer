# Release process

The release version is stored in `RELEASE_VERSION`. The build script produces a disposable, self-contained context in `dist/release-context`.

```sh
./scripts/build-release-context.sh
cd dist/release-context
docker buildx build --platform linux/amd64,linux/arm64 \
  -t themodcrafttmc/portainer:2.39.3.2.25 --push .
```

The script installs the locked frontend dependencies, builds the production UI once, builds static Linux backends for both AMD64 and ARM64 with the release version embedded, and copies only the runtime artifacts plus a small multi-architecture Dockerfile into the context. The generated build command publishes an immutable manifest containing both platforms.

Push the immutable version first and `latest` second:

```sh
docker buildx imagetools create \
  -t themodcrafttmc/portainer:latest \
  themodcrafttmc/portainer:2.39.3.2.25
```

Publish the immutable version before moving `latest`. Verify that both registry tags resolve to the same manifest digest and contain `linux/amd64` plus `linux/arm64`. Increment `RELEASE_VERSION` for every release; never reuse a published version tag.
