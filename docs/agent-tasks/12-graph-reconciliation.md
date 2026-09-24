# Tugas 12: Graph Reconciliation Read-Only

GitHub issue: #48. Branch: `flowak/issue-48`.

## Tujuan

Mendeteksi perbedaan graph normalized dengan snapshot JSON kompatibilitas tanpa mengubah data.

## Scope

- Tambahkan service atau command admin read-only untuk module atau project.
- Bandingkan count node/edge, ID yang hilang, dan field graph penting.
- Tambahkan integration test PostgreSQL untuk mismatch, idempotensi save, dan preservation child record.

## Batasan

- Jangan melakukan backfill, repair, atau delete otomatis.
- `workflow_nodes` dan `workflow_edges` tetap source of truth.
- Jangan mengubah migrasi yang sudah ada.

## Acceptance Criteria

- Output stabil, machine-readable, dan tanpa side effect.
- Operator dapat menentukan mismatch dan langkah rollback dari hasilnya.
- Test membuktikan comment/work item tidak hilang setelah graph update.

## Verifikasi

```powershell
cd backend
go test ./...
```
