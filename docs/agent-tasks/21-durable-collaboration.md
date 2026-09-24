# Tugas 21: Durable Collaboration Delivery

GitHub issue: #57. Branch: `flowak/issue-57`.

## Prasyarat

Issue #49 dan #51 sudah merged. Baca task 10.

## Tujuan

Membuat event/notification sinkronisasi tahan reconnect dan lengkap lifecycle retention-nya.

## Scope

- Tambahkan cursor event dan `Last-Event-ID` replay tenant-scoped.
- Tambahkan EventSource frontend untuk invalidasi cache dan reload server state.
- Implementasikan watcher subscription dan notification idempotency.
- Tambahkan retention/pagination untuk activity, notification, baseline, dan API run evidence.

## Batasan

- Jangan menambah collaborative cursor/presence atau WebSocket.
- Jangan mengirim payload sensitif melalui SSE.

## Acceptance Criteria

- Reconnect mendapat event terlewat atau invalidasi aman.
- Mention/watcher hanya menghasilkan satu notifikasi meski outbox retry.
- UI hanya mengklaim status sync yang benar-benar diketahui.

## Verifikasi

```powershell
cd backend
go test ./...
cd ../frontend
npm test
npm run lint
npm run build
```
