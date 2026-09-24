# Tugas 13: Tenant Integration and Session Coverage

GitHub issue: #49. Branch: `flowak/issue-49`.

## Tujuan

Menutup coverage integration untuk tenant isolation, role capability, dan lifecycle session.

## Scope

- Buat test table-driven untuk project, module, graph, work item, comment, notification, dan API runner.
- Uji IDOR lintas organisasi, viewer write denial, invitation expiry/reuse, refresh rotation, logout, dan revoke.
- Gunakan fixture database terisolasi; jangan memakai user/data lokal developer.

## Batasan

- Jangan mengubah business feature atau menambah role baru.
- Jangan melemahkan authorization hanya demi membuat test lulus.

## Acceptance Criteria

- Resource tenant lain tidak dapat dibaca atau diubah dengan ID yang ditebak.
- Session revoked tidak dapat refresh kembali.
- Semua test menyatakan HTTP status/error contract yang stabil.

## Verifikasi

```powershell
cd backend
go test ./...
```
