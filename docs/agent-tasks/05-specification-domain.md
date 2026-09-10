# Tugas 05: Refactor Domain Spesifikasi Bisnis, UI/UX, FE, dan BE

## Tujuan

Mengganti empat formulir panjang dan field berorientasi coding menjadi kontrak lintas disiplin yang terstruktur, dapat divalidasi, dan siap menjadi dokumentasi.

## Prasyarat

Tugas 01 dan 02 sudah digabung.

## Cakupan file utama

- `backend/db/migrations/`
- `backend/models/models.go`
- Handler specification baru atau handler graph terkait
- `frontend/src/domain/types.ts`
- `frontend/src/config/nodeTypes.ts`
- `frontend/src/components/inspector/BisnisTab.tsx`
- `frontend/src/components/inspector/UiuxTab.tsx`
- `frontend/src/components/inspector/FrontendTab.tsx`
- `frontend/src/components/inspector/BackendTab.tsx`

## Field Bisnis

Outcome, actors, trigger type dan trigger, preconditions, business inputs/outputs, process description, business rules dengan code/severity/order, decision outcomes yang terhubung ke edge, exception paths, SLA value+unit, priority, risk, acceptance criteria Given/When/Then, dan reference links.

## Field UI/UX

Owner/reviewer/readiness/due date, user goal, surface, screen/flow name, Figma frame URL, design version, screen states, interaction rows, content/messages, responsive intent, accessibility checklist, asset references, handoff checklist, dan evidence.

## Field Frontend

Owner/reviewer/readiness/due date, experience name, optional navigation route, entry/exit behavior, interactions, input requirements, validation/messages, UI state matrix, API references, analytics event intent, feature availability, handoff evidence, dan acceptance checklist.

Hapus dari UI: component implementation, framework, file name, boilerplate, code snippet, dan code generator.

## Field Backend

Owner/reviewer/readiness/due date, service capability, API contract references, request/response examples, error behavior, business validation, dependency references, idempotency/caching/security notes, observability intent, SLA, dan acceptance checklist. Struktur API request detail disiapkan untuk Tugas 08.

## Implementasi wajib

1. Pisahkan readiness/assignment facet dari work item.
2. Gunakan typed columns untuk filter utama dan child tables untuk repeatable rows; jangan menaruh semuanya di satu JSONB.
3. Tambahkan enum/check constraint atau server validation untuk priority, risk, status, SLA, URL, dan facet key.
4. Tambahkan completeness evaluator per facet berdasarkan field wajib dan checklist.
5. Field lanjutan memakai progressive disclosure; perubahan panjang disimpan saat blur/debounce dengan retry dan conflict response.
6. Migrasikan field lama yang dapat dipetakan. Simpan nilai ambigu di `legacy_notes` untuk review, jangan membuangnya.

## Acceptance criteria

- Tidak ada field coding di inspector FE/BE kecuali cURL dan API contract/reference yang relevan.
- Decision outcome tanpa edge dan rule tanpa acceptance/exception dapat dideteksi.
- UI/UX tidak membuat wireframe palsu dari nama screen; tampilkan link/evidence nyata.
- Completeness dihitung deterministik dari data server.
- Export lama dapat dimigrasikan atau menampilkan warning yang jelas.

## Verifikasi

Jalankan `go test ./...`, `npm run lint`, dan `npm run build`. Tambahkan test untuk validation, completeness, dan migration mapping.

