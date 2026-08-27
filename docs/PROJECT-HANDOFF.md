# Project Handoff — AffiliatorShopee

Dokumen ini adalah titik mulai untuk AI/developer berikutnya. Jika ada konflik dengan dokumen lama, gunakan kondisi source code dan dokumen ini sebagai referensi implementasi terbaru.

## Kondisi saat ini

- Backend Go dan frontend Vue/Vite berjalan dalam Docker Compose di OrbStack.
- Database PostgreSQL memakai migration otomatis saat app start.
- UI lokal: `http://localhost:8080`.
- Health check: `GET /healthz`.
- Posting ke X tetap manual melalui browser dan Chrome Extension helper.
- Tidak ada login atau penyimpanan identitas akun X.
- Media produk di-download ke local storage dan dikirim ke extension sebagai URL media lokal. Dari detail, user juga bisa menambah image/video URL baru dan menghapus media lokal yang sudah ada.

## Deploy lokal

Secret berada di `.env` lokal dan file tersebut di-ignore Git. Jangan memasukkan nilainya ke source code, dokumentasi, commit, atau log.

```bash
docker compose --env-file .env up -d --build
curl -i http://localhost:8080/healthz
```

Environment runtime aktif:

```text
OPENROUTER_BASE_URL=https://openrouter.ai/api/v1
OPENROUTER_MODEL=stealth/ox-alpha
OPENROUTER_API_KEY=<secret lokal, jangan ditulis di sini>
```

Config backend mendukung empat provider terpisah: model `openrouter/*` dikirim ke OpenRouter, `9router/*` ke base URL 9router, `opencode/*` ke OpenCode Zen dengan auth OpenCode, dan `codex/*` ke Codex CLI bridge di host. Daftar model di Settings diambil dari endpoint `/models` masing-masing provider dan memakai registry statis sebagai fallback. Model default lama tetap `stealth/ox-alpha`, tetapi pilihan tersimpan di browser.

Codex bridge lokal sudah dites dari dalam container pada `http://host.docker.internal:8787`. Model `codex/gpt-5.6-luna` diteruskan sebagai `gpt-5.6-luna` ke Codex CLI. Lihat [CODEX-CLI-BRIDGE.md](CODEX-CLI-BRIDGE.md).

## Struktur fitur utama

### Dashboard

- List produk dengan filter status, content model, cluster, search, dan pagination.
- Bulk pilih maksimal 10 produk.
- Bulk `Reformat AI` untuk caption baru bila diperlukan; operasi individual tetap tersedia di detail.
- Bulk `Buat varian caption` untuk variasi caption berbasis promo yang sudah ada.
- Dropdown AI tersimpan di `localStorage.ai_model`.
- Loading state mengunci tombol selama request AI agar tidak terjadi double-click.
- Content model dapat diubah dari list: `Trending`, `Branded`, `Murah`.
- Model AI dipilih dari halaman `/settings` dan disimpan di `localStorage.ai_model`.
- Model selector mengambil model dari OpenRouter, 9router, dan OpenCode; pencarian mempertahankan label provider.

### Product detail

- Raw text asli read-only.
- Promo text editable.
- Dropdown AI sendiri; pilihan tersinkron dengan dashboard melalui `localStorage.ai_model`.
- Dropdown content model: `Trending`, `Branded`, `Murah`.
- Produk manual dengan raw text memanggil Reformat AI otomatis setelah save.
- Detail memiliki `Reformat AI` utama yang selalu membangun promo dari raw text, serta `Reformat varian caption` yang memakai promo saat ini; varian berlaku untuk produk raw maupun import X.
- Source category ditampilkan sebagai `Raw text`, `Import X`, atau `Scrape Shopee`.
- Setiap produk memiliki `tracking_tag` unik yang dibuat saat save; produk lama di-backfill saat migration 009.
- Tracking tag ditampilkan di dashboard/detail dan bisa dicopy untuk dipakai saat membuat link affiliate Shopee.
- Loading spinner dan tombol terkunci selama AI bekerja.
- Link affiliate Shopee bisa diedit dari detail.
- Media lokal bisa ditambah dari URL image/video, dihapus, atau diunduh sebagai ZIP.
- Share ke X mengirim promo text dan media lokal.
- Jika promo text kosong setelah reset/gagal, tombol `Reformat AI` aktif kembali; `Save promo text` hanya aktif jika ada teks.

### Chrome Extension

Folder: `extension/x-helper/` dan `extension/shopee-scraper/`.

- Load masing-masing folder secara terpisah melalui `chrome://extensions`.
- `x-helper` membaca payload share dari halaman X dan membantu paste caption/media. Versi extension saat ini `1.8.3`.
- `shopee-scraper` membaca DOM, metadata, dan response network halaman detail Shopee yang sedang terbuka, mengambil raw text + media produk, lalu meng-insertkannya ke web app. Versi extension saat ini `1.2.9`; icon/logo/avatar kecil dan URL duplikat difilter.
- `x-helper` dan `shopee-scraper` adalah dua extension unpacked terpisah. Klik Reload di `chrome://extensions` setelah mengganti source extension.
- Untuk bridge Codex di macOS, double-click `tools/codex-bridge/start-codex-bridge.command`. Script membaca token dari `.env`, build binary jika perlu, dan menjalankan listener pada port `8787`; tutup Terminal atau tekan `Ctrl+C` untuk menghentikannya.
- Posting tetap harus dikonfirmasi manual oleh user di composer X.
- Jangan mengubah algoritma attach media tanpa reproduksi dan verifikasi di composer X; X dapat memproses upload media secara asynchronous.

