# Troubleshooting

## `expected one of "git"|"helm"|"oci"`

The generated API contract is stale. Confirm `SourceTypeVault` exists in `api/http/handler/gitops/sources/types.go`, run `make generate-api`, and commit the Swagger, OpenAPI, generated TypeScript, SDK, and Zod changes. Do not add `vault` only to `zod.gen.ts`.

## Connection test succeeds but secret reads fail

The health check treats Vault HTTP statuses below 500 as reachable; it does not prove the token can read a particular mount or path. Check the token policy, namespace, KV version, and secret mapping path. The mapping path must include the mount and must not include `/v1/`.

## KV v2 returns 404

Verify the configured KV version and mount name. Portainer converts `kv/app` to `/v1/kv/data/app`. Supplying a Vault UI path or omitting the mount produces the wrong API URL.

## Folder mappings fail behind a proxy

Portainer tries `LIST` first and retries with `GET ?list=true` only after 405 or 501. Check whether the proxy forwards one of those methods and preserves the query string and Vault headers. Other status codes are returned as folder-list failures.

## A folder mapping creates unexpected names

Only direct, non-folder children are expanded. A child with one value becomes `<child>`; a child with multiple values becomes `<child>_<key>`. Nested folders are skipped. Use explicit keyed mappings when stable custom names are required.

## Updating a source loses or exposes a token

The edit form should omit an unchanged token. The pointer-based update payload must leave stored authentication untouched when the token field is absent. Every response must pass through a sanitized DTO or `redactVaultSource`; add a regression test before changing this behavior.

## No Vault providers appear in stack creation

Check that the list request uses `type=vault`, that `sources.SourceType` accepts it, and that the response source has public type `"vault"`. Then inspect the generated Zod list query and response schemas. Finally verify the current user can read the source under its source access-control settings.
