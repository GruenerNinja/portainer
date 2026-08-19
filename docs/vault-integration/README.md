# Vault integration maintainer guide

This is the main entry point for Portainer's HashiCorp Vault integration. The integration stores a reusable Vault connection as a GitOps source and resolves configured Vault values into a stack's environment immediately before deployment.

The most important maintenance rule is: **after changing a Go API type, enum, payload, route annotation, or stack field, run `make generate-api` and commit every relevant generated API artifact.** Do not hand-edit the generated TypeScript client or Zod schemas.

Missing that step caused Vault source responses with `type: "vault"` to fail at runtime with:

```text
Invalid option: expected one of "git"|"helm"|"oci"
```

The Go API already accepted Vault, but the committed OpenAPI definition and generated Zod validator did not. See [Changing the integration](maintenance.md#api-schema-and-client-generation) for the complete update sequence.

## Guide contents

- [Architecture and data flow](architecture.md) explains the stored model, API, UI, Vault client, and deploy-time secret resolution.
- [Changing the integration](maintenance.md) lists the files that must move together for common changes, including source types, authentication, KV behavior, and stack mappings.
- [Testing and release checklist](testing.md) gives focused commands and a manual smoke test.
- [Troubleshooting](troubleshooting.md) maps common symptoms to likely causes.

## Current supported behavior

- HashiCorp Vault KV v1 and KV v2.
- Token authentication through `X-Vault-Token`.
- Optional Vault Enterprise/HCP namespace through `X-Vault-Namespace`.
- Optional TLS certificate verification bypass.
- A single key mapped to an environment variable.
- All keys at a secret path mapped to environment variables.
- Direct child secrets of a folder expanded when the requested path is not itself a secret.
- Vault values are fetched during deployment; resolved values are copied into the in-memory stack used by the deployer rather than written back to the stored stack.

## Operational contract

Use the Vault server base URL as the source address, for example `https://vault.example.com`. Do not use a Vault UI URL. Secret mapping paths include the KV mount, for example `kv/apps/my-stack`; callers should not prepend `/v1/`. For KV v2, the backend inserts `data` for reads and `metadata` for folder listings.

The stored token is sensitive. API responses must continue to redact it. Never log the Vault configuration, request headers, resolved values, or the final environment after secret injection.
