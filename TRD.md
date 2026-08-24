---
type: trd
project: AffiliatorShopee
status: aligned-with-prd
---

# Technical Requirements Document - AffiliatorShopee

## 1. Keputusan Teknis

| Layer | Teknologi |
|---|---|
| Backend | Go |
| Database | PostgreSQL |
| Frontend | Vue 3 + Vite + Tailwind CSS |
| HTTP Client | Fetch API |
| Routing | Vue Router |
| State | Pinia |
| AI | OpenRouter |
| Deployment | Docker Compose di Orbstack |

MVP hanya mencakup web app pribadi dan alur posting manual ke X. Web app tidak login, memilih, atau menyimpan identitas akun X. Tidak ada autentikasi operator, media storage backend, Threads, atau Chrome Extension pada MVP.

## 2. Arsitektur Sistem

```text
Browser (Vue 3 App)
        |
        | HTTP/JSON
        v
Go Backend API
   |             |
   | SQL         | HTTPS API
   v             v
PostgreSQL    OpenRouter
```

User tetap meng-upload gambar/video langsung di X dan menekan tombol Post secara manual. Backend hanya menyimpan data produk, caption, dan catatan posting.

## 3. Struktur Folder

```text
AffiliatorShopee/
├── cmd/
│   └── api/
│       └── main.go
├── internal/
│   ├── config/
│   │   └── config.go
│   ├── db/
│   │   ├── connection.go
│   │   └── migrations/
│   │       ├── 001_create_products.sql
│   │       ├── 002_create_post_logs.sql
│   │       └── 003_create_caption_variations.sql
│   ├── handler/
│   │   ├── product_handler.go
│   │   ├── ai_handler.go
│   │   ├── caption_handler.go
│   │   └── post_log_handler.go
│   ├── model/
│   │   ├── product.go
│   │   ├── post_log.go
│   │   └── caption_variation.go
│   ├── repository/
│   │   ├── product_repo.go
│   │   ├── post_log_repo.go
│   │   └── caption_variation_repo.go
│   └── service/
│       ├── product_service.go
│       ├── ai_service.go
│       ├── caption_service.go
│       ├── share_service.go
│       └── post_log_service.go
├── web/
│   ├── index.html
│   ├── package.json
│   └── src/
│       ├── main.js
│       ├── App.vue
│       ├── router/index.js
│       ├── stores/productStore.js
│       ├── stores/captionStore.js
│       ├── components/
│       │   ├── ProductList.vue
│       │   ├── ProductForm.vue
│       │   ├── ProductParser.vue
│       │   ├── BulkReformat.vue
│       │   ├── CaptionGenerator.vue
│       │   ├── HashtagSelector.vue
│       │   ├── ShareButton.vue
│       │   └── PostLogForm.vue
│       └── views/
│           ├── HomeView.vue
│           ├── ProductDetailView.vue
│           └── PostLogsView.vue
├── docker-compose.yml
├── Dockerfile
├── go.mod
├── go.sum
├── .env.example
├── README.md
├── PRD.md
├── TRD.md
└── TODO.md
```

## 4. Database Schema

### 4.1 Tabel products

