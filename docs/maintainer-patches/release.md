# Release process

The release version is stored in `RELEASE_VERSION`. The build script produces a disposable, self-contained context in `dist/release-context`.

```sh
./scripts/build-release-context.sh
cd dist/release-context
docker build --platform linux/arm64 -t themodcrafttmc/portainer:2.39.3.2.11 -t themodcrafttmc/portainer:latest .
```

The script installs the locked frontend dependencies, builds the production UI, builds a static Linux backend with the release version embedded, and copies only the runtime artifacts plus a small Dockerfile into the context. Set `TARGETARCH=amd64` or `TARGETARCH=arm64` before running it when building for a different target than the current host. The generated Docker command includes the matching `--platform` value so the base image and binary architecture cannot be mixed accidentally.

Push the immutable version first and `latest` second:

```sh
docker push themodcrafttmc/portainer:2.39.3.2.11
docker push themodcrafttmc/portainer:latest
```

Both tags must refer to the same locally built image. Verify with `docker image inspect` locally and compare registry digests after pushing. Increment `RELEASE_VERSION` for every release; never reuse a published version tag.
