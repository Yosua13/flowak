# Tugas 18: Kanban Completion and Accessibility

GitHub issue: #54. Branch: `flowak/issue-54`.

## Prasyarat

Issue #50 dan #51 sudah merged. Baca task 07.

## Tujuan

Melengkapi Kanban dengan metadata delivery, detail artifact, dan alternatif keyboard untuk transisi.

## Scope

- Tambahkan comment count, breadcrumb node lengkap, dan filter/group-by label bila model label tersedia.
- Tambahkan keyboard status transition yang memakai policy API sama dengan drag/drop.
- Lengkapi modal dengan checklist, watcher, attachment, activity, parent, reporter, dan metadata terkait.
- Tambahkan test rollback, focus, URL restore, count update, dan keyboard transition.

## Batasan

- Jangan membuat workflow status baru.
- Jangan mem-bypass `Blocked` reason atau `Done` resolution.

## Acceptance Criteria

- Drag dan keyboard menghasilkan policy/error server yang identik.
- Transisi invalid kembali ke status asal dan memberi penjelasan.
- Modal dan deep link membuka detail work item yang sama.

## Verifikasi

```powershell
cd frontend
npm test
npm run lint
npm run build
```
