# Tugas 02: Tenant Isolation, Authentication, dan RBAC

## Tujuan

Memastikan pengguna hanya melihat dan mengubah organisasi/proyek yang sah, serta memisahkan functional role dari project permission.

## Cakupan file utama

- `backend/handlers/auth.go`
- `backend/handlers/projects.go`
- `backend/handlers/users.go`
- `backend/middleware/auth.go`
- `backend/models/models.go`
- `backend/db/migrations/`
- `backend/main.go`
- `frontend/src/components/auth/`
- `frontend/src/components/team/TeamView.tsx`
- `frontend/src/config/roles.ts`

## Implementasi wajib

1. Tambahkan `organization_invitations` dan `user_sessions` melalui migrasi baru.
2. Registrasi publik tidak boleh memilih sendiri role PM atau privilege tinggi. User pertama organisasi boleh menjadi owner melalui alur bootstrap yang eksplisit.
3. Tambahkan invitation acceptance dengan token hash dan expiry.
4. Terapkan project role `owner`, `editor`, `commenter`, `viewer`; functional role PM/UIUX/FE/BE hanya untuk assignment dan default view.
5. Buat helper middleware/policy tunggal untuk resolve organization membership, project membership, dan capability.
6. Scope endpoint project, module, graph, member, dan user picker berdasarkan tenant aktif.
7. Endpoint `/users` tidak boleh mengembalikan user global lintas organisasi.
8. Tambahkan refresh session, revoke/logout server-side, dan password reset request flow minimal. Jangan mengirim informasi apakah email terdaftar.
9. Jangan menaruh token akses jangka panjang di state yang diekspor atau log. Ikuti pola session yang paling aman dan kompatibel dengan arsitektur saat ini.

## Matriks minimum

| Aksi | Owner | Editor | Commenter | Viewer |
|---|---|---|---|---|
| Kelola member/proyek | Ya | Tidak | Tidak | Tidak |
| Ubah graph/spec | Ya | Ya | Tidak | Tidak |
| Buat/ubah work item | Ya | Ya | Terbatas | Tidak |
| Komentar | Ya | Ya | Ya | Tidak |
| Lihat proyek | Ya | Ya | Ya | Ya |

## Acceptance criteria

- User organisasi A tidak dapat menebak ID resource organisasi B untuk membaca atau mengubahnya.
- Viewer yang memanggil endpoint PATCH langsung menerima 403.
- Registrasi biasa tidak dapat menghasilkan owner/PM privilege secara sewenang-wenang.
- Invitation expired/reused ditolak.
- Logout atau revoke membuat refresh session tidak dapat dipakai kembali.

## Verifikasi

Tambahkan integration test authorization matrix dan IDOR test, lalu jalankan:

```powershell
cd backend
go test ./...
```

## Di luar cakupan

- SSO enterprise.
- Billing dan plan enforcement lengkap.
- Presence real-time.

