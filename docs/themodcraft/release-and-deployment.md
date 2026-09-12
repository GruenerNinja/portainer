# Release and deployment

## Build and publish

The release source is `RELEASE_VERSION`. `scripts/build-release-context.sh` builds the frontend once and produces static Linux binaries for both AMD64 and ARM64 under `dist/release-context/`.

The generated `BUILD-IMAGE.txt` contains the exact commands for the release:

```sh
./scripts/build-release-context.sh
cd dist/release-context
docker buildx build --platform linux/amd64,linux/arm64 \
  -t themodcrafttmc/portainer:2.39.3.2.17 --push .
docker buildx imagetools create \
  -t themodcrafttmc/portainer:latest \
  themodcrafttmc/portainer:2.39.3.2.17
```

The immutable version is published and verified before `latest` is moved. Both registry references must resolve to the same OCI index digest and contain `linux/amd64` and `linux/arm64` manifests.

## Node 4 deployment

Node 4 runs the AMD64 image as a standalone Docker container named `portainer`. Preserve these runtime settings during replacement:

- restart policy `always`;
- host ports 8000 and 9443;
- volume `portainer_data:/data`;
- bind `/var/run/docker.sock:/var/run/docker.sock`;
- bridge networking.

Credentials stay outside this repository. The service-account descriptor refers to a macOS Keychain entry; never copy the password into scripts, documentation, shell history, or environment files.

Before replacement, pull the immutable tag and verify its selected architecture and registry digest. Stop and rename the prior container to `portainer-rollback-<version>`, then start the new container with the same settings. Verify `https://127.0.0.1:9443/api/status`, the running image digest, restart policy, and startup logs.

## Rollback

If startup or the HTTPS check fails:

1. stop and remove only the failed new `portainer` container;
2. rename the retained rollback container to `portainer`;
3. start it and verify `/api/status`.

The named `portainer_data` volume is never removed during deployment or rollback.
