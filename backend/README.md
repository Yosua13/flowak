# Flowak Backend ⚙️

Direktori ini berisi kode backend untuk **Flowak** yang ditulis menggunakan bahasa pemrograman **Go (Golang)** dengan framework **Gin-Gonic** dan database **PostgreSQL**.

Backend ini bertanggung jawab untuk mengelola otentikasi pengguna, manajemen proyek dan modul, kolaborasi tim (SaaS tenant & sharing), serta bertindak sebagai gerbang (gateway) proxy ke **Gemini AI API** untuk fitur-fitur pintar berbasis AI.

---

## 🛠️ Tech Stack & Library Utama

*   **Runtime**: Go v1.21+
*   **Web Framework**: [Gin-Gonic](https://gin-gonic.com/) (kinerja tinggi, router HTTP cepat, & middleware mudah)
*   **Database Driver**: `github.com/lib/pq` (klien driver PostgreSQL murni untuk Go)
*   **Keamanan & Otentikasi**:
    *   `golang.org/x/crypto/bcrypt` (hashing password pengguna secara aman)
    *   `github.com/golang-jwt/jwt/v5` (pembuatan dan validasi token JWT)
*   **Integrasi AI**: HTTP Client native yang memanggil REST API resmi Google Gemini.

---

## 📂 Struktur Direktori `/backend`

```text
├── config/            # Loader file .env dan konfigurasi global runtime
│   └── config.go
├── db/                # Koneksi database PostgreSQL & script migrasi otomatis
│   ├── migrations/    # File migrasi skema database (.sql)
│   └── db.go          # Fungsi ping, inisialisasi, dan eksekutor migrasi
├── handlers/          # Controller HTTP (API Endpoints handler)
│   ├── ai.go          # Integrasi prompt Gemini (generate-flow, mock-payload, audit-flow)
│   ├── auth.go        # Pendaftaran, login pengguna, & enkripsi password
│   ├── module_graph.go# Operasi CRUD untuk modul graf, node, edge, dan spesifikasi teknis
│   ├── projects.go    # Manajemen proyek, anggota tim, dan pembagian hak akses
│   └── users.go       # Data profil pengguna & statistik performa dasbor
├── middleware/        # Interceptor HTTP Gin
│   └── auth.go        # Validasi JWT Token & pembatasan akses berbasis peran (role authorization)
├── models/            # Struktur data (Go structs) penampung request & response payload
│   └── models.go
├── main.go            # Entry point aplikasi backend & definisi router
└── .env               # File konfigurasi lingkungan backend (database & API key)
```

---

## 🔌 API Endpoints Reference

Semua endpoint API dilindungi oleh middleware otentikasi JWT kecuali rute pendaftaran & login publik. Sertakan header `Authorization: Bearer <JWT_TOKEN>` pada setiap request terproteksi.

### 👤 Otentikasi & Pengguna (Auth & Users)
*   `POST /api/auth/register` (Publik) - Mendaftarkan pengguna baru dengan peran (`pm`, `uiux`, `frontend`, `backend`).
*   `POST /api/auth/login` (Publik) - Autentikasi email & password untuk mendapatkan token JWT.
*   `GET /api/users` - Mengambil daftar seluruh pengguna di organisasi.
*   `POST /api/users` (Hanya PM) - Membuat akun pengguna baru secara administratif.
*   `DELETE /api/users/:id` (Hanya PM) - Menghapus akun pengguna.
*   `GET /api/users/dashboard-stats` - Mengambil statistik pengerjaan tugas & performa SLA untuk dasbor.

### 📁 Manajemen Proyek (Projects)
*   `GET /api/projects` - Mengambil semua proyek yang dimiliki atau dibagikan dengan pengguna.
*   `POST /api/projects` - Membuat proyek baru.
*   `GET /api/projects/:id` - Mengambil detail informasi satu proyek tertentu.
*   `DELETE /api/projects/:id` - Menghapus proyek beserta modul di dalamnya.
*   `GET /api/projects/:id/members` - Melihat kolaborator tim pada suatu proyek.
*   `POST /api/projects/:id/members` - Menambahkan anggota baru ke proyek.
*   `DELETE /api/projects/:id/members/:userId` - Menghapus anggota tim dari proyek.

### 🌀 Modul & Graf Alur Kerja (Modules & Workflow)
*   `GET /api/projects/:id/modules` - Mengambil semua modul diagram di bawah satu proyek.
*   `POST /api/projects/:id/modules` - Membuat modul diagram baru.
*   `PUT /api/modules/:id` - Memperbarui status diagram (nama, deskripsi, posisi koordinat node graf, & sambungan edges).
*   `DELETE /api/modules/:id` - Menghapus modul diagram.

### 🔮 Integrasi AI Gemini (AI Proxies)
*   `POST /api/ai/generate-flow` - Menghasilkan graf node visual berdasarkan deskripsi prompt proses bisnis.
*   `POST /api/ai/mock-payload` - Menghasilkan mock request/response JSON secara otomatis berdasarkan deskripsi API.
*   `POST /api/ai/audit-flow` - Menganalisis diagram alur kerja untuk menemukan kelemahan integrasi, bottleneck, atau celah keamanan.

---

## ⚡ Perintah Menjalankan Aplikasi

Semua perintah dijalankan di dalam direktori `/backend/`.

1.  **Mengunduh Dependensi Go**:
    ```bash
    go mod download
    ```
2.  **Menjalankan Server Pengembangan**:
    ```bash
    go run main.go
    ```
3.  **Menjalankan Unit Test**:
    ```bash
    go test ./...
    ```
4.  **Kompilasi Aplikasi (Build Binary)**:
    ```bash
    go build -o flowak_backend main.go
    ```

---

## ⚙️ Variabel Lingkungan (`.env`)

Konfigurasi backend disimpan dalam file `.env`:

```env
PORT=3001
JWT_SECRET=<YOUR_JWT_SECRET>
GEMINI_API_KEY=<YOUR_GEMINI_API_KEY>

DB_HOST=localhost
DB_PORT=5432
DB_USER=<YOUR_DB_USER>
DB_PASSWORD=<YOUR_DB_PASSWORD>
DB_NAME=flowak
```
*Catatan: Backend secara otomatis mendeteksi jika database `flowak` belum ada pada PostgreSQL dan akan mencoba membuat database tersebut secara otomatis saat pertama kali dijalankan.*
