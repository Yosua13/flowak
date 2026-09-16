# Release Checklist

| Exit criterion | Evidence required | Status |
| --- | --- | --- |
| Graph child data preserved | graph handler tests and shadow mismatch report | CI gate |
| Tenant authorization | auth/IDOR matrix | CI gate |
| Work item consistency | Kanban, calendar, analytics, document fixture | CI gate |
| Runner secret safety | SSRF and redaction suite | CI gate |
| One graph source of truth | normalized read plus compatibility protocol | monitored |
| Documentation current | README, runbook, API contract | reviewed |

Feature flags: graph compatibility write, AI advisory, API runner, collaboration SSE. Each flag needs an owner, rollback condition, and release note. The SQLite artifact is retained until the data owner approves archival or deletion.
