# Tugas 09: Status, Kalender, Analitik, Dokumen, dan Export

## Tujuan

Menjadikan semua view turunan membaca source of truth yang sama: specification readiness, work item/history, evidence, dan published module version.

## Prasyarat

Tugas 04, 05, dan 07 sudah digabung.

## Cakupan file utama

- `frontend/src/components/status/StatusView.tsx`
- `frontend/src/components/calendar/CalendarView.tsx`
- `frontend/src/components/analytics/AnalyticsView.tsx`
- `frontend/src/components/doc/DocView.tsx`
- `frontend/src/services/exportService.ts`
- Backend query/report endpoint jika diperlukan

## Implementasi wajib

1. Status view membedakan facet readiness dari work item progress.
2. Kalender berasal dari `work_items.start_date/due_date`, facet review due date, dan milestone/version; jangan hanya mengambil satu due date per facet.
3. Analitik dihitung dari status history: throughput, cycle time, blocked time, aging WIP, overdue, review rework, completeness, dan traceability coverage.
4. Jangan menilai produktivitas individu atau membuat ranking anggota.
5. Hapus parsing SLA dari teks bebas; gunakan value dan unit typed.
6. Document view merender flow, specification, work item summary, comments decision yang relevan, evidence, dan version provenance dari revision yang sama.
7. Export JSON membawa `schemaVersion`; OpenAPI berasal dari kontrak API typed; cURL memakai redacted variables.
8. Loading/error/empty state harus jujur dan tidak menampilkan angka demo sebagai data aktual.

## Acceptance criteria

- Kartu Kanban, event kalender, dan angka analitik menunjuk work item yang sama.
- Cycle time dapat direkonstruksi dari history tanpa heuristik status facet.
- Dokumen published menyebut module version/baseline dan tidak berubah ketika draft berikutnya diedit.
- Export tidak memuat secret atau response sensitif.
- Filter project/module/facet konsisten lintas view.

## Verifikasi

Tambahkan fixture deterministik untuk memverifikasi perhitungan metric dan snapshot/export schema. Jalankan frontend lint/build dan backend tests bila endpoint ditambah.

## Di luar cakupan

- Custom dashboard builder.
- Time tracking individual.
- Billing analytics.

