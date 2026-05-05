#!/usr/bin/env bash
# nea-ai installer for Linux and macOS.
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/RDuuke/nea-ai/main/scripts/install.sh | bash
#
# Optional environment variables:
#   NEA_AI_VERSION   Pin a specific tag (e.g. v0.2.0). Defaults to latest release.
#   NEA_AI_BIN_DIR   Install directory. Defaults to $HOME/.local/bin.

set -euo pipefail

REPO="RDuuke/nea-ai"
BINARY="nea-ai"
DEFAULT_BIN_DIR="${HOME}/.local/bin"
BIN_DIR="${NEA_AI_BIN_DIR:-${DEFAULT_BIN_DIR}}"

err() {
  printf 'error: %s\n' "$*" >&2
  exit 1
}

info() {
  printf '==> %s\n' "$*"
}

require() {
  command -v "$1" >/dev/null 2>&1 || err "missing required command: $1"
}

detect_downloader() {
  if command -v curl >/dev/null 2>&1; then
    DOWNLOADER="curl"
  elif command -v wget >/dev/null 2>&1; then
    DOWNLOADER="wget"
  else
    err "need either curl or wget on PATH"
  fi
}

download() {
  local url="$1"
  local out="$2"
  if [ "${DOWNLOADER}" = "curl" ]; then
    curl -fsSL "${url}" -o "${out}"
  else
    wget -q -O "${out}" "${url}"
  fi
}

fetch_text() {
  local url="$1"
  if [ "${DOWNLOADER}" = "curl" ]; then
    curl -fsSL "${url}"
  else
    wget -q -O- "${url}"
  fi
}

detect_os() {
  local raw
  raw="$(uname -s)"
  case "${raw}" in
    Linux) OS="linux" ;;
    Darwin) OS="darwin" ;;
    *) err "unsupported OS: ${raw}" ;;
  esac
}

detect_arch() {
  local raw
  raw="$(uname -m)"
  case "${raw}" in
    x86_64|amd64) ARCH="amd64" ;;
    aarch64|arm64) ARCH="arm64" ;;
    *) err "unsupported architecture: ${raw}" ;;
  esac
}

resolve_version() {
  if [ -n "${NEA_AI_VERSION:-}" ]; then
    TAG="${NEA_AI_VERSION}"
    return
  fi
  local body
  body="$(fetch_text "https://api.github.com/repos/${REPO}/releases/latest")" \
    || err "failed to query latest release"
  TAG="$(printf '%s' "${body}" | grep -E '"tag_name"' | head -n1 | sed -E 's/.*"tag_name": *"([^"]+)".*/\1/')"
  [ -n "${TAG}" ] || err "could not parse tag_name from GitHub API response"
}

verify_checksum() {
  local archive="$1"
  local sums="$2"
  local archive_name
  archive_name="$(basename "${archive}")"
  local expected
  expected="$(grep " ${archive_name}\$" "${sums}" | awk '{print $1}')"
  [ -n "${expected}" ] || err "no checksum entry for ${archive_name}"
  local actual
  if command -v sha256sum >/dev/null 2>&1; then
    actual="$(sha256sum "${archive}" | awk '{print $1}')"
  elif command -v shasum >/dev/null 2>&1; then
    actual="$(shasum -a 256 "${archive}" | awk '{print $1}')"
  else
    err "need sha256sum or shasum on PATH"
  fi
  if [ "${expected}" != "${actual}" ]; then
    err "checksum mismatch for ${archive_name} (expected ${expected}, got ${actual})"
  fi
}

main() {
  require uname
  require tar
  require mkdir
  require install
  detect_downloader
  detect_os
  detect_arch
  resolve_version

  local version_no_v="${TAG#v}"
  local archive_name="${BINARY}_${version_no_v}_${OS}_${ARCH}.tar.gz"
  local archive_url="https://github.com/${REPO}/releases/download/${TAG}/${archive_name}"
  local sums_url="https://github.com/${REPO}/releases/download/${TAG}/checksums.txt"

  TMP_DIR="$(mktemp -d)"
  trap 'rm -rf "${TMP_DIR}"' EXIT

  info "downloading ${archive_name}"
  download "${archive_url}" "${TMP_DIR}/${archive_name}"
  download "${sums_url}" "${TMP_DIR}/checksums.txt"

  info "verifying sha256"
  verify_checksum "${TMP_DIR}/${archive_name}" "${TMP_DIR}/checksums.txt"

  info "extracting archive"
  tar -xzf "${TMP_DIR}/${archive_name}" -C "${TMP_DIR}"

  mkdir -p "${BIN_DIR}"
  install -m 0755 "${TMP_DIR}/${BINARY}" "${BIN_DIR}/${BINARY}"

  info "installed ${BINARY} ${TAG} -> ${BIN_DIR}/${BINARY}"

  case ":${PATH}:" in
    *":${BIN_DIR}:"*) ;;
    *)
      printf '\n%s is not in your PATH.\nAdd this line to your shell profile (~/.bashrc, ~/.zshrc):\n\n  export PATH="%s:$PATH"\n\n' "${BIN_DIR}" "${BIN_DIR}"
      ;;
  esac
}

main "$@"
