# Read-only stack sharing

This fork supports per-stack viewer grants in addition to the existing read-write resource-control grants. It is intended for service users, agentic accounts, and teams that manage their own stacks but need visibility into selected stacks owned by somebody else.

## Access model

| Relationship to stack                  | List and inspect | Read Compose file and stored environment configuration | Update, redeploy, start, stop, migrate, or delete | Change access control                   |
| -------------------------------------- | ---------------- | ------------------------------------------------------ | ------------------------------------------------- | --------------------------------------- |
| Read-write user or team                | Yes              | Yes                                                    | Yes, subject to the environment role              | Yes, under the existing ownership rules |
| Explicit read-only user or team        | Yes              | Yes                                                    | No                                                | No                                      |
| No stack grant                         | No               | No                                                     | No                                                | No                                      |
| Portainer or environment administrator | Yes              | Yes                                                    | Yes                                               | Yes                                     |

Use the **Standard user** environment role for an agent that must create and manage its own stacks. Assign its user or owning team read-write access to those stacks, then add read-only user/team grants only to the other stacks it may inspect. API keys inherit the permissions of their user.

Administrators configure viewer grants in a stack's **Access control** panel or during stack creation. Read-only users and teams are stored in the existing `UserAccesses` and `TeamAccesses` collections with `AccessLevel: 2`; existing read-write grants remain `AccessLevel: 1`.

For backward compatibility, legacy resource-control entries that omit `AccessLevel` and decode as `0` retain their historical read-write behavior.

## Security boundaries

- Stack list, inspect, and file endpoints accept either access level.
- All stack mutations and resource-control changes require read-write access.
- Read-only grants do not inherit to Docker containers, services, volumes, networks, configs, or secrets, preventing direct runtime mutation through the Docker proxy.
- Viewer stack responses are marked `ReadOnly: true`. The UI disables bulk selection, ownership changes, editor submission, Git redeploy, start/stop, and deletion controls.
- Git source details and redeploy webhook identifiers are omitted from viewer responses so a viewer cannot trigger the public webhook as a side channel.
- Compose files and stored environment values can themselves contain credentials. Grant viewer access only when the recipient is trusted to see the complete stack definition.

## API payload

Resource-control create and update requests accept two additional stack-only fields:

```json
{
  "users": [1],
  "teams": [2],
  "readOnlyUsers": [3],
  "readOnlyTeams": [4],
  "public": false,
  "administratorsOnly": false
}
```

A user or team cannot appear in both the read-write and read-only lists. Only administrators can change viewer grants; existing non-administrator owners may continue making ownership changes only while the viewer lists remain unchanged.

## Verification

```sh
GOCACHE=/tmp/portainer-go-cache go test \
  ./pkg/authorization \
  ./api/internal/authorization \
  ./api/http/security \
  ./api/http/handler/resourcecontrols \
  ./api/http/handler/stacks

pnpm typecheck
pnpm exec vitest run \
  app/react/portainer/access-control/utils.test.ts \
  app/react/portainer/access-control/AccessControlForm/AccessControlForm.test.tsx \
  app/react/portainer/access-control/AccessControlPanel/AccessControlPaneDetails.test.tsx \
  app/react/docker/stacks/ItemView/StackEditorTab/StackEditorTabInner.test.tsx
```
