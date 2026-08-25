# AffiliatorShopee

Web app pribadi untuk menyimpan produk affiliate Shopee, merapikan data dengan AI, membuat caption, dan membantu posting manual ke X.

## Dokumen

- [PRD](PRD.md) - kebutuhan produk dan scope MVP
- [TRD](TRD.md) - arsitektur teknis dan kontrak API
- [TODO](TODO.md) - tahapan implementasi
- [Owner Requirement](Owner-Requirement.md) - konteks dan keputusan produk
- [AI Coder Guide](AI-CODER-GUIDE.md) - workflow coding assistant dan review Luna

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

Core MVP sudah diimplementasikan. Test provider AI dan browser E2E tetap perlu dijalankan saat API key dan browser target tersedia. Ikuti `TODO.md` untuk progress implementasi.
