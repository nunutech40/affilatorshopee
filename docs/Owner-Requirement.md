---
type: owner-requirement
project: AffiliatorShopee
status: aligned-with-prd
---

# Owner Requirement — AffiliatorShopee

## Tujuan Produk

Web app pribadi untuk menyimpan produk-produk curated affiliate Shopee, lalu memudahkan pembuatan caption dan posting berulang tanpa harus menyusun caption dari nol setiap kali.

## Target User

- Owner sendiri sebagai operator
- Platform MVP: X
- Platform lain: roadmap

## Model Konten yang Didukung

| Model | Karakteristik |
|---|---|---|
| Trending | Demand/momen yang sedang panas, lalu produk sebagai jawaban |
| Branded | Reminder brand yang sudah dipercaya + diskon/deal/urgency faktual |
| Murah | Harga termurah + kegunaan nyata + proof/value, tanpa mengarang |

## Keputusan Penting: Tidak Auto-Posting Penuh

Owner belum menemukan cara otomatis posting yang aman dari deteksi bot platform dalam jangka panjang.

Solusi yang dipilih:

```text
Simpan produk mentah di web app
→ AI membuat promo text dari raw text dan menyimpan hasil
→ Buat varian caption bila perlu
→ Pilih caption + hashtag
→ Klik "Share ke X"
→ Tab X terbuka, caption terisi
→ Media di-download ke local storage
→ User upload media lokal manual
→ Klik Post
→ User catat posting jika berhasil
```

Jadi web app hanya membantu menyusun caption dan membuka tab posting. User tetap mengontrol klik terakhir.

## Fitur Utama

### 1. Manajemen Produk Curated

- Tambah produk dari raw text, import X, atau extension scraper Shopee
- Paste satu atau beberapa link gambar eksternal
- Paste link video eksternal jika ada, termasuk `.mp4`
- Pilih/ubah content model Trending, Branded, atau Murah; cluster tetap metadata terpisah
- Riwayat posting sederhana per produk
- Produk boleh diposting berulang kali

### 2. Caption Generator

- Generate caption dari template
- Banyak variasi caption untuk satu produk
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
- Riwayat posting berisi caption, hashtag, platform, dan waktu
- Media yang berhasil di-download tersedia di local storage dan dapat di-download sebagai ZIP
- Bukan automation penuh, hanya helper

### 5. Dashboard Sederhana

- Produk yang sudah/belum diposting
- Riwayat caption
- Jumlah dan waktu posting terakhir per produk

## Data Produk yang Disimpan

| Field | Keterangan |
|---|---|
| id | ID unik internal |
| raw_text | Data copas Shopee asli |
| product_name | Nama produk |
| shopee_link | Link affiliate |
| image_url | URL gambar utama |
| image_urls | URL gambar tambahan |
| video_url | URL video |
| normal_price | Harga normal |
| sale_price | Harga diskon |
| discount_percent | Persentase diskon |
| rating | Rating produk |
| sold_count | Jumlah terjual |
| review_count | Jumlah penilaian |
| cluster | Kategori/cluster |
| keyword | Keyword utama untuk hook |
| problem | Problem yang ingin diangkat |
| content_model | capture / cheap / trending |
| capture_angle | search / reply / trend / problem |
| benefit_1 | Benefit utama |
| benefit_2 | Benefit kedua |
| benefit_3 | Benefit ketiga |
| urgency | Voucher, stok, PO, flash sale |
| caption_template | Template default |
| hashtag_pool | Pilihan hashtag |
| status | raw / reformatted / ready |
| notes | Catatan tambahan |
| created_at | Waktu dibuat |
| updated_at | Waktu terakhir diubah |

## Prinsip Caption

- Berdasarkan ilmu market ekonomi dan model affiliate yang sudah disusun
- Hook dari keyword, problem, atau value
- Benefit maksimal 3 poin
- Proof berdasarkan data nyata: rating, terjual, review
- Harga/offer transparan
- Urgency hanya kalau punya dasar nyata
- CTA sederhana

## Teknologi Final

- Frontend: Vue 3 + Vite + Tailwind CSS + Pinia
- Backend: Go
- Database: PostgreSQL
- Share: Twitter Web Intent
- Media: URL di-download ke local storage, lalu user upload manual di tab X

## Path Project

```text
/Users/nununugraha/Documents/Programming/OtherPorject/AffiliatorShopee
```

## Status

Selaras dengan PRD dan TRD — siap dilanjutkan ke implementasi setelah kontrak teknis final.
