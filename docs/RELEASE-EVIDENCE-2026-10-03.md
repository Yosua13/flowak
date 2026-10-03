# Release evidence — 2026-10-03

This record contains only redacted, non-production identifiers. No connection string, credential, API body, or secret is included.

## Quality gates

| Gate | Evidence | Result |
| --- | --- | --- |
| Backend unit/integration/security/resilience | `cd backend; go test ./...` | Passed |
| Frontend contract/accessibility/resilience | `cd frontend; npm test` | Passed: 17 files, 41 tests |
| Frontend type/build | `npm run lint`; `npm run build` | Passed; existing bundle-size warning accepted |
| OpenAPI contract | `pwsh ./scripts/quality-gate.ps1 -Gate contract` | Passed |
| Release documentation | `pwsh ./scripts/quality-gate.ps1 -Gate release` | Passed |
| Migration rehearsal | `pwsh ./scripts/rehearse-migrations.ps1` | Passed against distinct disposable PostgreSQL 18 source and target databases |

## Migration and rollback rehearsal

- Environment: local disposable PostgreSQL 18 cluster; source and target names contained `rehearsal`; production was not accessed.
- Fixture: 1 project, 1 module, 2 normalized nodes, 1 edge, 1 work item, 1 status-history row, 1 comment, and 1 activity row.
- Backup SHA-256: `7B0E4899ABA54E4ACE643C47170797CFE0D03AC19F90DB02D59915FACB37C8AC`.
- Schema checksum after restore/migration: `9949842690d7e0d5655d1657ce7efa69`.
- All 14 forward-only migrations replayed in lexical order.
- Source and restored-target row counts matched across all 38 public tables.
- Rollback procedure was verified by restoring the backup into a distinct disposable target before replaying migrations.

## Shadow reconciliation

The release candidate was reconciled twice, read-only, against the restored target:

| Pass | Module | Nodes snapshot/normalized | Edges snapshot/normalized | Result |
| --- | --- | --- | --- | --- |
| 1 | `module_rehearsal` | 2 / 2 | 1 / 1 | `matches: true` |
| 2 | `module_rehearsal` | 2 / 2 | 1 / 1 | `matches: true` |

Both reports had empty missing, changed, and snapshot-issue arrays. Compatibility writes now default off. Set `GRAPH_COMPATIBILITY_WRITE_ENABLED=true` only as a documented rollback control; normalized tables remain authoritative and the legacy read fallback remains for one release.

## SQLite disposition and residual risk

- The ignored 32 KiB local SQLite artifact was archived outside the repository under `%LOCALAPPDATA%\Flowak\archives\flowak-20261003.db`.
- Archive SHA-256: `936119E589FA449D166A37378BB6D42AE4AD013B4384760594E5AF30522D9116`.
- PostgreSQL remains the only runtime database.
- Legacy snapshot columns are retained for the documented one-release read fallback and must be removed only in a later major forward migration.
- The frontend production bundle still emits the existing chunk-size advisory; correctness gates pass.

## Authorization and rollback

- Execution authorization: repository-owner instruction in task #33/#58/#22 on 2026-10-03.
- Rollback: enable the compatibility-write flag only if required, restore the verified PostgreSQL backup to a new database, validate row counts and reconciliation, then repoint the deployment.
- Secret rotation, incident response, retention, and runner allowlist procedures are defined in `RUNBOOK.md`.
