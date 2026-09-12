# Live Docker logs and session-cached node metrics

Release `2.39.3.2.23` removes repetitive Docker log requests and keeps recent Kubernetes node charts available during navigation.

## Docker log streaming

Container, Swarm service, and task log views connect to authenticated `GET /api/websocket/logs`. The handler validates the environment, resource type, resource identifier, timestamps, time range, and tail size before forwarding a follow-mode request through Portainer's existing Docker proxy. That proxy continues to apply environment and Docker resource-control authorization.

Docker's raw TTY stream and eight-byte multiplexed stdout/stderr frames are decoded incrementally in the browser, including frames split across WebSocket messages. Updates are batched to at most ten per second. Both the server and browser cap the initial tail at 10,000 lines, the client rejects frames above 16 MiB, and the displayed history is trimmed continuously. Navigating away or pausing **Live log stream** closes the connection; an unexpected clean disconnect reconnects after one second.

The WebSocket carries only log bytes already authorized for the current user. It does not put credentials in the URL or expose a separate unauthenticated runtime connection.

### Container 404 protection

Release `2.39.3.2.25` fixes intermittent container detail and log initialization 404s on multi-node agent environments. Modern requests carry their target node explicitly, and the Axios compatibility interceptor no longer replaces that value with an older node left in the legacy FIFO queue. Container action links preserve `nodeName`, and the remaining AngularJS log controller passes it directly to the typed container query.

Authenticated mutations now clear server-side overview responses before the realtime invalidation event tells active browser queries to refetch. This prevents a successful recreate or deletion from immediately returning a cached collection containing the old container ID. Read-only Kubernetes `SelfSubjectAccessReview` requests are excluded from mutation publication so permission checks cannot create invalidation/refetch loops.

## Kubernetes node chart cache

Node stats retain up to 600 validated chart samples in `sessionStorage`, partitioned by environment and encoded node name. Returning through forward/back navigation renders that history immediately while the normal metrics query resumes. The latest matching timestamp is not appended twice.

The cache is tab-session scoped, versioned, and tied to the node CPU count used for percentage calculations. Invalid JSON, non-finite values, schema mismatches, CPU changes, disabled browser storage, and quota failures all fall back safely to an empty chart. Application/container statistics are not persisted by this change.

## Focused verification

```sh
GOCACHE=/tmp/portainer-websocket-go-cache go test ./api/http/handler/websocket -run 'TestCreateDockerLogs|TestEventHub' -count=1
pnpm exec vitest run app/docker/helpers/logHelper/liveLogs.test.ts app/react/kubernetes/metrics/useAggregatedMetrics.test.ts
pnpm typecheck
pnpm build
```

In the browser, open a running container's Logs view and confirm the switch reads **Live log stream**, new lines arrive without three-second request polling, and pausing stops the connection. For Kubernetes, open Node stats, navigate away and back, and confirm the previous chart samples render while fresh metrics resume.
