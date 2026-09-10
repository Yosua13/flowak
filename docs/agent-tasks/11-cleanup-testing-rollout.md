# Tugas 11: Cleanup, Dokumentasi, Pengujian Sistem, dan Rollout

## Tujuan

Menutup transisi arsitektur, menghapus perilaku yang bertentangan dengan fokus produk, dan memastikan migrasi dapat dirilis tanpa kehilangan data.

## Prasyarat

Tugas 01 sampai 10 sudah digabung dan feature flag tersedia untuk area berisiko.

## Cakupan file utama

- `frontend/server.ts`
- `backend/handlers/ai.go`
- `backend/README.md`
- `frontend/README.md`
- Root README bila ada
- Seluruh test/config yang diperlukan
- Dokumen PRD/SRS/ERD/OpenAPI/runbook baru di `docs/`

## Cleanup wajib

1. Hapus code generation dari UI dan route produksi. AI hanya boleh membuat draft specification, audit finding, atau task suggestion yang memerlukan review manusia.
2. Deprecate lalu hapus Express standalone setelah seluruh route produksi dilayani Go. Jangan menjalankan dua backend dengan perilaku berbeda.
3. Arsipkan atau hapus referensi SQLite `backend/flowak.db` setelah pemilik data dikonfirmasi; PostgreSQL menjadi satu-satunya runtime database.
4. Setelah minimal dua release shadow verification tanpa mismatch, hentikan compatibility write `modules.nodes/edges`, pertahankan read fallback satu release, lalu drop pada major migration terpisah.
5. Hapus seed/demo data dari flow produksi atau tandai eksplisit sebagai sample workspace.
6. Hapus copy `Online`, `Real-time Active`, mock API response, wireframe palsu, dan metric heuristic lama.

## Dokumentasi wajib

- PRD: scope baru task tracking dan API runner terbatas, nilai jual traceability-native.
- SRS: server-first multi-tenant, domain WorkItem/FacetAssignment/Comment/Evidence/Environment/APIRequest/APIRun.
- ERD: relasi, constraints, indexes, delete/retention lifecycle.
- OpenAPI: seluruh endpoint versioned dan error contract.
- Runbook: backup/restore, migration, secret rotation, allowlist runner, incident response, retention.
- README: satu runtime produksi, setup aktual, command test, environment variables tanpa secret contoh nyata.

## Test gate wajib

1. Unit: graph invariant, hierarchy cycle, transitions, completeness, SLA, redaction.
2. Integration: tenant isolation, permission, graph preservation, pagination, sequence, outbox.
3. Contract: OpenAPI lint dan request/response fixtures.
4. E2E: invite-login-project-graph-spec-task-comment-Kanban-API run-document publish.
5. Security: IDOR, auth matrix, SSRF, secret leakage, XSS comment, upload validation, rate limit.
6. Performance: 150-node canvas, 500 work-item board, concurrent autosave, search latency, runner limits.
7. Accessibility: WCAG 2.2 AA target, modal focus, drag alternative, labels, error announcement, reduced motion.
8. Resilience: offline save failure, 409 conflict, SSE reconnect, partial runner failure, retry without duplication.

## Strategi rollout

- Gelombang Foundation: database safety, tenancy, auth, typed client.
- Gelombang Traceability Work: specs, work items, comments, node detail, Kanban.
- Gelombang Contract Validation: API contract dan secure runner.
- Gelombang SaaS Collaboration: versions, events, notifications, integrations.
- Gunakan backup, row counts, checksum, shadow reads, feature flags, reconciliation report, dan rollback plan per gelombang.

## Exit criteria

- Tidak ada kehilangan child data saat graph update.
- Semua endpoint tenant lulus authorization matrix.
- Work item dapat dibuat dari node dan konsisten di Kanban, kalender, analitik, dan dokumen.
- API runner lulus security suite dan tidak membocorkan secret.
- Tidak ada dua source of truth aktif tanpa protocol migrasi tertulis.
- `go test ./...`, frontend lint/build/test, contract, E2E, security, dan migration rehearsal lulus di CI.
- PRD, SRS, ERD, OpenAPI, README, dan runbook sesuai perilaku aktual.

## Laporan akhir

Sertakan tabel status setiap exit criterion, bukti command CI, hasil reconciliation database, daftar feature flag, langkah rollback, dan daftar risiko residual yang diterima product owner.

