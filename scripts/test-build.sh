#!/usr/bin/env bash
# Cross-platform build verification. Compiles for every supported OS/arch combo.
set -euo pipefail

cd "$(dirname "$0")/.."

green=$'\033[32m'; reset=$'\033[0m'

mkdir -p /tmp/sgai-builds
combos=("linux/amd64" "linux/arm64" "darwin/amd64" "darwin/arm64" "windows/amd64" "windows/arm64")

for combo in "${combos[@]}"; do
    GOOS="${combo%/*}"
    GOARCH="${combo#*/}"
    EXT=""
    [ "$GOOS" = "windows" ] && EXT=".exe"
    OUT="/tmp/sgai-builds/sgai-${GOOS}-${GOARCH}${EXT}"
    GOOS=$GOOS GOARCH=$GOARCH go build -trimpath -o "$OUT" ./cmd/system-general-ai
    SIZE=$(stat -c%s "$OUT" 2>/dev/null || stat -f%z "$OUT")
    printf '  %s✓%s %s (%s bytes)\n' "$green" "$reset" "$OUT" "$SIZE"
done

echo
echo "All cross-platform builds passed."

