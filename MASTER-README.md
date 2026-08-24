# AffiliatorShopee

Web app pribadi untuk menyimpan produk curated affiliate Shopee, merapikan data dengan AI, membuat caption, dan membantu posting manual ke X.

## Dokumen

- [Owner Requirement](Owner-Requirement.md)
- [PRD](PRD.md)
- [TRD](TRD.md)
- [TODO](TODO.md)

## Inti Produk

```text
Simpan data produk mentah
→ AI reformat dan simpan hasil
→ Generate caption
→ Pilih hashtag
→ Share ke X via Web Intent
→ User upload gambar dan klik Post
→ Catat riwayat posting
```

Tidak ada auto-posting penuh untuk menghindari deteksi bot.

## Scope MVP

- CRUD produk curated
- AI reformat maksimal 20 produk per request
- Caption template dan variasi
- Hashtag selector
- Share ke X via Web Intent dan clipboard
- Riwayat posting sederhana tanpa menyimpan identitas akun X

Threads, Chrome Extension, dan media storage adalah roadmap lanjutan.

## Tech Stack

- Backend: Go
- Database: PostgreSQL
- Frontend: Vue 3 + Vite + Tailwind CSS + Pinia
- AI: OpenRouter
- Deployment: Docker Compose
