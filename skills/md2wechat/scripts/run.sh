#!/bin/bash
#
# md2wechat - Minimal binary provisioner
# Philosophy: Simple, fast, single responsibility
#

set -e

# =============================================================================
# CONFIGURATION
# =============================================================================

REPO="geekjourneyx/md2wechat-skill"
BINARY_NAME="md2wechat"

# Cache directory (tool-specific, not Claude's cache)
CACHE_DIR="${XDG_CACHE_HOME:-${HOME}/.cache}/md2wechat"
BIN_DIR="${CACHE_DIR}/bin"
VERSION_FILE="${CACHE_DIR}/.version"

# Minimum binary size (100KB - protects against corrupted downloads)
MIN_BINARY_SIZE=102400

get_version() {
    if [[ -n "${MD2WECHAT_SKILL_VERSION:-}" ]]; then
        printf '%s\n' "${MD2WECHAT_SKILL_VERSION}"
        return 0
    fi

    local script_dir repo_root version_file
    script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
    repo_root="$(cd "${script_dir}/../../.." && pwd)"
    version_file="${repo_root}/VERSION"

    if [[ -f "${version_file}" ]]; then
        tr -d '[:space:]' < "${version_file}"
        return 0
    fi

    printf '2.0.1\n'
}

VERSION="$(get_version)"

# =============================================================================
# PLATFORM DETECTION
# =============================================================================

detect_platform() {
    local os arch

    os=$(uname -s | tr '[:upper:]' '[:lower:]')
    arch=$(uname -m)

    case "$arch" in
        x86_64|amd64) arch="amd64" ;;
        arm64|aarch64) arch="arm64" ;;
        *) echo "Unsupported architecture: $arch" >&2; return 1 ;;
    esac

    case "$os" in
        darwin|linux) echo "${os}-${arch}" ;;
        msys*|mingw*|cygwin*) echo "windows-${arch}" ;;
        *) echo "Unsupported OS: $os" >&2; return 1 ;;
    esac
}

# =============================================================================
# BINARY MANAGEMENT
# =============================================================================

get_binary_path() {
    local platform=$1
    local path="${BIN_DIR}/${BINARY_NAME}-${platform}"
    [[ "$platform" == windows-* ]] && path="${path}.exe"
    echo "$path"
}

is_cache_valid() {
    local binary=$1
    [[ -x "$binary" ]] && [[ -f "$VERSION_FILE" ]] && [[ "$(cat "$VERSION_FILE" 2>/dev/null)" == "$VERSION" ]]
}

extract_version_from_json() {
    local output=$1
    printf '%s\n' "$output" | sed -n 's/.*"version":[[:space:]]*"\([^"]*\)".*/\1/p' | head -n1
}

binary_matches_version() {
    local candidate=$1
    local resolved_candidate script_path output actual_version

    [[ -x "$candidate" ]] || return 1

    resolved_candidate="$(cd "$(dirname "$candidate")" && pwd)/$(basename "$candidate")"
    script_path="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/$(basename "${BASH_SOURCE[0]}")"
    if [[ "$resolved_candidate" == "$script_path" ]]; then
        return 1
    fi

    output="$("$candidate" version --json 2>/dev/null)" || return 1
    actual_version="$(extract_version_from_json "$output")"
    [[ -n "$actual_version" ]] || return 1
    [[ "$actual_version" == "$VERSION" ]]
}

get_release_base_url() {
    if [[ -n "${MD2WECHAT_SKILL_RELEASE_BASE_URL:-}" ]]; then
        printf '%s\n' "${MD2WECHAT_SKILL_RELEASE_BASE_URL}"
        return 0
    fi

    printf 'https://github.com/%s/releases/download/v%s\n' "${REPO}" "${VERSION}"
}

verify_checksum() {
    local checksums_file=$1
    local binary_file=$2
    local binary_name=$3

    if command -v sha256sum >/dev/null 2>&1; then
        (cd "$(dirname "$binary_file")" && sha256sum -c "$checksums_file" --ignore-missing --status)
        return 0
    fi

    if command -v shasum >/dev/null 2>&1; then
        local expected actual
        expected="$(awk -v file="$binary_name" '$2 == file { print $1 }' "$checksums_file")"
        [[ -n "$expected" ]] || return 1
        actual="$(shasum -a 256 "$binary_file" | awk '{print $1}')"
        [[ "$expected" == "$actual" ]]
        return 0
    fi

    echo "  Error: sha256sum or shasum required for checksum verification" >&2
    return 1
}

