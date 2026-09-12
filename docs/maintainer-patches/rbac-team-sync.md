# Environment roles and synchronized team leaders

This fork makes the complete built-in environment role set available without changing the reported Portainer edition:

| ID | Role                      | Intended access                                      |
| -- | ------------------------- | ---------------------------------------------------- |
| 1  | Environment administrator | Full control of every resource in the environment    |
| 2  | Helpdesk                  | Read-only access to every resource                   |
| 3  | Standard user             | Full control of assigned resources                   |
| 4  | Read-only user            | Read-only access to assigned resources               |
| 5  | Operator                  | Inspect and operate existing resources without CRUD  |

The numeric IDs are persisted in user and team access policies and must remain stable. `restorePredefinedRoles` runs as a `2.45.0` migration to add the formerly absent Operator record and repair all predefined names, priorities, and authorization maps. The restrictive precedence order is Environment administrator (1), Operator (2), Helpdesk (3), Standard user (4), and Read-only user (5).

The UI intentionally enables role creation, editing, the role column, and the effective-access viewer without changing the global edition flag. Keep both the legacy environment access screen and the React environment-group access screen aligned when upstream changes either implementation.

When LDAP team sync is enabled, membership add/remove controls remain disabled because the identity provider owns membership. A Portainer administrator can still promote a synchronized member to team leader or demote a leader, and a synchronized team leader retains the Users and Teams navigation needed to exercise that role. LDAP synchronization does not overwrite the role on an existing membership.

## Authentik OAuth role synchronization

OAuth/OIDC team synchronization uses a trusted claim from the configured UserInfo resource endpoint. It deliberately does not use claims copied only from an unverified ID token for authorization decisions.

1. In Authentik, expose a `groups` claim from the Portainer provider. The claim should be a string or an array of group names and must be available from the provider's UserInfo endpoint.
2. In **Settings > Authentication > OAuth > Team membership**, enable **Automatic team membership** and set the claim name to `groups`.
3. Add anchored regular-expression mappings, for example `^authentik Admins$` to the Portainer `Admin` team and `^portainer-agentic$` to the `Agentic` team.
4. Grant each Portainer team the intended environment or environment-group role. Authentik owns who belongs to a team; Portainer continues to own what that team may do.
5. Have the affected user sign out and sign in again. Synchronization runs atomically during every OAuth login. Existing team-leader status is preserved while a claim remains mapped, memberships for currently configured mappings are removed when their claim disappears, and unrelated local memberships are left untouched.

The optional default team is assigned only when none of the configured claim mappings match. Invalid regular expressions and mappings to deleted teams are rejected when authentication settings are saved.

Run the focused checks after an upstream merge:

```sh
go test ./api/internal/authorization ./api/datastore/migrator ./pkg/authorization \
  ./api/http/handler/auth ./api/http/handler/settings ./api/oauth \
  ./api/http/handler/teammemberships ./api/http/handler/teams ./api/http/handler/users
npm run typecheck
NODE_OPTIONS=--localstorage-file=/tmp/portainer-vitest-localstorage npm test -- --run \
  app/react/portainer/access-control/AccessManagement/AccessDatatable/AccessDatatable.test.tsx \
  app/react/portainer/environments/environment-groups/ItemView/tabs/AccessTab.test.tsx \
  app/react/portainer/users/teams/ItemView/TeamAssociationSelector/TeamMembersList/TeamMembersList.test.tsx \
  app/react/sidebar/SettingsSidebar.test.tsx
```
