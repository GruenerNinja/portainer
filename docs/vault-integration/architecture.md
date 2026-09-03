# Architecture and data flow

## Components

| Responsibility                                                | Primary files                                                                                                                                           |
| ------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Persistent source and stack mapping models                    | `api/portainer.go`                                                                                                                                      |
| Source type conversion and list filtering                     | `api/http/handler/gitops/sources/types.go`, `list.go`                                                                                                   |
| Create, update, read, redact, and connection-test handlers    | `api/http/handler/gitops/sources/create_vault.go`, `update_vault.go`, `get.go`, `source_connection.go`, `handler.go`                                    |
| Vault HTTP requests, KV path conversion, and folder expansion | `api/gitops/secrets/vault.go`                                                                                                                           |
| Periodic token renewal worker                                 | `api/gitops/secrets/renewer.go`                                                                                                                         |
| Deploy-time mapping resolution                                | `api/stacks/deployments/secret_mappings.go` and the deployer entry points in the same package                                                           |
| Source create/edit UI                                         | `app/react/portainer/gitops/sources/`                                                                                                                   |
| Stack mapping UI and payload conversion                       | `app/react/common/stacks/SecretMappingsFieldset.tsx`, `app/react/common/stacks/queries/useCreateStack/`, and `app/react/common/stacks/EditGitSettings/` |
| Generated API contract and runtime validators                 | `api/docs/swagger.yaml`, `api/docs/openapi.yaml`, and `app/react/portainer/generated-api/portainer/`                                                    |

Paths in the last row are generated artifacts. Change Go declarations and Swagger annotations, then regenerate them.

## Source lifecycle

1. An administrator creates or tests a Vault source through `/gitops/sources/vault` or `/gitops/sources/vault/test`.
2. `VaultSourceCreatePayload.Validate` currently allows only the `token` authentication method and requires a non-empty token.
3. `BuildVaultSource` defaults KV version `0` to version `2`, trims user-entered public address/internal address/namespace/name values, and stores the configuration as `portainer.SourceTypeVault`.
4. Read responses expose connection metadata but not the token. Create and update responses call `redactVaultSource` before serialization.
5. Updates and stored-source connection tests use the generic `/{id}` and `/{id}/test` routes, dispatching on the stored numeric `portainer.SourceType`.

Create and pre-save test routes are administrator-only. Other source routes use normal authenticated source access checks. Preserve that distinction deliberately if authorization changes.

## Deployment flow

1. A stack stores zero or more `StackSecretMapping` values: `sourceId`, `path`, optional `key`, and optional target `name`.
2. Each local or remote deployer calls `stackWithResolvedSecrets` before invoking the deployment implementation.
3. The resolver loads the referenced source and rejects non-Vault sources.
4. With a `key`, one Vault value is read and written to `name`; if `name` is empty, the key is used.
5. Without a `key`, all values at the path are injected. If a read returns 404, the backend lists direct children and reads each non-folder child.
6. Existing environment variables with the same name are replaced; new variables are appended.
7. The stored stack is not mutated. Resolution returns a shallow stack copy with a separate environment slice for that deployment.

Folder expansion names a child containing one value after the child secret. For a child with several values, it uses `<child>_<key>`. Nested folders are skipped rather than recursively expanded.

## Vault HTTP behavior

All requests use a dedicated client with a 15-second timeout. When an internal address is configured, the backend tries it first and falls back to the public/domain address if the request fails. This keeps secret resolution working while the public reverse proxy is being redeployed. `tlsSkipVerify` clones the default transport and disables certificate verification only for that client. Token and namespace headers are applied centrally by `applyVaultHeaders`.

For a mapping path such as `kv/app`:

| Operation   | KV v1        | KV v2                 |
| ----------- | ------------ | --------------------- |
| Read        | `/v1/kv/app` | `/v1/kv/data/app`     |
| List folder | `/v1/kv/app` | `/v1/kv/metadata/app` |

Folder listing first uses the Vault `LIST` method. A 405 or 501 response retries with `GET ...?list=true` for proxies that do not pass `LIST`.

## Periodic token renewal

Portainer starts one Vault token renewal worker with the application lifecycle. It checks all stored Vault sources immediately at startup and every hour afterward. For each source, it calls `auth/token/lookup-self`. Renewable periodic tokens are renewed through `auth/token/renew-self` when their remaining TTL is at or below half of their configured period. Non-periodic and non-renewable tokens are left unchanged. Lookup and renewal use the same internal-first address fallback, namespace, TLS, and token-header behavior as secret reads. Logs identify only the source and renewed TTL; they never include the token.
