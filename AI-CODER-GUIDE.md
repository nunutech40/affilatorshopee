# AI Coder Guide - AffiliatorShopee

Dokumen ini mengatur penggunaan AI coding assistant, termasuk DeepSeek Flash, untuk mengimplementasikan project. Dokumen ini tidak mengatur model AI runtime untuk fitur reformat produk.

## Model yang Berbeda

- **Coding model:** DeepSeek Flash boleh digunakan untuk membuat dan mengubah source code.
- **Review model:** Luna melakukan review akhir terhadap code, security, schema, API, dan testing.
- **Runtime model:** fitur AI reformat memakai `OPENROUTER_MODEL` di backend. Model ini dikonfigurasi terpisah dan tidak otomatis berubah ketika coding model diganti.

## Pekerjaan yang Boleh Dikerjakan DeepSeek Flash

- Setup Go module, struktur folder, config, dan Docker.
- Migration PostgreSQL sesuai `TRD.md`.
- Model, repository, service, handler, dan router.
- CRUD products, post logs, dan caption variations.
- AI service OpenRouter dengan mock HTTP test.
- Caption template engine, formatter, hashtag validation, dan character count.
- Vue views, components, Pinia stores, API client, loading state, dan error state.
- Unit test, API test, migration test, dan frontend test.
- README teknis dan checklist `TODO.md`.

## Pekerjaan yang Tidak Boleh Diputuskan Sendiri

- Mengubah scope MVP atau menambahkan fitur roadmap.
- Mengembalikan fitur account, Threads, Chrome Extension, atau media storage ke MVP.
- Mengubah schema, status transition, atau API contract tanpa memperbarui PRD dan TRD.
- Menghapus atau menimpa `raw_text`.
- Mengarang proof, harga, rating, sold count, review count, atau urgency.
- Mengirim API key, password, file `.env`, atau data sensitif ke prompt atau log.
- Memakai API key OpenRouter asli dalam automated test.
- Menjalankan migration destructive, menghapus database, atau mengakses production.
- Melakukan `git reset --hard`, `git checkout --`, force push, atau menghapus file tanpa kebutuhan task.
- Commit atau push perubahan sebelum review Luna, kecuali owner meminta secara eksplisit.

## Workflow Wajib

1. Baca `PRD.md`, `TRD.md`, `TODO.md`, dan file yang relevan.
2. Kerjakan satu phase atau satu vertical slice saja.
3. Ikuti kontrak yang sudah ada; jangan menebak requirement baru.
4. Gunakan mock untuk OpenRouter dan data fixture untuk test.
5. Jalankan formatter dan test yang relevan.
6. Periksa `git diff` dan `git diff --check`.
7. Laporkan file yang berubah, command test, hasil test, dan risiko yang tersisa.
8. Hentikan pekerjaan bila menemukan konflik requirement; jangan menyelesaikannya dengan asumsi.
9. Luna melakukan review akhir sebelum merge atau push.

## Guardrail Implementasi

- Semua endpoint harus mengikuti response envelope dan HTTP status code di `TRD.md`.
- Produk baru berstatus `raw`; AI menghasilkan `reformatted`; edit manual dapat menghasilkan `ready`.
- Produk boleh dicatat posting berkali-kali dan tidak memiliki status `posted`.
- Web app tidak login, memilih, membaca, atau menyimpan identitas akun X.
- Share X dibuka melalui user gesture dan clipboard hanya sebagai fallback.
- `raw_text` selalu dipertahankan.
- AI output harus diparse dan divalidasi sebelum database di-update.
- Request AI dibatasi maksimal 20 produk dan memiliki timeout serta concurrency limit.
- Caption final, character count, clipboard, share URL, dan post log harus menggunakan text yang konsisten.
- Image/video MVP hanya berupa URL eksternal; backend tidak mengunduh atau menyimpan file media.
- MVP hanya memakai `content_model` `capture` dan `cheap`.

## Definition of Ready untuk Luna Review

Review dapat dimulai bila:

- `go test ./...` berhasil.
- Frontend test dan `npm run build` berhasil.
- Migration dapat dijalankan pada database kosong.
- Docker Compose berhasil start di localhost.
- Test AI tidak membutuhkan API key asli.
- Tidak ada secret atau file `.env` di diff.
- Acceptance criteria PRD dapat didemonstrasikan melalui browser.
