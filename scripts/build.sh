#!/usr/bin/env bash
# Build flclash-web for all desktop targets.
# Requires: Go >= 1.27 (GOTOOLCHAIN=auto fetches it), Node.js >= 20.
set -euo pipefail
root="$(cd "$(dirname "$0")/.." && pwd)"
cd "$root"

echo "== 1/3 building frontend =="
(cd frontend && npm install --no-audit --no-fund && npm run build)

echo "== 2/3 embedding frontend =="
rm -rf backend/internal/api/dist
cp -r frontend/dist backend/internal/api/dist

echo "== 3/3 building binaries =="
mkdir -p dist
cd backend
export GOTOOLCHAIN=auto CGO_ENABLED=0
for target in windows-amd64 linux-amd64 linux-arm64 darwin-amd64 darwin-arm64; do
    os="${target%-*}"
    arch="${target#*-}"
    ext=""
    [ "$os" = "windows" ] && ext=".exe"
    out="../dist/flclash-web-$os-$arch$ext"
    echo "building $out"
    GOOS="$os" GOARCH="$arch" go build -trimpath -ldflags "-s -w" -o "$out" .
done
echo "done -> dist/"
