# Tugas 04: Work Item, Hierarchy, Comment, dan Evidence

## Tujuan

Membuat unit eksekusi yang terpisah dari readiness facet sehingga PM dapat membuat story, task, bug, review, research, dan subtask pada node.

## Prasyarat

Tugas 01 dan 02 sudah digabung.

## Cakupan file utama

- `backend/db/migrations/`
- `backend/models/models.go`
- `backend/handlers/` dan `backend/main.go`
- `frontend/src/domain/types.ts`
- `frontend/src/services/` untuk kontrak work item

## Model wajib

`work_items` minimal memiliki: `id`, `project_id`, `module_id`, nullable `node_id`, nullable `facet_key`, nullable `parent_id`, project sequence/key, `type`, `title`, `description`, `priority`, nullable `points`, `status`, nullable `assignee_id`, `reporter_id`, `start_date`, `due_date`, `blocked_reason`, `resolution`, `row_version`, timestamps, dan soft-delete marker.

Tambahkan tabel untuk `work_item_status_history`, `work_item_links`, `attachments` metadata, `watchers`, dan acceptance criteria/checklist bila belum disediakan tugas spesifikasi.

## Aturan domain

- Type: Story, Task, Bug, Review, Research, Subtask.
- Point: nullable atau 1, 2, 3, 5, 8, 13.
- Status: Backlog, Ready, In Progress, In Review, Blocked, Done, Canceled.
- Parent harus satu proyek, tidak boleh self-parent atau cycle.
- Task umum proyek boleh tanpa node, tetapi harus memiliki alasan/scope proyek.
- `Blocked` mewajibkan alasan; `Done` mewajibkan resolution sesuai policy.
- Comments harus menargetkan tepat satu `node_id` atau `work_item_id`, mendukung reply, mention, edit-own, resolve, dan timestamp.
- Perubahan status, assignee, point, parent, due date, dan evidence menghasilkan activity/history.

## API minimum

- `GET/POST /api/projects/{id}/work-items`
- `GET/PATCH /api/work-items/{key}`
- `POST /api/work-items/{key}/transitions`
- `POST /api/work-items/{key}/comments`
- `GET/POST /api/nodes/{id}/comments`

List harus memiliki filter eksplisit, sort, dan cursor pagination.

## Acceptance criteria

- PM dapat membuat Bug pada node dengan parent, 5 point, assignee, deadline, description, dan comment.
- Key proyek bersifat unik dan aman pada concurrent create.
- Cycle hierarchy ditolak.
- Transition invalid ditolak dengan field error yang dapat ditampilkan UI.
- Comment node tidak hilang setelah node dipindahkan atau graph disimpan.
- Status history cukup untuk menghitung cycle time.

## Verifikasi

Tambahkan unit test domain dan integration test repository/handler, lalu jalankan `go test ./...` dari `backend`.

## Di luar cakupan

- Tampilan Kanban dan modal.
- Sprint planning, custom workflow, dan automation rule engine.

