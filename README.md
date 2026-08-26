# AffiliatorShopee

Web app pribadi untuk menyimpan produk affiliate Shopee, merapikan data dengan AI, membuat caption, dan membantu posting manual ke X.

## Dokumen

- [Dokumentasi](docs/) - seluruh PRD, TRD, TODO, handoff, dan prompt AI

## Alur MVP

```text
Simpan data Shopee mentah
→ AI reformat dan simpan hasil
→ Reformat AI sesuai content model
→ Buat varian caption bila perlu
→ Download gambar/video ke local storage atau tambah dari halaman detail
→ Share ke X
→ User upload media lokal dan klik Post
→ Catat riwayat posting
```

MVP tidak melakukan auto-posting dan tidak mengelola akun sosial. Media di-download ke local storage agar dikirim ke Chrome Extension helper dan di-upload manual ke X. Media dapat ditambah atau dihapus dari halaman detail. Extension adalah satu kesatuan (load unpacked `extension/`).

Akun X yang digunakan mengikuti session browser yang sedang login. Web app tidak membaca, memilih, atau menyimpan identitas akun tersebut. Produk yang sama boleh dibagikan dan dicatat berkali-kali.

## Status

Core MVP sudah diimplementasikan dan dideploy lokal via Docker/OrbStack. Ikuti [docs/PROJECT-HANDOFF.md](docs/PROJECT-HANDOFF.md) untuk kondisi terbaru.

## Struktur proyek

- `frontend/` — aplikasi Vue/Vite.
- `backend/` — API Go, handler, service, repository, model, storage, dan migration PostgreSQL.
- `extension/` — Chrome extension untuk composer X.
- Setiap produk memiliki tracking tag unik untuk pencocokan report klik/penjualan affiliate.
- `docs/` — PRD, TRD, prompt AI, handoff, dan dokumentasi operasional.
- `scripts/` — utilitas lokal.
