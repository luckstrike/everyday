#!/bin/sh
set -eu
cd "$(dirname "$0")/.."
GO=${GO:-go}
if ! command -v "$GO" >/dev/null 2>&1; then
  echo "Go 1.22+ is required. On macOS: brew install go" >&2
  exit 1
fi
if [ ! -d frontend/node_modules ]; then
  echo "Install frontend dependencies first: make install" >&2
  exit 1
fi
# Build first, then run the binary directly so Ctrl-C can stop the actual server.
bin_dir=$(mktemp -d "${TMPDIR:-/tmp}/everyday.XXXXXX")
backend_pid=
frontend_pid=
cleanup() {
  trap - EXIT INT TERM
  [ -z "$backend_pid" ] || kill "$backend_pid" 2>/dev/null || true
  [ -z "$frontend_pid" ] || kill "$frontend_pid" 2>/dev/null || true
  wait 2>/dev/null || true
  rm -rf "$bin_dir"
}
trap cleanup EXIT
trap 'exit 130' INT TERM
(cd backend && "$GO" build -o "$bin_dir/server" ./cmd/server)
"$bin_dir/server" &
backend_pid=$!
(cd frontend && exec node node_modules/vite/bin/vite.js --host 127.0.0.1) &
frontend_pid=$!
# Exit if either process fails instead of leaving half of the application running.
while kill -0 "$backend_pid" 2>/dev/null && kill -0 "$frontend_pid" 2>/dev/null; do
  sleep 1
done
exit 1
