# Tugas 14: Frontend State Boundaries

GitHub issue: #50. Branch: `flowak/issue-50`.

## Tujuan

Memecah `useStore.ts` menjadi batas auth, project, graph, specification, work item, notification, dan UI yang dapat diuji.

## Scope

- Pindahkan server state ke service/query layer yang konsisten.
- Pertahankan Zustand untuk UI state dan adapter selector lama selama migrasi.
- Pastikan cancellation saat project/module berganti dan optimistic rollback tetap berfungsi.
- Tambahkan test adapter dan error/cache behavior.

## Batasan

- Jangan redesign layar atau mengganti library canvas.
- Jangan jadikan localStorage/IndexedDB source of truth kedua.

## Acceptance Criteria

- `useStore.ts` tidak lagi mengatur seluruh network CRUD lintas domain.
- Komponen lama tetap compile melalui adapter transisi.
- Status save tetap berasal dari state nyata.

## Verifikasi

```powershell
cd frontend
npm test
npm run lint
npm run build
```
