# TheModCraft Portainer customizations

This directory is the maintenance handoff for behavior added by this fork. It explains the design, security boundaries, release identity, verification, and deployment process without requiring readers to reconstruct the changes from Git history.

## Contents

- [Recent features](../recent-features/README.md): table of work added, restored, or consolidated during the last four weeks.
- [Overview response cache](overview-cache.md): stale-while-revalidate caching for Docker, Podman, and Kubernetes overview pages.
- [Custom versioning](custom-versioning.md): footer version, Docker Hub update checks, and the boundary between the fork release and the upstream API/database version.
- [Release and deployment](release-and-deployment.md): AMD64/ARM64 image production, Node 4 deployment, and rollback.
- [Verification](verification.md): focused tests and release checks.
- [Read-only stack sharing](../maintainer-patches/stack-read-only-access.md): owner/editor versus viewer permissions for selected stacks and agentic accounts.

The broader patch catalog remains in [the maintainer patch guide](../maintainer-patches/README.md).

## Design principles

1. Cache only collection data used by overview pages. Inspect/configuration pages and mutations always reach the runtime.
2. Cache data before Portainer's Docker authorization filter, then filter an independent response copy for each user.
3. Never share Kubernetes results across users because Kubernetes service-account RBAC is applied upstream.
4. Keep the upstream API/database compatibility version separate from this fork's release version.
5. Publish immutable multi-architecture version tags before moving `latest`.
6. Treat stack viewer grants as configuration visibility only; never inherit them as Docker-resource mutation access or expose redeploy webhook tokens.
