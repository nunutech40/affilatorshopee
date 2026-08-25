# AffiliatorShopee

Web app pribadi untuk menyimpan produk curated affiliate Shopee, merapikan data dengan AI, membuat caption, dan membantu posting manual ke X.

## Dokumen

- [Owner Requirement](Owner-Requirement.md)
- [PRD](PRD.md)
- [TRD](TRD.md)
- [TODO](TODO.md)
- [AI Coder Guide](AI-CODER-GUIDE.md)

## Inti Produk

```text
Simpan data produk mentah
→ AI reformat dan simpan hasil
→ Generate caption
→ Pilih hashtag
→ Download gambar/video ke local storage
→ Share ke X via Web Intent
→ User upload media lokal dan klik Post
→ Catat riwayat posting
```

Tidak ada auto-posting penuh untuk menghindari deteksi bot.

## Scope MVP

- CRUD produk curated
- AI reformat maksimal 10 produk per request (pilih model OpenCode, default `muse-spark-1.2-contributor-free` gratis)
- Caption template dan variasi
- Hashtag selector
- Share ke X via Web Intent dan clipboard + Chrome Extension helper auto-paste (Manifest V3, satu kesatuan)
- Download image/video URL ke local storage (banyak URL + tombol `+ Add image URL`, termasuk `.mp4`)
- Riwayat posting sederhana tanpa menyimpan identitas akun X
- Lihat metadata media dan download ZIP

Threads adalah roadmap lanjutan. S3/CDN adalah pengganti local storage di masa depan.

## Tech Stack

- Backend: Go
- Database: PostgreSQL
- Frontend: Vue 3 + Vite + Tailwind CSS + Pinia
- AI: OpenRouter
- Deployment: Docker Compose
