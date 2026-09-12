# Overview response cache

## Request flow

```text
Docker / Podman / Kubernetes
            |
            v
  raw overview cache
            |
            v
 authorization filtering
            |
            v
      Portainer API / UI
```

The shared cache is implemented in `api/http/proxy/cache/`. Docker and Podman use it in `api/http/proxy/factory/docker/transport.go`; Kubernetes uses it in `api/http/proxy/factory/kubernetes/transport.go`.

## Cache behavior

- A cold request waits for the runtime once because no valid response exists yet.
- A warm request returns the cached response immediately and starts a coalesced refresh in the background.
- Active entries are refreshed hourly even when no UI request is waiting.
- Entries unused for two hours are evicted.
- At most 512 responses and 128 MiB per response are cached.
- Background refresh concurrency is limited to four requests.
- Failed and non-2xx refreshes do not replace the last successful response.
- Authenticated API mutations clear the overview cache before realtime query
  invalidation is published, so a refetch cannot receive identifiers for
  resources that the mutation replaced or removed.
- Concurrent cold or refresh requests for the same key are collapsed with singleflight.
- Cached responses are replayed as independent header/body copies so later filtering cannot mutate the stored value.
- Deleting or replacing an environment proxy invalidates that environment's Docker and Kubernetes entries.

## Docker and Podman scope

Only exact `GET` collection routes are cached, including version-prefixed Docker API paths:

- `/containers/json`
- `/images/json`
- `/networks`
- `/nodes`
- `/services`
- `/tasks`
- `/volumes`

The cache key includes environment ID, agent target, method, path, canonical query string, `Accept`, and `Accept-Encoding`. One raw runtime response can therefore be reused while the existing Portainer resource-control filter still produces a user-specific result.

The Docker dashboard has a separate aggregate cache in `api/http/handler/docker/dashboard_cache.go`. It caches container, image, service, volume, network, and container-stat aggregates by environment URL and agent target, then applies user access control when creating the response.

## Kubernetes scope and isolation

Cluster-scoped and namespaced `GET` collection routes are cached. Examples include nodes, namespaces, pods, deployments, and services. Watch requests, individual-resource paths, `configmaps`, and `secrets` are excluded.

Kubernetes cache keys include the Portainer user ID and role. The upstream request already carries that user's Kubernetes service-account token, so sharing an administrator's raw response with another user would cross the RBAC boundary.

## Explicit exclusions

- Inspect/detail routes such as `/containers/{id}/json` or `/networks/{id}`.
- Configuration and secret data.
- Logs, consoles, stats streams, Kubernetes watches, and other streaming responses.
- `POST`, `PUT`, `PATCH`, and `DELETE` requests.
