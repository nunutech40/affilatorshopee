---
type: trd
project: AffiliatorShopee
status: aligned-with-prd
---

> **Implementation update 2026-08-31:** source code sudah melampaui sebagian requirement lama di dokumen ini. Runtime AI mendukung OpenRouter, 9router, OpenCode, dan Codex CLI bridge lokal; model ditemukan dinamis dengan fallback registry, request maksimal 10 produk, content model aktif `trending`, `branded`, `cheap`, dan `curated`, serta media di-download ke local storage dan dikirim melalui extension. Scraper Shopee dan helper X dipisah menjadi dua extension MV3. Fitur purge testing menyimpan archive ringan untuk tracking tag dan menghapus media lokal. Referensi handoff faktual: `PROJECT-HANDOFF.md` dan `CODEX-CLI-BRIDGE.md`.

# Technical Requirements Document - AffiliatorShopee

> Kontrak API normatif dan daftar endpoint lengkap berada di [openapi.yaml](openapi.yaml). Perubahan endpoint wajib memperbarui file tersebut.

## 1. Keputusan Teknis

| Layer | Teknologi |
|---|---|
| Backend | Go |
| Database | PostgreSQL |
| Frontend | Vue 3 + Vite + Tailwind CSS |
| HTTP Client | Fetch API |
| Routing | Vue Router |
| State | Pinia |
| AI | OpenRouter, 9router, OpenCode (OpenAI-compatible), Codex CLI bridge |
| Deployment | Docker Compose di Orbstack |

MVP mencakup web app pribadi + dua Chrome Extension terpisah (X helper dan Shopee scraper) serta alur posting manual ke X. Web app tidak login, memilih, atau menyimpan identitas akun X. Media dari URL gambar/video di-download ke local storage (`STORAGE_PATH`) saat produk disimpan. Tidak ada autentikasi operator atau Threads pada MVP.

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
PostgreSQL    AI provider terpilih
```

User tetap meng-upload gambar/video langsung di X dan menekan tombol Post secara manual. Backend menyimpan data produk, caption, catatan posting, serta file media yang di-download dari URL eksternal ke local storage.

## 3. Struktur Folder

```text
AffiliatorShopee/
├── cmd/
│   └── api/
│       └── main.go
├── backend/
│   ├── config/
│   │   └── config.go
│   ├── db/
│   │   ├── connection.go
│   │   └── migrations/
│   │       ├── 001_create_products.up.sql
│   │       ├── 001_create_products.down.sql
│   │       ├── 002_create_post_logs.up.sql
│   │       ├── 002_create_post_logs.down.sql
│   │       ├── 003_create_caption_variations.up.sql
│   │       ├── 003_create_caption_variations.down.sql
│   │       ├── 004_create_product_media.up.sql
│   │       └── 004_create_product_media.down.sql
│   ├── handler/
│   │   ├── product_handler.go
│   │   ├── ai_handler.go
│   │   ├── caption_handler.go
│   │   ├── post_log_handler.go
│   │   └── media_handler.go
│   ├── model/
│   │   ├── product.go
│   │   ├── post_log.go
│   │   ├── caption_variation.go
│   │   └── media_file.go
│   ├── repository/
│   │   ├── product_repo.go
│   │   ├── post_log_repo.go
│   │   ├── caption_variation_repo.go
│   │   └── media_repo.go
│   ├── storage/
│   │   └── storage.go
│   └── service/
│       ├── product_service.go
│       ├── ai_service.go
│       ├── caption_service.go
│       ├── share_service.go
│       ├── post_log_service.go
│       └── media_service.go
├── frontend/
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
├── backend/go.mod
├── go.sum
├── .env.example
├── README.md
└── docs/
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
    tracking_tag VARCHAR(64) NOT NULL UNIQUE,
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
    content_model VARCHAR(20) CHECK (content_model IN ('capture', 'cheap', 'trending', 'branded')),
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

`image_url`, `image_urls`, dan `video_url` menyimpan URL eksternal yang diberikan user. Saat produk dibuat, backend mencoba men-download setiap URL ke local storage (`STORAGE_PATH/products/{product_id}/`) dan menyimpan metadata di tabel `product_media`. Form input mendukung banyak image URL via tombol `+ Add image URL` dan satu video URL (termasuk `.mp4`). Download yang gagal tidak membatalkan pembuatan produk; kegagalan dilaporkan pada response create.

