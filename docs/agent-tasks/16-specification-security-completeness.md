# Tugas 16: Specification Security and Completeness

GitHub issue: #52. Branch: `flowak/issue-52`.

## Prasyarat

Issue #49 sudah merged. Baca task 05 dan 08.

## Tujuan

Menghapus secret mentah dari facet dan membuat validasi/completeness spesifikasi deterministik.

## Scope

- Hapus input token/API key mentah dari graph/facet UI dan gunakan variable reference aman.
- Lengkapi validasi server untuk priority, risk, SLA, URL, facet key, dan field wajib.
- Tambahkan UX repeatable row untuk rule, interaction, validation, state matrix, dan reference.
- Simpan nilai legacy ambigu di `legacy_notes`.

## Batasan

- Jangan menyimpan secret di JSON graph, export, cURL, event, atau AI prompt.
- Jangan membangun full API runner di task ini.

## Acceptance Criteria

- Completeness berasal dari evaluator server yang deterministic.
- Error validasi dapat dirender UI.
- Tidak ada secret mentah pada payload persistence.

## Verifikasi

```powershell
cd backend
go test ./...
cd ../frontend
npm test
npm run lint
npm run build
```
