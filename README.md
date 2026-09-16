# Flowak

Flowak is a traceability-native SaaS workspace for business flow, specifications, work items, evidence, and controlled API contracts. It is not a code generator: AI can propose a draft flow or audit findings for human review.

## Production architecture

One Go/Gin runtime serves the API and compiled Vite application. PostgreSQL is the only runtime database. Normalized workflow tables are authoritative; the legacy module JSON snapshot is compatibility-only during the documented migration window.

## Development

1. Configure PostgreSQL and copy safe placeholder values into `backend/.env`.
2. Run `go run main.go` from `backend`.
3. Run `npm install` then `npm run dev` from `frontend`.

Verification: `go test ./...` in `backend`; `npm run lint`, `npm test`, and `npm run build` in `frontend`.

See [production runbook](docs/RUNBOOK.md) and [release checklist](docs/RELEASE-CHECKLIST.md) for migration, rollback, secret rotation, API-runner allowlist, retention, and incident procedures.
