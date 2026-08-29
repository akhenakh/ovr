#!/usr/bin/env bash
# Generate the Homebrew formula for a release and push it to akhenakh/homebrew-tap.
#
# Usage: scripts/brew-bump.sh <tag>    (e.g. scripts/brew-bump.sh v0.2)
#
# Requires write access to akhenakh/homebrew-tap, either via a
# TAP_GITHUB_TOKEN (a PAT with contents:write on the tap) or an SSH key.
set -euo pipefail

TAG="${1:?usage: $0 <tag>}"
VERSION="${TAG#v}"

BIN="ovr"
CLASS="Ovr"
REPO="akhenakh/ovr"
DESC="CLI tool to pipe anything into and apply transformations with an advanced UI"
LICENSE="MIT"

TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT

curl -fsSL "https://github.com/${REPO}/releases/download/${TAG}/${BIN}_${VERSION}_checksums.txt" \
  -o "$TMP/checksums.txt"

sha() {
  awk -v f="$1" '$2 == f { print $1 }' "$TMP/checksums.txt"
}

DARWIN_AMD64="$(sha "${BIN}_Darwin_x86_64.tar.gz")"
DARWIN_ARM64="$(sha "${BIN}_Darwin_arm64.tar.gz")"
LINUX_AMD64="$(sha "${BIN}_Linux_x86_64.tar.gz")"
LINUX_ARM64="$(sha "${BIN}_Linux_arm64.tar.gz")"

UIBIN="ovrui"
UI_DARWIN_ARM64="$(sha "${UIBIN}_Darwin_arm64.tar.gz")"
UI_LINUX_AMD64="$(sha "${UIBIN}_Linux_x86_64.tar.gz")"
UI_LINUX_ARM64="$(sha "${UIBIN}_Linux_arm64.tar.gz")"

for v in DARWIN_AMD64 DARWIN_ARM64 LINUX_AMD64 LINUX_ARM64 \
         UI_DARWIN_ARM64 UI_LINUX_AMD64 UI_LINUX_ARM64; do
  if [[ -z "${!v}" ]]; then
    echo "error: missing checksum for ${v}" >&2
    exit 1
  fi
done

cat > "$TMP/${BIN}.rb" <<EOF
# typed: false
# frozen_string_literal: true

class ${CLASS} < Formula
  desc "${DESC}"
  homepage "https://github.com/${REPO}"
  license "${LICENSE}"

  on_macos do
    on_arm do
      url "https://github.com/${REPO}/releases/download/${TAG}/${BIN}_Darwin_arm64.tar.gz"
      sha256 "${DARWIN_ARM64}"
    end
    on_intel do
      url "https://github.com/${REPO}/releases/download/${TAG}/${BIN}_Darwin_x86_64.tar.gz"
      sha256 "${DARWIN_AMD64}"
    end
  end

  on_linux do
    on_arm do
      url "https://github.com/${REPO}/releases/download/${TAG}/${BIN}_Linux_arm64.tar.gz"
      sha256 "${LINUX_ARM64}"
    end
    on_intel do
      url "https://github.com/${REPO}/releases/download/${TAG}/${BIN}_Linux_x86_64.tar.gz"
      sha256 "${LINUX_AMD64}"
    end
  end

  # ovrui ships in its own archives, as a resource
  on_linux do
    on_arm do
      resource "ovrui" do
        version "${TAG}"
        url "https://github.com/${REPO}/releases/download/${TAG}/${UIBIN}_Linux_arm64.tar.gz"
        sha256 "${UI_LINUX_ARM64}"
      end
    end
    on_intel do
      resource "ovrui" do
        version "${TAG}"
        url "https://github.com/${REPO}/releases/download/${TAG}/${UIBIN}_Linux_x86_64.tar.gz"
        sha256 "${UI_LINUX_AMD64}"
      end
    end
  end

  on_macos do
    on_arm do
      resource "ovrui" do
        version "${TAG}"
        url "https://github.com/${REPO}/releases/download/${TAG}/${UIBIN}_Darwin_arm64.tar.gz"
        sha256 "${UI_DARWIN_ARM64}"
      end
    end
  end

  def install
    bin.install "ovr"
    on_linux do
      resource("ovrui").stage { bin.install "ovrui" }
    end
    on_macos do
      on_arm do
        resource("ovrui").stage { bin.install "ovrui" }
      end
    end
  end
end
EOF

TAP_DIR="$TMP/tap"
if [[ -n "${TAP_GITHUB_TOKEN:-}" ]]; then
  git clone "https://x-access-token:${TAP_GITHUB_TOKEN}@github.com/akhenakh/homebrew-tap.git" "$TAP_DIR"
else
  git clone "git@github.com:akhenakh/homebrew-tap.git" "$TAP_DIR"
fi
mkdir -p "$TAP_DIR/Formula"
cp "$TMP/${BIN}.rb" "$TAP_DIR/Formula/${BIN}.rb"

git -C "$TAP_DIR" config user.email "akh@inair.space"
git -C "$TAP_DIR" config user.name "Fabrice Aneche"
git -C "$TAP_DIR" add "Formula/${BIN}.rb"
git -C "$TAP_DIR" commit -m "Brew formula update for ${BIN} ${TAG}" || true
git -C "$TAP_DIR" push origin main
