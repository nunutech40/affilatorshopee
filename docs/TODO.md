# TODO - AffiliatorShopee

Project: AffiliatorShopee
Deskripsi: Web app pribadi untuk menyimpan produk affiliate Shopee, merapikan data dengan AI, membuat caption, dan membantu posting manual ke X.
Target: MVP yang bisa dipakai sendiri.
Stack: Go, PostgreSQL, Vue 3 + Vite + Tailwind + Pinia, AI provider OpenRouter/9router/OpenCode.
Deployment: Docker Compose di Orbstack.

Dokumen sumber:

- `PRD.md` - kebutuhan produk dan scope MVP
- `TRD.md` - arsitektur teknis, API, database schema, dan testing
- `Owner-Requirement.md` - konteks dan keputusan produk
- `AI-CODER-GUIDE.md` - guardrail coding assistant dan review Luna

---

## Phase 0: Keputusan dan Kontrak

- [x] Tetapkan stack final: Go, PostgreSQL, Vue 3, Vite, Tailwind, Pinia
- [x] Tetapkan scope MVP: produk, AI reformat, caption, hashtag, share X, post log
- [x] Keluarkan manajemen identitas akun X dan Threads dari MVP; media lokal (download URL ke `STORAGE_PATH`) dan dua Chrome Extension terpisah tetap bagian MVP
- [x] Tetapkan produk boleh diposting berulang kali
- [x] Tetapkan post log tidak menyimpan identitas akun X
- [x] Tetapkan AI reformat langsung menyimpan hasil tanpa preview
- [x] Finalisasi API request/response dan HTTP status code
- [x] Finalisasi template caption dan placeholder
- [x] Pilih `golang-migrate/migrate` sebagai migration tool

## Phase 1: Project Setup

- [x] Repository Git tersedia di folder project
- [x] Inisialisasi module Go: `go mod init github.com/nunutech40/affilatorshopee`
- [x] Buat struktur folder sesuai `TRD.md`
- [x] Buat `Dockerfile` untuk Go backend
- [x] Buat `docker-compose.yml` dengan service `db` dan `app`
- [x] Buat `.env.example` tanpa secret nyata
- [x] Buat project Vue dengan Vite di folder `web`
- [x] Install Tailwind CSS, Vue Router, Pinia, dan test runner frontend
- [x] Tambahkan endpoint `GET /healthz`
- [x] Verifikasi `docker compose up --build` berhasil

## Phase 2: Database dan Backend Core

- [x] Buat migration `001_create_products.up.sql` dan `001_create_products.down.sql`
- [x] Buat migration `002_create_post_logs.up.sql` dan `002_create_post_logs.down.sql`
- [x] Buat migration `003_create_caption_variations.up.sql` dan `003_create_caption_variations.down.sql`
- [x] Buat migration `004_create_product_media.up.sql` dan `004_create_product_media.down.sql`
- [x] Buat koneksi PostgreSQL dan menjalankan migration saat startup
- [x] Buat model `Product`, `PostLog`, `CaptionVariation`, dan `MediaFile`
- [x] Buat config loader untuk `PORT`, `DATABASE_URL`, `STORAGE_PATH`, provider AI, model, dan `ENV`
- [x] Implementasikan `ProductRepository`: Create, GetByID, List, Update, Delete
- [x] Implementasikan `PostLogRepository`: Create, List
- [x] Implementasikan `CaptionVariationRepository`: Create, List
- [x] Implementasikan validasi input produk
- [x] Enforce transisi status `raw -> reformatted`, `raw -> ready`, dan `reformatted -> ready`
- [x] Validasi field minimum sebelum status `ready`
- [x] Simpan image/video URL dan download file ke local storage (`STORAGE_PATH`) saat create produk
- [x] Tambahkan master jenis barang multi-label dan relasi `product_niches`
- [x] Pisahkan jenis barang `Buku` dan `Pengembangan Diri`, lalu petakan produk existing
- [x] Implementasikan `MediaRepository` dan `MediaService` dengan validasi private IP dan batas ukuran
- [x] Implementasikan endpoint `GET /api/products/{id}/media`, `POST /api/products/{id}/media`, `DELETE /api/products/{id}/media/{mediaID}`, dan `GET /api/products/{id}/media/download` (ZIP)
- [x] Setup router HTTP dan CORS localhost
- [x] Implementasikan API products
- [x] Implementasikan API post logs
- [x] Sertakan `post_count` dan `last_posted_at` pada product list/detail
- [ ] Verifikasi CRUD product dan post log dengan automated API tests
- [ ] Test aggregate `post_count` dan `last_posted_at` pada product list/detail