```sql
CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE products (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    raw_text TEXT NOT NULL,
    product_name VARCHAR(255),
    shopee_link TEXT NOT NULL,
    image_url TEXT,
    image_urls TEXT[],
    video_url TEXT,
    normal_price INTEGER CHECK (normal_price IS NULL OR normal_price >= 0),
    sale_price INTEGER CHECK (sale_price IS NULL OR sale_price >= 0),
    discount_percent INTEGER CHECK (discount_percent IS NULL OR discount_percent BETWEEN 0 AND 100),
    rating NUMERIC(2,1) CHECK (rating IS NULL OR rating BETWEEN 0 AND 5),
    sold_count VARCHAR(50),
    review_count VARCHAR(50),
    keyword VARCHAR(255),
    problem VARCHAR(255),
    cluster VARCHAR(100),
    content_model VARCHAR(20) CHECK (content_model IN ('capture', 'cheap')),
    capture_angle VARCHAR(20) CHECK (capture_angle IN ('search', 'reply', 'trend', 'problem')),
    CHECK (capture_angle IS NULL OR content_model = 'capture'),
    benefit_1 VARCHAR(255),
    benefit_2 VARCHAR(255),
    benefit_3 VARCHAR(255),
    urgency VARCHAR(255),
    caption_template VARCHAR(50) NOT NULL DEFAULT 'direct_product'
        CHECK (caption_template IN ('direct_product', 'keyword_recommendation', 'problem_specific', 'cheap_value')),
    hashtag_pool TEXT[],
    notes TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'raw'
        CHECK (status IN ('raw', 'reformatted', 'ready')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CHECK (sale_price IS NULL OR normal_price IS NULL OR sale_price <= normal_price),
    CHECK (hashtag_pool IS NULL OR cardinality(hashtag_pool) <= 3)
);

CREATE INDEX idx_products_status ON products(status);
CREATE INDEX idx_products_cluster ON products(cluster);
CREATE INDEX idx_products_content_model ON products(content_model);
CREATE INDEX idx_products_created_at ON products(created_at DESC);

CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER products_set_updated_at
BEFORE UPDATE ON products
FOR EACH ROW EXECUTE FUNCTION set_updated_at();
```

`raw_text` selalu dipertahankan sebagai sumber asli. Hasil AI mengisi atau memperbarui field terstruktur, tetapi tidak boleh menghapus data mentah.

`image_url`, `image_urls`, dan `video_url` pada MVP hanya menyimpan URL eksternal yang diberikan user. Backend tidak mengunduh, mem-proxy, atau menyimpan file media tersebut.

`content_model` menggantikan field `model` lama agar dapat mewakili dua pendekatan konten MVP: `capture` dan `cheap`. `capture_angle` hanya digunakan bila modelnya `capture`. Model `branded` ditunda ke roadmap.

Status `posted` tidak disimpan pada produk. Satu produk boleh dicatat dan diposting berulang kali; riwayatnya disimpan di `post_logs`.

Transisi status:

- Product baru: `raw`.
- AI reformat: `raw` menjadi `reformatted`.
- Edit manual dengan data lengkap: `raw` atau `reformatted` menjadi `ready`.
- Posting tidak mengubah status product.

Field minimum untuk status `ready` adalah `product_name`, `shopee_link`, `cluster`, `content_model`, dan minimal satu dari `benefit_1`, `benefit_2`, atau `benefit_3`. Jika `content_model` adalah `capture`, `capture_angle` juga wajib diisi. Caption hanya dapat dibuat untuk status `reformatted` atau `ready`.

### 4.2 Tabel post_logs

```sql
CREATE TABLE post_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    platform VARCHAR(30) NOT NULL DEFAULT 'x' CHECK (platform = 'x'),
    caption TEXT NOT NULL,
    hashtags TEXT[],
    notes TEXT,
    posted_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CHECK (hashtags IS NULL OR cardinality(hashtags) <= 3)
);

CREATE INDEX idx_post_logs_product_id ON post_logs(product_id);
CREATE INDEX idx_post_logs_posted_at ON post_logs(posted_at DESC);
```

Identitas akun X tidak disimpan. Riwayat dapat berisi beberapa posting untuk produk dan platform yang sama.

### 4.3 Tabel caption_variations

```sql
CREATE TABLE caption_variations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    label VARCHAR(50) NOT NULL,
    template VARCHAR(50) NOT NULL,
    caption TEXT NOT NULL,
    hashtags TEXT[],
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CHECK (cardinality(hashtags) <= 3)
);

CREATE INDEX idx_caption_variations_product_id ON caption_variations(product_id);

CREATE TRIGGER caption_variations_set_updated_at
BEFORE UPDATE ON caption_variations
FOR EACH ROW EXECUTE FUNCTION set_updated_at();
```