## Content model dan strategi prompt

### Trending

Trend Capture: `DEMAND/TREND → DESIRE → PRODUCT → BENEFIT → PROOF → OFFER → CTA`.

Hook tidak boleh menyalin judul panjang. Hook memakai barang + value utama dari deskripsi + konteks kebutuhan. Jangan mengarang trend atau memakai “sekarang” jika tidak ada dasar raw.

### Branded

Brand Reminder: brand sudah memiliki trust. Hook mengingatkan brand + deal/diskon. Fokus utama: trust brand, discount, urgency. Jangan mengajar kualitas dari nol.

### Murah

Generic/Cheap Alternative: fokus harga/deal, kegunaan nyata, proof rating/terjual, dan value yang benar-benar berasal dari raw. Jangan mengganti proof dengan filler seperti “bahan berkualitas” atau “harga terjangkau”.

Semua model wajib memakai rating/bintang dan jumlah terjual jika tersedia, harga/diskon/voucher jika tersedia, harga termurah saja jika ada banyak varian, serta icon `✅`, `⭐️`, `🔥`, `💸`, `⚡`, `👇`.

Dokumentasi prompt lengkap: [AI-PROMPT-DOCUMENTATION.md](AI-PROMPT-DOCUMENTATION.md).

## Alur request AI

```text
Vue dashboard/detail
  → POST /api/ai/reformat
  → ProductService.Reformat (maksimal 10 product ID)
  → AIService.Reformat
  → provider terpilih (OpenRouter / 9router / OpenCode / Codex CLI bridge)
  → parse JSON array
  → normalizePromoLayout
  → validateAIResults
  → simpan reformatted_text dan field hasil
```

`Reformat AI` mengabaikan `CURRENT_PROMO` dan membuat ulang dari raw. `Reformat AI varian caption` memakai raw sebagai sumber fakta dan current promo sebagai caption dasar.

Timeout HTTP AI saat ini 3 menit karena Ox Alpha dapat lebih lambat daripada request pendek. UI harus tetap menampilkan loading selama request berjalan.

## File penting

- `../backend/internal/service/ai_service.go` — prompt, request provider, parse response, normalisasi caption.
- `../backend/internal/service/media_service.go` — download, penamaan berurutan, deduplikasi, dan penghapusan media lokal.
- `../backend/internal/service/model_registry.go` — fallback model registry.
- `../backend/internal/config/config.go` — environment provider AI.
- `../backend/internal/service/product_service.go` — validasi status dan penyimpanan hasil AI.
- `../backend/internal/db/migrations/008_add_source_category.up.sql`, `010_add_scrape_shopee_source.up.sql` — source category `raw_text`, `import_x`, atau `scrape_shopee`.
- `../backend/internal/db/migrations/009_add_tracking_tag.up.sql` — tracking tag unik per produk.
- `../frontend/src/views/HomeView.vue` — dashboard.
- `../frontend/src/views/ProductDetailView.vue` — detail, model selector, reformat, share.
- `../frontend/src/components/ModelSelector.vue` — dropdown AI bersama.
- `../frontend/src/stores/productStore.js` — API client dan loading state request AI.
- `../frontend/src/components/ShareButton.vue` — payload share ke extension/X.
- `extension/x-helper/background.js`, `extension/x-helper/content.js` — helper paste dan attach media.
- `extension/shopee-scraper/content.js` — ekstraksi produk Shopee dan pengiriman payload ke web app.
- `../backend/internal/db/migrations/007_add_branded_content_model.up.sql` — dukungan content model branded.

## Verifikasi setelah perubahan

```bash
env GOCACHE=/private/tmp/affiliator-go-cache go build ./...
cd frontend && npm run build
cd ..
docker compose --env-file .env up -d --build
curl -i http://localhost:8080/healthz
curl http://localhost:8080/api/ai/models
```

Jangan menjalankan reformat AI live hanya untuk test tanpa persetujuan, karena request dapat memakai quota provider dan mengubah data produk.

## Catatan risiko

- Model Ox Alpha dapat terkena rate limit/upstream timeout. Jangan menurunkan timeout kembali ke 45 detik tanpa pengujian.
- Backend saat ini menunggu response JSON lengkap; belum memakai streaming token.
- Auto-reformat saat create dilakukan oleh `ProductNewView.vue` setelah API create berhasil; jika AI gagal, produk tetap tersimpan dan error ditampilkan.
- `shopee_link` dikirim eksplisit sebagai fakta AI dan normalizer mengganti URL Shopee hasil model dengan link affiliate produk yang tersimpan. Ini mencegah UUID produk atau URL hasil halusinasi masuk ke caption.
- Media detail memakai endpoint `POST /api/products/{id}/media` untuk download URL baru dan `DELETE /api/products/{id}/media/{mediaID}` untuk menghapus file + metadata.
- Normalizer dapat mengubah hasil AI secara agresif. Jika output Hermes berbeda, review `normalizePromoLayout` selain prompt.
- Model registry dinamis mencoba `/models` pada OpenRouter, 9router, dan OpenCode; registry statis menjadi fallback jika discovery provider gagal.
- `.env` harus tetap untracked dan tidak boleh dibagikan.
