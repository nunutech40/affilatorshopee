---
type: trd
project: AffiliatorShopee
status: draft
---

# Technical Requirements Document — AffiliatorShopee

## 1. Stack

| Layer | Teknologi |
|---|---|
| Backend | Go |
| Database | PostgreSQL |
| Frontend | Vue 3 + Vite + Tailwind CSS |
| HTTP Client | Fetch API |
| Routing | Vue Router |
| State | Pinia |
| Chrome Extension | Manifest V3 |
| Storage | Local filesystem (abstraction untuk CDN/S3 di masa depan) |
| Deployment | Docker Compose di Orbstack |

## 2. Arsitektur Sistem

```text
┌─────────────────────────────────────────┐
│           Browser (Chrome)              │
│  ┌─────────────┐  ┌─────────────────┐  │
│  │  Vue 3 App  │  │ Chrome Extension │  │
│  │  (Vite)     │  │  (Manifest V3)  │  │
│  └──────┬──────┘  └─────────────────┘  │
└─────────┼───────────────────────────────┘
          │ HTTP/JSON
┌─────────▼───────────────────────────────┐
│         Go Backend API                  │
│  ┌─────────┐ ┌─────────┐ ┌──────────┐  │
│  │ Handler │ │ Service │ │ Repository│  │
│  └────┬────┘ └────┬────┘ └────┬─────┘  │
│       └───────────┴───────────┘         │
│              │ SQL                       │
│       ┌──────▼──────┐                   │
│       │ PostgreSQL  │                   │
│       │   (Docker)  │                   │
│       └─────────────┘                   │
│                                         │
│  ┌─────────────────────────────────┐    │
│  │  Local Filesystem Storage       │    │
│  │  /data/uploads/products/{id}/   │    │
│  │  (bisa diganti CDN/S3 nanti)    │    │
│  └─────────────────────────────────┘    │
└─────────────────────────────────────────┘
```

## 3. Storage Abstraction

Interface storage sejak awal agar mudah ganti ke S3/CDN:

```go
type Storage interface {
    Save(productID string, filename string, data []byte) (string, error)
    GetURL(productID string, filename string) string
    Delete(productID string, filename string) error
    List(productID string) ([]string, error)
}
```

Implementasi awal: `LocalStorage`.
Implementasi masa depan: `S3Storage`.

## 4. Database Schema

### 4.1 Tabel products

```sql
CREATE TABLE products (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    raw_text TEXT,
    product_name VARCHAR(255),
    shopee_link TEXT,
    image_url TEXT,
    image_urls TEXT[],
    video_url TEXT,
    normal_price INTEGER,
    sale_price INTEGER,
    discount_percent INTEGER,
    rating NUMERIC(2,1),
    sold_count VARCHAR(50),
    review_count VARCHAR(50),
    cluster VARCHAR(100),
    model VARCHAR(20) CHECK (model IN ('cheap', 'branded')),
    benefit_1 VARCHAR(255),
    benefit_2 VARCHAR(255),
    benefit_3 VARCHAR(255),
    urgency VARCHAR(255),
    caption_template VARCHAR(50) DEFAULT 'direct_product',
    hashtag_pool TEXT[],
    notes TEXT,
    status VARCHAR(20) DEFAULT 'raw' CHECK (status IN ('raw', 'reformatted', 'ready', 'posted')),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### 4.2 Tabel accounts

```sql
CREATE TABLE accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    platform VARCHAR(50) NOT NULL,
    account_name VARCHAR(100) NOT NULL,
    account_type VARCHAR(50) CHECK (account_type IN ('capture', 'cheap', 'branded')),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### 4.3 Tabel post_logs

```sql
CREATE TABLE post_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID NOT NULL REFERENCES products(id),
    account_id UUID NOT NULL REFERENCES accounts(id),
    caption TEXT NOT NULL,
    hashtags TEXT[],
    posted_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### 4.4 Tabel caption_variations

```sql
CREATE TABLE caption_variations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID NOT NULL REFERENCES products(id),
    label VARCHAR(50) NOT NULL,
    caption TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

## 5. API Specification

