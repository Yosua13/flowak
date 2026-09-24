# Audit Delivery Issue #22-#33

Audit ini membandingkan issue GitHub yang masih terbuka dengan implementasi pada branch `develop` pada 17 September 2026. Status `partial` berarti fondasi atau sebagian UI sudah ada, tetapi acceptance criteria issue belum dapat diklaim selesai.

| Issue | Status | Bukti implementasi | Gap yang harus ditutup |
| --- | --- | --- | --- |
| #22 Epic Traceability Workspace | Partial | Migrasi normalized graph, work item, specs, Kanban, runner, derived views, dan collaboration sudah ada. | Epic tidak boleh ditutup sebelum semua gap #23-#33 dan test gate release tertutup. |
| #23 Graph safety | Largely implemented | `05_workflow_graph_safety.sql`, `module_graph.go`, dan test conflict/preservation. | Tambahkan command admin read-only untuk reconciliation snapshot JSON versus normalized graph serta integration test database nyata. |
| #24 Tenant auth and RBAC | Largely implemented | `06_tenant_auth_rbac.sql`, middleware capability, session dan invitation handler. | Lengkapi integration test semua endpoint tenant dan rotasi/revoke session pada skenario produksi. |
| #25 Frontend state foundation | Partial | Typed `apiClient`, query cache, graph optimistic rollback, dan save state tersedia. | `useStore.ts` masih menjadi store lintas domain besar; pisahkan auth, projects, graph, UI, dan server state secara bertahap. |
| #26 Work items, evidence, and collaboration | Partial | `07_work_items.sql`, work-item CRUD, hierarchy validation, status history, dan comments. | Tambahkan API/UI untuk attachment metadata, evidence, checklist, links, dan watcher. Tambahkan sort eksplisit pada list. |
| #27 Cross-discipline specifications | Partial | `08_domain_specifications.sql`, typed facet persistence, dan tab Bisnis/UIUX/Frontend/Backend. | Hapus input secret mentah dari Backend tab; lengkapi validator completeness dan repeatable-row UX untuk semua facet. |
| #28 Node detail surface | Partial | `NodeDetailSurface.tsx` dan inspector task/comment data. | Lengkapi full-page detail/link, activity tab, focus-return, dan viewport accessibility verification. |
| #29 Work-item Kanban | Partial | Board, filters, modal detail, deep-link, optimistic transition, serta migration compatibility untuk history. | Tambahkan label/group-by, comment count, checklist/watchers/attachments in modal, dan alternatif keyboard status movement. |
| #30 Secure API Contract Runner | Partial | Schema `10_api_contract_runner.sql`, server-side runner, SSRF policy, redaction tests. | Tambahkan CRUD environment/request/parameter/example dan Request Builder UI nyata. Jangan simpan Bearer/API key sebagai text pada graph. |
| #31 Derived views and export | Largely implemented | `derived_views.go`, Status, Calendar, Analytics, Document, dan export service memakai source data turunan. | Tambahkan integration test terhadap pagination/filter dan fixture export contract. |
| #32 Events, versions, and notifications | Partial | Event outbox, baseline publish/restore, notifications, dan SSE endpoint tersedia. | Tambahkan persistent SSE resume (`Last-Event-ID`), frontend EventSource invalidation, watcher subscription, retention job, dan pagination. |
| #33 Cleanup, testing, and rollout | Partial | Runtime Go tunggal, cleanup fallback AI, Docker deployment, README dan runbook. | Tambahkan OpenAPI aktual, ERD, automated E2E/security/performance/accessibility gates, serta migration rehearsal/reconciliation report. |

## Hotfix Kanban 2026-09-17

Gejala: kartu drag-and-drop mendapat HTTP 500 lalu kembali ke kolom asal.

Penyebab yang terverifikasi: PostgreSQL menolak prepared statement update dengan `42P08` karena placeholder status `$1` digunakan sebagai assignment `VARCHAR` dan juga pada ekspresi `CASE` tanpa type cast eksplisit. Pembuatan task tetap berhasil; semua update status gagal sebelum history ditulis.

Perbaikan: query transition sekarang memakai `$1::varchar` pada assignment dan setiap `CASE`, sehingga tipe parameter konsisten pada driver PostgreSQL. Query telah diverifikasi sebagai prepared statement dalam transaction rollback. `12_work_item_transition_history_compatibility.sql` tetap menambah `from_status` secara idempotent sebagai hardening untuk instalasi lama, dan backend mencatat kegagalan update atau audit transition ke log server tanpa mengirim detail database ke browser.

## Urutan Pengerjaan Berikutnya

1. Deploy hotfix Kanban ini dan restart backend agar migrasi 12 berjalan. Verifikasi `Backlog -> Ready` dan lihat log bila masih gagal.
2. Selesaikan #29 dan #26 bersama: evidence/checklist/watcher harus tersedia di detail work item, bukan hanya sebagai tabel.
3. Selesaikan #30: pindahkan contract runner dari field graph ke CRUD contract dan variable reference yang terenkripsi.
4. Selesaikan #32: SSE resume/invalidation dan retention sebelum membuat klaim sinkronisasi antar klien.
5. Tutup #25 dan #33 melalui pemecahan store serta test gate CI yang dapat diulang.
