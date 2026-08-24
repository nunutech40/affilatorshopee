# TODO - AffiliatorShopee

Project: AffiliatorShopee  
Deskripsi: Web app pribadi untuk menyimpan produk affiliate Shopee, merapikan data dengan AI, membuat caption, dan membantu posting manual ke X.
Target: MVP yang bisa dipakai sendiri.  
Stack: Go, PostgreSQL, Vue 3 + Vite + Tailwind + Pinia, OpenRouter AI.
Deployment: Docker Compose di Orbstack.

Dokumen sumber:

- `PRD.md` - kebutuhan produk dan scope MVP
- `TRD.md` - arsitektur teknis, API, database schema, dan testing
- `Owner-Requirement.md` - konteks dan keputusan produk

---

## Phase 0: Keputusan dan Kontrak

- [x] Tetapkan stack final: Go, PostgreSQL, Vue 3, Vite, Tailwind, Pinia
- [x] Tetapkan scope MVP: produk, AI reformat, caption, hashtag, share X, post log
- [x] Keluarkan manajemen identitas akun X, Threads, Chrome Extension, dan media storage dari MVP
- [x] Tetapkan produk boleh diposting berulang kali
- [x] Tetapkan post log tidak menyimpan identitas akun X
- [x] Tetapkan AI reformat langsung menyimpan hasil tanpa preview
- [ ] Finalisasi API request/response dan HTTP status code
- [ ] Finalisasi template caption dan placeholder
- [ ] Pilih satu migration tool

## Phase 1: Project Setup

- [x] Repository Git tersedia di folder project
- [ ] Inisialisasi module Go: `go mod init github.com/nununugraha/affiliatorshopee`
- [ ] Buat struktur folder sesuai `TRD.md`
- [ ] Buat `Dockerfile` untuk Go backend
- [ ] Buat `docker-compose.yml` dengan service `db` dan `app`
- [ ] Buat `.env.example` tanpa secret nyata
- [ ] Buat project Vue dengan Vite di folder `web`
- [ ] Install Tailwind CSS, Vue Router, Pinia, dan test runner frontend
- [ ] Tambahkan endpoint `GET /healthz`
- [ ] Verifikasi `docker compose up --build` berhasil

## Phase 2: Database dan Backend Core

- [ ] Buat migration `001_create_products.sql`
- [ ] Buat migration `002_create_post_logs.sql`
- [ ] Buat migration `003_create_caption_variations.sql`
- [ ] Buat koneksi PostgreSQL dan menjalankan migration saat startup
- [ ] Buat model `Product`, `PostLog`, dan `CaptionVariation`
- [ ] Buat config loader untuk `PORT`, `DATABASE_URL`, `AI_API_KEY`, `OPENROUTER_MODEL`, dan `ENV`
- [ ] Implementasikan `ProductRepository`: Create, GetByID, List, Update, Delete
- [ ] Implementasikan `PostLogRepository`: Create, List
- [ ] Implementasikan `CaptionVariationRepository`: Create, List
- [ ] Implementasikan validasi input produk
- [ ] Setup router HTTP dan CORS localhost
- [ ] Implementasikan API products
- [ ] Implementasikan API post logs
- [ ] Verifikasi CRUD product dan post log dengan automated API tests

## Phase 3: AI Reformat

- [ ] Buat `AIService` dengan HTTP client ke OpenRouter
- [ ] Buat prompt reformat sesuai `TRD.md`
- [ ] Parse response JSON secara strict
- [ ] Validasi ID, enum, angka, hashtag, dan field hasil AI
- [ ] Pastikan `raw_text` tidak pernah tertimpa
- [ ] Implementasikan `POST /api/ai/reformat`
- [ ] Tolak request dengan lebih dari 20 `product_ids`
- [ ] Simpan hasil AI langsung dan ubah status menjadi `reformatted`
- [ ] Laporkan partial failure per produk
- [ ] Test response valid, invalid, timeout, dan partial failure dengan mock server

