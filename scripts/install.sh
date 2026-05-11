#!/usr/bin/env bash
# system-general-ai bootstrap — Linux / macOS
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/lucianorepetti/system-general-ai/main/scripts/install.sh | bash
#
# This script:
#   1. Detects OS + arch
#   2. Downloads the matching binary from GitHub Releases
#   3. Places it at ~/.local/bin/system-general-ai (+ symlink sgai)
#   4. Ensures ~/.local/bin is on PATH (adds to ~/.profile if missing)
#   5. Runs `system-general-ai install` with default settings

set -euo pipefail

REPO="lucianorepetti/system-general-ai"
BIN_NAME="system-general-ai"
TARGET_DIR="${HOME}/.local/bin"

err() { printf '\033[31merror:\033[0m %s\n' "$*" >&2; exit 1; }
info() { printf '\033[34m::\033[0m %s\n' "$*"; }
ok() { printf '\033[32m✓\033[0m %s\n' "$*"; }

detect_os_arch() {
  local os arch
  case "$(uname -s)" in
    Linux)  os="linux" ;;
    Darwin) os="darwin" ;;
    *) err "Unsupported OS: $(uname -s). This script supports Linux and macOS." ;;
  esac
  case "$(uname -m)" in
    x86_64|amd64) arch="amd64" ;;
    arm64|aarch64) arch="arm64" ;;
    *) err "Unsupported architecture: $(uname -m)." ;;
  esac
  printf '%s_%s' "$os" "$arch"
}

latest_tag() {
  curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" \
    | grep '"tag_name"' \
    | head -n1 \
    | sed -E 's/.*"tag_name": *"([^"]+)".*/\1/'
}

ensure_path() {
  case ":${PATH}:" in
    *":${TARGET_DIR}:"*) return ;;
  esac
  local profile="${HOME}/.profile"
  [ -f "${HOME}/.zshrc" ] && profile="${HOME}/.zshrc"
  [ -f "${HOME}/.bashrc" ] && [ -z "${ZSH_VERSION:-}" ] && profile="${HOME}/.bashrc"
  printf '\n# Added by system-general-ai installer\nexport PATH="%s:$PATH"\n' "${TARGET_DIR}" >> "${profile}"
  info "Added ${TARGET_DIR} to PATH in ${profile}. Open a new shell to pick it up."
}

main() {
  command -v curl >/dev/null 2>&1 || err "curl is required."

  local platform tag url archive tmpdir
  platform="$(detect_os_arch)"
  tag="$(latest_tag)"
  [ -n "${tag}" ] || err "Could not determine latest release tag from ${REPO}."

  info "Installing ${BIN_NAME} ${tag} for ${platform}"

  archive="${BIN_NAME}_${tag}_${platform}.tar.gz"
  url="https://github.com/${REPO}/releases/download/${tag}/${archive}"

  tmpdir="$(mktemp -d)"
  trap 'rm -rf "${tmpdir}"' EXIT

  curl -fsSL "${url}" -o "${tmpdir}/${archive}" \
    || err "Failed to download ${url}"
  tar -xzf "${tmpdir}/${archive}" -C "${tmpdir}" \
    || err "Failed to extract archive."

  mkdir -p "${TARGET_DIR}"
  install -m 0755 "${tmpdir}/${BIN_NAME}" "${TARGET_DIR}/${BIN_NAME}" \
    || err "Failed to install binary to ${TARGET_DIR}."
  ln -sf "${TARGET_DIR}/${BIN_NAME}" "${TARGET_DIR}/sgai"

  ok "Binary installed at ${TARGET_DIR}/${BIN_NAME} (alias: sgai)"

  ensure_path

  info "Running zero-config install ..."
  "${TARGET_DIR}/${BIN_NAME}" install || err "system-general-ai install failed."

  ok "Done. Restart Claude Code to apply."
}

main "$@"