### 5.1 Response Format

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
  "error": "error message"
}
```

### 5.2 Products API

#### List Products

```http
GET /api/products?cluster=&model=&status=&search=
```

#### Create Product

```http
POST /api/products
Content-Type: application/json

{
  "raw_text": "...",
  "shopee_link": "...",
  "image_url": "..."
}
```

#### Get Product

```http
GET /api/products/{id}
```

#### Update Product

```http
PUT /api/products/{id}
Content-Type: application/json
```

#### Delete Product

```http
DELETE /api/products/{id}
```

### 5.3 AI Reformat API

#### Bulk Reformat

```http
POST /api/ai/reformat
Content-Type: application/json

{
  "product_ids": ["uuid1", "uuid2"]
}
```

Maksimal 20 product_ids per request.

Response:

```json
{
  "success": true,
  "data": [
    { "product_id": "uuid1", "fields": { ... } },
    { "product_id": "uuid2", "fields": { ... } }
  ],
  "error": null
}
```

### 5.4 Caption API

#### Generate Caption

```http
POST /api/captions/generate
Content-Type: application/json

{
  "product_id": "uuid",
  "template": "direct_product",
  "hashtags": ["#BajuAnak"]
}
```

#### Generate Variations

```http
POST /api/captions/variations
Content-Type: application/json

{
  "product_id": "uuid",
  "count": 3
}
```

### 5.5 Share API

#### Share to X

```http
GET /api/share/x?caption={encoded_caption}
```

Redirect ke Twitter Web Intent.

### 5.6 Post Logs API

#### Create Post Log

```http
POST /api/post-logs
Content-Type: application/json

{
  "product_id": "uuid",
  "account_id": "uuid",
  "caption": "...",
  "hashtags": ["#BajuAnak"]
}
```

### 5.7 Accounts API

#### List Accounts

```http
GET /api/accounts
```

#### Create Account

```http
POST /api/accounts
Content-Type: application/json

{
  "platform": "x",
  "account_name": "akun_cheap_1",
  "account_type": "cheap"
}
```

### 5.8 Media API

#### Upload Media

```http
POST /api/products/{id}/media
Content-Type: multipart/form-data

file: <binary>
```

#### List Media

```http
GET /api/products/{id}/media
```

#### Download All Media as ZIP

```http
GET /api/products/{id}/media/download
```

## 6. Backend Folder Structure

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
│   │       ├── 002_create_accounts.sql
│   │       ├── 003_create_post_logs.sql
│   │       └── 004_create_caption_variations.sql
│   ├── handler/
│   │   ├── product_handler.go
│   │   ├── parser_handler.go
│   │   ├── caption_handler.go
│   │   ├── account_handler.go
│   │   ├── post_log_handler.go
│   │   └── media_handler.go
│   ├── model/
│   │   ├── product.go
│   │   ├── account.go
│   │   └── post_log.go
│   ├── repository/
│   │   ├── product_repo.go
│   │   ├── account_repo.go
│   │   └── post_log_repo.go
│   ├── service/
│   │   ├── product_service.go
│   │   ├── ai_service.go
│   │   ├── caption_service.go
│   │   ├── share_service.go
│   │   └── media_service.go
│   └── storage/
│       ├── storage.go
│       └── local_storage.go
├── web/
│   ├── index.html
│   ├── src/
│   │   ├── main.js
│   │   ├── App.vue
│   │   ├── router/
│   │   │   └── index.js
│   │   ├── stores/
│   │   │   └── productStore.js
│   │   ├── components/
│   │   │   ├── ProductList.vue
│   │   │   ├── ProductForm.vue
│   │   │   ├── ProductParser.vue
│   │   │   ├── CaptionGenerator.vue
│   │   │   ├── MediaUploader.vue
│   │   │   └── ShareButton.vue
│   │   └── views/
│   │       ├── HomeView.vue
│   │       ├── ProductDetailView.vue
│   │       └── AccountsView.vue
│   └── package.json
├── extension/
│   ├── manifest.json
│   ├── background.js
│   ├── content.js
│   └── popup.html
├── docker-compose.yml
├── Dockerfile
├── go.mod
├── go.sum
├── .env.example
└── README.md
```

## 7. Caption Template Engine

### 7.1 Placeholders

