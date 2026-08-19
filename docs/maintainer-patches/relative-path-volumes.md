# Relative-path volume deployments

Relative-path stacks require their Git working tree and user-managed files on the target environment. The remote deployer must continue to set:

- `--flat` and `-k`;
- `--source-dir` when the compose file is below the repository root;
- `--deployment-dir` and `--cleanup-deployment-files` when configured;
- the compose destination and original entry-point path.

Repull, prune, and force-recreate flags are additive. They must never replace or reorder away the path arguments. See `api/stacks/deployments/deployer_remote.go`, `relative_path_config.go`, and `compose_unpacker_cmd_builder.go`.

Run the deployment package tests after any GitOps, stack update, or compose-unpacker command change.
