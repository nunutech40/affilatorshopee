#!/bin/bash
set -euo pipefail

# ===== YANG DITAMBAHIN =====
NEW_SRC="/Users/nununugraha/Documents/Programming/OtherPorject"
NEW_DST="/source/Programming"
# ===========================

echo "== Container yang jalan =="
docker ps --format 'table {{.Names}}\t{{.Image}}'
read -rp "Nama container Hermes (Enter = autodetect): " CNAME
[ -z "${CNAME}" ] && CNAME=$(docker ps --format '{{.Names}} {{.Image}}' \
  | grep -i hermes | awk '{print $1}' | head -1)
echo ">> Pakai container: $CNAME"
docker inspect "$CNAME" > /dev/null   # pastiin ada

J=$(docker inspect "$CNAME")

# --- bedah konfigurasi container lama ---
IMAGE=$(echo "$J" | python3 -c 'import json,sys;print(json.load(sys.stdin)[0]["Config"]["Image"])')

BINDS=$(echo "$J" | python3 -c '
import json,sys
for b in (json.load(sys.stdin)[0]["HostConfig"].get("Binds") or []): print("-v", b)')

PORTS=$(echo "$J" | python3 -c '
import json,sys
for k,v in (json.load(sys.stdin)[0]["HostConfig"].get("PortBindings") or {}).items():
    for m in v or []: print("-p %s:%s" % (m["HostPort"], k.split("/")[0]))')

ENVS=$(echo "$J" | python3 -c '
import json,sys,shlex
skip={"PATH","HOME","HOSTNAME","TERM","SHLVL","PWD"}
for e in (json.load(sys.stdin)[0]["Config"].get("Env") or []):
    if "=" in e and e.split("=",1)[0] not in skip: print("--env", shlex.quote(e))')

EXTRA=$(echo "$J" | python3 -c '
import json,sys,shlex
c=json.load(sys.stdin)[0]["HostConfig"]
rp=(c.get("RestartPolicy") or {}).get("Name","")
if rp and rp!="no": print("--restart", rp)
net=c.get("NetworkMode","")
if net and net not in ("default","bridge"): print("--network", net)
for h in (c.get("ExtraHosts") or []): print("--add-host", h)')

CMD=$(echo "$J" | python3 -c '
import json,sys,shlex
cmd=json.load(sys.stdin)[0]["Config"].get("Cmd") or []
print(shlex.join(cmd))')

echo ""
echo "================= COMMAND BARU ================="
RUN_CMD="docker run -d --name $CNAME $EXTRA $PORTS $ENVS $BINDS \
-v '$NEW_SRC:$NEW_DST' '$IMAGE' $CMD"
echo "$RUN_CMD"
echo "================================================"
read -rp "Lanjut stop container lama + recreate? (y/N): " GO
[ "$GO" != "y" ] && { echo "Batal."; exit 0; }

# --- eksekusi ---
docker stop "$CNAME"
docker rename "$CNAME" "$CNAME-old-backup"
eval "$RUN_CMD"

echo ""
echo ">> Beres! Container baru jalan, mount lama + baru lengkap."
echo ">> Container lama disimpan sebagai '$CNAME-old-backup' (hapus belakangan aja)."
docker ps --filter "name=$CNAME" --format 'table {{.Names}}\t{{.Status}}'