## 5. API Contract

### 5.1 Format Response

Success:

```json
{
  "success": true,
  "data": {},
  "error": null
}
```

Error:

```json
{
  "success": false,
  "data": null,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Request tidak valid",
    "fields": {
      "shopee_link": "wajib diisi"
    }
  }
}
```

Gunakan HTTP status code yang sesuai: `200` untuk sukses, `201` untuk create, `204` untuk delete tanpa body, `400` untuk input invalid, `404` untuk data tidak ditemukan, `409` untuk conflict, dan `500` untuk error internal.

### 5.2 Products API

```http
GET /api/products?cluster=&content_model=&status=&search=&page=1&limit=20
```

Default `page` adalah 1 dan default `limit` adalah 20. `limit` minimum 1 dan maksimum 100. `search` melakukan pencarian case-insensitive pada `product_name`, `keyword`, `cluster`, dan `raw_text`.

List response harus berisi `items`, `page`, `limit`, dan `total`. Setiap item juga berisi `post_count` dan `last_posted_at` yang dihitung dari `post_logs`.

```http
POST /api/products
Content-Type: application/json
```

Request minimum:

```json
{
  "raw_text": "data copas Shopee yang masih berantakan",
  "shopee_link": "https://shopee.co.id/...",
  "image_url": "https://...",
  "notes": ""
}
```

Produk baru selalu disimpan dengan status `raw`.

Response create dan detail menggunakan object product di dalam `data`. Contoh ringkas:

```json
{
  "success": true,
  "data": {
    "id": "uuid",
    "raw_text": "data copas Shopee",
    "shopee_link": "https://shopee.co.id/...",
    "status": "raw",
    "post_count": 0,
    "last_posted_at": null
  },
  "error": null
}
```

```http
GET /api/products/{id}
PATCH /api/products/{id}
DELETE /api/products/{id}
```

`PATCH` menerima allowlist field produk terstruktur: `product_name`, `shopee_link`, `image_url`, `image_urls`, `video_url`, harga, rating, `sold_count`, `review_count`, `keyword`, `problem`, `cluster`, `content_model`, `capture_angle`, benefit, `urgency`, `caption_template`, `hashtag_pool`, `notes`, dan `status`. Field nullable dapat dihapus dengan nilai JSON `null` kecuali `raw_text`, yang immutable untuk menjaga sumber asli.

Status hanya boleh berubah melalui transisi `raw -> reformatted`, `raw -> ready`, atau `reformatted -> ready`. Request yang mencoba mengubah `ready` kembali ke `raw` atau `reformatted` ditolak kecuali ada endpoint reset khusus di masa depan.

Delete mengembalikan `204` tanpa body. Get, patch, dan create mengembalikan product object dalam format response standar.

`post_count` dan `last_posted_at` adalah field read-only hasil aggregate `post_logs`, bukan kolom pada `products`.

Nilai `null` berarti data memang belum tersedia; nilai `0` hanya boleh dipakai untuk angka yang benar-benar diketahui bernilai nol. Backend wajib memvalidasi panjang string, array, URL `http/https`, enum template/model, hubungan harga, dan field minimum sebelum menerima status `ready`.

### 5.3 AI Reformat API

```http
POST /api/ai/reformat
Content-Type: application/json
```

Request:

```json
{
  "product_ids": ["uuid1", "uuid2"]
}
```

Aturan:

- `product_ids` wajib berisi 1-20 ID.
- Produk yang diproses harus berstatus `raw`.
- Hasil AI langsung divalidasi dan disimpan ke database.
- Status yang berhasil diproses berubah menjadi `reformatted`.
- `raw_text` tidak pernah ditimpa.
- Jika sebagian produk gagal, response harus melaporkan hasil per produk.
- Tidak ada preview atau approval step di antara AI dan penyimpanan.

