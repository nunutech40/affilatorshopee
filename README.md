# AffiliatorShopee

Web app pribadi untuk menyimpan produk affiliate Shopee, merapikan data dengan AI, membuat caption, dan membantu posting manual ke X.

## Dokumen

- [Dokumentasi](docs/) - seluruh PRD, TRD, TODO, handoff, dan prompt AI

## Alur utama

```text
Scrape Shopee / import X / paste raw text
→ simpan produk + media lokal + tracking tag
→ reformat AI satu kali saat save (raw text saja)
→ edit link affiliate dan promo
→ reformat varian caption bila perlu
→ Share ke X → cek media → Post manual → catat posting
```

MVP tidak melakukan auto-posting dan tidak mengelola akun sosial. Media di-download ke local storage agar dikirim ke Chrome Extension helper dan di-upload manual ke X. Media dapat ditambah atau dihapus dari halaman detail. Extension dipisah per fungsi di dalam `extension/`.

Akun X yang digunakan mengikuti session browser yang sedang login. Web app tidak membaca, memilih, atau menyimpan identitas akun tersebut. Produk yang sama boleh dibagikan dan dicatat berkali-kali.

## Status

Core MVP sudah diimplementasikan dan dideploy lokal via Docker/OrbStack. Jika reformat AI gagal, produk tetap tersimpan sebagai raw dan dapat dicoba ulang dari detail. Ikuti [docs/PROJECT-HANDOFF.md](docs/PROJECT-HANDOFF.md) untuk kondisi terbaru.

## Struktur proyek

- `frontend/` — aplikasi Vue/Vite.
- `backend/` — API Go, handler, service, repository, model, storage, dan migration PostgreSQL.
- `extension/x-helper/` — Chrome extension untuk composer X dan upload media saat share.
- `extension/shopee-scraper/` — Chrome extension untuk scrape halaman detail Shopee (raw text, image, dan video) lalu mengirimkannya ke web app.
- Setiap produk memiliki tracking tag unik alfanumerik untuk pencocokan report klik/penjualan affiliate.
- `docs/` — PRD, TRD, prompt AI, handoff, dan dokumentasi operasional.

## Scrape produk Shopee

1. Load unpacked folder `extension/shopee-scraper/` di `chrome://extensions`.
2. Buka halaman detail produk Shopee dengan pola `/product/.../...`.
3. Klik ikon extension → **Ambil data halaman ini** → **Kirim ke web app**.
4. Pastikan tab `http://localhost:8080/products/new` sudah terbuka. Form akan terisi raw text dan URL media, tanpa link affiliate.
5. Klik **Kirim ke web app**, lalu simpan produk. AI menghasilkan caption dan tracking tag; jika gagal, produk tetap raw. Isi/ganti link affiliate dari detail.

Scraping membaca DOM, metadata, dan response network halaman yang sedang dibuka, tanpa membaca cookie atau membypass login. Extension scraper saat ini versi `1.2.0` dan dipisah dari extension X helper.
