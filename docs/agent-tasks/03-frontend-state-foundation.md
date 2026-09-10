# Tugas 03: Fondasi State dan API Frontend

## Tujuan

Memecah store besar menjadi batas domain yang dapat diuji dan menjadikan server sebagai sumber data kolaboratif tanpa mengubah tampilan secara drastis.

## Cakupan file utama

- `frontend/src/store/useStore.ts`
- `frontend/src/infra/persistence.ts`
- `frontend/src/domain/types.ts`
- `frontend/src/domain/invariants.ts`
- `frontend/src/App.tsx`
- `frontend/package.json`
- File baru di `frontend/src/features/`, `frontend/src/services/`, atau `frontend/src/store/`

## Implementasi wajib

1. Petakan state menjadi `auth`, `projects`, `graph`, `specifications`, `workItems`, dan `ui`.
2. Pertahankan Zustand hanya untuk UI/client state yang sesuai. Tambahkan server-state layer yang konsisten, disarankan TanStack Query.
3. Buat typed API client tunggal untuk base URL, auth/session, JSON parsing, cancellation, error normalization, dan 409 handling.
4. Pisahkan query cache dari graph editing state. Jangan menyalin seluruh response server ke beberapa store.
5. Definisikan status save nyata: `idle`, `saving`, `saved`, `offline`, `failed`, `conflict`.
6. Batasi persistence lokal pada preferensi dan draft/offline queue yang jelas. IndexedDB/localStorage tidak boleh menjadi source of truth kedua tanpa protocol sync.
7. Pertahankan public selector/actions compatibility sementara agar migrasi component dapat bertahap.
8. Tambahkan test untuk error normalization, cache invalidation, optimistic rollback, dan conflict state.

## Acceptance criteria

- `useStore.ts` tidak lagi menangani auth, network CRUD, graph, notification, dan seluruh UI dalam satu implementation body.
- Tidak ada label `Online` atau `Real-time Active` berdasarkan nilai hard-coded.
- Request dapat dibatalkan saat project/module berpindah.
- Optimistic mutation rollback ketika server menolak.
- Component lama tetap dapat dikompilasi selama adapter migrasi tersedia.

## Verifikasi

```powershell
cd frontend
npm run lint
npm run build
```

Jalankan test frontend yang tersedia; bila belum ada runner, tambahkan konfigurasi test minimal hanya jika diperlukan untuk kode baru.

## Di luar cakupan

- Redesign inspector dan Kanban.
- Implementasi backend endpoint baru.
- Mengubah library canvas pada tugas ini.

