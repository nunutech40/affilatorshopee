# Codex CLI bridge

Bridge ini menjalankan Codex CLI di host macOS agar backend dapat memakai sesi
login lokal tanpa memasang `~/.codex` atau token ke container. Default bridge
hanya listen di `127.0.0.1`; jangan membuka ke seluruh jaringan.

Jalankan di host:

```sh
cd tools/codex-bridge
CODEX_BRIDGE_TOKEN='buat-token-lokal-yang-panjang' go run .
```

Untuk menjalankan dari Finder, double-click `start-codex-bridge.command` di
folder ini. Script akan membaca `CODEX_BRIDGE_TOKEN` dari `.env`, build binary
jika diperlukan, lalu menjalankan bridge pada port `8787`. Biarkan jendela
Terminal tetap terbuka selama fitur Reformat AI dipakai; tutup jendela atau
tekan `Ctrl+C` untuk menghentikannya.

Jika runtime Docker tidak dapat menjangkau loopback host, set
`CODEX_BRIDGE_ADDR` ke interface host yang memang dipakai Docker dan lindungi
dengan firewall; token Bearer tetap wajib.

Lalu isi token yang sama pada `.env` aplikasi:

```dotenv
CODEX_BRIDGE_URL=http://host.docker.internal:8787
CODEX_BRIDGE_TOKEN=buat-token-lokal-yang-panjang
```

Endpoint yang tersedia:

- `GET /healthz`
- `POST /v1/execute` (endpoint legacy, tetap tersedia)
- `POST /v1/chat/completions` (OpenAI-compatible)

Contoh request chat:

```sh
curl -sS http://127.0.0.1:8787/v1/chat/completions \
  -H 'Authorization: Bearer buat-token-lokal-yang-panjang' \
  -H 'Content-Type: application/json' \
  -d '{"model":"codex/gpt-5.6-luna","messages":[{"role":"system","content":"Jawab singkat."},{"role":"user","content":"Halo"}],"stream":false}'
```

Prefix `codex/` boleh dipakai oleh client dan dihapus bridge sebelum model
dikirim ke CLI. `temperature`, `max_tokens`, dan parameter chat lain yang tidak
didukung Codex diterima untuk kompatibilitas tetapi tidak diteruskan ke CLI.
Auth Codex tetap dibaca oleh CLI host, bukan oleh aplikasi.