| Placeholder | Sumber Data |
|---|---|
| `{product_name}` | product_name |
| `{benefit_1}` | benefit_1 |
| `{benefit_2}` | benefit_2 |
| `{benefit_3}` | benefit_3 |
| `{rating}` | rating |
| `{sold_count}` | sold_count |
| `{review_count}` | review_count |
| `{normal_price}` | normal_price |
| `{sale_price}` | sale_price |
| `{discount_percent}` | discount_percent |
| `{urgency}` | urgency |
| `{shopee_link}` | shopee_link |
| `{hashtags}` | hashtag_pool dipilih |

### 7.2 Helper Functions

| Fungsi | Contoh |
|---|---|
| `format_price` | 39999 → Rp39.999 |
| `format_number` | 1000000 → 1.000.000 |
| `join_hashtags` | array → string |

### 7.3 Template Examples

**Direct Product:**

```text
Cari {product_name}?

✅ {benefit_1}
✅ {benefit_2}
✅ {benefit_3}
✅ {rating}⭐️
✅ {sold_count} terjual
🔥 {urgency}

Cek di sini 👇
{shopee_link}

{hashtags}
```

**Keyword + Recommendation:**

```text
Lagi cari {product_name}?

Ini salah satu yang gue shortlist.

Kenapa masuk shortlist:
✅ {benefit_1}
✅ {benefit_2}
✅ {benefit_3}
✅ {rating}⭐️ | {sold_count} terjual

Harganya juga masih masuk akal.

Cek:
{shopee_link}

{hashtags}
```

**Cheap / Value:**

```text
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

## 8. AI Reformat Service

### 8.1 Input

POST body:

```json
{
  "product_ids": ["uuid1", "uuid2"]
}
```

### 8.2 Prompt ke AI via OpenRouter

Endpoint: `https://openrouter.ai/api/v1/chat/completions`

Model default: `google/gemini-flash-1.5`

Alternatif: `openai/gpt-4o-mini`

Headers:

```text
Authorization: Bearer {OPENROUTER_API_KEY}
HTTP-Referer: http://localhost:8080
X-Title: AffiliatorShopee
```

Prompt:

```text
Kamu adalah asisten kurasi produk affiliate.
Saya akan memberikan data mentah dari Shopee.
Tugas kamu:
1. Rapikan nama produk
2. Ekstrak harga normal, harga diskon, diskon %
3. Ekstrak rating dan jumlah terjual
4. Tentukan cluster/kategori
5. Pilih 3 benefit utama
6. Tentukan urgency (stok, voucher, PO, flash sale, dll)
7. Tentukan model: cheap atau branded
8. Pilih 1-3 hashtag relevan

Output hanya JSON array, tanpa penjelasan.
Jika data tidak ditemukan, isi dengan null.
Jangan mengarang data yang tidak ada di input.

Format output:
[
  {
    "product_id": "...",
    "product_name": "...",
    "normal_price": 0,
    "sale_price": 0,
    "discount_percent": 0,
    "rating": 0.0,
    "sold_count": "...",
    "review_count": "...",
    "cluster": "...",
    "model": "cheap",
    "benefit_1": "...",
    "benefit_2": "...",
    "benefit_3": "...",
    "urgency": "...",
    "hashtag_pool": ["#..."]
  }
]
```

### 8.3 Batasan

- Maksimal 20 produk per request
- AI key di backend
- Validasi hasil AI sebelum disimpan
- Status produk berubah dari `raw` ke `reformatted`

## 9. Chrome Extension

### 9.1 Manifest V3

```json
{
  "manifest_version": 3,
  "name": "AffiliatorShopee Helper",
  "version": "1.0.0",
  "permissions": ["activeTab", "storage", "scripting", "clipboardWrite"],
  "host_permissions": ["https://twitter.com/*", "https://x.com/*", "https://threads.net/*"],
  "background": {
    "service_worker": "background.js"
  },
  "content_scripts": [
    {
      "matches": ["https://twitter.com/*", "https://x.com/*", "https://threads.net/*"],
      "js": ["content.js"]
    }
  ],
  "action": {
    "default_popup": "popup.html"
  }
}
```

### 9.2 Flow

