# TheModCraft feature planning backlog

This is the discovery backlog for the planning window from **September 13 through October 12, 2026**. The current `2.39.3.2.24` release is the stabilization baseline. Items in this document are ideas to investigate, not scheduled implementation work or release commitments.

## Planning goals

1. Collect operational problems before choosing solutions.
2. Prefer features that are meaningfully different from documented Portainer Community and Business functionality.
3. Extend the fork's strongest areas: agentic accounts, least-privilege access, GitOps, Vault, auditability, and realtime operation.
4. Preserve upstream API and datastore compatibility unless a documented migration provides a clear benefit.
5. Require an authorization model, threat model, rollback path, and maintenance plan before implementation begins.

## Candidate backlog

| ID | Candidate | Problem it should solve | Smallest useful version | Important questions and risks | Initial priority |
| --- | --------- | ----------------------- | ----------------------- | ----------------------------- | ---------------- |
| TMC-P01 | Agentic Operations Guard | Agentic and service accounts need useful automation access without receiving broad, permanent human permissions. | Short-lived, stack-scoped capability leases; allowed actions; expiry; rate and mutation budgets; automatic denial of privileged, host-level, or cross-owner changes. | Capability revocation, token storage, replay resistance, API compatibility, emergency override, and avoiding a second inconsistent RBAC system. | Highest |
| TMC-P02 | Mutation plan and approval workflow | Operators cannot see the full impact of a destructive action or require a second person to approve it. | A dry-run plan for stack updates and deletion, policy outcomes of allow/deny/approval-required, an approval inbox, expiry, and complete activity-log records. | Docker has limited native dry-run support; plans must be explicitly marked when some runtime effects are estimates. Approvals must be race-safe and bound to an immutable request digest. | Highest |
| TMC-P03 | Three-way GitOps drift inspector | Commit-hash checks do not explain differences between Git, the last successful deployment, and live runtime state. | A stack page that compares desired Compose, last-applied configuration, and live resources, then classifies image, environment, mount, network, scale, and ownership drift. | Redaction of secrets, Compose normalization, resources changed outside Portainer, and avoiding false-positive drift from runtime defaults. | Highest |
| TMC-P04 | Permission what-if simulator | Current effective access does not answer why an action is allowed or what a proposed role/group change would do. | Select a user or Authentik group, resource, and action; display an allow/deny explanation and simulate unsaved team or role mappings. | Explanations must use the same authorization engine as enforcement so they cannot silently disagree. | High |
| TMC-P05 | Docker team quotas and guardrails | Docker teams and agentic accounts can be ownership-scoped but do not have Kubernetes-style resource budgets. | Per-team limits for running containers, requested CPU/RAM, published ports, volume usage, and allowed registries, enforced on Portainer mutations. | Existing workloads, direct Docker-socket changes, accurate storage accounting, and the behavior when a team is already over quota. | High |
| TMC-P06 | Vault secret dependency and rotation graph | Operators cannot easily determine which stacks depend on a Vault path or what must restart after rotation. | Redacted dependency graph from Vault paths to workflows/stacks/services, impact preview, rotation status, and controlled redeploy actions. | Never persist or emit secret values; handle dynamic leases and aliases; require strict authorization on metadata because names can also be sensitive. | High |
| TMC-P07 | Stack leases and automatic cleanup | Preview, test, and agent-created stacks are easy to leave running indefinitely. | Optional expiry time, owner/team, warning period, stop-before-delete policy, renewal, and cleanup audit event. | Protect persistent volumes, Git-managed production stacks, dependencies, and recently renewed leases. | Medium |
| TMC-P08 | Incident freeze mode | During an incident, automation and ordinary users can continue changing the affected workload while responders investigate. | Time-bounded stack mutation lock, named incident, audited break-glass override, and a captured bundle of configuration metadata, recent events, logs, and metrics. | Avoid locking out recovery operations; bound captured data; redact credentials and sensitive logs. | Medium |
| TMC-P09 | Cross-platform deployment checkpoints | Compose editor history and platform-specific rollback do not provide a single restorable checkpoint around every mutation. | Save Compose, environment-variable metadata, image digests, ownership, source revision, and runtime inventory before supported stack mutations. | Checkpoints cannot guarantee rollback of external state or persistent data; the UI must state exactly what can be restored. | Medium |
| TMC-P10 | React-first frontend modernization | The UI still boots through AngularJS and mounts React through many adapters, increasing bundle size, framework coupling, and the cost of maintaining routes and shared state. | Establish bundle and navigation baselines, prohibit new AngularJS code, lazy-load one low-risk route family, and write a vertical-slice migration playbook. | A broad rewrite would be risky and provide weak short-term value. Migrations must preserve authorization behavior, URLs, browser history, and upstream mergeability. | High |

The strongest product direction is to combine **TMC-P01**, **TMC-P02**, and **TMC-P04** as a coherent **Safe Agentic Operations** capability:

1. An account requests an operation using a short-lived capability.
2. Portainer evaluates real authorization and produces an explainable plan.
3. Policy allows, denies, or requests human approval.
4. A checkpoint is captured when the operation supports restoration.
5. Execution status is streamed over the existing authenticated WebSocket infrastructure.
6. The request, decision, approval, execution, and result are written to the activity log.

## Frontend modernization track

