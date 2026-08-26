# AffiliatorShopee

Web app pribadi untuk menyimpan produk curated affiliate Shopee, merapikan data dengan AI, membuat caption, dan membantu posting manual ke X.

## Dokumen

- [Owner Requirement](Owner-Requirement.md)
- [PRD](PRD.md)
- [TRD](TRD.md)
- [TODO](TODO.md)
- [AI Coder Guide](AI-CODER-GUIDE.md)
- [Project Handoff](PROJECT-HANDOFF.md)
- [AI Prompt Documentation](AI-PROMPT-DOCUMENTATION.md)

## Inti Produk

```text
Simpan data produk mentah
→ AI reformat dan simpan hasil
→ Reformat AI sesuai content model
→ Buat varian caption bila perlu
→ Download gambar/video ke local storage atau tambah dari detail
→ Share caption + media ke X via Extension helper
→ User klik Post manual
→ Catat riwayat posting
```

Tidak ada auto-posting penuh untuk menghindari deteksi bot.

## Scope MVP

- CRUD produk curated
- AI reformat maksimal 10 produk per request (OpenRouter, default `stealth/ox-alpha`)
- Content model Trending, Branded, dan Murah
- Reformat utama dan varian caption terpisah
- Share ke X via Web Intent dan clipboard + Chrome Extension helper auto-paste (Manifest V3, satu kesatuan)
- Download image/video URL ke local storage (banyak URL + tombol `+ Add image URL`, termasuk `.mp4`)
- Edit link affiliate Shopee dari product detail
- Tambah dan hapus media lokal dari product detail
- Riwayat posting sederhana tanpa menyimpan identitas akun X
- Lihat metadata media dan download ZIP

Threads adalah roadmap lanjutan. S3/CDN adalah pengganti local storage di masa depan.

## Tech Stack

- Backend: Go
- Database: PostgreSQL
- Frontend: Vue 3 + Vite + Tailwind CSS + Pinia
- AI: OpenRouter
- Deployment: Docker Compose
