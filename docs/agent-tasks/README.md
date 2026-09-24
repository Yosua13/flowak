# Paket Eksekusi Refactor Flowak

Folder ini memecah `Rencana_Pengembangan_Flowak_Traceability_Workspace.docx` menjadi pekerjaan kecil untuk agent AI berbiaya lebih rendah. Setiap agent hanya boleh mengerjakan satu dokumen pada satu branch dan wajib membaca dokumen ini serta dokumen tugasnya.

## Aturan global

- Jangan mengubah file di luar cakupan kecuali perubahan kecil diperlukan agar build lulus; jelaskan pengecualian di laporan akhir.
- Jangan menghapus atau menimpa perubahan yang sudah ada di working tree.
- PostgreSQL adalah database runtime. Jangan menambah ketergantungan pada `backend/flowak.db`.
- Tabel normalized adalah source of truth. JSON graph lama hanya boleh menjadi compatibility snapshot selama migrasi.
- Jangan menyimpan token, API key, cookie, atau secret di graph, log, export, response body, maupun prompt AI.
- Field produk berfokus pada flow, dokumentasi, kontrak, handoff, task, dan evidence. Jangan menambah field untuk source code, framework, nama file, atau boilerplate.
- Semua endpoint tenant harus memverifikasi membership organisasi/proyek pada server.
- Gunakan migrasi forward-only yang idempotent. Jangan mengubah migrasi lama yang mungkin sudah dijalankan.
- Pertahankan pola UI dan stack yang sudah ada kecuali dokumen tugas memerintahkan perubahan.
- Sebelum selesai, jalankan `go test ./...` dari `backend` dan `npm run lint` dari `frontend`. Jalankan test tambahan yang disebutkan dalam tugas.

## Urutan dan dependensi

| Gelombang | Dokumen | Dapat paralel | Prasyarat |
|---|---|---|---|
| 1 | `01-database-graph-safety.md` | Ya | Tidak ada |
| 1 | `02-tenant-auth-rbac.md` | Ya | Tidak ada |
| 1 | `03-frontend-state-foundation.md` | Ya | Tidak ada |
| 2 | `04-work-items-comments.md` | Tidak | 01, 02 |
| 2 | `05-specification-domain.md` | Ya | 01, 02 |
| 3 | `06-node-detail-surface.md` | Ya | 03, 04, 05 |
| 3 | `07-kanban-work-items.md` | Ya | 03, 04 |
| 3 | `08-secure-api-runner.md` | Ya | 01, 02, 05 |
| 4 | `09-derived-views-documents.md` | Ya | 04, 05, 07 |
| 4 | `10-events-versions-notifications.md` | Ya | 01, 02, 04 |
| 5 | `11-cleanup-testing-rollout.md` | Tidak | 01 sampai 10 |
| 6 | `12-graph-reconciliation.md` | Ya | 01 |
| 6 | `13-tenant-integration-coverage.md` | Ya | 02 |
| 6 | `14-frontend-state-boundaries.md` | Ya | 03 |
| 7 | `15-work-item-delivery-artifacts.md` | Ya | 04, 13 |
| 7 | `16-specification-security-completeness.md` | Ya | 05, 13 |
| 8 | `17-node-detail-accessibility.md` | Ya | 06, 14, 15, 16 |
| 8 | `18-kanban-completion.md` | Ya | 07, 14, 15 |
| 8 | `19-api-contract-builder.md` | Ya | 08, 15, 16 |
| 9 | `20-derived-view-contract-tests.md` | Ya | 09, 15, 18, 19 |
| 9 | `21-durable-collaboration.md` | Ya | 10, 13, 15 |
| 10 | `22-release-quality-gates.md` | Tidak | 12 sampai 21 |

## Pemetaan Issue Audit

| Issue GitHub | Dokumen | Branch yang disarankan |
|---|---|---|
| #48 | `12-graph-reconciliation.md` | `flowak/issue-48` |
| #49 | `13-tenant-integration-coverage.md` | `flowak/issue-49` |
| #50 | `14-frontend-state-boundaries.md` | `flowak/issue-50` |
| #51 | `15-work-item-delivery-artifacts.md` | `flowak/issue-51` |
| #52 | `16-specification-security-completeness.md` | `flowak/issue-52` |
| #53 | `17-node-detail-accessibility.md` | `flowak/issue-53` |
| #54 | `18-kanban-completion.md` | `flowak/issue-54` |
| #55 | `19-api-contract-builder.md` | `flowak/issue-55` |
| #56 | `20-derived-view-contract-tests.md` | `flowak/issue-56` |
| #57 | `21-durable-collaboration.md` | `flowak/issue-57` |
| #58 | `22-release-quality-gates.md` | `flowak/issue-58` |

## Cara Menjalankan Agent

Jalankan satu agent untuk satu issue dan satu branch. Agent wajib membaca issue GitHub, dokumen tugas yang dipetakan, dan perubahan pada branch target sebelum mengedit.

```text
Kerjakan GitHub issue #NN pada repository Yosua13/flowak.
Baca docs/agent-tasks/NN-nama-tugas.md dan seluruh kode yang relevan.
Implementasikan hanya scope issue, tambahkan test, jalankan verifikasi,
commit profesional ke branch flowak/issue-NN, push, lalu buat PR menuju develop.
Jangan merge dan jangan mengubah bagian di luar scope.
```

Jangan menjalankan issue dengan prasyarat yang belum merged. Setelah tiap PR merged, agent berikutnya harus membuat branch dari `develop` terbaru.

## Kontrak serah terima agent

Laporan akhir setiap agent harus berisi:

1. Ringkasan perilaku yang berubah.
2. Daftar file yang diubah dan alasan singkat.
3. Migrasi atau compatibility concern.
4. Test yang ditambahkan dan hasil perintah verifikasi.
5. Risiko atau pekerjaan lanjutan yang sengaja tidak dikerjakan.

Jangan menulis hanya "selesai". Jika test gagal karena kondisi awal repository, sertakan command, pesan error inti, dan bukti bahwa perubahan agent tidak memperburuknya.

