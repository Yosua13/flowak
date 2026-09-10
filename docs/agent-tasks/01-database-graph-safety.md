# Tugas 01: Database dan Keamanan Sinkronisasi Graph

## Tujuan

Menghilangkan risiko kehilangan komentar, task, history, dan evidence ketika graph disimpan. Jadikan `workflow_nodes` dan `workflow_edges` sebagai source of truth dengan update incremental.

## Konteks masalah

`syncModuleGraph` saat ini menghapus seluruh edge dan node lalu memasukkannya kembali. Foreign key cascade dapat menghapus data turunan. `modules.nodes` dan `modules.edges` juga menduplikasi data normalized.

## Cakupan file utama

- `backend/handlers/module_graph.go`
- `backend/handlers/module_graph_test.go`
- `backend/db/migrations/`
- `backend/models/models.go`
- `frontend/src/infra/persistence.ts` hanya jika kontrak graph berubah
- `frontend/src/domain/types.ts` hanya untuk `rowVersion` atau tombstone

## Implementasi wajib

1. Tambahkan migrasi baru, jangan edit migrasi 01-04.
2. Tambahkan `row_version`, `deleted_at`, dan timestamp yang diperlukan pada node/edge.
3. Ubah penyimpanan graph menjadi upsert per ID di dalam transaksi.
4. Hapus record hanya jika client mengirim operasi delete eksplisit. Gunakan soft delete pada tahap migrasi.
5. Tolak edge yang source/target-nya tidak ada, berasal dari module lain, atau membuat referensi tidak valid.
6. Gunakan optimistic locking; konflik versi harus menghasilkan HTTP 409 dengan error code stabil.
7. Endpoint read harus menyusun graph dari tabel normalized.
8. Pertahankan penulisan snapshot JSON hanya sebagai compatibility write sementara dan tandai dengan komentar deprecation. Jangan membacanya sebagai source utama.
9. Tambahkan pemeriksaan backfill/mismatch yang dapat dijalankan saat migrasi atau sebagai helper admin tanpa memodifikasi data secara diam-diam.

## Acceptance criteria

- Mengubah posisi atau label satu node tidak menghapus record turunan.
- Menyimpan graph yang sama dua kali bersifat idempotent.
- Dua update dengan versi lama menghasilkan satu sukses dan satu 409.
- Edge lintas module ditolak.
- Delete node eksplisit mempertahankan audit/tombstone sesuai retention design.
- Test membuktikan komentar atau record turunan dummy tetap ada setelah graph update.

## Verifikasi

```powershell
cd backend
go test ./...
```

Tambahkan test repository/handler untuk upsert, conflict, invalid edge, idempotency, dan preservation of children.

## Di luar cakupan

- UI inspector baru.
- Work item schema lengkap.
- Menghapus kolom JSON lama pada release ini.

