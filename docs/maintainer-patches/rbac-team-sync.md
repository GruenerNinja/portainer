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

Run the focused checks after an upstream merge:

```sh
go test ./api/internal/authorization ./api/datastore/migrator ./pkg/authorization \
  ./api/http/handler/teammemberships ./api/http/handler/teams ./api/http/handler/users
npm run typecheck
NODE_OPTIONS=--localstorage-file=/tmp/portainer-vitest-localstorage npm test -- --run \
  app/react/portainer/access-control/AccessManagement/AccessDatatable/AccessDatatable.test.tsx \
  app/react/portainer/environments/environment-groups/ItemView/tabs/AccessTab.test.tsx \
  app/react/portainer/users/teams/ItemView/TeamAssociationSelector/TeamMembersList/TeamMembersList.test.tsx \
  app/react/sidebar/SettingsSidebar.test.tsx
```