## Phase 3: AI Reformat

- [x] Buat `AIService` dengan client OpenAI-compatible untuk OpenRouter, 9router, dan OpenCode
- [x] Buat prompt reformat sesuai `TRD.md`
- [x] Parse response JSON secara strict
- [x] Validasi ID, enum, angka, hashtag, dan field hasil AI
- [x] Pastikan `raw_text` tidak pernah tertimpa
- [x] Implementasikan `POST /api/ai/reformat`
- [x] Tolak request dengan lebih dari 10 `product_ids`
- [x] Simpan hasil AI langsung dan ubah status menjadi `reformatted`
- [x] Laporkan partial failure per produk
- [x] Reformat otomatis satu kali saat save produk yang memiliki raw text; gagal tetap tersimpan raw dan bisa di-retry dari detail
- [ ] Test response valid, invalid, timeout, dan partial failure dengan mock server

## Phase 4: Caption Generator

- [x] Implementasikan registry template:
  - [x] `direct_product`
  - [x] `keyword_recommendation`
  - [x] `problem_specific`
  - [x] `cheap_value`
- [x] Implementasikan placeholder replacement dan fallback field
- [x] Hilangkan baris placeholder yang nilainya kosong
- [x] Implementasikan format Rupiah dan angka
- [x] Implementasikan character count Unicode
- [x] Validasi hashtag maksimal 3
- [x] Implementasikan `POST /api/captions/generate`
- [x] Implementasikan `POST /api/captions/variations`
- [x] Implementasikan `GET /api/products/{id}/caption-variations`
- [x] Implementasikan `PATCH` dan `DELETE` caption variation
- [x] Test semua template, field kosong, hashtag, dan batas karakter
- [x] Pastikan final caption dan character count sudah mencakup hashtag

## Phase 5: Share dan Post Log

- [x] Implementasikan builder Twitter Web Intent URL
- [x] Implementasikan `GET /api/share/x?caption=...`
- [x] Buka share URL dengan `window.open()` dalam user gesture, bukan `fetch`
- [x] Implementasikan copy caption ke clipboard di frontend
- [x] Implementasikan form `Catat Posting`
- [x] Pastikan post log tidak menyimpan identitas akun X
- [x] Test URL encoding untuk spasi, newline, emoji, dan karakter khusus
- [x] Test posting berulang pada produk yang sama
- [ ] Test share saat clipboard ditolak atau popup diblokir browser

## Phase 6: Frontend Core

- [x] Buat Pinia `productStore`
- [x] Buat Pinia `captionStore`
- [x] Buat `HomeView.vue` dengan filter status, cluster, model, dan search
- [x] Tambahkan filter sumber data dan jenis barang, termasuk `Uncategorized`
- [x] Buat `ProductList.vue`
- [x] Buat `ProductParser.vue` untuk paste raw text
- [x] Buat `ProductForm.vue` untuk edit produk
- [x] Buat `BulkReformat.vue` dengan batas maksimal 10 produk
- [x] Pisahkan Chrome Extension menjadi `x-helper` dan `shopee-scraper` dengan icon berbeda
- [x] Implementasikan scraper Shopee berbasis halaman terbuka, DOM, metadata, dan response network
- [x] Buat `ProductDetailView.vue`
- [x] Tambahkan edit link affiliate serta tambah/hapus media dari product detail
- [x] Buat `CaptionGenerator.vue`
- [x] Buat `HashtagSelector.vue`
- [x] Buat `ShareButton.vue`
- [x] Buat `PostLogForm.vue`
- [x] Buat `PostLogsView.vue`
- [x] Tampilkan dan edit jenis barang produk dari detail; tampilkan label pada list
- [x] Setup base URL API
- [x] Tampilkan loading, success, dan error state
- [x] Rapikan list Bank Konten dengan judul/excerpt terpotong, metadata ringkas, dan detail responsive
- [ ] Verifikasi alur input sampai share melalui browser

