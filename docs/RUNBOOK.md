# Flowak Production Runbook

## Runtime

Production has one runtime: Go/Gin serves `/api` and compiled Vite assets. PostgreSQL is the only runtime database. `modules.nodes` and `modules.edges` remain a deprecated compatibility snapshot until two shadow releases show no mismatch; normalized workflow tables are authoritative.

## Deploy and rollback

1. Take a PostgreSQL backup and record row counts for `workflow_nodes`, `workflow_edges`, `work_items`, `comments`, and `activity_logs`.
2. Run `go test ./...`, frontend lint/test/build, then apply forward-only migrations on a staging copy.
3. Run the graph mismatch helper and compare row counts/checksums before enabling the release.
4. Roll back application binaries only. Do not roll back a migration; restore the database backup only through an approved incident procedure.

## Security and incidents

Rotate `JWT_SECRET`, database credentials, and API-runner encryption keys through the deployment secret store. Never place values in `.env.example`, graph metadata, exports, events, or logs. Disable the API runner immediately when an allowlist/SSRF incident is suspected; preserve redacted activity metadata and notify the tenant owner.

## Retention

Keep activity, notification/outbox, version, and API-run evidence according to the tenant retention policy. Purge only expired records through an audited maintenance job; verify backups before a purge.
