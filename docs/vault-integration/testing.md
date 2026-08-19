# Testing and release checklist

## Focused automated tests

Run backend tests for the Vault client, source handlers, and deploy-time mappings:

```sh
go test ./api/gitops/secrets ./api/http/handler/gitops/sources ./api/stacks/deployments -count=1
```

Run the focused frontend tests after form, mapping, query, or payload changes:

```sh
pnpm vitest run \
  app/react/portainer/gitops/sources \
  app/react/common/stacks \
  app/react/docker/stacks/CreateView/CreateStackForm
```

Verify the generated contract and formatting:

```sh
make docs-sync-check
make generate-api
git diff --check
```

After generation, confirm `zSourcesSourceType` and `zGitOpsSourcesListQuery` in `app/react/portainer/generated-api/portainer/zod.gen.ts` both accept `vault`.

## Manual smoke test

1. Start a Vault development/test instance with a known token.
2. Enable a KV v1 or KV v2 mount and write a string secret.
3. In Portainer, create a Vault source with the server base address, matching KV version, token, and namespace if applicable.
4. Test the connection, save the source, reload the source list, and open its settings page. A reload is important because it exercises generated response validation.
5. Create a Git-backed stack and add a Vault mapping for one key. Deploy and confirm the application receives the value without the value appearing in the stored stack configuration or API response.
6. Repeat with an empty key to test all-key expansion.
7. Map a folder and verify direct-child naming for both single-value and multi-value child secrets.
8. Update the Vault source without entering a replacement token and confirm the existing credential remains usable; then replace it and retest.
9. Exercise a denied source/stack action as a non-administrator to verify access control.

## Release checklist

- Go numeric enum values were appended, not reordered.
- Public string enums, parsing, and conversion switches include Vault.
- Swagger/OpenAPI and generated TypeScript/Zod artifacts are committed.
- Vault endpoints and fields appear in the generated SDK and types.
- All credentials are redacted from create, get, update, test errors, and logs.
- KV v1, KV v2, namespace, TLS, and folder-list behavior have tests appropriate to the change.
- All stack deployment paths still call the secret resolver.
- No resolved value is persisted or logged.
- Source and stack access-control behavior was reviewed.
- The manual reload test produces no Zod `invalid_value` error.
