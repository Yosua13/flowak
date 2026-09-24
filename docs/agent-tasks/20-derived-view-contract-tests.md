# Tugas 20: Derived View Contract Tests

GitHub issue: #56. Branch: `flowak/issue-56`.

## Prasyarat

Issue #51, #54, dan #55 sudah merged.

## Tujuan

Mengunci konsistensi derived view dan export dengan fixture deterministik.

## Scope

- Buat fixture work item, history, facet, evidence, dan baseline.
- Uji project/module/facet filter, cursor pagination, metric, calendar, document, dan export.
- Uji schema version serta redaction secret pada semua output.

## Batasan

- Jangan menambah ranking produktivitas individu atau dashboard builder.
- Jangan mengubah formula metric tanpa test fixture.

## Acceptance Criteria

- Cycle dan blocked time dapat direkonstruksi dari history.
- Semua view menunjuk record sama untuk fixture identik.
- Export tidak memuat secret atau response sensitif.

## Verifikasi

```powershell
cd backend
go test ./...
cd ../frontend
npm test
npm run lint
npm run build
```
