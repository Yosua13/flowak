# Flowak 🌊

**Flowak** (*Flow to Action*) adalah platform SaaS kolaboratif berbasis *traceability-native* yang dirancang untuk memodelkan proses bisnis, menyusun spesifikasi teknis lintas peran, mengelola tiket pekerjaan (*Work Items*), serta menguji kontrak API secara terkontrol.

Flowak memadukan kanvas diagram interaktif dengan otomatisasi cerdas berbasis AI advisory, antrean pesan asinkron, dan sistem *caching* berperforma tinggi untuk memastikan koordinasi proyek perangkat lunak berjalan sinkron dari tahap analisis kebutuhan hingga eksekusi teknis.

---

## 🚀 Tech Stack

Flowak dibangun di atas fondasi teknologi modern dengan pemisahan tanggung jawab yang tegas:

### 1. Frontend
* **Core**: [React 19](https://react.dev/) & [TypeScript](https://www.typescriptlang.org/)
* **Build Tool**: [Vite](https://vite.dev/) (dengan proxy API internal untuk pengembangan)
* **Styling**: [Tailwind CSS v4](https://tailwindcss.com/)
* **State Management**: [Zustand](https://github.com/pmndrs/zustand)
* **Motion & Animasi**: [Motion](https://motion.dev/)
* **Ikon**: [Lucide React](https://lucide.dev/)

### 2. Backend
* **Runtime**: [Go (Golang)](https://go.dev/) v1.21+
* **Web Framework**: [Gin-Gonic](https://gin-gonic.com/)
* **Arsitektur**: Clean Architecture (Hexagonal / Ports & Adapters) dengan pola **1 File = 1 Fungsi** (*Single-Function per File*)
* **Otentikasi & Keamanan**: JWT (*JSON Web Token*) dengan Argon2/Bcrypt & Redis Session Blacklisting

### 3. Data & Infrastruktur
* **Primary Database**: [PostgreSQL](https://www.postgresql.org/) (skema ternormalisasi, multi-tenant RBAC, & audit log)
* **In-Memory Cache & Lock**: [Redis](https://redis.io/) (Hot Graph Cache, Distributed Lock Redlock, dan JWT Blacklist)
* **Message Broker & Event-Driven**: [RabbitMQ](https://www.rabbitmq.com/) (Topic Exchange `flowak.events`, antrean pekerjaan asinkron, dan SSE event broadcast)
* **AI Provider**: Google Gemini AI API (*Advisory only* untuk asistensi alur kerja dan audit desain sistem)

---

## 🏛️ Arsitektur Sistem

Backend Flowak menerapkan prinsip **Clean Architecture** yang terbagi dalam beberapa lapisan modular:

```mermaid
flowchart TD
    subgraph Client ["Client Layer"]
        Web[Frontend React 19 SPA]
    end

    subgraph Transport ["1. Transport Layer"]
        HTTP[Gin HTTP Handlers]
        SSE[SSE Stream Hub]
    end

    subgraph UseCase ["2. Application / Use Case Layer"]
        UC_Auth[Auth & Session Use Cases]
        UC_Work[WorkItem & Kanban Use Cases]
        UC_Graph[Module Graph Sync Use Cases]
        UC_AI[AI Flow Enqueue Use Cases]
        UC_API[API Contract Runner Use Cases]
    end

    subgraph Domain ["3. Core Domain Layer (Pure Go)"]
        D_Model[Domain Entities]
        D_State[Status State Machine Validation]
        D_Cycle[Graph Cycle Detection]
        D_Err[Sentinel Domain Errors]
    end

    subgraph Infra ["4. Infrastructure & Persistence Layer"]
        PG[(PostgreSQL Database)]
        RD[(Redis Cache & Distributed Lock)]
        MQ{RabbitMQ Message Broker}
    end

    subgraph Worker ["5. Background Worker Daemon"]
        C_AI[AI Jobs Consumer]
        C_Runner[API Runner Consumer]
    end

    Web <-->|HTTP / REST| HTTP
    Web <--|Server-Sent Events| SSE
    HTTP --> UseCase
    UseCase --> Domain
    UseCase --> PG
    UseCase --> RD
    UseCase -->|Publish Events & Jobs| MQ

    MQ -->|flowak.sse.broadcast| SSE
    MQ -->|flowak.ai.jobs| C_AI
    MQ -->|flowak.api_runner.jobs| C_Runner
    C_AI --> PG
    C_Runner --> PG
```

### Karakteristik Arsitektur Utama:
1. **Pemisahan Server & Worker**: Tugas komputasi berat (AI generation dan eksekusi cURL API target runner) tidak memblokir HTTP worker, melainkan didelegasikan ke antrean RabbitMQ dan diproses oleh daemon worker independen.
2. **Synchronous Fallback**: Sistem dilengkapi mode fallback otomatis; jika antrean RabbitMQ atau worker tidak aktif, sistem tetap dapat merespons permintaan secara sinkron tanpa galat.
3. **Optimistic Locking & Concurrency Safety**: Perubahan kanvas graf dan transisi status Kanban dilindungi dengan *row versioning* dan *distributed lock* Redis.
4. **Pola 1 File = 1 Fungsi**: Menjamin kemudahan pemeliharaan (*maintainability*), navigasi kode yang intuitif, serta mencegah terjadinya *merge conflict*.

---

## 🔄 Alur Aplikasi (*Application Workflow*)

```mermaid
sequenceDiagram
    autonumber
    actor User as Pengguna / Tim
    participant Canvas as Canvas UI (FE)
    participant API as API Server (BE)
    participant Broker as RabbitMQ Broker
    participant Worker as Worker Daemon
    participant DB as PostgreSQL / Redis

    User->>Canvas: 1. Input ide / kebutuhan alur proses bisnis
    Canvas->>API: 2. Request AI Generate Flow
    API->>Broker: 3. Enqueue job ke antrean flowak.ai.jobs
    API-->>Canvas: 4. Respons instan job_id (Pending)
    Broker->>Worker: 5. Konsumsi pesan & eksekusi Gemini AI
    Worker->>DB: 6. Simpan hasil graf (Nodes & Edges) ke DB
    Canvas->>API: 7. Polling status / menerima hasil flow lengkap
    User->>Canvas: 8. Telaah & sesuaikan rincian proses, peran (UI/UX, FE, BE), dan SLA
    User->>Canvas: 9. Transisi tugas di Kanban Board
    Canvas->>API: 10. Validasi status state-machine & update DB
    API->>Broker: 11. Broadcast event perubahan via SSE
    Broker-->>Canvas: 12. Update Kanban & status kolaborator secara real-time
    User->>Canvas: 13. Uji endpoint API pada server-side runner
    Canvas->>API: 14. Eksekusi contract runner dengan policy keamanan
```

1. **Pemodelan Alur Kerja (Workflow Modeling)**: Pengguna mendesain alur bisnis interaktif di kanvas visual atau meminta asisten AI merancang draf langkah proses (nodes) beserta alirannya (edges).
2. **Spesifikasi Multidisiplin**: Setiap langkah proses diperinci ke dalam dokumen analisis (aktor, input, proses, output, aturan bisnis, SLA) serta peran teknis (UI/UX, Frontend, Backend).
3. **Pelacakan Kanban Real-Time**: Langkah kerja dipetakan menjadi tiket *work items* di papan Kanban. Setiap perpindahan status divalidasi oleh *state machine* domain dan disiarkan lintas pengguna via Server-Sent Events (SSE).
4. **Eksekusi Kontrak API Aman**: Pengembang dapat menguji kontrak endpoint langsung dari kanvas dengan kebijakan isolasi host target, pencatatan durasi, serta redaksi rahasia/token.

---

## 📖 Panduan Menjalankan Aplikasi

Petunjuk instalasi, konfigurasi, dan cara menjalankan masing-masing komponen aplikasi telah dipisahkan ke dalam direktori masing-masing:

* 👉 **Panduan Backend**: Silakan baca [`backend/README.md`](backend/README.md) untuk cara menjalankan server HTTP Go, Redis, RabbitMQ, dan worker daemon.
* 👉 **Panduan Frontend**: Silakan baca [`frontend/README.md`](frontend/README.md) untuk cara memasang dependensi dan menjalankan development server React Vite.