`content_model` menggantikan field `model` lama agar dapat mewakili pendekatan konten: `cheap`, `trending`, `branded`, dan `curated`. `capture` tetap diterima backend untuk data legacy, tetapi tidak lagi ditawarkan pada form/filter baru. `capture_angle` hanya digunakan bila modelnya `capture`.

Status `posted` tidak disimpan pada produk. Satu produk boleh dicatat dan diposting berulang kali; riwayatnya disimpan di `post_logs`.

Transisi status:

- Product baru: `raw`.
- AI reformat: `raw` menjadi `reformatted`.
- Edit manual dengan data lengkap: `raw` atau `reformatted` menjadi `ready`.
- Posting tidak mengubah status product.

Field minimum untuk status `ready` adalah `product_name`, `shopee_link`, `cluster`, `content_model` (`capture`/`cheap`/`trending`/`branded`/`curated`), dan minimal satu dari `benefit_1`, `benefit_2`, atau `benefit_3`. Jika `content_model` adalah `capture`, `capture_angle` juga wajib diisi. Caption hanya dapat dibuat untuk status `reformatted` atau `ready`.

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
GET /api/products?cluster=&content_model=&status=&search=&clicked=&sort=newest&page=1&limit=20
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
  "image_urls": ["https://...", "https://..."],
  "video_url": "https://.../video.mp4",
  "notes": ""
}
```

Produk baru selalu disimpan dengan status `raw`. Semua URL gambar/video yang valid akan dicoba di-download ke local storage; kegagalan per URL dilaporkan tanpa membatalkan produk.

Response create mengembalikan object gabungan:

```json
{
  "success": true,
  "data": {
    "product": { "id": "uuid", "status": "raw" },
    "media": {
      "downloaded": [{ "filename": "image-01.jpg", "media_type": "image" }],
      "failed": [{ "source_url": "https://...", "code": "MEDIA_DOWNLOAD_FAILED" }]
    }
  },
  "error": null
}
```

Response detail menggunakan object product di dalam `data`. Contoh ringkas:

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
GET /api/ai/models
POST /api/ai/reformat
Content-Type: application/json
```

`GET /api/ai/models` menggabungkan model dinamis dari OpenRouter, 9router, dan OpenCode dengan model statis/fallback dari registry aplikasi, termasuk model Codex CLI. Setiap pilihan membawa identitas provider agar model bernama sama tidak tertukar. Auth runtime hanya dibaca backend dari environment, mount auth OpenCode, atau session Codex CLI host melalui bridge; tidak boleh ditulis ke source code.

Request reformat:

```json
{
  "product_ids": ["uuid1", "uuid2"],
  "model": "stealth/ox-alpha"
}
```

Aturan:

- `product_ids` wajib berisi 1-10 ID (batas 10 untuk hemat token, UI disable >10).
- `model` opsional; jika kosong memakai `OPENROUTER_MODEL` (default `stealth/ox-alpha`).
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

### 5.7 Media API

```http
GET /api/products/{id}/media
POST /api/products/{id}/media
GET /api/products/{id}/media/download
DELETE /api/products/{id}/media/{mediaID}
```

`GET /api/products/{id}/media` mengembalikan daftar metadata file lokal untuk produk tersebut. `POST /api/products/{id}/media` menerima `image_urls` dan/atau `video_url`, lalu mengunduh media ke local storage. `DELETE /api/products/{id}/media/{mediaID}` menghapus metadata dan file lokal. `GET /api/products/{id}/media/download` mengembalikan ZIP berisi semua file media yang berhasil di-download. Batasan: image maksimal 20 MB, video maksimal 200 MB, timeout download 2 menit, dan URL private/local ditolak. Penamaan media baru melanjutkan nomor media yang sudah ada dan URL duplikat diabaikan.

Contoh menambah media:

```json
{
  "image_urls": ["https://cdn.example.com/image.jpg"],
  "video_url": "https://cdn.example.com/video.mp4"
}
```

