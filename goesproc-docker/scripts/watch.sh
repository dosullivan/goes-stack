apk add --no-cache docker-cli netcat-openbsd &&
cat >/usr/local/bin/watchdog.sh <<'SH'
#!/bin/sh
set -eu

host="${TARGET_HOST:?set TARGET_HOST}"
port="${TARGET_PORT:?set TARGET_PORT}"
interval="${CHECK_INTERVAL:-5}"
up_need="${UP_STREAK:-2}"
down_need="${DOWN_STREAK:-3}"
svc="${SERVICE_NAME:-goesproc}"
cooldown="${COOLDOWN_SEC:-30}"

consec_up=0
consec_down=0
was_down=0     # flip to 1 once we've confirmed a down state
last_restart=0

log() { echo "[watchdog] $*"; }

while :; do
  if nc -z -w2 "$host" "$port" >/dev/null 2>&1; then
    consec_up=$((consec_up+1))
    consec_down=0
    if [ "$was_down" -eq 1 ] && [ "$consec_up" -ge "$up_need" ]; then
now=$(date +%s)
if [ $((now - last_restart)) -ge "$cooldown" ]; then
  log "target ${host}:${port} recovered (up ${consec_up}x) → restarting ${svc}"
  docker restart "$svc" >/dev/null 2>&1 || log "restart failed"
  last_restart="$now"
else
  log "recovered but in cooldown; skipping restart"
fi
was_down=0
    fi
  else
    consec_down=$((consec_down+1))
    consec_up=0
    if [ "$consec_down" -ge "$down_need" ]; then
if [ "$was_down" -eq 0 ]; then
  log "target ${host}:${port} considered DOWN (fail ${consec_down}x)"
fi
was_down=1
    fi
  fi
  sleep "$interval"
done
SH
chmod +x /usr/local/bin/watchdog.sh
exec /usr/local/bin/watchdog.sh