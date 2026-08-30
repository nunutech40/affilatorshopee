# AffiliatorShopee

Web app pribadi untuk menyimpan produk affiliate Shopee, merapikan data dengan AI, membuat caption, dan membantu posting manual ke X.

## Dokumen

- [Dokumentasi](docs/) - seluruh PRD, TRD, TODO, handoff, prompt AI, dan Codex bridge
- [API Reference (OpenAPI/Swagger)](docs/openapi.yaml) - kontrak seluruh endpoint backend
- [API Reference (OpenAPI/Swagger)](docs/openapi.yaml) - kontrak seluruh endpoint backend

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
3. Klik ikon extension → **Ambil data halaman ini** → **Insert ke web app**.
4. Pastikan tab `http://localhost:8080/products/new` sudah terbuka. Form akan terisi raw text dan URL media, tanpa link affiliate.
5. Klik **Insert ke web app**, lalu simpan produk. AI menghasilkan caption dan tracking tag; jika gagal, produk tetap raw. Isi/ganti link affiliate dari detail.

Scraping membaca DOM, metadata, dan response network halaman yang sedang dibuka, tanpa membaca cookie atau membypass login. Extension scraper saat ini versi `1.3.0`, memprioritaskan gambar galeri produk, menangkap harga aktif/harga normal/diskon dari blok harga, serta memfilter icon/logo/avatar dan URL duplikat. Extension scraper dipisah dari extension X helper dan harus di-reload dari `chrome://extensions` setelah update.

## Codex CLI lokal

Model `codex/*` dapat dipilih dari Settings melalui bridge Codex CLI di host.
Lihat [docs/CODEX-CLI-BRIDGE.md](docs/CODEX-CLI-BRIDGE.md).

Bridge juga menyediakan endpoint OpenAI-compatible `POST /v1/chat/completions`
di `http://host.docker.internal:8787/v1`. Hermes mengakses endpoint ini secara
langsung dengan `CODEX_BRIDGE_TOKEN`, lalu bridge menjalankan Codex CLI memakai
session login Codex lokal. Jalur ini tidak melewati OpenRouter atau 9router dan
bridge harus tetap aktif selama request berjalan. Pada macOS, double-click
`tools/codex-bridge/start-codex-bridge.command`; jendela Terminal yang terbuka
adalah proses listener dan dapat ditutup untuk menghentikannya. Gunakan
`CGO_ENABLED=0` bila diperlukan.
