# Tugas 08: API Contract Builder dan Secure API Runner

## Tujuan

Membuat pengalaman seperti Postman dalam batas fokus Flowak: request berasal dari kontrak Backend node, dapat dijalankan melalui proxy aman, dan hasilnya dapat disimpan sebagai evidence.

## Prasyarat

Tugas 01, 02, dan 05 sudah digabung.

## Cakupan file utama

- `backend/db/migrations/`
- `backend/models/models.go`
- Handler/service runner baru dan `backend/main.go`
- `frontend/src/components/inspector/BackendTab.tsx`
- `frontend/src/services/curl.ts`
- `frontend/src/domain/types.ts`
- Component baru request builder/response viewer

## Model minimum

- `environments`: project, name, approved base URL, default flag, runner policy.
- `environment_variables`: key, encrypted value, secret flag, updater, timestamp.
- `api_requests`: node, name, method, relative path, body type/template, timeout, version.
- `api_request_parameters`: location path/query/header/auth, key, template, description, enabled, secret reference.
- `api_response_examples`: status, content type, body, error flag.
- `api_runs`: request, environment, actor, status, duration, response size, redacted metadata, optional retained body, evidence flag, expiry.

## UI minimum

Request bar berisi method, environment, relative path, resolved URL, Send/Cancel. Tab Params, Authorization, Headers, Body, Examples, Response, History, dan cURL. Response menampilkan status, duration, size, headers, pretty/raw body, truncation, dan Save as evidence.

Authorization MVP: no auth, API key, Bearer, Basic, inherit. Secret selalu berupa variable reference dan tidak pernah disalin sebagai teks mentah.

## Guardrail keamanan wajib

1. Request dijalankan server-side, bukan fetch langsung browser.
2. Hanya HTTP/HTTPS; tolak localhost, loopback, link-local, private network, multicast, metadata cloud, dan hostname/IP terlarang kecuali runner private yang dirancang terpisah.
3. Resolve DNS dan validasi seluruh hasil; validasi ulang setiap redirect. Redirect default off.
4. Gunakan host allowlist untuk SaaS publik.
5. Batasi method, timeout, body, response size, concurrency, dan rate per tenant.
6. Enkripsi secret dengan master key/KMS terpisah. Jangan kembalikan nilai penuh setelah disimpan.
7. Redact Authorization, Cookie, Set-Cookie, API key, dan secret dari cURL, logs, response viewer, screenshots, dan AI prompt.
8. Catat actor, target host, policy decision, status, duration, dan request ID tanpa payload sensitif.

## Acceptance criteria

- Request valid ke host allowlisted menghasilkan response nyata dan history metadata.
- Target localhost atau metadata cloud ditolak sebelum koneksi.
- DNS rebinding/redirect ke private address ditolak.
- cURL memakai placeholder variable, bukan secret.
- Cancel dan timeout membebaskan resource.
- Hasil run dapat disimpan sebagai evidence yang terkait node/request/version.

## Verifikasi

Tambahkan security/integration test dengan local fake servers untuk allow/deny, redirect, timeout, size limit, redaction, dan auth. Jalankan backend tests serta frontend lint/build.

## Di luar cakupan

- Collection marketplace, scripting, monitor berkala, GraphQL, gRPC, WebSocket, dan binary multipart pada MVP.

