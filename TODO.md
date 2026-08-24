# TODO — AffiliatorShopee

Project: AffiliatorShopee  
Deskripsi: Web app pribadi untuk menyimpan produk affiliate Shopee, generate caption, dan membantu posting ke banyak akun/platform.  
Target: MVP yang bisa dipakai sendiri.  
Stack: Go, PostgreSQL, Vue 3 + Vite + Tailwind + Pinia, Chrome Extension Manifest V3, OpenRouter AI.  
Deployment: Docker Compose di Orbstack.

Dokumen pendukung:
- `PRD.md` — kebutuhan produk, flowchart, acceptance criteria
- `TRD.md` — arsitektur teknis, API spec, database schema, stack
- `Owner-Requirement.md` — hasil brainstorming awal
- `MVP-Plan.md` — rencana MVP sebelumnya

---

## Phase 1: Project Setup

- [ ] Inisialisasi repo Git di folder `AffiliatorShopee/`
- [ ] Copy `PRD.md`, `TRD.md`, `Owner-Requirement.md`, `MVP-Plan.md` ke root repo
- [ ] Backend: `go mod init github.com/nununugraha/affiliatorshopee`
- [ ] Backend: buat struktur folder `cmd/api`, `internal/config`, `internal/db/migrations`, `internal/handler`, `internal/model`, `internal/repository`, `internal/service`, `internal/storage`, `web`, `extension`
- [ ] Database: buat file migrasi `001_create_products.sql`
- [ ] Database: buat file migrasi `002_create_accounts.sql`
- [ ] Database: buat file migrasi `003_create_post_logs.sql`
- [ ] Database: buat file migrasi `004_create_caption_variations.sql`
- [ ] Docker: buat `Dockerfile` untuk Go backend
- [ ] Docker: buat `docker-compose.yml` dengan service `db` (PostgreSQL) dan `app` (Go)
- [ ] Environment: buat `.env.example`
- [ ] Frontend: `npm create vite@latest web -- --template vue`
- [ ] Frontend: install Tailwind CSS
- [ ] Frontend: install Vue Router dan Pinia
- [ ] Frontend: setup struktur folder `web/src/components`, `web/src/views`, `web/src/router`, `web/src/stores`
- [ ] Verifikasi: `docker-compose up --build` berjalan tanpa error

## Phase 2: Backend Core

- [ ] Model: buat `Product` struct
- [ ] Model: buat `Account` struct
- [ ] Model: buat `PostLog` struct
- [ ] Model: buat `CaptionVariation` struct
- [ ] Config: load environment variables `PORT`, `DATABASE_URL`, `STORAGE_PATH`, `AI_API_KEY`, `OPENROUTER_MODEL`
- [ ] DB: buat koneksi PostgreSQL
- [ ] DB: jalankan migrasi saat startup
- [ ] Repository: implementasi `ProductRepository` (Create, GetByID, List, Update, Delete)
- [ ] Repository: implementasi `AccountRepository` (Create, List)
- [ ] Repository: implementasi `PostLogRepository` (Create, ListByProduct)
- [ ] Service: implementasi `ProductService`
- [ ] Handler: implementasi `ProductHandler`
- [ ] Handler: implementasi `AccountHandler`
- [ ] Router: setup HTTP router (chi atau gin)
- [ ] Router: tambah endpoint dasar dan CORS
- [ ] Verifikasi: endpoint `GET /api/products` dan `POST /api/products` bisa di-test via curl/Postman

## Phase 3: AI Reformat

- [ ] Service: buat `AIService` dengan HTTP client ke OpenRouter
- [ ] Service: prompt reformat produk sesuai TRD section 8.2
- [ ] Service: parse response JSON dari OpenRouter
- [ ] Service: validasi hasil AI sebelum update database
- [ ] Handler: buat endpoint `POST /api/ai/reformat`
- [ ] Handler: validasi maksimal 20 `product_ids`
- [ ] Handler: update field produk dan ubah status ke `reformatted`
- [ ] Verifikasi: bulk reformat 1–20 produk berhasil

## Phase 4: Caption Generator

- [ ] Service: buat `CaptionService` dengan template engine
- [ ] Service: load template Direct Product, Keyword + Recommendation, Problem-specific, Cheap/Value
- [ ] Service: implementasi placeholder replacement
- [ ] Service: helper format harga dan join hashtag
- [ ] Handler: buat endpoint `POST /api/captions/generate`
- [ ] Handler: buat endpoint `POST /api/captions/variations`
- [ ] Verifikasi: generate caption dari produk dummy menghasilkan output sesuai template

## Phase 5: Share & Post Log

- [ ] Service: buat `ShareService` untuk generate Twitter Web Intent URL
- [ ] Handler: buat endpoint `GET /api/share/x?caption=...` yang redirect ke Twitter intent
- [ ] Handler: buat endpoint `POST /api/post-logs`
- [ ] Verifikasi: share ke X membuka tab dengan caption ter-encode
- [ ] Verifikasi: post log tersimpan di database

## Phase 6: Chrome Extension

- [ ] Extension: buat `extension/manifest.json` Manifest V3
- [ ] Extension: buat `extension/background.js`
- [ ] Extension: buat `extension/content.js` untuk deteksi textarea X/Threads dan paste caption
- [ ] Extension: buat `extension/popup.html` sederhana
- [ ] Verifikasi: extension bisa di-load unpacked dan paste caption ke composer X

## Phase 7: Frontend

- [ ] Store: buat Pinia store `productStore`
- [ ] Store: buat Pinia store `accountStore`
- [ ] Store: buat Pinia store `captionStore`
- [ ] View: buat `HomeView.vue` — daftar produk dengan filter status
- [ ] Component: buat `ProductList.vue`
- [ ] Component: buat `ProductForm.vue` untuk tambah produk
- [ ] Component: buat `ProductParser.vue` textarea paste raw_text
- [ ] Component: buat `BulkReformat.vue` checkbox + tombol bulk reformat
- [ ] View: buat `ProductDetailView.vue`
- [ ] Component: buat `CaptionGenerator.vue`
- [ ] Component: buat `HashtagSelector.vue`
- [ ] Component: buat `ShareButton.vue`
- [ ] View: buat `AccountsView.vue`
- [ ] View: buat `PostLogsView.vue`
- [ ] Config: setup base URL API di frontend
- [ ] Verifikasi: frontend bisa CRUD produk, reformat AI, generate caption, dan share

## Phase 8: Integration & Polish

- [ ] Integrasi: frontend call backend untuk semua fitur utama
- [ ] Build: frontend `npm run build` menghasilkan folder `web/dist`
- [ ] Backend: serve static `web/dist`
- [ ] Docker: `docker-compose up --build` berhasil menjalankan full stack
- [ ] Testing: CRUD produk
- [ ] Testing: AI reformat bulk max 20
- [ ] Testing: generate caption dan variasi
- [ ] Testing: share ke X
- [ ] Testing: extension paste caption
- [ ] Dokumentasi: update `README.md` dengan cara run

---

## Catatan untuk AI Coder

- Bacalah `PRD.md` dan `TRD.md` sebelum memulai setiap phase.
- Setiap phase sebaiknya di-commit ke Git.
- Gunakan OpenRouter dengan model default `google/gemini-flash-1.5` untuk fitur AI reformat.
- Jangan bangun fitur di luar scope MVP tanpa persetujuan owner.
- Media (gambar/video) tetap di-upload manual oleh user ke platform target.
- Tidak ada auto-posting penuh untuk menghindari deteksi bot.
