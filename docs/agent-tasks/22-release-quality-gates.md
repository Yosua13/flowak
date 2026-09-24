# Tugas 22: Release Documentation and Quality Gates

GitHub issue: #58. Branch: `flowak/issue-58`.

## Prasyarat

Task 12 sampai 21 sudah merged dan release candidate dapat dijalankan.

## Tujuan

Membuat release Flowak dapat diuji, dioperasikan, dan dipulihkan secara reproducible.

## Scope

- Lengkapi OpenAPI versioned, ERD, dan dokumentasi perilaku aktual.
- Tambahkan CI gates unit, integration, contract, E2E, security, performance, accessibility, dan resilience.
- Tambahkan migration rehearsal: backup, row count, checksum, reconciliation, feature flag, rollback, dan laporan residual risk.

## Batasan

- Jangan mengklaim feature yang belum ada.
- Jangan menjalankan rehearsal terhadap database produksi.

## Acceptance Criteria

- CI dapat dijalankan ulang dari branch bersih dan menyimpan bukti hasilnya.
- Runbook menjelaskan backup/restore, secret rotation, incident response, retention, dan rollback.
- Rehearsal memakai salinan non-produksi dan tidak kehilangan record.

## Verifikasi

- Jalankan workflow CI pada branch uji.
- Jalankan rehearsal migrasi pada database non-produksi.
- Tautkan hasil command dan artefak laporan di PR.