ID duplikat dalam request ditolak dengan `DUPLICATE_PRODUCT_ID`. ID yang tidak ditemukan atau bukan `raw` masuk ke daftar `failed` tanpa mengubah produk tersebut. Setiap product diproses dan disimpan secara atomic; kegagalan satu product tidak membatalkan product lain yang valid.

Response:

```json
{
  "success": true,
  "data": {
    "processed": [
      { "product_id": "uuid1", "status": "reformatted" }
    ],
    "failed": [
      {
        "product_id": "uuid2",
        "code": "AI_INVALID_OUTPUT",
        "message": "Hasil AI tidak lolos validasi"
      }
    ]
  },
  "error": null
}
```

### 5.4 Caption API

```http
POST /api/captions/generate
Content-Type: application/json
```

Request:

```json
{
  "product_id": "uuid",
  "template": "direct_product",
  "hashtags": ["#BajuAnak"]
}
```

Response:

```json
{
  "success": true,
  "data": {
    "caption": "...",
    "character_count": 240,
    "over_limit": false,
    "template": "direct_product"
  },
  "error": null
}
```

Caption dapat dibuat untuk produk berstatus `reformatted` atau `ready`. Hashtag harus berjumlah 0-3 item setelah normalisasi.

Field `caption` selalu merupakan final text yang sudah mengandung hashtag. Array `hashtags` tetap dikembalikan terpisah untuk kebutuhan UI dan post log. Character count dihitung dari final text yang sama.

```http
POST /api/captions/variations
Content-Type: application/json
```

Request:

```json
{
  "product_id": "uuid",
  "template": "direct_product",
  "count": 3,
  "hashtags": ["#BajuAnak"]
}
```

`count` dibatasi 2-3. Variasi dibuat oleh template service, disimpan di `caption_variations`, dan tidak terikat ke identitas akun X.

Response variations mengembalikan array variation object. Setiap variation menyimpan caption final yang sudah mengandung hashtag.

```http
GET /api/products/{id}/caption-variations
PATCH /api/caption-variations/{id}
DELETE /api/caption-variations/{id}
```

`PATCH` hanya menerima `label`, `caption`, dan `hashtags`. Update variation mengubah `updated_at`. Delete mengembalikan `204`.

Create, list, dan patch mengembalikan variation object dalam `data`. `product_id` pada patch/delete harus sesuai dengan variation yang dituju; variation dari produk lain tidak boleh dapat diubah.

### 5.5 Share API

```http
GET /api/share/x?caption={encoded_caption}
```

Endpoint melakukan redirect ke:

```text
https://twitter.com/intent/tweet?text={url_encoded_caption}
```

Frontend juga menyalin caption ke clipboard sebagai fallback. Caption harus di-encode dengan URL query encoder, bukan konkatenasi string mentah.

`ShareButton` harus membuka URL intent melalui `window.open()` atau navigasi langsung di dalam user gesture, bukan memanggil endpoint redirect melalui `fetch`. Gunakan `noopener,noreferrer` dan tetap tampilkan tombol copy manual bila popup atau clipboard ditolak browser.

### 5.6 Post Logs API

```http
POST /api/post-logs
Content-Type: application/json
```

Request:

```json
{
  "product_id": "uuid",
  "platform": "x",
  "caption": "...",
  "hashtags": ["#BajuAnak"],
  "notes": ""
}
```

```http
GET /api/post-logs?product_id={uuid}&page=1&limit=20
```

Create response mengembalikan post log object di dalam `data`; list response mengembalikan `items`, `page`, `limit`, dan `total`. Membuat post log tidak mengubah status produk. Endpoint ini hanya mencatat konfirmasi manual user.

## 6. Caption Template Engine

### 6.1 Template Registry

Template MVP yang wajib tersedia:

- `direct_product`
- `keyword_recommendation`
- `problem_specific`
- `cheap_value`

Setiap template memiliki daftar placeholder wajib dan opsional. Placeholder dengan nilai kosong harus dihilangkan bersama label atau barisnya, bukan menghasilkan teks `null` atau placeholder mentah.

