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

## Scope implementasi saat ini

- CRUD produk curated
- AI reformat maksimal 10 produk per request (OpenRouter, 9router, OpenCode, atau Codex CLI bridge; default `stealth/ox-alpha`)
- Content model Trending, Branded, dan Murah
- Reformat utama dan varian caption terpisah
- Share ke X via Web Intent dan Chrome Extension helper auto-paste (Manifest V3)
- Scrape produk Shopee memakai extension terpisah; hasil scrape di-insert ke web app tanpa link affiliate
- Download image/video URL ke local storage (banyak URL + tombol `+ Add image URL`, termasuk `.mp4`)
- Edit link affiliate Shopee dari product detail
- Tambah dan hapus media lokal dari product detail
- Riwayat posting sederhana tanpa menyimpan identitas akun X
- Lihat metadata media dan download ZIP
- Media detail tampil sebagai grid preview; bisa dipilih massal untuk dihapus. Scraper memfilter icon/logo/avatar kecil dan URL duplikat.
- Jenis barang bersifat multi-label: dapat dikelola dari Settings, diubah dari detail produk, ditampilkan di list, dan dipakai untuk filter dashboard. Produk tanpa label tersedia melalui filter `Uncategorized`.
- Master jenis barang saat ini memisahkan `Buku` dan `Pengembangan Diri` (sebelumnya satu label gabungan).
- Dashboard juga dapat memfilter sumber input: `X`, `Shopee`, atau `Copas`.
- Tampilan Bank Konten diringkas agar list tetap rapi: judul dan excerpt panjang memakai ellipsis, sementara raw konten lengkap tetap tersedia di detail.
- Detail Bank Konten memakai layout responsive dengan textarea raw yang dibatasi tingginya serta metadata, popularitas, niche, dan jenis barang di sidebar.

Threads, riset konten populer per niche untuk X/Facebook, reformat AI per kanal, dan share ke Facebook adalah roadmap lanjutan. Implementasinya harus dipisahkan per platform agar flow, prompt, parser, dan extension tidak saling bertabrakan. S3/CDN adalah pengganti local storage di masa depan.

## Roadmap: Bank Konten X per Niche

Bank konten adalah modul terpisah dari library produk. Konten disimpan per niche, dengan versi asli yang tidak diubah, versi bersih, hasil reformat AI khusus X, statistik snapshot, dan varian per akun.

Niche awal:

- Sukses & Kesuksesan
- Fashion Pria
- Hubungan / Relasi Pria Wanita
- Gym, Lari & Exercise
- Affiliate

User dapat menambah atau mengubah master niche. Satu niche dapat memiliki banyak konten. Satu konten juga dapat memiliki banyak label `Jenis barang` yang sama dengan produk, dan satu label jenis barang dapat dipakai oleh banyak konten (many-to-many); niche konten dan jenis barang tetap dua master yang berbeda.

Alur yang direncanakan: cari konten populer per niche → simpan URL, konten asli, media, dan statistik → bersihkan tanpa menghapus sumber asli → reformat dengan prompt X → buat varian per akun → share melalui helper X. Extension X Research dapat menangkap satu post atau thread yang terlihat pada halaman detail X; thread digabung berurutan dengan media unik. Riset awal dapat memakai halaman pencarian/session X dan seleksi user; collector X API menjadi opsi tahap berikutnya.

Implementasi awal tersedia di menu `Bank konten`: pengguna memilih niche, kategori konten, lalu preset keyword berbasis sinonim/OR. Aplikasi menyusun query dengan grouping, `lang:in` (kode bahasa Indonesia di X), pengecualian repost dan balasan, sebelum membuka pencarian X dengan filter Populer, Terbaru, atau Media. Query custom mengambil alih preset dan menonaktifkan pilihan bawaan sampai dibersihkan. Extension `extension/x-research` menangkap post yang sedang dibuka beserta URL, teks, author, media, tanggal, dan statistik yang terlihat untuk direview sebelum disimpan; X API dan ranking otomatis tetap menjadi tahap berikutnya.

## Tech Stack

- Backend: Go
- Database: PostgreSQL
- Frontend: Vue 3 + Vite + Tailwind CSS + Pinia
- AI: OpenRouter + 9router + OpenCode + Codex CLI bridge (provider/model dipilih dari Settings)
- Codex CLI memakai session login Codex di host melalui `tools/codex-bridge/`; lihat [CODEX-CLI-BRIDGE.md](CODEX-CLI-BRIDGE.md)
- Codex bridge menyediakan `POST /v1/chat/completions` OpenAI-compatible dan menghubungkan Hermes langsung ke Codex CLI lokal; bridge tidak melewati OpenRouter/9router dan harus tetap aktif.
- Deployment: Docker Compose
