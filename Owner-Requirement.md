---
type: owner-requirement
project: AffiliatorShopee
status: draft
---

# Owner Requirement — AffiliatorShopee

## Tujuan Produk

Web app untuk menyimpan produk-produk curated affiliate Shopee, lalu memudahkan posting ke banyak akun/platform tanpa harus menyusun caption dari nol setiap kali.

## Target User

- Pemilik banyak akun affiliate
- Akun tanpa branding kuat
- Akun branding
- Platform: X, Threads, dan kemungkinan lainnya

## Model Konten yang Didukung

| Model | Channel | Karakteristik |
|---|---|---|
| 4 Capture Models | X Akun 1 | Search, Reply, Trend, Problem Capture |
| Curated Cheap/Value | X Akun 2 | Produk murah, berguna, deal |
| Curated Branded | Threads | Brand dikenal, diskon, voucher, promo |

## Keputusan Penting: Tidak Auto-Posting Penuh

Owner belum menemukan cara otomatis posting yang aman dari deteksi bot platform dalam jangka panjang.

Solusi yang dipilih:

```text
Simpan produk di web app
→ Pilih produk + caption + hashtag
→ Klik "Share ke X"
→ Tab X terbuka, caption terisi
→ User login di browser yang sudah login akun X
→ Upload gambar manual
→ Klik Post
```

Jadi web app hanya membantu menyusun caption dan membuka tab posting. User tetang mengontrol klik terakhir.

## Fitur Utama

### 1. Manajemen Produk Curated

- Tambah produk dengan data lengkap
- Upload atau paste link gambar
- Kategorikan cluster
- Status posting per akun/platform
- Tandai produk sudah diposting atau belum

### 2. Caption Generator

- Generate caption dari template
- Banyak variasi caption untuk satu produk
- Bisa bedakan caption sedikit per akun
- Template yang tersedia:
  - Direct Product
  - Keyword + Recommendation
  - Problem-specific
  - Cheap/Value

### 3. Hashtag Selector

- Pilih hashtag berdasarkan cluster/kategori
- Bisa custom hashtag
- Tidak memakai hashtag berlebihan
- Batasi 1–3 hashtag

### 4. Posting Helper

- Tombol "Share ke X" dengan Twitter Web Intent
- Caption sudah terisi otomatis
- User tinggal upload gambar dan klik Post
- Bisa pilih akun mana yang akan posting
- Riwayat posting per akun

### 5. Multi-akun Support

- Daftar akun affiliate
- Riwayat posting per akun
- Bukan automation penuh, hanya helper

### 6. Dashboard Sederhana

- Produk yang sudah/belum diposting
- Riwayat caption
- Status posting per platform

## Data Produk yang Disimpan

| Field | Keterangan |
|---|---|
| product_id | ID unik internal |
| product_name | Nama produk |
| shopee_link | Link affiliate |
| image_url | Gambar produk |
| normal_price | Harga normal |
| sale_price | Harga diskon |
| discount_percent | Persentase diskon |
| rating | Rating produk |
| sold_count | Jumlah terjual |
| review_count | Jumlah penilaian |
| cluster | Kategori/cluster |
| model | cheap / branded |
| benefit_1 | Benefit utama |
| benefit_2 | Benefit kedua |
| benefit_3 | Benefit ketiga |
| urgency | Voucher, stok, PO, flash sale |
| caption_template | Template default |
| hashtag_pool | Pilihan hashtag |
| status | draft / ready / posted |
| notes | Catatan tambahan |

## Prinsip Caption

- Berdasarkan ilmu market ekonomi dan model affiliate yang sudah disusun
- Hook dari keyword, problem, atau value
- Benefit maksimal 3 poin
- Proof berdasarkan data nyata: rating, terjual, review
- Harga/offer transparan
- Urgency hanya kalau punya dasar nyata
- CTA sederhana

## Teknologi yang Disarankan

- Frontend: Next.js atau HTML + Tailwind
- Backend: Node.js/Express atau Go
- Database: SQLite untuk MVP
- Share: Twitter Web Intent
- Gambar: user upload manual di tab X

## Path Project

```text
/Users/nununugraha/Documents/Programming/OtherPorject/AffiliatorShopee
```

## Status

Draft — menunggu MVP plan dan implementasi.
