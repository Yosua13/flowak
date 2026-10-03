# Release checklist

- [x] Clean branch passes backend unit/integration/security/resilience tests.
- [x] Clean branch passes frontend component/accessibility/resilience tests, lint, and production build.
- [x] `docs/openapi.v1.json` passes `scripts/quality-gate.ps1 -Gate contract`.
- [x] CI artifacts are linked in the release pull request.
- [x] A non-production migration rehearsal report is recorded; no production database was used.
- [x] Backup, restore, secret rotation, runner allowlist, incident, and retention steps were reviewed in the runbook.
- [x] Graph reconciliation output was reviewed twice for the migrated fixture scope.
- [x] API-runner response bodies and all credentials are absent from release evidence.
- [x] Residual risks and execution authorization are recorded in `RELEASE-EVIDENCE-2026-10-03.md`.

## Exit criteria and rollout controls

| Exit criterion | Evidence required | Status |
| --- | --- | --- |
| Graph child data preserved | graph handler tests and shadow mismatch report | CI gate |
| Tenant authorization | auth/IDOR matrix | CI gate |
| Work item consistency | Kanban, calendar, analytics, document fixture | CI gate |
| Runner secret safety | SSRF and redaction suite | CI gate |
| One graph source of truth | normalized read, two shadow passes, compatibility write disabled by default | verified |
| Documentation current | README, runbook, API contract | reviewed |

Current rollout controls are the rollback-only `GRAPH_COMPATIBILITY_WRITE_ENABLED` flag, AI advisory behavior, API-runner environment approval, and collaboration SSE. There is no centralized feature-flag service: each release records an owner, rollback condition, and release note. The ignored local SQLite artifact was archived outside the repository with a recorded checksum.
