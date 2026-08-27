# Codex CLI Bridge

AffiliatorShopee dapat memakai session login Codex CLI di host macOS melalui
bridge lokal. Backend Docker hanya memanggil bridge dengan Bearer token.

## Menjalankan bridge

```sh
cd tools/codex-bridge
CGO_ENABLED=0 \
CODEX_BRIDGE_TOKEN='token-lokal-baru' \
CODEX_BRIDGE_ADDR='0.0.0.0:8787' \
go run .
```

Cara praktis di macOS: double-click
`tools/codex-bridge/start-codex-bridge.command`. Script mengambil
`CODEX_BRIDGE_TOKEN` dari `.env`, build binary jika diperlukan, lalu menjalankan
listener. Biarkan Terminal terbuka saat Reformat AI digunakan; tutup Terminal
atau tekan `Ctrl+C` untuk menghentikan bridge.

`CGO_ENABLED=0` diperlukan pada setup macOS ini agar bridge tidak gagal dengan
error `missing LC_UUID`. Jika `codex` tidak ditemukan oleh PATH, isi
`CODEX_CLI_PATH` dengan lokasi binary Codex CLI.

Tes koneksi:

```sh
curl -H 'Authorization: Bearer token-lokal-baru' http://127.0.0.1:8787/healthz
```

## Endpoint OpenAI-compatible

Selain endpoint legacy `/v1/execute`, bridge menyediakan `POST
/v1/chat/completions` dengan format Chat Completions OpenAI. Role `system`,
`user`, dan `assistant` sebelumnya digabung berurutan menjadi satu prompt tanpa
mengubah isi message. Model boleh ditulis sebagai `gpt-5.6-luna` atau
`codex/gpt-5.6-luna`; prefix `codex/` dihapus sebelum menjalankan CLI.

Bridge menjalankan `codex exec --ephemeral --sandbox read-only
--skip-git-repo-check --json --model <model>` dan mengembalikan hanya event
`item.completed` dengan `item.type == "agent_message"` sebagai
`choices[0].message.content`. Semua request wajib memakai
`Authorization: Bearer <CODEX_BRIDGE_TOKEN>`.

## Konfigurasi Hermes dan aplikasi

Hermes harus diarahkan langsung ke bridge lokal, bukan OpenRouter atau 9router:

```yaml
model:
  default: gpt-5.6-luna
  provider: custom
  base_url: http://host.docker.internal:8787/v1
  api_key: ${CODEX_BRIDGE_TOKEN}
```

Tambahkan ke `.env` lokal dan jangan commit:

```dotenv
CODEX_BRIDGE_URL=http://host.docker.internal:8787
CODEX_BRIDGE_TOKEN=token-lokal-baru
```

Contoh environment bridge:

```dotenv
CODEX_BRIDGE_TOKEN=token-lokal-baru
```

Alurnya adalah `Hermes → bridge lokal → Codex CLI → session login Codex`.
Bridge harus tetap aktif selama request AI berlangsung. Jalur ini berbeda dari
provider Codex melalui OpenRouter atau 9router: bridge lokal tidak mengirim
request ke gateway tersebut.

```sh
docker compose --env-file .env up -d --build
```

Pilih `Codex → GPT-5.6 Luna (Codex CLI)` di Settings. ID aplikasinya
`codex/gpt-5.6-luna`; backend meneruskan `gpt-5.6-luna` ke bridge.

Bridge harus tetap hidup selama Reformat AI. Jika berhenti, produk tetap
tersimpan sebagai raw dan dapat direformat ulang setelah bridge aktif.
