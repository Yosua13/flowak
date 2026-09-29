# Tugas 15: Work Item Delivery Artifacts

GitHub issue: #51. Branch: `flowak/issue-51`.

## Prasyarat

Issue #49 sudah merged. Baca juga task 04 dan 07.

## Tujuan

Mengoperasionalkan checklist, watcher, work item link, attachment metadata, dan evidence pada detail work item.

## Scope

- Tambah endpoint tenant-scoped untuk CRUD checklist, watcher, link, dan attachment metadata.
- Tambahkan service/domain type serta UI detail work item.
- Tambahkan sort eksplisit dan cursor stabil pada list work item.
- Catat activity untuk perubahan execution artifact.

## Batasan

- Metadata attachment saja; jangan membuat storage provider baru.
- Jangan mengubah graph atau status workflow.

## Acceptance Criteria

- Semua artifact memeriksa capability project pada server.
- Count komentar/attachment dapat diperbarui tanpa reload halaman penuh.
- Sort dan cursor tidak menghasilkan duplikasi atau kartu hilang.

## Verifikasi

```powershell
cd backend
go test ./...
cd ../frontend
npm test
npm run lint
npm run build
```
