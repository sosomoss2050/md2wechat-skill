#!/usr/bin/env bash
#
# md2wechat OpenClaw runtime wrapper
#
# This wrapper only executes an already-installed runtime.
# It does not download binaries at execution time.
#

set -euo pipefail

RUNTIME_DIR="${MD2WECHAT_OPENCLAW_RUNTIME_DIR:-${HOME}/.openclaw/tools/md2wechat}"

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

Install or reinstall the OpenClaw skill package so the runtime is provisioned first.
You can also set MD2WECHAT_OPENCLAW_RUNTIME=/absolute/path/to/md2wechat.
EOF
        exit 1
    fi

    exec "$runtime" "$@"
}

main "$@"
