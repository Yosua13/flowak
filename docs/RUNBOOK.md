# Flowak operations runbook

## Preconditions

Use PostgreSQL only. Set `APP_ENV=production`, a unique `JWT_SECRET` of at least 32 random characters, a non-default database password, and explicit `ALLOWED_ORIGINS`. Do not commit `.env` files or place real secrets in graph/specification fields.

## Backup and restore

Before migration or incident work, take a logical backup from the target environment:

```powershell
pg_dump --format=custom --file flowak-before-change.dump $env:FLOWAK_DATABASE_URL
pg_restore --list flowak-before-change.dump
```

Restore only to a new non-production database first. Verify table counts and normalized-graph reconciliation before switching traffic. Never run `scripts/rehearse-migrations.ps1` against production.

## Migration and rollback

Migrations are forward-only and idempotent. The runtime applies them at startup. For a failed rollout, stop traffic, restore the verified backup to a replacement database, set the deployment's database URL to that replacement, and redeploy the prior image. Do not edit or rerun old migration files to force a rollback.

`modules.nodes`/`edges` are compatibility snapshots only. Use `backend/cmd/graph-reconcile` to produce a read-only mismatch report before choosing a rollback or repair plan.

Compatibility writes are disabled by default after the two successful 2026-10-03 shadow reconciliation passes. `GRAPH_COMPATIBILITY_WRITE_ENABLED=true` is a temporary rollback control only. Keep the legacy read fallback for one release; remove the columns only through a later major forward migration.

## Secret rotation

Rotate database credentials and runner environment variables in the secret store, deploy with the new values, verify health and login, then revoke the previous values. Rotating `JWT_SECRET` invalidates existing access sessions; schedule it as an incident or maintenance operation and communicate a re-login requirement. Never include a secret value in tickets, event payloads, exports, or incident logs.

## Incident response

1. Preserve deployment revision, UTC timestamps, request IDs, and redacted logs.
2. Contain: disable affected API-runner environments or revoke sessions/credentials as needed.
3. Assess tenant scope using project IDs, not user-supplied client state.
4. Restore a non-production copy first, reconcile normalized graph data, and document the decision.
5. Record cause, impact, remediation, and residual risk in the release evidence template.

## Retention

Startup pruning removes expired event/notification outbox records, activity older than its expiry, and expired non-evidence API runs. Baselines and evidence are intentionally retained until an approved policy exists. Monitor database growth and approve any change to their lifecycle explicitly.
