# Flowak Frontend 🎨

Direktori ini berisi aplikasi antarmuka pengguna (Frontend) untuk **Flowak**, yang dibangun menggunakan **React 19**, **Vite**, **TypeScript**, dan **Tailwind CSS v4**.

Frontend ini dapat dijalankan dalam mode **Standalone** (dengan server Express lokal sebagai proxy AI) atau dihubungkan ke backend Go untuk lingkungan produksi penuh.

---

## 🚀 Fitur Utama Frontend

1.  **Kanvas Alur Kerja (Workflow Canvas)**:
    *   Menggambar node proses bisnis (`terminator`, `process`, `decision`, `actor`, `system`).
    *   Menghubungkan langkah-masing proses secara visual dengan garis alur (*edges*).
    *   Memindahkan node secara drag-and-drop dengan layout grid yang rapi.
2.  **Papan Kanban (Kanban Board)**:
    *   Visualisasi tugas teknis berdasarkan peran (`UI/UX`, `Frontend`, `Backend`).
    *   Mengatur status tugas dari `Planned` ➡️ `In Progress` ➡️ `Review` ➡️ `Done`.
3.  **Spesifikasi Teknis (Technical Specs / Doc View)**:
    *   Editor dokumen spesifikasi proses bisnis terintegrasi untuk aktor, input, proses, output, SLA, dan business rules.
4.  **Kolaborasi Tim & Penjadwalan**:
    *   Fitur tim untuk melihat kontributor aktif.
    *   Kalender untuk melacak tenggat waktu (due dates) tugas.
5.  **Dasbor Analitik & Status**:
    *   Analitik tingkat penyelesaian tugas dan performa SLA.
    *   Status server & indikator latensi sistem.

---

## 🛠️ Tech Stack & Dependensi Utama

*   **UI Library**: [React 19](https://react.dev/)
*   **Build Tool**: [Vite](https://vite.dev/)
*   **Styling**: [Tailwind CSS v4](https://tailwindcss.com/) (menggunakan `@tailwindcss/vite` plugin baru)
*   **State Management**: [Zustand](https://github.com/pmndrs/zustand) (mengelola status global kanvas, filter, otentikasi, dan workspace)
*   **Animasi**: [Motion](https://motion.dev/) (sebelumnya Framer Motion) untuk transisi mikro yang halus
*   **Ikon**: [Lucide React](https://lucide.dev/)
*   **Standalone Server**: [Express](https://expressjs.com/) (menangani Vite SSR/dev middleware dan endpoint proxy AI)

---

## 📂 Struktur Direktori `/frontend/src`

```text
├── components/          # Komponen UI
│   ├── analytics/       # Visualisasi grafik dan metrik SLA
│   ├── auth/            # Halaman Login dan Registrasi
│   ├── calendar/        # Kalender pelacakan deadline tugas
│   ├── canvas/          # Kanvas interaktif, kartu node, & layer koneksi (edges)
│   ├── dashboard/       # Dasbor Hub Proyek
│   ├── doc/             # Panel penulisan dokumen spesifikasi
│   ├── inspector/       # Detail inspector sebelah kanan untuk edit data node aktif
│   ├── kanban/          # Papan tugas per-langkah berdasarkan role developer
│   ├── status/          # Halaman status server & deployment
│   ├── team/            # Pengaturan anggota tim
│   ├── AppShell.tsx     # Layout utama pembungkus aplikasi (Sidebar + Topbar)
│   ├── Sidebar.tsx      # Navigasi lensa (views)
│   └── Topbar.tsx       # Header aplikasi, pencarian, & notifikasi
│
├── config/              # Konfigurasi statis aplikasi
│   ├── nodeTypes.ts     # Konfigurasi warna, ikon, & tipe node
│   ├── roles.ts         # Konfigurasi warna dan label untuk UX, FE, BE
│   └── seedData.ts      # Data demo offline
│
├── domain/              # Definisi entitas data & interface TypeScript
├── infra/               # Klien HTTP untuk interaksi API backend
├── services/            # Logika integrasi API dan helper eksternal
├── store/
│   └── useStore.ts      # State Store global berbasis Zustand (satu file utama terintegrasi)
├── index.css            # Desain sistem global & kustomisasi Tailwind
└── main.tsx             # Titik masuk pemuatan DOM React
```

---

## ⚡ Perintah Pengembangan & Build

Semua perintah dijalankan di dalam direktori `/frontend/`.

| Perintah | Deskripsi |
| :--- | :--- |
| `npm install` | Memasang semua modul dependensi proyek. |
| `npm run dev` | Menjalankan server pengembangan Vite + Express standalone di **`http://localhost:5173`**. |
| `npm run build` | Mengompilasi aplikasi React ke `/dist` dan mem-bundle `server.ts` menggunakan esbuild ke `/dist/server.cjs`. |
| `npm run start` | Menjalankan aplikasi hasil kompilasi produksi via Express (`node dist/server.cjs`). |
| `npm run lint` | Melakukan pemeriksaan tipe TypeScript secara statis. |
| `npm run clean` | Menghapus direktori hasil kompilasi `/dist`. |

---

## 🔌 Variabel Lingkungan (`.env`)

Buat file `.env` di direktori `/frontend/` untuk konfigurasi:

```env
# Kunci API Gemini untuk pemanggilan AI langsung dalam mode standalone
GEMINI_API_KEY="<YOUR_GEMINI_API_KEY>"

# URL deployment aplikasi
APP_URL="http://localhost:3002"

# Port server Express standalone
FRONTEND_SERVER_PORT=3002
```