Format canonical MVP:

```text
[direct_product]
Cari {product_name}?

✅ {benefit_1}
✅ {benefit_2}
✅ {benefit_3}
✅ {rating}⭐️
✅ {sold_count} terjual
{urgency}

Cek di sini 👇
{shopee_link}

{hashtags}

[keyword_recommendation]
Lagi cari {keyword}?

Ini salah satu yang gue shortlist.

Kenapa masuk shortlist:
✅ {benefit_1}
✅ {benefit_2}
✅ {benefit_3}
✅ {rating}⭐️ | {sold_count} terjual

Harganya masih masuk akal.

Cek:
{shopee_link}

{hashtags}

[problem_specific]
Punya masalah {problem}?

{product_name} ini bisa jadi salah satu opsi.

✅ {benefit_1}
✅ {benefit_2}
✅ {benefit_3}

Cek detailnya:
{shopee_link}

{hashtags}

[cheap_value]
Cari {product_name} murah tapi nggak murahan?

Yang ini menarik 👀

✅ {benefit_1}
✅ {benefit_2}
✅ {rating}⭐️
✅ {sold_count} terjual
✅ {sale_price}

Kalau budget lo sekitar {sale_price}, ini worth checking.

👇
{shopee_link}

{hashtags}
```

`problem_specific` membutuhkan field `problem`. Template lain memakai fallback `keyword -> product_name`. `benefit_2`, `benefit_3`, rating, sold count, urgency, dan harga boleh kosong; renderer menghilangkan baris yang tidak memiliki data.

### 6.2 Placeholder

| Placeholder | Sumber |
|---|---|
| `{product_name}` | `product_name` |
| `{keyword}` | `keyword`, fallback ke `product_name` |
| `{problem}` | `problem` |
| `{benefit_1}` | `benefit_1` |
| `{benefit_2}` | `benefit_2` |
| `{benefit_3}` | `benefit_3` |
| `{rating}` | `rating` |
| `{sold_count}` | `sold_count` |
| `{review_count}` | `review_count` |
| `{normal_price}` | `normal_price`, format Rupiah |
| `{sale_price}` | `sale_price`, format Rupiah |
| `{discount_percent}` | `discount_percent` |
| `{urgency}` | `urgency`, hanya jika ada bukti |
| `{shopee_link}` | `shopee_link` |
| `{hashtags}` | hashtag pilihan user |

Tidak ada placeholder `proof` terpisah. Proof dibentuk hanya dari rating, jumlah terjual, dan jumlah review yang tersedia.

### 6.3 Helper

- `format_price(39999)` menghasilkan `Rp39.999`.
- `format_number(1000000)` menghasilkan `1.000.000`.
- `join_hashtags(array)` menghasilkan string hashtag dengan spasi.
- Character count menggunakan jumlah rune/karakter Unicode, bukan jumlah byte.

## 7. AI Reformat Service

### 7.1 Provider

- Endpoint: `https://openrouter.ai/api/v1/chat/completions`
- Model wajib diambil dari `OPENROUTER_MODEL`; tidak ada model provider yang di-hardcode oleh backend.
- API key hanya dibaca backend dari `AI_API_KEY`.
- HTTP client memiliki timeout dan tidak mengirim API key ke frontend.
- `raw_text` diperlakukan sebagai untrusted content, dikirim di dalam delimiter prompt, dan tidak ditulis ke log aplikasi.
- Kirim hanya data yang diperlukan untuk reformat; jangan mengirim secret atau data pribadi ke provider AI.

### 7.2 Output AI

AI diminta mengembalikan JSON array dengan field:

