# GitOps cleanup, re-pull, and force redeploy

`LoadWorkflowMap` bulk-loads only workflows that still exist. A stale stack/source reference therefore no longer makes the complete GitOps listing fail. The normal detach/cleanup path can remove the orphan later.

The UI's **Re-pull image and redeploy** option is one combined guarantee:

- pull the current image;
- prune orphaned resources when selected;
- force recreation even when the tag or service spec did not change.

Local Swarm passes `PullImage` and `ForceRecreate` to libstack. Remote Swarm passes `-f` and `--force-recreate` to compose-unpacker. Scheduled GitOps with `ForceUpdate` now also forwards force recreation for Compose.

## Regression tests

Run:

```sh
GOCACHE=/tmp/portainer-maintainer-go-cache go test ./api/gitops/workflows ./api/exec ./api/stacks/deployments -count=1
```

The key tests cover a missing workflow ID, local Swarm option forwarding, and the combination of remote relative paths with pull/prune/force flags.
