# Tugas 19: API Contract Builder

GitHub issue: #55. Branch: `flowak/issue-55`.

## Prasyarat

Issue #49, #51, dan #52 sudah merged. Baca task 08.

## Tujuan

Mengganti simulator Backend tab dengan CRUD contract dan Request Builder server-side yang aman.

## Scope

- Tambahkan CRUD tenant-scoped untuk environment, variable reference, request, parameter, example, dan run history.
- Buat Request Builder dengan Params, Authorization, Headers, Body, Examples, Response, History, cURL serta Send/Cancel.
- Simpan secret sebagai encrypted/reference value dan response sebagai evidence ketika user berwenang.

## Batasan

- Browser tidak boleh mengakses target API langsung.
- Jangan menambahkan collection marketplace, GraphQL, gRPC, atau WebSocket.

## Acceptance Criteria

- Target private/localhost/metadata dan redirect berbahaya ditolak.
- cURL dan activity tidak memuat secret.
- Response menampilkan status, duration, size, header, truncation, dan request ID.

## Verifikasi

```powershell
cd backend
go test ./...
cd ../frontend
npm test
npm run lint
npm run build
```
