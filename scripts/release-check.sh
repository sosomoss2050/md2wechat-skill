#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$repo_root"

fail() {
  echo "release-check: $*" >&2
  exit 1
}

require_file() {
  local path="$1"
  [[ -f "$path" ]] || fail "missing required file: $path"
}

require_file "VERSION"
require_file "CHANGELOG.md"
require_file "README.md"
require_file "docs/INSTALL.md"
require_file "scripts/install.sh"
require_file "scripts/install.ps1"
require_file "scripts/install-openclaw.sh"
require_file "platforms/openclaw/md2wechat/SKILL.md"
require_file "platforms/openclaw/md2wechat/scripts/run.sh"
require_file ".github/workflows/release.yml"
require_file ".github/workflows/ci.yml"
require_file "docs/AGENTS.md"
require_file "docs/OPENCLAW.md"

version="$(tr -d '[:space:]' < VERSION)"
[[ "$version" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]] || fail "VERSION must be SemVer, got: $version"

changelog_version="$(sed -n 's/^## \[\([0-9][^]]*\)\].*/\1/p' CHANGELOG.md | head -n1)"
[[ -n "$changelog_version" ]] || fail "failed to read top CHANGELOG version"
[[ "$changelog_version" == "$version" ]] || fail "VERSION ($version) does not match CHANGELOG top version ($changelog_version)"

grep -q 'MD2WECHAT_VERSION' scripts/install.sh || fail "scripts/install.sh must support MD2WECHAT_VERSION"
grep -q 'MD2WECHAT_RELEASE_BASE_URL' scripts/install.sh || fail "scripts/install.sh must support MD2WECHAT_RELEASE_BASE_URL"
grep -q 'MD2WECHAT_INSTALL_DIR' scripts/install.sh || fail "scripts/install.sh must support MD2WECHAT_INSTALL_DIR"
grep -q 'MD2WECHAT_VERSION' scripts/install.ps1 || fail "scripts/install.ps1 must support MD2WECHAT_VERSION"
grep -q 'MD2WECHAT_RELEASE_BASE_URL' scripts/install.ps1 || fail "scripts/install.ps1 must support MD2WECHAT_RELEASE_BASE_URL"
grep -q 'MD2WECHAT_INSTALL_DIR' scripts/install.ps1 || fail "scripts/install.ps1 must support MD2WECHAT_INSTALL_DIR"
grep -q 'MD2WECHAT_NONINTERACTIVE' scripts/install.ps1 || fail "scripts/install.ps1 must support MD2WECHAT_NONINTERACTIVE"
grep -q 'MD2WECHAT_NO_PATH_UPDATE' scripts/install.ps1 || fail "scripts/install.ps1 must support MD2WECHAT_NO_PATH_UPDATE"
grep -q 'MD2WECHAT_RELEASE_BASE_URL' scripts/install-openclaw.sh || fail "scripts/install-openclaw.sh must support MD2WECHAT_RELEASE_BASE_URL"
grep -q 'MD2WECHAT_VERSION' scripts/install-openclaw.sh || fail "scripts/install-openclaw.sh must support MD2WECHAT_VERSION"
grep -q 'MD2WECHAT_OPENCLAW_INSTALL_DIR' scripts/install-openclaw.sh || fail "scripts/install-openclaw.sh must support MD2WECHAT_OPENCLAW_INSTALL_DIR"
grep -q 'MD2WECHAT_OPENCLAW_RUNTIME_DIR' scripts/install-openclaw.sh || fail "scripts/install-openclaw.sh must support MD2WECHAT_OPENCLAW_RUNTIME_DIR"
grep -q 'MD2WECHAT_NONINTERACTIVE' scripts/install-openclaw.sh || fail "scripts/install-openclaw.sh must support MD2WECHAT_NONINTERACTIVE"
! grep -q 'releases/latest/download' scripts/install-openclaw.sh || fail "scripts/install-openclaw.sh must not silently fall back to releases/latest/download"
grep -q 'checksums.txt' scripts/install.sh || fail "scripts/install.sh must verify checksums.txt"
grep -q 'checksums.txt' scripts/install.ps1 || fail "scripts/install.ps1 must verify checksums.txt"
grep -q 'checksums.txt' scripts/install-openclaw.sh || fail "scripts/install-openclaw.sh must verify checksums.txt"
grep -q 'md2wechat-openclaw-skill.tar.gz' scripts/install-openclaw.sh || fail "scripts/install-openclaw.sh must install the OpenClaw release archive"
grep -q 'md2wechat-linux-amd64' scripts/install-openclaw.sh || fail "scripts/install-openclaw.sh must install a versioned md2wechat runtime"
grep -q 'upload-artifact' .github/workflows/release.yml || fail "release workflow must upload a release artifact"
grep -q 'download-artifact' .github/workflows/release.yml || fail "release workflow must download a release artifact"
grep -q 'install.sh' .github/workflows/release.yml || fail "release workflow must publish install.sh"
grep -q 'install.ps1' .github/workflows/release.yml || fail "release workflow must publish install.ps1"
grep -q 'install-openclaw.sh' .github/workflows/release.yml || fail "release workflow must publish install-openclaw.sh"
grep -q 'md2wechat-openclaw-skill.tar.gz' .github/workflows/release.yml || fail "release workflow must publish the OpenClaw skill archive"
grep -q 'platforms/openclaw/md2wechat' .github/workflows/release.yml || fail "release workflow must package the OpenClaw-specific skill"
grep -q 'version --json' .github/workflows/release.yml || fail "release workflow must smoke version --json"
grep -q 'MD2WECHAT_RELEASE_BASE_URL' .github/workflows/release.yml || fail "release workflow must smoke the installers from the same bundle"
grep -q '"install"' platforms/openclaw/md2wechat/SKILL.md || fail "OpenClaw skill metadata must declare install resources"
grep -q 'releases/download/v' README.md || fail "README must point install instructions at fixed-version release assets"
grep -q 'releases/download/v' docs/INSTALL.md || fail "docs/INSTALL.md must point install instructions at fixed-version release assets"
grep -q 'releases/download/v' docs/OPENCLAW.md || fail "docs/OPENCLAW.md must point OpenClaw install instructions at fixed-version release assets"
! grep -q 'raw.githubusercontent.com/geekjourneyx/md2wechat-skill/main/scripts/install.sh' README.md || fail "README must not point the md2wechat installer at main"
! grep -q 'raw.githubusercontent.com/geekjourneyx/md2wechat-skill/main/scripts/install.ps1' README.md || fail "README must not point the md2wechat installer at main"
! grep -q 'raw.githubusercontent.com/geekjourneyx/md2wechat-skill/main/scripts/install-openclaw.sh' README.md || fail "README must not point the OpenClaw installer at main"
! grep -q 'raw.githubusercontent.com/geekjourneyx/md2wechat-skill/main/scripts/install.sh' docs/INSTALL.md || fail "docs/INSTALL.md must not point the md2wechat installer at main"
! grep -q 'raw.githubusercontent.com/geekjourneyx/md2wechat-skill/main/scripts/install.ps1' docs/INSTALL.md || fail "docs/INSTALL.md must not point the md2wechat installer at main"
! grep -q 'raw.githubusercontent.com/geekjourneyx/md2wechat-skill/main/scripts/install-openclaw.sh' docs/OPENCLAW.md || fail "docs/OPENCLAW.md must not point the OpenClaw installer at main"
grep -q 'VERSION' Makefile || fail "Makefile must reference VERSION"
grep -q 'artifact smoke' docs/AGENTS.md || fail "docs/AGENTS.md must mention artifact smoke"

echo "release-check: OK (version $version)"