## Phase 4: Caption Generator

- [ ] Implementasikan registry template:
  - [ ] `direct_product`
  - [ ] `keyword_recommendation`
  - [ ] `problem_specific`
  - [ ] `cheap_value`
- [ ] Implementasikan placeholder replacement dan fallback field
- [ ] Hilangkan baris placeholder yang nilainya kosong
- [ ] Implementasikan format Rupiah dan angka
- [ ] Implementasikan character count Unicode
- [ ] Validasi hashtag maksimal 3
- [ ] Implementasikan `POST /api/captions/generate`
- [ ] Implementasikan `POST /api/captions/variations`
- [ ] Implementasikan `GET /api/products/{id}/caption-variations`
- [ ] Test semua template, field kosong, hashtag, dan batas karakter

## Phase 5: Share dan Post Log

- [ ] Implementasikan builder Twitter Web Intent URL
- [ ] Implementasikan `GET /api/share/x?caption=...`
- [ ] Implementasikan copy caption ke clipboard di frontend
- [ ] Implementasikan form `Catat Posting`
- [ ] Pastikan post log tidak menyimpan identitas akun X
- [ ] Test URL encoding untuk spasi, newline, emoji, dan karakter khusus
- [ ] Test posting berulang pada produk yang sama

## Phase 6: Frontend Core

- [ ] Buat Pinia `productStore`
- [ ] Buat Pinia `captionStore`
- [ ] Buat `HomeView.vue` dengan filter status, cluster, model, dan search
- [ ] Buat `ProductList.vue`
- [ ] Buat `ProductParser.vue` untuk paste raw text
- [ ] Buat `ProductForm.vue` untuk edit produk
- [ ] Buat `BulkReformat.vue` dengan batas maksimal 20 produk
- [ ] Buat `ProductDetailView.vue`
- [ ] Buat `CaptionGenerator.vue`
- [ ] Buat `HashtagSelector.vue`
- [ ] Buat `ShareButton.vue`
- [ ] Buat `PostLogForm.vue`
- [ ] Buat `PostLogsView.vue`
- [ ] Setup base URL API
- [ ] Tampilkan loading, success, dan error state
- [ ] Verifikasi alur input sampai share melalui browser

## Phase 7: Integration dan Release MVP

- [ ] Integrasikan semua frontend call ke backend
- [ ] Build frontend dengan `npm run build`
- [ ] Backend serve `web/dist`
- [ ] Tambahkan graceful shutdown
- [ ] Tambahkan database readiness dan migration failure handling
- [ ] Jalankan seluruh stack melalui `docker compose up --build`
- [ ] Test persistence PostgreSQL setelah container restart
- [ ] Test backup dan restore database lokal
- [ ] Update `README.md` dengan setup, env, run, dan troubleshooting
- [ ] Pastikan tidak ada secret di repository
- [ ] Tandai MVP selesai setelah memenuhi Definition of Done di `TRD.md`

## Phase 8: Roadmap Setelah MVP

- [ ] Chrome Extension untuk membantu paste caption
- [ ] Integrasi Threads
- [ ] Media storage backend atau S3/CDN
- [ ] Scheduling dan analytics
- [ ] Auth dan multi-user

---

## Catatan untuk AI Coder

- Gunakan `PRD.md` sebagai source of truth produk dan `TRD.md` sebagai source of truth teknis.
- Jangan membuat fitur Threads, extension, atau media storage sebelum Phase 8.
- Data mentah harus selalu dipertahankan.
- AI boleh memperbaiki data, tetapi tidak boleh mengarang proof, harga, rating, atau urgency.
- Produk tidak berubah menjadi status `posted`; posting dicatat sebagai `post_logs` dan boleh berulang.
- User tetap meng-upload media dan menekan Post secara manual di X.
- API key OpenRouter hanya berada di backend.
- Setiap phase besar sebaiknya dibuat dalam commit terpisah.
