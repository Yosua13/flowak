# Graph reconciliation

This command reports differences between the legacy `modules.nodes` and
`modules.edges` JSON snapshots and active normalized workflow rows. It does not
write, repair, delete, or run migrations; normalized rows are the source of
truth.

Run it from `backend` with one scope:

```powershell
go run ./cmd/graph-reconcile --module-id module_123
go run ./cmd/graph-reconcile --project-id project_123
```

The result is JSON with counts, missing IDs, changed key graph fields, and
invalid snapshot diagnostics. Use a database account with read-only access for
operations. A mismatch is evidence for a controlled recovery runbook, not an
instruction for this command to mutate production data.

For an incident, retain the report and affected snapshot before starting a
separate repair. Perform any approved repair in its own transaction, validate
with this command before commit, and roll that repair transaction back if the
report is not as expected. This command itself has no state to roll back.
