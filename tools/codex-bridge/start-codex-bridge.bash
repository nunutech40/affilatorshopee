#!/bin/bash

# Jalankan Codex bridge untuk aplikasi AffiliatorShopee.
# File ini aman dijalankan dari Finder maupun Terminal.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/../.." && pwd)"
ENV_FILE="$PROJECT_DIR/.env"
BRIDGE_BIN="$SCRIPT_DIR/affiliatorshopee-codex-bridge"

if [[ ! -f "$ENV_FILE" ]]; then
  echo "File .env tidak ditemukan: $ENV_FILE"
  echo "Buat .env dulu dan isi CODEX_BRIDGE_TOKEN yang sama dengan konfigurasi Docker."
  read -r -p "Tekan Enter untuk menutup..."
  exit 1
fi

# Ambil hanya token bridge dari .env; variabel lain tidak dieksekusi sebagai shell.
CODEX_BRIDGE_TOKEN="$(sed -n 's/^CODEX_BRIDGE_TOKEN[[:space:]]*=[[:space:]]*//p' "$ENV_FILE" | tail -n 1)"
CODEX_BRIDGE_TOKEN="${CODEX_BRIDGE_TOKEN#\"}"
CODEX_BRIDGE_TOKEN="${CODEX_BRIDGE_TOKEN%\"}"
CODEX_BRIDGE_TOKEN="${CODEX_BRIDGE_TOKEN#\'}"
CODEX_BRIDGE_TOKEN="${CODEX_BRIDGE_TOKEN%\'}"

if [[ -z "$CODEX_BRIDGE_TOKEN" ]]; then
  echo "CODEX_BRIDGE_TOKEN kosong di $ENV_FILE"
  read -r -p "Tekan Enter untuk menutup..."
  exit 1
fi

if ! command -v go >/dev/null 2>&1; then
  echo "Go tidak ditemukan di PATH."
  read -r -p "Tekan Enter untuk menutup..."
  exit 1
fi

# Build otomatis bila binary belum ada atau source lebih baru.
if [[ ! -x "$BRIDGE_BIN" || "$SCRIPT_DIR/main.go" -nt "$BRIDGE_BIN" ]]; then
  echo "Membangun Codex bridge..."
  (cd "$SCRIPT_DIR" && CGO_ENABLED=0 go build -o "$BRIDGE_BIN" .)
fi

export CODEX_BRIDGE_TOKEN
export CODEX_BRIDGE_ADDR="${CODEX_BRIDGE_ADDR:-0.0.0.0:8787}"

echo "Codex bridge aktif di http://127.0.0.1:8787"
echo "Docker memakai host.docker.internal:8787"
echo "Tutup jendela Terminal ini atau tekan Ctrl+C untuk menghentikan."
echo

exec "$BRIDGE_BIN"
