# Non-production migration rehearsal

This rehearsal proves that a copy of a database can be migrated without record loss. It is intentionally blocked for production-looking connection strings and must be run by an operator who controls disposable PostgreSQL instances.

## Procedure

1. Create a disposable source and target database with names containing `rehearsal`, `staging`, `test`, or `dev`.
2. Set `FLOWAK_REHEARSAL_SOURCE_DATABASE_URL` and `FLOWAK_REHEARSAL_DATABASE_URL` to those databases.
3. Run `pwsh ./scripts/rehearse-migrations.ps1`.
4. Attach the generated report and backup manifest to the PR. Compare source/target row counts, schema checksum, and the graph reconciliation report.
5. Test rollback by restoring the generated backup into a new disposable database; do not rollback by editing old migrations.

The script checks every native command exit code and stops on SQL, dump, or restore failure. A successful report includes the backup checksum, schema checksum, migration count, and per-table row-count snapshot. The completed 2026-10-03 rehearsal is recorded in `RELEASE-EVIDENCE-2026-10-03.md`.

## Feature flags and residual risk

There is no application feature-flag service in the current runtime. Rollout controls are deployment revision, database backup, API-runner environment approval, and read-only graph reconciliation. The release owner must explicitly record this limitation and sign the residual-risk section in `RELEASE-EVIDENCE.md`.
