# GitHub Container Registry authentication

GitHub Container Registry uses registry type `8`, URL `ghcr.io`, a GitHub username, and a personal access token. A classic PAT needs at least `read:packages`; an organization may also require SSO authorization.

## Code ownership

- `api/portainer.go`: `GithubRegistryData`, `GithubRegistry`, and the field on `Registry`.
- `api/http/handler/registries/registry_create.go`: accepts type `8` and persists GitHub organization metadata.
- `api/http/handler/registries/registry_update.go`: updates GitHub metadata without replacing a stored token when the password field is blank.
- `app/portainer/models/registryTypes.js` and `app/portainer/models/registry.js`: Angular/API model mapping.
- `app/portainer/components/forms/registry-form-github/`: create form.
- `app/portainer/registry-management/views/create/` and `edit/`: provider selection and editing.
- `app/react/portainer/registries/`: provider label, icon, and selector option.

The JSON field spelling is intentional: API responses use `Github.UseOrganisation` and `Github.OrganisationName`; the create form uses lower-camel local model fields and `RegistryCreateRequest` converts them.

After modifying the payload or enum, run `make generate-api` and verify `GithubRegistryData` plus enum value `8` remain in both OpenAPI files and generated TypeScript.
