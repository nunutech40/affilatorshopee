---
type: plan
project: AffiliatorShopee
status: draft
---

# MVP Plan — AffiliatorShopee

## MVP Scope

Versi pertama hanya fokus pada:

1. CRUD produk curated
2. Caption generator dari template
3. Hashtag selector
4. Share ke X via Twitter Web Intent
5. Riwayat posting sederhana

## User Flow MVP

```text
1. User login ke web app
2. User tambah produk (form)
3. User lihat daftar produk
4. User pilih produk
5. User pilih template dan generate caption
6. User pilih hashtag
7. User pilih akun target
8. User klik "Share ke X"
9. Browser membuka tab baru ke twitter.com/intent/tweet?text=...
10. User upload gambar manual dan klik Post
11. Web app mencatat produk sudah diposting di akun tersebut
```

## Skema Database SQLite

### Tabel products

```sql
CREATE TABLE products (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  product_name TEXT NOT NULL,
  shopee_link TEXT NOT NULL,
  image_url TEXT,
  normal_price INTEGER,
  sale_price INTEGER,
  discount_percent INTEGER,
  rating REAL,
  sold_count TEXT,
  review_count TEXT,
  cluster TEXT,
  model TEXT CHECK(model IN ('cheap', 'branded')),
  benefit_1 TEXT,
  benefit_2 TEXT,
  benefit_3 TEXT,
  urgency TEXT,
  caption_template TEXT,
  hashtag_pool TEXT,
  notes TEXT,
  status TEXT DEFAULT 'draft',
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

### Tabel accounts

```sql
CREATE TABLE accounts (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  platform TEXT NOT NULL,
  account_name TEXT NOT NULL,
  account_type TEXT CHECK(account_type IN ('capture', 'cheap', 'branded')),
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

### Tabel post_logs

```sql
CREATE TABLE post_logs (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  product_id INTEGER NOT NULL,
  account_id INTEGER NOT NULL,
  caption TEXT NOT NULL,
  hashtags TEXT,
  posted_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (product_id) REFERENCES products(id),
  FOREIGN KEY (account_id) REFERENCES accounts(id)
);
```

## Endpoint API Minimal

| Method | Endpoint | Fungsi |
|---|---|---|
| GET | /api/products | Daftar produk |
| POST | /api/products | Tambah produk |
| GET | /api/products/:id | Detail produk |
| PUT | /api/products/:id | Update produk |
| DELETE | /api/products/:id | Hapus produk |
| POST | /api/generate-caption | Generate caption dari produk + template |
| GET | /api/accounts | Daftar akun |
| POST | /api/post-logs | Catat posting |

## Halaman Frontend

| Halaman | Fungsi |
|---|---|
| / | Dashboard daftar produk |
| /products/new | Form tambah produk |
| /products/:id | Detail + generate caption |
| /accounts | Kelola akun |

## Caption Generator

Format template yang diisi variabel:

### Direct Product

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

### Keyword + Recommendation

```text
Lagi cari {keyword}?

Ini salah satu yang gue shortlist.

Kenapa masuk shortlist:
✅ {benefit_1}
✅ {benefit_2}
✅ {benefit_3}
✅ {proof}

Harganya juga masih masuk akal.

Cek:
{shopee_link}

{hashtags}
```

### Cheap / Value

```text
Cari {keyword} murah tapi nggak murahan?

Yang ini menarik 👀

✅ {benefit_1}
✅ {benefit_2}
✅ {rating}⭐️
✅ {sold_count} terjual
✅ Rp{sale_price}

Kalau budget lo sekitar Rp{sale_price}, ini worth checking.

👇
{shopee_link}

{hashtags}
```

## Share ke X

Gunakan Twitter Web Intent:

```
https://twitter.com/intent/tweet?text={encoded_caption}
```

Gambar tidak bisa di-attach otomatis. User upload manual di tab X.

## Validasi Karakter X

Tampilkan jumlah karakter caption sebelum share.
Jika lebih dari 280 karakter, beri peringatan.

## Stack Rekomendasi

- Backend: Node.js + Express
- Database: SQLite
- Frontend: Next.js atau HTML + Tailwind
- Deploy: Vercel atau local server

## Tahapan Implementasi

### Sprint 1

- Setup project
- Buat database SQLite
- CRUD produk
- Tampilkan daftar produk

### Sprint 2

- Form tambah produk lengkap
- Generate caption dari template
- Pilih hashtag

### Sprint 3

- Tombol share ke X
- Catat post_logs
- Halaman dashboard sederhana

### Sprint 4

- Kelola akun
- Filter produk per cluster/model
- Status posted/belum

## Catatan

- Fokus MVP: helper posting, bukan automation penuh
- Semua posting akhir tetap dikontrol user
- Data produk harus akurat dan berbasis proof nyata