The September 12, 2026 source snapshot is already predominantly React and TypeScript by non-generated production source volume, but the application runtime is still AngularJS-first:

| Signal | Current snapshot | Meaning |
| ------ | ---------------- | ------- |
| React implementation | About 150,000 lines in `app/react` across 2,577 TypeScript/TSX files | Most feature implementation is now in the React tree. |
| Legacy implementation | About 43,100 lines across 494 JavaScript and 136 HTML files | A meaningful AngularJS controller, service, template, and routing surface remains. |
| Source-volume estimate | Approximately 78% React/TypeScript and 22% JavaScript/AngularJS templates | This measures source volume, not screens, runtime cost, or migration completeness. |
| Application root | `app/index.html` declares `ng-app="portainer"`, and `app/index.js` creates the AngularJS root module | AngularJS still owns application startup and the hybrid shell. |
| React bridge | 168 `r2a(...)` registrations and 135 `withUIRouter(...)` uses | Many React components still depend on AngularJS lifecycle and hybrid-router integration. |
| Route registration | 163 JavaScript and 21 TypeScript `$stateRegistryProvider.register(...)` calls | Route ownership remains heavily centered in the AngularJS configuration layer, even when a route renders React. |
| Loading strategy | Webpack creates initial `vendor` and `main` chunks; no production route-level dynamic imports were found | Users download a large hybrid application before the first view can become useful. |

The initial proposal is an incremental **strangler migration**, not a frontend rewrite:

1. Capture reproducible bundle, cold-load, warm-navigation, memory, and render baselines for representative Docker, Kubernetes, administration, and stack routes.
2. Add guardrails that reject new AngularJS controllers, services, templates, and bridge registrations unless explicitly approved for an upstream merge.
3. Introduce route-level lazy loading and split vendors by stable cache lifetime before converting large numbers of components.
4. Select one low-risk route family and move its route definition, data access, authorization checks, error boundaries, and layout into a React-owned vertical slice.
5. Move shared state toward typed API clients and query caches. WebSocket events should invalidate or update those caches rather than creating a separate live-data state model.
6. Reduce bridge roots and digest-boundary crossings, then remove AngularJS services only after their callers have migrated.
7. Remove AngularJS, its router bridge, and compatibility dependencies only after the final route and startup shell have moved.

Success should be measured by shipped JavaScript and navigation latency, not converted-file count. The discovery exit criteria are a reviewed architecture decision, a bundle report, performance budgets, one representative migration design, compatibility tests, and an estimate of recurring upstream-merge cost.

## One-month discovery plan

| Week | Focus | Expected output |
| ---- | ----- | --------------- |
| Sep 13–19 | Observation and collection | Record actual operator, team, and agentic-account pain points; do not change production behavior. |
| Sep 20–26 | Product comparison and grouping | Check candidate overlap with current upstream CE/Business releases and combine duplicate ideas. |
| Sep 27–Oct 3 | Security and architecture | Threat-model the leading candidates, identify authorization and datastore boundaries, and estimate upstream-merge cost. Read-only prototypes are allowed; no production rollout. |
| Oct 4–12 | Prioritization | Score the candidates, choose at most one implementation theme, and write its acceptance criteria, migration, test, release, and rollback plan. |

## Scoring rubric

Score each dimension from 1 to 5. Record the evidence behind the score rather than relying on the total alone.

| Dimension | A high score means |
| --------- | ------------------ |
| User value | It removes a frequent or high-risk operational problem. |
| Differentiation | The capability is not already documented in Portainer CE or Business. |
| Fork fit | It reuses and strengthens existing TheModCraft features. |
| Security improvement | It reduces authority, exposure, ambiguity, or recovery time. |
| Technical confidence | The behavior can be implemented and tested reliably across supported environments. |
| Maintenance cost | Reverse-scored: a high score means low recurring merge and support cost. |

## Planning-ready gate

A candidate is ready to leave discovery only when it has:

- a concrete user and problem statement;
- explicit goals and non-goals;
- a comparison against the current upstream release;
- authorization rules and a threat model;
- data model, API, and UI boundaries;
- compatibility and migration behavior;
- failure, rollback, and disaster-recovery behavior;
- backend, frontend, security, and end-to-end test plans;
- an upstream-merge maintenance estimate;
- measurable acceptance criteria.

## Feature intake template

Add new ideas to the candidate table and capture the following before prioritizing them:

```text
Name:
Submitted:
Problem observed:
Affected users/teams:
Current workaround:
Frequency and impact:
Proposed smallest useful behavior:
Security considerations:
Possible overlap with Portainer CE/Business:
Evidence or example:
Open questions:
```

## Comparison baseline

Recheck these official sources during the final prioritization week because upstream functionality changes frequently:

- [Portainer roles and effective access](https://docs.portainer.io/admin/user/roles)
- [Portainer access control](https://docs.portainer.io/advanced/access-control)
- [Portainer GitOps update behavior](https://docs.portainer.io/faqs/troubleshooting/stacks-deployments-and-updates/how-do-automatic-updates-for-stacks-applications-work)
- [Portainer fleet governance policies](https://docs.portainer.io/admin/environments/policies)
- [Portainer Kubernetes namespace quotas](https://docs.portainer.io/user/kubernetes/namespaces/add)
- [Portainer account and API tokens](https://docs.portainer.io/user/account-settings)
- [Portainer current feature overview](https://docs.portainer.io/whats-new)
