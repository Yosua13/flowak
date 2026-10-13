# Flowak Frontend 🎨

Direktori ini berisi kode aplikasi antarmuka pengguna (Frontend) untuk **Flowak**, yang dibangun sebagai *Single-Page Application* (SPA) menggunakan **React 19**, **Vite**, **TypeScript**, dan **Tailwind CSS v4**.

Frontend Flowak menyediakan kanvas pemodelan diagram alur kerja interaktif, papan pelacakan Kanban real-time, editor dokumen spesifikasi bisnis multidisiplin, serta antarmuka pengujian kontrak API server-side.

---

## 🚀 Fitur Utama Antarmuka

1. **Kanvas Alur Kerja Visual (*Workflow Canvas*)**:
   * Menggambar dan mengatur node proses bisnis (`terminator`, `process`, `decision`, `actor`, `system`).
   * Menghubungkan langkah-langkah proses menggunakan garis aliran koneksi (*edges*).
   * Fitur tata letak otomatis (*auto-align*), pemindahan interaktif (*drag-and-drop*), dan grid alignment.
   * Integrasi prompt AI untuk meng-generate struktur modul alur kerja secara instan maupun mengaudit cacat desain alur.

2. **Papan Kanban Lintas Peran (*Multi-Role Kanban Board*)**:
   * Visualisasi tiket pekerjaan yang terhubung langsung dengan langkah proses di kanvas.
   * Pemisahan tugas berdasarkan peran teknis (`PM`, `UI/UX`, `Frontend`, `Backend`).
   * Transisi status tugas (`Planned` ➡️ `In Progress` ➡️ `Review` ➡️ `Done` / `Blocked`) dengan validasi *state-machine* dan catatan audit.
   * Pembaruan data otomatis secara real-time melalui Server-Sent Events (SSE).

3. **Editor Spesifikasi Fungsional (*Technical Specs View*)**:
   * Dokumentasi komprehensif pada tiap langkah alur: Aktor, Input, Proses, Output, Aturan Bisnis (*Rules*), Sistem Terkait, dan SLA.
   * Panel inspektor detail untuk mengatur penanggung jawab (*assignee*), rute komponen UI, dan endpoint API backend.

4. **Pengujian Kontrak API (*API Contract Runner*)**:
   * Pengujian request HTTP langsung dari kanvas dengan manajemen environment terisolasi.
   * Visualisasi status respons, waktu eksekusi (*latency*), ukuran payload, dan redaksi otomatis untuk data sensitif/kredensial.

5. **Dasbor Analitik & Kolaborasi**:
   * Grafik tingkat penyelesaian pekerjaan, metrik kepatuhan SLA, dan status kolaborator aktif.

---

## 🛠️ Tech Stack & Pustaka Utama

* **Core Framework**: [React 19](https://react.dev/) & [React DOM 19](https://react.dev/)
* **Bahasa**: [TypeScript](https://www.typescriptlang.org/)
* **Build Tool & Dev Server**: [Vite](https://vite.dev/) (dengan proxy terintegrasi ke backend port `3001`)
* **Styling & Design System**: [Tailwind CSS v4](https://tailwindcss.com/) via plugin `@tailwindcss/vite`
* **State Management**: [Zustand](https://github.com/pmndrs/zustand) (mengelola status modul aktif, nodes, edges, filter, dan preferensi sesi)
* **Animasi & Transisi**: [Motion](https://motion.dev/)
* **Ikon**: [Lucide React](https://lucide.dev/)
* **Unit Testing**: [Vitest](https://vitest.dev/)

---

## 📂 Struktur Direktori `/frontend/src`

```text
frontend/src/
├── components/          # Komponen UI modular
│   ├── analytics/       # Visualisasi grafik performa dan kepatuhan SLA
│   ├── auth/            # Halaman login dan registrasi pengguna
│   ├── calendar/        # Kalender tenggat waktu (due dates) tiket pekerjaan
│   ├── canvas/          # Kanvas visual diagram, kartu node, & layer koneksi edges
│   ├── dashboard/       # Dasbor manajemen modul dan proyek
│   ├── doc/             # Editor spesifikasi bisnis dan teknis
│   ├── inspector/       # Panel samping inspeksi dan konfigurasi detail node
│   ├── kanban/          # Papan tugas Kanban per peran pengembang
│   ├── status/          # Tampilan status kesehatan server & koneksi
│   ├── team/            # Manajemen anggota proyek dan pembagian hak akses
│   ├── AppShell.tsx     # Tata letak utama aplikasi (Sidebar + Topbar)
│   ├── Sidebar.tsx      # Navigasi utama perspektif (Canvas, Kanban, Docs, dll.)
│   └── Topbar.tsx       # Header aplikasi, pencarian, status AI, & profil
│
├── config/              # Konfigurasi konstan dan opsi visual
│   ├── nodeTypes.ts     # Warna, ikon, dan definisi tipe node
│   ├── roles.ts         # Warna badge dan penamaan peran (PM, UI/UX, FE, BE)
│   └── seedData.ts      # Data inisial dan fallback demonstrasi
│
├── domain/              # Definisi entitas data TypeScript
├── infra/               # Klien HTTP fetch dan konfigurasi jaringan
├── services/            # Integrasi modul API (proyek, modul, kanban, api contracts)
├── store/               # State store global terpadu berbasis Zustand
├── index.css            # Token desain global, typography, dan konfigurasi Tailwind
└── main.tsx             # Entry point pemuatan DOM React
```

---

## ⚡ Panduan Menjalankan Frontend

### 1. Prasyarat Sistem
Pastikan komputer Anda telah terpasang:
* **Node.js** versi 18.x atau lebih baru
* **npm** (atau package manager yang setara)

---

### 2. Memasang Dependensi
Buka terminal pada direktori `/frontend`, lalu jalankan:
```powershell
npm install
```

---

### 3. Menjalankan Server Pengembangan (Development Mode)
Jalankan dev server Vite:
```powershell
npm run dev
```
* Buka browser pada alamat: **`http://localhost:5173`**.
* Dev server Vite telah dikonfigurasi dengan proxy otomatis; semua request API `/api/*` akan langsung diteruskan ke backend Go di `http://localhost:3001`.

---

### 4. Menjalankan Pengecekan Tipe & Pengujian
Untuk memeriksa konsistensi tipe TypeScript atau menjalankan unit test:
```powershell
# Pengecekan tipe statis TypeScript
npm run lint

# Menjalankan unit test dengan Vitest
npm test
```

---

### 5. Kompilasi Produksi (Production Build)
Untuk membangun bundel aplikasi statis siap produksi:
```powershell
npm run build
```
* Hasil kompilasi akan tersimpan di direktori `/frontend/dist`.
* Pada lingkungan produksi terpadu, direktori `/frontend/dist` ini akan disajikan secara otomatis oleh server backend Go melalui route *fallback SPA*.
* Anda juga dapat menguji hasil build secara lokal menggunakan:
  ```powershell
  npm run preview
  ```