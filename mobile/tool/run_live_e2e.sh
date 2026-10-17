#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
TMPDB="$(mktemp -d)"
DB="$TMPDB/e2e.sqlite3"
MEDIA="$TMPDB/media"
mkdir -p "$MEDIA"

echo "TMPDB=$TMPDB DB=$DB MEDIA=$MEDIA"

# ensure static/dist exists (go embed may need it)
if [ ! -d "$ROOT/static/dist" ] || [ -z "$(ls -A "$ROOT/static/dist" 2>/dev/null)" ]; then
  echo "Building static/dist..."
  (cd "$ROOT" && npm run build 2>&1 | tail -20 || echo "npm build failed or not needed")
fi

export SECRET_KEY=devsecret
export ALLOWED_HOSTS=localhost,127.0.0.1
export CSRF_TRUSTED_ORIGINS=http://127.0.0.1:6270
export DATABASE_FILE="$DB"
export MEDIA_ROOT="$MEDIA"
export PORT=6270

echo "Starting Go server..."
go run . >"$TMPDB/server.log" 2>&1 &
PID=$!
echo "PID $PID"

cleanup() {
  echo "Killing server $PID"
  kill $PID 2>/dev/null || true
  wait $PID 2>/dev/null || true
  echo "--- server.log tail ---"
  tail -n 100 "$TMPDB/server.log" || true
  rm -rf "$TMPDB"
}
trap cleanup EXIT

echo "Waiting for /api/v1/ping..."
for i in $(seq 1 30); do
  if curl -sf http://127.0.0.1:6270/api/v1/ping >/dev/null 2>&1; then
    echo "Server ready"
    break
  fi
  sleep 0.5
  if [ "$i" -eq 30 ]; then
    echo "Server failed to start"
    cat "$TMPDB/server.log"
    exit 1
  fi
done

echo "Running e2e test..."
(cd "$ROOT/mobile" && SANDWITCHES_E2E_URL=http://127.0.0.1:6270 flutter test test/e2e/live_e2e_test.dart)
STATUS=$?
# prevent trap double rm; keep log on success briefly then cleanup will remove
exit $STATUS
