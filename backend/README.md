# Flowak Backend ⚙️

Direktori ini berisi kode backend untuk **Flowak** yang dibangun menggunakan bahasa pemrograman **Go (Golang)** dengan prinsip **Clean Architecture (Hexagonal / Ports & Adapters)** dan aturan ketat **1 File = 1 Fungsi** (*Single-Function per File*).

Backend bertanggung jawab mengelola otentikasi & sesi pengguna, validasi domain graf proses bisnis, transaksi Kanban dengan *optimistic locking*, *caching* & *distributed locks* via Redis, pengiriman event real-time via SSE, serta pemrosesan *background job* asinkron untuk AI dan API contract runner melalui RabbitMQ.

---

## 🛠️ Tech Stack & Pustaka Utama

* **Runtime**: Go v1.21+
* **Web Framework**: [Gin-Gonic](https://gin-gonic.com/) (router HTTP berkecepatan tinggi)
* **Database Driver**: `github.com/lib/pq` (driver PostgreSQL murni untuk Go)
* **In-Memory Cache & Lock**: `github.com/redis/go-redis/v9` (klien Redis v9)
* **Message Broker Client**: `github.com/rabbitmq/amqp091-go` (klien resmi RabbitMQ AMQP 0-9-1)
* **Keamanan & Kriptografi**:
  * `golang.org/x/crypto/bcrypt` (hashing password)
  * `github.com/golang-jwt/jwt/v5` (pembuatan dan verifikasi token JWT)
* **Mock Testing**: `github.com/DATA-DOG/go-sqlmock`

---

## 📂 Struktur Direktori `/backend`

Arsitektur backend dibagi menjadi beberapa lapis modular:

```text
backend/
├── cmd/
│   ├── worker/              # Daemon worker mandiri untuk antrean RabbitMQ (AI & API Runner)
│   │   └── main.go
│   └── graph-reconcile/     # CLI tool rekonsiliasi graf alur kerja
│       └── main.go
│
├── config/                  # Pengaturan konfigurasi runtime
│   └── config.go
│
├── db/                      # Koneksi database & migrasi SQL otomatis
│   ├── migrations/          # File migrasi terurut (01..15)
│   └── db.go
│
├── internal/
│   ├── domain/              # Layer 1: Domain Entities, State Machine, & Repository Interfaces
│   │   ├── ai/              # Entitas AIJob & status
│   │   ├── auth/            # Entitas User & validasi otentikasi
│   │   ├── common/          # Sentinel domain errors terpusat
│   │   ├── module/          # Graf entitas (Node, Edge), Cycle Detector, & SLA Parser
│   │   └── workitem/        # Entitas WorkItem & Status State Machine
│   │
│   ├── usecase/             # Layer 2: Application Use Cases (Pola 1 File = 1 Fungsi)
│   │   ├── ai/              # Enqueue flow, Process AI job, Get job status
│   │   ├── apicontract/     # Enqueue API run & Process API run job
│   │   ├── auth/            # Register, Login, Logout, Refresh token
│   │   ├── module/          # Sync graph, Get graph, Publish baseline, Reconcile
│   │   └── workitem/        # Create, List, Detail, Transition status, Update
│   │
│   ├── repository/          # Layer 3: Implementasi Akses Data (PostgreSQL & Redis)
│   │   ├── postgres/        # Query repositories (ai, auth, module, workitem)
│   │   └── redis/           # Redis Cache, Redlock Distributed Lock, & Token Blacklist
│   │
│   ├── infra/               # Layer 3: External Adapters & Broker
│   │   └── rabbitmq/        # Koneksi, Topologi Exchange/Queue, Producer, & Consumer
│   │
│   └── transport/           # Layer 4: HTTP Delivery (Gin Engine)
│       └── http/
│           ├── handler/     # Modular HTTP Handlers (ai, auth, module, workitem)
│           ├── middleware/  # Auth JWT, CORS, RBAC
│           ├── response/    # Standardized unified JSON error & success writers
│           └── routes/      # Registrasi rute modular
│
├── handlers/                # Kompatibilitas handler (API runner, tenant helpers)
├── middleware/              # Middleware otentikasi & tenant authorization
├── models/                  # Struct model data transport legacy
├── main.go                  # Entrypoint utama server HTTP API Gin
└── Dockerfile               # Kontainerisasi server produksi
```

---

## ⚡ Panduan Menjalankan Backend

### 1. Prasyarat Sistem
Pastikan perangkat Anda telah terpasang:
* **Go** versi 1.21 atau lebih baru
* **PostgreSQL** (Port default `5432`)
* **Redis** (Port default `6379`)
* **RabbitMQ** (Port default `5672`, Management UI port `15672`)

*(Tips: Anda dapat menjalankan Redis & RabbitMQ secara cepat menggunakan Docker):*
```powershell
docker run -d --name flowak-redis -p 6379:6379 redis:alpine
docker run -d --name flowak-rabbitmq -p 5672:5672 -p 15672:15672 rabbitmq:3-management
```

---

### 2. Memasang Dependensi Go
Buka terminal di direktori `/backend`, lalu jalankan:
```powershell
go mod download
```

---

### 3. Menjalankan Server Utama (HTTP API)
Jalankan server API web Flowak:
```powershell
go run main.go
```
* Server akan otomatis menguji koneksi ke PostgreSQL, mengeksekusi migrasi tabel terbaru, menghubungkan Redis & RabbitMQ, lalu mulai melayani request pada port default **`3001`** (`http://localhost:3001`).

---

### 4. Menjalankan Background Worker Daemon
Buka terminal baru di direktori `/backend`, lalu jalankan daemon worker untuk memproses pekerjaan asinkron AI dan API contract runner:
```powershell
go run cmd/worker/main.go
```
* Worker daemon akan mulai mendengarkan antrean `flowak.ai.jobs` dan `flowak.api_runner.jobs` secara terisolasi dengan dukungan *graceful shutdown* (`Ctrl+C`).

---

### 5. Menjalankan CLI Rekonsiliasi Graf (Opsional)
Untuk memvalidasi dan mencocokkan konsistensi data graf kanvas terhadap baris tabel ternormalisasi:
```powershell
# Rekonsiliasi berdasarkan Project ID
go run cmd/graph-reconcile/main.go --project-id=<PROJECT_UUID>

# Atau rekonsiliasi berdasarkan Module ID
go run cmd/graph-reconcile/main.go --module-id=<MODULE_UUID>
```

---

### 6. Menjalankan Unit & Integration Tests
Untuk menjalankan seluruh suite pengujian otomatis di seluruh paket backend:
```powershell
go test -v -count=1 ./...
```
