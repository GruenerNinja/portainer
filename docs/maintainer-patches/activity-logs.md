# User activity logs

The Community Edition implementation records authorized, state-changing API calls (`POST`, `PUT`, `PATCH`, and `DELETE`). It stores timestamp, username, method, and path. Request bodies, authorization headers, cookies, registry passwords, Git credentials, and Vault secrets are never captured.

Entries live in the BoltDB `activity_logs` bucket and are pruned to seven days when a new entry is written. Database export/import includes the retained entries.

## Code ownership

- `api/dataservices/activitylog/`: persistent CRUD service and retention.
- `api/http/security/bouncer.go`: records mutations after authentication/authorization context has been established.
- `api/http/handler/useractivity/`: administrator-only JSON and CSV endpoints.
- `api/http/server.go` and `api/http/handler/handler.go`: route wiring.
- `app/react/portainer/logs/ActivityLogsView/`: live CE data instead of teaser rows/overlay.

Endpoints:

- `GET /api/useractivity/logs`
- `GET /api/useractivity/logs.csv`

Supported query parameters are `offset`, `limit`, `sortBy`, `sortDesc`, `keyword`, `after`, and `before`. Keep response field names lower-camel because the existing React table consumes that contract.

When extending the payload, treat it as security-sensitive: add metadata explicitly and never serialize the request body.
