# Tugas 07: Kanban Berbasis Work Item dan Popup Detail

## Tujuan

Mengganti kartu otomatis dari facet menjadi board pekerjaan nyata serta membuka detail lengkap saat kartu diklik.

## Prasyarat

Tugas 03 dan 04 sudah digabung.

## Cakupan file utama

- `frontend/src/components/kanban/KanbanView.tsx`
- Component baru work item card, board filters, quick add, dan detail modal
- `frontend/src/components/calendar/CalendarView.tsx` hanya untuk adapter link bila diperlukan
- `frontend/src/domain/types.ts`
- `frontend/src/services/`
- `frontend/src/index.css`

## Board wajib

- Kolom default: Backlog, Ready, In Progress, In Review, Blocked, Done.
- Canceled disembunyikan secara default dan dapat dibuka melalui filter.
- Header: module selector, search, filter assignee/type/facet/priority/label, group-by, dan quick add.
- Kolom menampilkan WIP count, total point, overdue count, dan add action.
- Kartu menampilkan key, type icon, title, node breadcrumb, priority, assignee, point, due date, comment count, dan blocked marker.

## Popup saat kartu diklik

- Modal 760-920 px atau responsif pada mobile.
- Header: key, type, inline title, status, previous/next, copy link, close.
- Main: description, acceptance checklist, child tasks, activity, comment composer.
- Sidebar: node/facet, parent, assignee, reporter, priority, points, dates, labels, watchers.
- Node preview menampilkan outcome, actor, facet readiness, dan tombol buka node drawer.
- Comments mendukung mention, reply, edit/delete milik sendiri, resolve, attachment, timestamp.
- Deep link `/projects/{projectId}/work-items/{key}` harus membuka modal/detail yang sama.

## Perilaku wajib

1. Drag memakai optimistic update dan rollback saat policy server menolak.
2. Transition `Blocked` tanpa alasan ditolak melalui dialog field, bukan silent failure.
3. Sediakan alternatif keyboard untuk perpindahan status.
4. Quick add meminta title dahulu; node mengikuti filter atau dipilih setelah create.
5. Jangan menghasilkan tiga task otomatis untuk setiap node/facet.
6. Escape tidak menutup modal jika ada edit belum tersimpan; beri dialog yang jelas.

## Acceptance criteria

- Bug baru pada satu node muncul pada Backlog dengan seluruh metadata yang diminta user.
- Klik kartu dari Kanban membuka data yang sama dengan deep link.
- Comment baru memperbarui count tanpa reload penuh.
- Filter dan pagination tidak kehilangan kartu ketika cache diinvalidasi.
- Drag invalid kembali ke kolom asal dan menjelaskan sebab.

## Verifikasi

Jalankan lint/build dan test untuk filter, quick add, modal, transition rollback, comment, focus trap, dan URL restore.