`PATCH /api/products/{id}` tetap digunakan untuk mengubah `shopee_link`. Setiap produk juga memiliki `tracking_tag` unik yang dibuat saat save dan ditampilkan di dashboard/detail untuk dicopy ke proses pembuatan link affiliate Shopee. Saat AI menghasilkan caption, backend mengirim link affiliate sebagai fakta eksplisit dan normalizer selalu mengganti URL Shopee di output dengan `shopee_link` produk.

### Purge produk testing dan tracking archive

`DELETE /api/products/{id}` adalah delete biasa dan perilakunya tidak berubah. Purge testing memakai endpoint terpisah `POST /api/products/purge-testing` dan hanya menerima model `trending` atau `cheap` serta daftar ID terpilih. UI checklist “semua” hanya memilih item yang terlihat pada halaman aktif; pindah halaman mengosongkan selection. Produk `curated` tidak menjadi target purge.

Purge menghapus row produk beserta data berat dan media lokal, tetapi menyalin identitas minimum (`product_id`, `shopee_link`, `tracking_tag`, nama/model, dan agregat) ke `product_tracking_archive`. Import CSV klik dan komisi mencocokkan tracking tag ke `products` maupun archive. Dengan demikian event CSV tetap tercatat dan produk testing yang kembali mendapat klik/penjualan masih dapat dikenali untuk di-scrape ulang.

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

- Provider dipilih dari model yang disimpan di Settings: OpenRouter (`OPENROUTER_BASE_URL`), 9router (`NINEROUTER_BASE_URL`), OpenCode (endpoint Zen), atau Codex CLI bridge (`CODEX_BRIDGE_URL`).
- Model default dan fallback diambil dari environment/registry; UI tidak mengirim API key.
- API key hanya dibaca backend dari environment atau file auth OpenCode yang di-mount read-only.
- Codex bridge memakai session login Codex CLI di host dan meneruskan model tanpa prefix `codex/`.
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
- Memilih `content_model` dari `trending`, `branded`, atau `cheap`; `capture` tetap diterima sebagai legacy.
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
- `BulkReformat.vue` - pilih maksimal 10 produk dan panggil AI.
- `CaptionGenerator.vue` - pilih template dan generate caption.
- `HashtagSelector.vue` - pilih 0-3 hashtag.
- `ShareButton.vue` - copy clipboard dan buka X.
- `PostLogForm.vue` - catat posting setelah user selesai posting di X.
- `SettingsView.vue` - kelola master jenis barang.

### 8.4 Jenis Barang

- Produk dapat memiliki beberapa jenis barang melalui tabel relasi `product_niches`.
- Master jenis barang dikelola melalui `GET/POST /api/niches` dan `DELETE /api/niches/{id}`.
- Relasi produk diganti sekaligus melalui `PUT /api/products/{id}/niches` dengan payload `{ "niche_ids": ["..."] }`.
- Dashboard menerima `niche_id`; nilai `uncategorized` berarti produk yang belum memiliki relasi pada `product_niches`.
- UI menggunakan istilah “Jenis barang”; nama `niche` dipertahankan pada nama endpoint dan tabel untuk kompatibilitas internal.

### 8.5 Bank Konten X per Niche (Implemented)

`content_niches` adalah master terpisah dari tabel `niches` yang saat ini dipakai sebagai master Jenis Barang produk. Jangan memakai `product_niches` untuk mengelompokkan niche konten.

Model data yang direncanakan:

- `content_niches`: `id`, `name`, `slug`, `description`, `status`, timestamps; unique pada `slug`.
- `content_items`: `id`, `platform`, `external_post_id`, `canonical_url`, `author_handle`, `original_text`, `media`, `published_at`, `source_query`, `status`, timestamps; unique pada `(platform, external_post_id)` atau canonical URL.
- `content_item_niches`: relasi many-to-many content ↔ content niche.
- `content_item_product_types`: relasi many-to-many content ↔ master Jenis Barang (`niches`).
- `content_stat_snapshots`: `content_item_id`, `captured_at`, `like_count`, `repost_count`, `reply_count`, `bookmark_count`, `view_count`, dan raw metrics JSON.
- `content_clean_versions`, `content_reformats`, dan `content_reformat_variants`: versi turunan yang selalu menyimpan `content_item_id`; `original_text` tidak pernah ditimpa.

Endpoint yang direncanakan:

