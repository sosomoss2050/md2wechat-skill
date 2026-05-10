#!/usr/bin/env bash

set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

echo "🔍 Running local/CI quality gates..."

echo "0) Check version consistency (fail fast)"
_ver="$(tr -d '[:space:]' < VERSION)"
_pkg="$(node -e "process.stdout.write(require('./package.json').version)" 2>/dev/null || echo "MISSING")"
_mp="$(sed -n 's/.*"version": "\([0-9][^"]*\)".*/\1/p' .claude-plugin/marketplace.json | head -n1)"
_cl="$(grep -m1 '^## \[[0-9]' CHANGELOG.md | sed 's/^## \[\([^]]*\)\].*/\1/')"
_ok=1
[[ -z "$_mp" ]] && _mp="MISSING"
[[ "$_pkg" != "$_ver" ]] && echo "   ❌ package.json ($_pkg) != VERSION ($_ver)" && _ok=0
[[ "$_mp"  != "$_ver" ]] && echo "   ❌ marketplace.json ($_mp) != VERSION ($_ver)" && _ok=0
[[ "$_cl"  != "$_ver" ]] && echo "   ❌ CHANGELOG.md top ($_cl) != VERSION ($_ver)" && _ok=0
if [[ "$_ok" == 0 ]]; then
  echo "   Fix: align all version files to VERSION=$_ver before proceeding."
  exit 1
fi
echo "   ✓ All versions aligned at $_ver"

echo "1) Check formatting"
unformatted="$(gofmt -l .)"
if [[ -n "$unformatted" ]]; then
  echo "Unformatted files:"
  printf '%s\n' "$unformatted"
  exit 1
fi

echo "2) Run go vet"
go vet ./...

echo "3) Run golangci-lint"
bash scripts/run-golangci-lint.sh

echo "4) Run tests"
CGO_ENABLED=1 go test -count=1 ./...

echo "5) Validate npm package contents"
npm_config_cache="${TMPDIR:-/tmp}/md2wechat-npm-cache" npm pack --json --dry-run >/dev/null

echo "6) Run release checks"
make release-check

echo "quality-gates: OK"
