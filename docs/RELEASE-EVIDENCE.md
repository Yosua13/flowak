# Release evidence template

| Gate | Command or workflow URL | Result | Owner/date |
|---|---|---|---|
| Backend unit/integration/security/resilience | `go test ./...` | Pending | |
| Frontend test/accessibility/resilience | `npm test` | Pending | |
| Frontend type/build/performance budget review | `npm run lint`, `npm run build` | Pending | |
| OpenAPI contract | `pwsh ./scripts/quality-gate.ps1 -Gate contract` | Pending | |
| Migration rehearsal | `pwsh ./scripts/rehearse-migrations.ps1` | Pending | |
| Graph reconciliation | `go run ./cmd/graph-reconcile ...` | Pending | |

## Rehearsal record

- Source and target are non-production: `Pending`
- Backup manifest/checksum: `Pending`
- Row-count comparison: `Pending`
- Reconciliation result: `Pending`
- Restore-to-new-copy result: `Pending`

## Residual risk and sign-off

Record only identifiers and redacted summaries. Do not paste tokens, cookies, API bodies, or database connection strings.

- Accepted residual risk:
- Mitigation/rollback revision:
- Release owner:
- Date/time (UTC):
