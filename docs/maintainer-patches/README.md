# Maintainer patch guide

This is the entry point for the features maintained by this fork. Keep it next to the existing [Vault integration guide](../vault-integration/README.md).

## Feature map

| Area                            | What this fork adds or repairs                                                                                | Maintenance guide                               |
| ------------------------------- | ------------------------------------------------------------------------------------------------------------- | ----------------------------------------------- |
| GitHub Container Registry       | First-class `ghcr.io` registry type, PAT authentication, organization scope, create/edit UI, API persistence  | [GHCR](github-container-registry.md)            |
| User activity logs              | Persistent seven-day audit trail for authenticated mutations, administrator list/filter/export UI             | [Activity logs](activity-logs.md)               |
| GitOps lifecycle                | Ignores deleted workflow references and preserves usable stacks/sources while cleanup catches up              | [GitOps and redeploy](gitops-redeploy.md)       |
| Re-pull and force redeploy      | Makes the combined action pull images and force service/container recreation for local and remote deployments | [GitOps and redeploy](gitops-redeploy.md)       |
| Relative-path volumes           | Keeps source/deployment directory flags intact during remote repull, prune, and forced redeployment           | [Relative paths](relative-path-volumes.md)      |
| Browser security and Cloudflare | Supported Permissions Policy, Cloudflare Web Analytics CSP, English-only i18n fallback, edge settings         | [Browser and Cloudflare](browser-cloudflare.md) |
| Release packaging               | Produces a standalone Docker build context and documents version/tag publishing                               | [Release process](release.md)                   |

## Upgrade checklist

When merging a new upstream Portainer release:

1. Resolve conflicts without dropping the data fields, API routes, registry type `8`, GitOps flags, or relative-path arguments listed in these guides.
2. Run `make generate-api` whenever Go API annotations or payloads change.
3. Run the focused Go and frontend tests listed in each guide.
4. Run `git diff --check` and a production frontend build.
5. Update `RELEASE_VERSION`, run `scripts/build-release-context.sh`, then build the image from `dist/release-context`.

These are clean-room Community Edition implementations. They do not patch license checks or load proprietary Business Edition code.
