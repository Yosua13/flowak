# Tugas 17: Node Detail Accessibility

GitHub issue: #53. Branch: `flowak/issue-53`.

## Prasyarat

Issue #50, #51, dan #52 sudah merged.

## Tujuan

Menyediakan surface detail node yang sama untuk drawer, deep link, dan halaman penuh dengan aksesibilitas benar.

## Scope

- Tambahkan route full-page yang mempertahankan project/module/node context.
- Lengkapi activity, loading, empty, unauthorized, deleted, dan stale states.
- Kembalikan focus ke node asal saat close dan lindungi focus trap modal/drawer.
- Tambahkan component test serta verifikasi viewport 1440x900, 1024x768, dan 390x844.

## Batasan

- Jangan redesign seluruh application shell.
- Jangan membuat duplicate data fetching antara drawer dan full-page.

## Acceptance Criteria

- Deep link dan drawer merender data node yang konsisten.
- Keyboard navigation tidak kehilangan focus.
- Canvas controls dan detail tidak overlap secara tidak koheren.

## Verifikasi

```powershell
cd frontend
npm test
npm run lint
npm run build
```