```json
[
  {
    "product_id": "uuid",
    "product_name": "...",
    "normal_price": 0,
    "sale_price": 0,
    "discount_percent": 0,
    "rating": 0.0,
    "sold_count": "...",
    "review_count": "...",
    "keyword": "...",
    "problem": "...",
    "cluster": "...",
    "content_model": "cheap",
    "capture_angle": null,
    "benefit_1": "...",
    "benefit_2": "...",
    "benefit_3": "...",
    "urgency": "...",
    "hashtag_pool": ["#..."]
  }
]
```

Prompt wajib menginstruksikan AI untuk:

- Tidak mengarang harga, rating, jumlah terjual, review, urgency, atau benefit yang tidak didukung data.
- Mengisi `null` jika data tidak tersedia.
- Memilih `content_model` dari `capture` atau `cheap`.
- Mengisi `capture_angle` hanya untuk model `capture`.
- Menghasilkan JSON valid tanpa markdown atau penjelasan tambahan.

### 7.3 Validasi Sebelum Save

- JSON harus dapat diparse.
- Semua `product_id` response harus cocok dengan request.
- Tidak boleh ada ID duplikat atau ID di luar request.
- Harga tidak boleh negatif.
- `discount_percent` harus 0-100.
- `rating` harus 0-5.
- `content_model` dan `capture_angle` harus sesuai enum.
- Hashtag maksimal 3 dan harus dinormalisasi.
- Produk hanya di-update setelah response lolos validasi.
- Request bersamaan tidak boleh menimpa hasil yang lebih baru; update memakai conditional status atau row lock.
- Provider timeout dan error tidak boleh mengubah data produk.

## 8. Frontend

### 8.1 Routes

| Route | Fungsi |
|---|---|
| `/` | Dashboard produk dan filter |
| `/products/new` | Paste dan simpan produk raw |
| `/products/:id` | Detail, edit, reformat, caption, share |
| `/post-logs` | Riwayat posting sederhana |

### 8.2 Komponen

- `ProductList.vue` - daftar produk dengan filter.
- `ProductForm.vue` - edit field produk.
- `ProductParser.vue` - textarea data Shopee mentah.
- `BulkReformat.vue` - pilih maksimal 20 produk dan panggil AI.
- `CaptionGenerator.vue` - pilih template dan generate caption.
- `HashtagSelector.vue` - pilih 0-3 hashtag.
- `ShareButton.vue` - copy clipboard dan buka X.
- `PostLogForm.vue` - catat posting setelah user selesai posting di X.

MVP tidak memiliki flow pemilihan akun; akun X yang digunakan mengikuti session login di browser user.

### 8.3 State

- `productStore` - list, filter, CRUD, dan AI reformat.
- `captionStore` - caption aktif dan variasi.

## 9. Docker Compose

Compose MVP harus menyediakan PostgreSQL dan app Go. Port hanya di-bind ke localhost karena MVP tidak memiliki autentikasi.

```yaml
services:
  db:
    image: postgres:15
    environment:
      POSTGRES_USER: ${POSTGRES_USER:-affiliator}
      POSTGRES_PASSWORD: ${POSTGRES_PASSWORD:?set POSTGRES_PASSWORD}
      POSTGRES_DB: ${POSTGRES_DB:-affiliator}
    volumes:
      - postgres_data:/var/lib/postgresql/data
    ports:
      - "127.0.0.1:5432:5432"
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U ${POSTGRES_USER:-affiliator} -d ${POSTGRES_DB:-affiliator}"]
      interval: 5s
      timeout: 5s
      retries: 10

  app:
    build: .
    environment:
      DATABASE_URL: ${DATABASE_URL}
      PORT: 8080
      AI_API_KEY: ${AI_API_KEY}
      OPENROUTER_MODEL: ${OPENROUTER_MODEL}
      ENV: ${ENV:-development}
    ports:
      - "127.0.0.1:8080:8080"
    depends_on:
      db:
        condition: service_healthy

volumes:
  postgres_data:
```

## 10. Environment Variables

