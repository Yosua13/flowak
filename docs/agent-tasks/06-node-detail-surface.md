# Tugas 06: Node Detail Surface dan Inspector Baru

## Tujuan

Membuat satu detail node yang konsisten untuk Canvas, modal preview, dan halaman penuh. Ringkasan, pekerjaan aktif, dan masalah harus terlihat sebelum formulir facet.

## Prasyarat

Tugas 03, 04, dan 05 sudah digabung.

## Cakupan file utama

- `frontend/src/components/inspector/Inspector.tsx`
- Seluruh tab di `frontend/src/components/inspector/`
- `frontend/src/components/canvas/Canvas.tsx`
- `frontend/src/components/canvas/NodeCard.tsx`
- Component baru `NodeDetailSurface` dan subcomponent terkait
- `frontend/src/index.css`

## Tata letak target

- Drawer kanan di Canvas dengan lebar fleksibel 420-640 px dan opsi buka sebagai halaman penuh.
- Header tetap: nama, type, key, owner, risk, completeness, open-task count, save state, menu, close.
- Tab/section: Ringkasan, Bisnis, UI/UX, Frontend, Backend, Task, Diskusi, Aktivitas.
- Ringkasan menjadi landing tab: outcome, actor, trigger, next paths, readiness tiap facet, overdue/blocked/open tasks, dan quick actions.
- Progressive disclosure untuk field lanjutan; jangan tampilkan empat formulir panjang sekaligus.
- Satu implementasi data/detail dapat dirender sebagai drawer, modal preview, atau full page.

## Perilaku wajib

1. Autosave menampilkan Saving, Saved, Offline, Failed, atau Conflict berdasarkan state nyata.
2. Field teks panjang debounce/blur; select/toggle/date dapat disimpan segera.
3. Conflict menampilkan compare/reload/retry, bukan menimpa data diam-diam.
4. Keyboard dan focus management benar; close mengembalikan fokus ke node asal.
5. Link detail dapat dibagikan dan mempertahankan project/module/node context.
6. Loading, empty, error, unauthorized, deleted, dan stale-version state tersedia.
7. Gunakan icon library yang sudah ada dan tooltip untuk icon tidak umum.

## Acceptance criteria

- Klik node membuka ringkasan yang dapat dipahami tanpa scroll panjang.
- Task dan comment node dapat diakses tanpa pindah ke halaman lain.
- Drawer tidak menutupi kontrol Canvas secara tidak koheren pada desktop dan mobile.
- Data yang sama muncul konsisten pada drawer dan full page.
- Tidak ada label real-time palsu atau wireframe generatif palsu.

## Verifikasi

Jalankan lint/build dan component test. Verifikasi manual minimal pada viewport 1440x900, 1024x768, dan 390x844 untuk overflow, focus trap, serta text fitting.

## Di luar cakupan

- Implementasi runner HTTP.
- Redesign seluruh shell aplikasi.

