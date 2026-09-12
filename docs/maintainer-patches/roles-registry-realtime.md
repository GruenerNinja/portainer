# Custom roles, GHCR browsing, and realtime UI updates

Release `2.39.3.2.22` completes the Community implementation of custom environment roles, provides working GitHub Container Registry repository and tag browsing, fixes read-only-only stack access validation, and reduces repetitive UI requests.

## Role identifiers and custom roles

`AccessPolicy.RoleId` is an open integer contract:

- `0` is valid for access policies that do not assign an environment role, including registry policy records;
- `1` through `5` are Portainer's protected built-in roles;
- `6` and above are administrator-managed custom roles.

The administrator Roles page can create and edit custom roles, select their authorizations, and set their conflict priority. Higher priority numbers win when multiple environment or group policies apply. A custom role can be deleted only after it has been removed from every user/team access policy. Built-in roles cannot be edited or deleted and are restored during startup migrations.

The access pickers read roles from `/api/roles`; do not reintroduce a hard-coded list or narrow role IDs to a generated enum.

## GitHub Container Registry

Authenticated `ghcr.io` entries can open Community-owned repository and read-only tag-list views. The access-checked backend uses the GitHub Packages API for GHCR catalogs and ORAS for repository tags, keeping registry credentials on the server. The Browse action is a Community feature in this fork and must not be wrapped in `REGISTRY_MANAGEMENT` or another Business feature indicator. Registry mutation actions remain absent from these read-only browser views.

## Realtime query synchronization

The authenticated `/api/websocket/events` endpoint emits only `{"type":"invalidate"}` after mutating API requests complete. It never includes paths, resource identifiers, usernames, or payload data. Browsers reconnect automatically and invalidate active React Query data when the signal arrives.

The default client cache keeps data fresh for 30 seconds and retained for 10 minutes, disables refetch-on-window-focus, and keeps the previous result visible while a changed query loads. Explicit page refresh and user-configured polling continue to work.

## Focused verification

```sh
GOCACHE=/tmp/portainer-rolefix-go-cache go test ./api/http/handler/roles
GOCACHE=/tmp/portainer-rolefix-go-cache go test ./api/http/handler/websocket -run TestEventHubPublishesCoalescedInvalidations
GOCACHE=/tmp/portainer-rolefix-go-cache go test ./api/http/handler/registries -run '^TestRegistryRepositor'
pnpm test app/react/portainer/users/RolesView/role-schema.test.ts --run
pnpm typecheck
```

In the browser, confirm that Roles shows Add role, that opening the editor lists permissions, and that built-in rows have neither Edit nor selectable delete controls. On Registries, confirm that an authenticated GHCR row has an enabled Browse action without a Business Feature label, then open a repository and verify its read-only tag list.
