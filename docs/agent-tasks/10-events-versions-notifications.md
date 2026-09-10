# Tugas 10: Activity Events, Versions, Notifications, dan Kolaborasi

## Tujuan

Mengaktifkan tabel kolaborasi yang sudah dirintis dan mengganti klaim real-time palsu dengan status sinkronisasi serta event yang benar.

## Prasyarat

Tugas 01, 02, dan 04 sudah digabung.

## Cakupan file utama

- `backend/db/migrations/`
- Handler/service activity, version, notification, dan event stream
- `backend/main.go`
- `frontend/src/components/team/TeamView.tsx`
- `frontend/src/components/Topbar.tsx`
- `frontend/src/components/inspector/Inspector.tsx` atau NodeDetailSurface baru
- Frontend notification/event service

## Implementasi wajib

1. Definisikan event names dan payload versioned untuk graph/spec/work item/comment/version changes.
2. Tulis activity log di transaction boundary yang sama dengan perubahan utama atau melalui transactional outbox.
3. Tambahkan module publish baseline, version metadata, diff summary, actor, dan restore-as-new-draft; jangan menimpa history.
4. Tambahkan notification outbox, mention notification, watcher subscription, unread state, dan retry idempotent.
5. Gunakan SSE sebagai tahap awal untuk notification/sync invalidation. WebSocket hanya jika collaborative editing/presence benar-benar diperlukan.
6. UI menampilkan `Saving`, `Saved`, `Offline`, `Failed`, atau `Conflict`. Hapus `Online`/`Real-time Active` statis.
7. Activity feed harus human-readable tetapi menyimpan event typed; before/after sensitif wajib difilter.
8. Terapkan retention dan pagination untuk activity, notification, version snapshot, dan API run evidence.

## Acceptance criteria

- Mention menghasilkan satu notification meski outbox retry.
- Client lain menerima invalidation/event dan mengambil data server terbaru.
- Publish menghasilkan baseline immutable yang dapat dibandingkan.
- Restore membuat draft/version baru dan tercatat di audit.
- UI tidak mengklaim user online tanpa presence service.
- Secret dan payload sensitif tidak muncul di activity log.

## Verifikasi

Tambahkan test outbox idempotency, SSE reconnect/cursor, unread update, publish/diff/restore, tenant isolation, dan retention. Jalankan backend test serta frontend lint/build.

## Di luar cakupan

- Cursor presence, collaborative text editing, dan video/chat.
- Push notification mobile.

