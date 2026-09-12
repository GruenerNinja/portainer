# Custom release versioning

This fork has two distinct version concepts and intentionally keeps them separate.

## Fork release version

`RELEASE_VERSION` is the authoritative TheModCraft release identifier. The build script injects it into `pkg/build.ImageTag`; released images use tags such as `2.39.3.2.16`.

The authenticated `GET /api/system/version` response exposes this build tag as `ServerVersion` when it is a valid numeric release. Development builds without a release tag fall back to the upstream API version.

The sidebar footer and build-information dialog display this server version. They no longer display the upstream `2.45.0 LTS` string as the installed fork release.

## Update discovery

The backend checks the public tags for `themodcrafttmc/portainer` on Docker Hub instead of checking upstream Portainer GitHub releases. It:

1. fetches up to 100 recently updated tags every six hours;
2. ignores aliases and non-numeric tags such as `latest`;
3. compares every numeric component, including the fork's five-part versions;
4. reports an update only when a newer maintained image tag exists.

The UI's update link opens the matching Docker Hub tag. Browser-side version data remains cached for 24 hours, with a hard browser refresh causing a new API request.

## Upstream compatibility version

`portainer.APIVersion` and the database schema version remain upstream `2.45.0`. They are used for API compatibility, migrations, agents, and datastore safety; changing them to the fork release number would incorrectly trigger schema behavior.