- `GET/POST/PATCH/DELETE /api/content-niches`
- `GET/POST /api/content-items`
- `GET /api/content-items?content_niche_id=&platform=x&status=`
- `GET/POST /api/content-items/{id}/stats`
- `POST /api/content-items/{id}/clean`
- `POST /api/content-items/{id}/reformat`
- `PUT /api/content-items/{id}/product-types` dengan `{ "niche_ids": ["..."] }`

State content: `discovered → reviewed → cleaned → reformatted → ready_to_share`. Collector awal mengambil konten yang terlihat pada halaman pencarian/session X dan dipilih user melalui extension; tidak mengakses konten privat atau melewati pembatasan platform. Collector berbasis X API dapat ditambahkan kemudian. X Advanced Search mendukung filter kata, hashtag, akun, bahasa, lokasi, dan tanggal; Search Posts API menyediakan recent/full archive sesuai akses API dan dapat mengembalikan waktu publikasi serta public metrics. Lihat dokumentasi resmi [Advanced Search X](https://help.x.com/id/using-x/x-advanced-search) dan [Search Posts API](https://docs.x.com/x-api/posts/search/introduction).

MVP tidak memiliki flow pemilihan akun; akun X yang digunakan mengikuti session login di browser user.

Implementasi UI list/detail: `ContentBankView.vue` menampilkan judul maksimum dua baris dan excerpt ringkas dengan ellipsis, ditambah chip status/source/niche serta statistik popularitas. `ContentItemDetailView.vue` mempertahankan raw konten lengkap, membatasi tinggi textarea, dan memakai sidebar sticky pada desktop yang kembali normal pada layar kecil. `X Research` menangkap post tunggal atau thread yang terlihat pada halaman detail; thread disimpan sebagai satu `original_text` berurutan dan media dideduplikasi. Perubahan ini tidak membutuhkan migrasi karena memakai field raw/media yang sudah ada.

### 8.3 State

- `productStore` - list, filter, CRUD, dan AI reformat.
- `captionStore` - caption aktif dan variasi.

## 9. Chrome Extension (Manifest V3) — Satu Kesatuan dengan Web App

- `extension/manifest.json` — `activeTab`, `storage`, `scripting`, `clipboardWrite`, `clipboardRead`, host `x.com`, `twitter.com`, `threads.net`
- `extension/background.js` — simpan `lastCaption` via `chrome.storage`, message `AFFILIATOR_SET_CAPTION`
- `extension/content.js` — di `x.com/intent/tweet` parse `?text=` + fallback `chrome.storage`/`clipboardRead`, deteksi composer (`[data-testid="tweetTextarea_0"]`, `div[contenteditable]`, `DraftEditor`), `insertText` + `InputEvent`, `MutationObserver` + polling 12x, notifikasi toast
- `extension/popup.html` / `popup.js` — tombol `Paste caption sekarang` (fallback manual), instruksi load unpacked
- `frontend/src/components/ShareButton.vue` — saat `Share ke X` juga `chrome.storage.local.set` + `chrome.runtime.sendMessage` agar extension auto-paste tanpa user `Cmd+V`

Cara load: `chrome://extensions` → Developer mode → Load unpacked → folder `extension`

## 10. Docker Compose

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
      STORAGE_PATH: ${STORAGE_PATH:-/app/data/uploads}
      OPENROUTER_API_KEY: ${OPENROUTER_API_KEY}
      NINEROUTER_API_KEY: ${NINEROUTER_API_KEY}
      OPENCODE_API_KEY: ${OPENCODE_API_KEY}
      OPENROUTER_MODEL: ${OPENROUTER_MODEL}
      ENV: ${ENV:-development}
    ports:
      - "127.0.0.1:8080:8080"
    volumes:
      - ./data/uploads:/app/data/uploads
    depends_on:
      db:
        condition: service_healthy

volumes:
  postgres_data:
```

## 11. Environment Variables

```bash
PORT=8080
DATABASE_URL=postgres://affiliator:change-me@db:5432/affiliator?sslmode=disable
POSTGRES_USER=affiliator
POSTGRES_PASSWORD=change-me
POSTGRES_DB=affiliator
STORAGE_PATH=/app/data/uploads
OPENROUTER_API_KEY=sk-or-v1-...
OPENROUTER_MODEL=stealth/ox-alpha
NINEROUTER_BASE_URL=https://9router.103-59-94-121.nip.io/v1
NINEROUTER_API_KEY=9r-...
OPENCODE_BASE_URL=https://opencode.ai/zen/v1
OPENCODE_MODEL=muse-spark-1.2-contributor-free
ENV=development
```

Jangan commit file `.env` atau API key ke repository.

## 12. Security dan Reliability

- MVP hanya untuk local/private use dan tidak memiliki auth.
- Bind port API dan database ke `127.0.0.1`.
- Jangan expose service ke internet sebelum auth dan authorization tersedia.
- API key/provider auth hanya berada di backend.
 - Validasi ukuran body, panjang string, URL, enum, angka, dan jumlah hashtag.
 - Validasi media URL hanya `http/https` tanpa private IP dan batas ukuran 20 MB image / 200 MB video.
 - CORS hanya mengizinkan origin development lokal yang ditentukan, misalnya `http://localhost:5173`.
- Batasi request AI dengan rate limit dan concurrency limit agar biaya terkendali.
- Validasi environment wajib saat startup dan fail-fast bila `DATABASE_URL` atau konfigurasi provider aktif tidak valid.
- Gunakan timeout, error handling, dan logging terstruktur untuk setiap provider; 429/5xx harus menjadi error yang terlihat, bukan output kosong yang sukses.
- Jangan menyimpan raw product text di log aplikasi secara default.
- Semua query database menggunakan parameter binding.
- Caption di-encode dengan aman sebelum dipakai pada URL intent.

## 13. Error Handling

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

## 14. Testing Strategy

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
- AI bulk 1-10 produk dan menolak lebih dari 10.
- Raw text tetap utuh setelah reformat.
- AI malformed response, timeout, dan partial failure.
- Template dengan field kosong.
- Character count Unicode dan caption di atas batas.
- Hashtag 0-3 dan penolakan hashtag ke-4.
- URL encoding untuk spasi, emoji, newline, dan karakter khusus.
- Posting berulang pada produk yang sama.
- Migration dan database readiness Docker.

## 15. Migration Files

- `001_create_products.up.sql` / `001_create_products.down.sql`
- `002_create_post_logs.up.sql` / `002_create_post_logs.down.sql`
- `003_create_caption_variations.up.sql` / `003_create_caption_variations.down.sql`
- `004_create_product_media.up.sql` / `004_create_product_media.down.sql`
- `015_add_niches.up.sql` / `015_add_niches.down.sql`
- `016_split_book_niche.up.sql` / `016_split_book_niche.down.sql`

Gunakan `golang-migrate/migrate` untuk menjalankan migration SQL. Migration harus memiliki version tracking, dijalankan sebelum app menerima traffic, dan membuat startup gagal bila migration tidak berhasil.

## 16. Roadmap Setelah MVP

- Integrasi Threads melalui adapter, payload, composer, dan extension yang terpisah dari X.
- Riset posting X populer per niche: collector, penyimpanan sumber, pengelompokan niche, dan pipeline reformat AI khusus X.
- Riset konten Facebook populer per niche: collector, penyimpanan sumber, pengelompokan niche, dan pipeline reformat AI khusus Facebook.
- Share atau bantu posting konten Facebook melalui adapter dan extension Facebook; jangan memakai parser/payload X secara langsung.
- S3/CDN untuk media (menggantikan local storage).
- Bank konten X per niche dengan niche master, capture, versi asli/bersih/reformat, snapshot statistik, dan varian akun.
- Relasi many-to-many bank konten dengan master Jenis Barang produk.
- Scheduling dan analytics.
- Auth, multi-user, dan admin panel.

## 17. Definition of Done MVP

MVP dianggap selesai bila user dapat:

1. Menyimpan data Shopee mentah.
2. Memilih maksimal 10 produk dan menjalankan AI reformat.
3. Melihat hasil reformat yang sudah tersimpan dan mengeditnya.
4. Membuat caption dan variasi dengan hashtag.
5. Membuka X dengan caption terisi atau menyalinnya manual.
6. Mencatat posting berulang untuk produk yang sama.
7. Menjalankan seluruh stack melalui Docker Compose di localhost.
