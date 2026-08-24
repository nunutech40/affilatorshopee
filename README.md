# AffiliatorShopee

Web app pribadi untuk menyimpan produk affiliate Shopee, merapikan data dengan AI, membuat caption, dan membantu posting manual ke X.

## Dokumen

- [PRD](PRD.md) - kebutuhan produk dan scope MVP
- [TRD](TRD.md) - arsitektur teknis dan kontrak API
- [TODO](TODO.md) - tahapan implementasi
- [Owner Requirement](Owner-Requirement.md) - konteks dan keputusan produk
- [AI Coder Guide](AI-CODER-GUIDE.md) - workflow DeepSeek Flash dan review Luna

## Alur MVP

```text
Simpan data Shopee mentah
→ AI reformat dan simpan hasil
→ Generate caption
→ Pilih hashtag
→ Share ke X
→ User upload media dan klik Post
→ Catat riwayat posting
```

MVP tidak melakukan auto-posting, tidak mengelola akun sosial, dan tidak meng-upload media ke platform secara otomatis.

Akun X yang digunakan mengikuti session browser yang sedang login. Web app tidak membaca, memilih, atau menyimpan identitas akun tersebut. Produk yang sama boleh dibagikan dan dicatat berkali-kali.

## Status

Dokumentasi MVP sudah siap untuk implementasi. Source code belum dibuat.

Setelah Phase 1 selesai, jalankan local stack dengan:

```bash
cp .env.example .env
docker compose up --build
```

Buka `http://localhost:8080` di browser.
