# Release checklist

- [ ] Clean branch passes backend unit/integration/security/resilience tests.
- [ ] Clean branch passes frontend component/accessibility/resilience tests, lint, and production build.
- [ ] `docs/openapi.v1.json` passes `scripts/quality-gate.ps1 -Gate contract`.
- [ ] CI artifacts are linked in the PR.
- [ ] A non-production migration rehearsal report is attached; no production database was used.
- [ ] Backup, restore, secret rotation, runner allowlist, incident, and retention steps were reviewed in the runbook.
- [ ] Graph reconciliation output is reviewed for every migrated project/module in scope.
- [ ] API-runner response bodies and all credentials are absent from release evidence.
- [ ] Residual risks and release-owner approval are recorded in `RELEASE-EVIDENCE.md`.

## Exit criteria and rollout controls

| Exit criterion | Evidence required | Status |
| --- | --- | --- |
| Graph child data preserved | graph handler tests and shadow mismatch report | CI gate |
| Tenant authorization | auth/IDOR matrix | CI gate |
| Work item consistency | Kanban, calendar, analytics, document fixture | CI gate |
| Runner secret safety | SSRF and redaction suite | CI gate |
| One graph source of truth | normalized read plus compatibility protocol | monitored |
| Documentation current | README, runbook, API contract | reviewed |

Current rollout controls are graph compatibility write, AI advisory behavior, API-runner environment approval, and collaboration SSE. There is no centralized feature-flag service: each release must record an owner, rollback condition, and release note in the evidence template. The ignored local SQLite artifact is not a supported runtime database and remains untouched pending explicit data-owner archival approval.