```bash
PORT=8080
DATABASE_URL=postgres://affiliator:change-me@db:5432/affiliator?sslmode=disable
POSTGRES_USER=affiliator
POSTGRES_PASSWORD=change-me
POSTGRES_DB=affiliator
AI_API_KEY=sk-or-v1-...
OPENROUTER_MODEL=replace-with-tested-openrouter-model
ENV=development
```

Jangan commit file `.env` atau API key ke repository.

## 11. Security dan Reliability

- MVP hanya untuk local/private use dan tidak memiliki auth.
- Bind port API dan database ke `127.0.0.1`.
- Jangan expose service ke internet sebelum auth dan authorization tersedia.
- API key hanya berada di backend.
- Validasi ukuran body, panjang string, URL, enum, angka, dan jumlah hashtag.
- CORS hanya mengizinkan origin development lokal yang ditentukan, misalnya `http://localhost:5173`.
- Batasi request AI dengan rate limit dan concurrency limit agar biaya terkendali.
- Validasi environment wajib saat startup dan fail-fast bila `DATABASE_URL` atau `AI_API_KEY` tidak valid.
- Gunakan timeout, error handling, dan logging terstruktur untuk OpenRouter.
- Jangan menyimpan raw product text di log aplikasi secara default.
- Semua query database menggunakan parameter binding.
- Caption di-encode dengan aman sebelum dipakai pada URL intent.

## 12. Error Handling

Backend harus:

- Mengembalikan format error standar.
- Tidak membocorkan detail error internal atau secret.
- Membedakan validation error, not found, provider error, dan database error.
- Mengembalikan hasil per produk untuk bulk AI agar kegagalan sebagian terlihat.

Frontend harus:

- Menampilkan error yang actionable.
- Disable tombol selama request berjalan.
- Menampilkan status proses AI per produk.
- Menyediakan fallback copy-paste manual bila share intent gagal.

## 13. Testing Strategy

| Jenis | Scope | Tool |
|---|---|---|
| Unit | Template, validator, service, URL builder | Go testing |
| Repository integration | Migration dan query PostgreSQL | Go testing + test database |
| API integration | Semua endpoint dan status code | Go testing / httptest |
| AI contract | JSON valid/invalid, timeout, partial failure | Mock HTTP server |
| Frontend | Form, loading/error state, caption flow | Vitest |
| End-to-end | Input sampai share dan post log | Browser test/manual |

Test minimum:

- CRUD product dan validasi field.
- Status `raw` ke `reformatted` ke `ready`.
- AI bulk 1-20 produk dan menolak lebih dari 20.
- Raw text tetap utuh setelah reformat.
- AI malformed response, timeout, dan partial failure.
- Template dengan field kosong.
- Character count Unicode dan caption di atas batas.
- Hashtag 0-3 dan penolakan hashtag ke-4.
- URL encoding untuk spasi, emoji, newline, dan karakter khusus.
- Posting berulang pada produk yang sama.
- Migration dan database readiness Docker.

## 14. Migration Files

- `001_create_products.sql`
- `002_create_post_logs.sql`
- `003_create_caption_variations.sql`

Gunakan `golang-migrate/migrate` untuk menjalankan migration SQL. Migration harus memiliki version tracking, dijalankan sebelum app menerima traffic, dan membuat startup gagal bila migration tidak berhasil.

## 15. Roadmap Setelah MVP

- Chrome Extension untuk membantu paste caption.
- Integrasi Threads.
- Media storage di backend atau S3/CDN.
- Scheduling dan analytics.
- Auth, multi-user, dan admin panel.

## 16. Definition of Done MVP

MVP dianggap selesai bila user dapat:

1. Menyimpan data Shopee mentah.
2. Memilih maksimal 20 produk dan menjalankan AI reformat.
3. Melihat hasil reformat yang sudah tersimpan dan mengeditnya.
4. Membuat caption dan variasi dengan hashtag.
5. Membuka X dengan caption terisi atau menyalinnya manual.
6. Mencatat posting berulang untuk produk yang sama.
7. Menjalankan seluruh stack melalui Docker Compose di localhost.