```text
1. User di web app klik "Share ke X"
2. Web app copy caption ke clipboard
3. Web app buka tab baru ke twitter.com/intent/tweet?text=...
4. Content script di tab X aktif
5. Content script paste caption dari clipboard ke textarea
6. User upload gambar/video manual
7. User klik Post
```

### 9.3 content.js

Tugas:

- Deteksi textarea composer X/Threads
- Paste caption dari clipboard
- Tampilkan notifikasi sukses/gagal

## 10. Frontend

### 10.1 Stack Detail

- Vue 3 Composition API
- Vite build tool
- Tailwind CSS
- Vue Router
- Pinia state management
- Fetch API

### 10.2 Halaman

| Halaman | Fungsi |
|---|---|
| / | Dashboard produk |
| /products/new | Tambah produk via form/parser |
| /products/:id | Detail, edit, generate caption |
| /accounts | Kelola akun |
| /post-logs | Riwayat posting |

### 10.3 Komponen Utama

| Komponen | Fungsi |
|---|---|
| ProductList.vue | Daftar produk dengan filter |
| ProductForm.vue | Form tambah/edit produk |
| ProductParser.vue | Textarea paste ngawur |
| BulkReformat.vue | Checkbox produk + tombol reformat |
| CaptionGenerator.vue | Generate caption dan variasi |
| HashtagSelector.vue | Pilih hashtag |
| MediaUploader.vue | Upload/list media |
| ShareButton.vue | Share ke platform |

### 10.4 State Management (Pinia)

Stores:

- `productStore` — daftar produk, filter, status
- `accountStore` — daftar akun
- `captionStore` — caption dan variasi

## 11. Docker Compose

```yaml
version: '3.8'

services:
  db:
    image: postgres:15
    environment:
      POSTGRES_USER: affiliator
      POSTGRES_PASSWORD: secret
      POSTGRES_DB: affiliator
    volumes:
      - postgres_data:/var/lib/postgresql/data
    ports:
      - "5432:5432"

  app:
    build: .
    environment:
      DATABASE_URL: postgres://affiliator:secret@db:5432/affiliator?sslmode=disable
      PORT: 8080
      STORAGE_PATH: /app/data/uploads
      AI_API_KEY: ${AI_API_KEY}
    volumes:
      - ./data/uploads:/app/data/uploads
    ports:
      - "8080:8080"
    depends_on:
      - db

volumes:
  postgres_data:
```

## 12. Environment Variables

```bash
PORT=8080
DATABASE_URL=postgres://affiliator:secret@db:5432/affiliator?sslmode=disable
STORAGE_PATH=/app/data/uploads
AI_API_KEY=sk-or-v1-...
OPENROUTER_MODEL=google/gemini-flash-1.5
ENV=development
```

## 13. Security

- Tidak ada auth di MVP
- Jangan expose port ke internet tanpa autentikasi
- AI API key hanya di backend
- Validasi input di backend
- Sanitasi caption sebelum URL encode
- Batasi upload file size
- Hanya terima file gambar/video yang umum

## 14. Error Handling

### Backend

- Semua error di-wrap dalam response format standar
- Log error di server
- Jangan expose detail error internal ke client

### Frontend

- Tampilkan pesan error dari API
- Disable tombol saat loading
- Validasi form sebelum submit

## 15. Testing Strategy

| Jenis | Scope | Tool |
|---|---|---|
| Unit test | Service dan parser | Go testing |
| Integration test | API endpoint | Postman / curl |
| Manual test | End-to-end flow | Browser |

## 16. Migration Files

- `001_create_products.sql`
- `002_create_accounts.sql`
- `003_create_post_logs.sql`
- `004_create_caption_variations.sql`

Gunakan library migrasi `golang-migrate/migrate` atau `pressly/goose`.

## 17. Roadmap Teknis

| Fase | Fitur |
|---|---|
| MVP | CRUD produk, AI reformat, caption generator, share intent, extension paste, local storage |
| V2 | Scheduling post, bulk share, analytics sederhana |
| V3 | Auth, multi-user, admin panel |
| V4 | SaaS-ready, S3/CDN, managed DB |

## 18. Catatan

TRD ini akan diupdate saat implementasi berjalan, terutama di bagian Chrome Extension dan AI integration setelah pengujian pertama.