## Phase 7: Integration dan Release MVP

- [x] Integrasikan semua frontend call ke backend
- [x] Build frontend dengan `npm run build`
- [x] Backend serve `frontend/dist`
- [x] Tambahkan graceful shutdown
- [x] Tambahkan database readiness dan migration failure handling
- [x] Jalankan seluruh stack melalui `docker compose up --build`
- [ ] Test persistence PostgreSQL setelah container restart
- [ ] Test migration lock, failure handling, dan update timestamp
- [ ] Test CORS, body limit, rate limit AI, dan validasi environment
- [ ] Test backup dan restore database lokal
- [x] Update `README.md` dengan setup, env, run, dan troubleshooting
- [x] Pastikan tidak ada secret di repository
- [ ] Tandai MVP selesai setelah memenuhi Definition of Done di `TRD.md`

## Phase 8: Roadmap Setelah MVP

- [ ] Share ke Threads melalui flow dan extension khusus Threads
- [ ] Buat master niche konten terpisah dari Jenis Barang; isi niche awal: Sukses & Kesuksesan, Fashion Pria, Hubungan / Relasi Pria Wanita, Gym, Lari & Exercise, Affiliate
- [ ] Tambahkan CRUD niche konten dan kemampuan user menambah niche baru
- [x] Buat bank konten X: simpan konten asli, URL/post ID, media, author, tanggal, query/sumber, dan status kurasi
- [x] Sediakan flow riset-first: pilih niche, gunakan query awal, lalu buka hasil Top X sebelum menyimpan konten
- [x] Breakdown riset manual menjadi kategori konten dan keyword per niche, lalu susun query dengan filter dasar
- [x] Tambahkan extension X Research untuk menangkap post/thread yang dipilih secara manual sebelum disimpan
- [ ] Riset posting X yang populer berdasarkan niche secara otomatis; MVP saat ini membuka pencarian Top X dan tetap memakai seleksi manual
- [ ] Simpan snapshot statistik konten dan sinyal kenapa konten dianggap populer
- [ ] Bersihkan konten tanpa menimpa konten asli
- [ ] Reformat hasil riset X dengan AI memakai prompt/parser khusus kanal X
- [ ] Buat varian konten per akun dan share lewat helper X yang sudah ada
- [ ] Hubungkan konten ke master Jenis Barang dengan relasi many-to-many; satu konten boleh punya banyak label
- [ ] Tambahkan collector X API sebagai otomasi opsional setelah alur manual stabil
- [ ] Riset konten Facebook yang populer berdasarkan niche, simpan dan kelompokkan hasilnya per niche
- [ ] Reformat hasil riset Facebook dengan AI memakai prompt/aturan khusus Facebook
- [ ] Share atau bantu posting konten terkurasi ke Facebook melalui flow/extension Facebook
- [ ] S3/CDN untuk media (pengganti local storage)
- [ ] Scheduling dan analytics
- [ ] Auth dan multi-user

---

## Catatan untuk AI Coder

- Gunakan `PRD.md` sebagai source of truth produk dan `TRD.md` sebagai source of truth teknis.
- Jangan mencampur implementasi roadmap Threads/Facebook dengan flow X MVP; kanal, riset niche, prompt AI, dan extension harus dipisahkan per platform.
- Data mentah harus selalu dipertahankan.
- AI boleh memperbaiki data, tetapi tidak boleh mengarang proof, harga, rating, atau urgency.
- Produk tidak berubah menjadi status `posted`; posting dicatat sebagai `post_logs` dan boleh berulang.
- User tetap meng-upload media dan menekan Post secara manual di X.
- API key/provider auth hanya berada di backend; model di UI hanya identifier pilihan.
- Setiap phase besar sebaiknya dibuat dalam commit terpisah.
- DeepSeek Flash boleh mengerjakan implementasi sesuai `AI-CODER-GUIDE.md`; Luna melakukan review akhir sebelum merge atau push.
- Jangan mengubah `OPENROUTER_MODEL` hanya karena coding model diganti.
