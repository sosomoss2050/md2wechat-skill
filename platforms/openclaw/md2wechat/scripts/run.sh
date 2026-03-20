#!/usr/bin/env bash
#
# md2wechat OpenClaw runtime wrapper
#
# This wrapper only executes an already-installed runtime.
# It does not download binaries at execution time.
#

set -euo pipefail

RUNTIME_DIR="${MD2WECHAT_OPENCLAW_RUNTIME_DIR:-${HOME}/.openclaw/tools/md2wechat}"
SKILL_VERSION="${MD2WECHAT_OPENCLAW_SKILL_VERSION:-2.0.1}"

extract_version_from_json() {
    local output=$1
    printf '%s\n' "$output" | sed -n 's/.*"version":[[:space:]]*"\([^"]*\)".*/\1/p' | head -n1
}

runtime_matches_version() {
    local candidate=$1
    local output actual_version

    [[ -x "$candidate" ]] || return 1

    output="$("$candidate" version --json 2>/dev/null)" || return 1
    actual_version="$(extract_version_from_json "$output")"
    [[ -n "$actual_version" ]] || return 1
    [[ "$actual_version" == "$SKILL_VERSION" ]]
}

find_runtime() {
    local candidate

    if [[ -n "${MD2WECHAT_OPENCLAW_RUNTIME:-}" && -x "${MD2WECHAT_OPENCLAW_RUNTIME}" ]]; then
        printf '%s\n' "${MD2WECHAT_OPENCLAW_RUNTIME}"
        return 0
    fi

    candidate="${RUNTIME_DIR}/md2wechat"
    if [[ -x "$candidate" ]]; then
        printf '%s\n' "$candidate"
        return 0
    fi

    candidate="${RUNTIME_DIR}/bin/md2wechat"
    if [[ -x "$candidate" ]]; then
        printf '%s\n' "$candidate"
        return 0
    fi

    if command -v md2wechat >/dev/null 2>&1; then
        command -v md2wechat
        return 0
    fi

    return 1
}

main() {
    local runtime
    if ! runtime="$(find_runtime)"; then
        cat >&2 <<'EOF'
md2wechat runtime not found.

Expected one of:
  - ~/.openclaw/tools/md2wechat/md2wechat
  - ~/.openclaw/tools/md2wechat/bin/md2wechat
  - md2wechat on PATH

clawhub install md2wechat currently installs only the skill shell and may not provision the runtime.
Install the fixed-version OpenClaw bundle and runtime first, for example:

  export MD2WECHAT_RELEASE_BASE_URL="https://github.com/geekjourneyx/md2wechat-skill/releases/download/v2.0.1"
  curl -fsSL "${MD2WECHAT_RELEASE_BASE_URL}/install-openclaw.sh" | bash

Or manually place the runtime at:
  ~/.openclaw/tools/md2wechat/md2wechat

You can also set MD2WECHAT_OPENCLAW_RUNTIME=/absolute/path/to/md2wechat.
EOF
        exit 1
    fi

    if ! runtime_matches_version "$runtime"; then
        cat >&2 <<EOF
md2wechat runtime version mismatch.

Expected: ${SKILL_VERSION}
Found runtime: ${runtime}

Reinstall the fixed-version OpenClaw bundle and runtime, for example:

  export MD2WECHAT_RELEASE_BASE_URL="https://github.com/geekjourneyx/md2wechat-skill/releases/download/v${SKILL_VERSION}"
  curl -fsSL "\${MD2WECHAT_RELEASE_BASE_URL}/install-openclaw.sh" | bash

If you intentionally want to run a different but matching runtime, set:
  MD2WECHAT_OPENCLAW_RUNTIME=/absolute/path/to/md2wechat
  MD2WECHAT_OPENCLAW_SKILL_VERSION=${SKILL_VERSION}
EOF
        exit 1
    fi

    exec "$runtime" "$@"
}

main "$@"