download_binary() {
    local platform=$1
    local binary=$2
    local bin_name="${BINARY_NAME}-${platform}"
    [[ "$platform" == windows-* ]] && bin_name="${bin_name}.exe"

    local release_base_url url checksums_url
    release_base_url="$(get_release_base_url)"
    url="${release_base_url}/${bin_name}"
    checksums_url="${release_base_url}/checksums.txt"
    local temp_file="${binary}.tmp"
    local temp_checksums="${binary}.checksums.tmp"

    echo "  Downloading md2wechat v${VERSION} for ${platform}..." >&2

    # Check for download tool
    if command -v curl &>/dev/null; then
        if ! curl -fsSL --connect-timeout 15 --max-time 120 -o "$temp_file" "$url" 2>/dev/null; then
            rm -f "$temp_file" "$temp_checksums"
            echo "  Error: Failed to download runtime from GitHub Releases" >&2
            return 1
        fi
        if ! curl -fsSL --connect-timeout 15 --max-time 120 -o "$temp_checksums" "$checksums_url" 2>/dev/null; then
            rm -f "$temp_file" "$temp_checksums"
            echo "  Error: Failed to download checksums.txt from GitHub Releases" >&2
            return 1
        fi
    elif command -v wget &>/dev/null; then
        if ! wget -q --timeout=120 -O "$temp_file" "$url" 2>/dev/null; then
            rm -f "$temp_file" "$temp_checksums"
            echo "  Error: Failed to download runtime from GitHub Releases" >&2
            return 1
        fi
        if ! wget -q --timeout=120 -O "$temp_checksums" "$checksums_url" 2>/dev/null; then
            rm -f "$temp_file" "$temp_checksums"
            echo "  Error: Failed to download checksums.txt from GitHub Releases" >&2
            return 1
        fi
    else
        echo "  Error: curl or wget required" >&2
        echo "  Install: brew install curl (macOS) or apt install curl (Linux)" >&2
        return 1
    fi

    if [[ ! -f "$temp_file" ]]; then
        rm -f "$temp_checksums"
        echo "  Error: Runtime download did not produce a file" >&2
        return 1
    fi

    # Validate file size
    local size
    size=$(wc -c < "$temp_file" 2>/dev/null | tr -d ' ') || size=0
    if [[ $size -lt $MIN_BINARY_SIZE ]]; then
        rm -f "$temp_file" "$temp_checksums"
        echo "  Error: Download incomplete or corrupted" >&2
        return 1
    fi

    if ! verify_checksum "$temp_checksums" "$temp_file" "$bin_name"; then
        rm -f "$temp_file" "$temp_checksums"
        echo "  Error: Checksum verification failed" >&2
        return 1
    fi

    mv "$temp_file" "$binary"
    rm -f "$temp_checksums"
    chmod +x "$binary"
    echo "$VERSION" > "$VERSION_FILE"
    echo "  Ready!" >&2
}

ensure_binary() {
    local platform
    platform=$(detect_platform) || return 1

    local binary
    binary=$(get_binary_path "$platform")

    # Fast path: valid cache
    if is_cache_valid "$binary"; then
        echo "$binary"
        return 0
    fi

    # Try local development binary
    local script_dir
    script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
    local local_binary="${script_dir}/bin/${BINARY_NAME}-${platform}"
    [[ "$platform" == windows-* ]] && local_binary="${local_binary}.exe"

    if binary_matches_version "$local_binary"; then
        echo "$local_binary"
        return 0
    fi

    # Try an existing installed CLI on PATH before downloading a new runtime
    if command -v md2wechat >/dev/null 2>&1; then
        local path_binary
        path_binary="$(command -v md2wechat)"
        if binary_matches_version "$path_binary"; then
            echo "$path_binary"
            return 0
        fi
        echo "  Found md2wechat on PATH, but version does not match v${VERSION}" >&2
    fi

    # Download
    mkdir -p "$BIN_DIR" 2>/dev/null || {
        echo "  Error: Cannot create cache directory: $BIN_DIR" >&2
        return 1
    }

    if download_binary "$platform" "$binary"; then
        echo "$binary"
        return 0
    fi

    echo "" >&2
    echo "  Download failed. Try:" >&2
    echo "    • If md2wechat is already installed, add it to PATH and rerun the skill" >&2
    echo "    • If GitHub Releases CDN is blocked, set MD2WECHAT_SKILL_RELEASE_BASE_URL to a reachable mirror" >&2
    echo "    • Download manually: https://github.com/${REPO}/releases" >&2
    echo "" >&2
    return 1
}

# =============================================================================
# MAIN
# =============================================================================

main() {
    local binary
    binary=$(ensure_binary) || exit 1
    exec "$binary" "$@"
}

main "$@"
