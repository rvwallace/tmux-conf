#!/usr/bin/env bash
# Cloud Agent install for tmux-conf. Idempotent. Run from /workspace.
set -euo pipefail

export DEBIAN_FRONTEND=noninteractive
export PATH="/usr/local/go/bin:/usr/local/bin:${HOME}/.local/bin:${HOME}/.bun/bin:${HOME}/go/bin:${HOME}/.cargo/bin:${PATH}"

cd /workspace

sudo apt-get update
sudo apt-get install -y --no-install-recommends \
  bison \
  btop \
  build-essential \
  ca-certificates \
  curl \
  fd-find \
  fzf \
  git \
  jq \
  libevent-dev \
  libncurses-dev \
  neovim \
  pkg-config \
  python3 \
  ripgrep \
  tar \
  tree \
  xclip \
  zoxide

# The Tokyo Night theme uses message-format, which tmux 3.7 added. The image
# tmux is older and appears earlier on PATH, so install 3.8 and link it ahead.
TMUX_VERSION="3.8"
TMUX_SHA256="e79c699c7e949dccd0a4a125e17b8d1261e311b16979dbc8eb34542f3966d82e"
if ! /usr/local/bin/tmux -V 2>/dev/null | grep -q "tmux ${TMUX_VERSION}"; then
  tmp="$(mktemp -d)"
  curl -fsSL -o "${tmp}/tmux-${TMUX_VERSION}.tar.gz" \
    "https://github.com/tmux/tmux/releases/download/${TMUX_VERSION}/tmux-${TMUX_VERSION}.tar.gz"
  echo "${TMUX_SHA256}  ${tmp}/tmux-${TMUX_VERSION}.tar.gz" | sha256sum -c -
  tar -C "${tmp}" -xzf "${tmp}/tmux-${TMUX_VERSION}.tar.gz"
  (
    cd "${tmp}/tmux-${TMUX_VERSION}"
    ./configure --prefix=/usr/local
    make -j"$(nproc)"
    sudo make install
  )
  rm -rf "${tmp}"
fi
sudo mkdir -p /usr/local/cargo/bin "${HOME}/.local/bin"
sudo ln -sfn /usr/local/bin/tmux /usr/local/cargo/bin/tmux
ln -sfn /usr/local/bin/tmux "${HOME}/.local/bin/tmux"
hash -r

sudo ln -sfn "$(command -v fdfind)" /usr/local/bin/fd

GO_VERSION="1.26.5"
GO_SHA256="5c2c3b16caefa1d968a94c1daca04a7ca301a496d9b086e17ad77bb81393f053"
if ! /usr/local/go/bin/go version 2>/dev/null | grep -q "go${GO_VERSION} "; then
  tmp="$(mktemp)"
  curl -fsSL -o "${tmp}" "https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz"
  echo "${GO_SHA256}  ${tmp}" | sha256sum -c -
  sudo rm -rf /usr/local/go
  sudo tar -C /usr/local -xzf "${tmp}"
  rm -f "${tmp}"
fi
sudo ln -sfn /usr/local/go/bin/go /usr/local/bin/go
sudo ln -sfn /usr/local/go/bin/gofmt /usr/local/bin/gofmt

if ! command -v bun >/dev/null 2>&1; then
  curl -fsSL https://bun.sh/install | bash
fi
sudo ln -sfn "${HOME}/.bun/bin/bun" /usr/local/bin/bun

if ! command -v uv >/dev/null 2>&1; then
  curl -LsSf https://astral.sh/uv/install.sh | sh
fi
if [ -x "${HOME}/.local/bin/uv" ]; then
  sudo ln -sfn "${HOME}/.local/bin/uv" /usr/local/bin/uv
elif [ -x "${HOME}/.cargo/bin/uv" ]; then
  sudo ln -sfn "${HOME}/.cargo/bin/uv" /usr/local/bin/uv
else
  echo "uv installer finished without a binary" >&2
  exit 1
fi

install_archive_bin() {
  local url="$1"
  local sha256="$2"
  local archive_name="$3"
  local bin_name="$4"
  local inner_path="$5"
  if command -v "${bin_name}" >/dev/null 2>&1; then
    return 0
  fi
  local tmp
  tmp="$(mktemp -d)"
  curl -fsSL -o "${tmp}/${archive_name}" "${url}"
  echo "${sha256}  ${tmp}/${archive_name}" | sha256sum -c -
  tar -C "${tmp}" -xzf "${tmp}/${archive_name}"
  sudo install -m 755 "${tmp}/${inner_path}" "/usr/local/bin/${bin_name}"
  rm -rf "${tmp}"
}

install_deb() {
  local url="$1"
  local sha256="$2"
  local bin_name="$3"
  if command -v "${bin_name}" >/dev/null 2>&1; then
    return 0
  fi
  local tmp
  tmp="$(mktemp --suffix=.deb)"
  curl -fsSL -o "${tmp}" "${url}"
  echo "${sha256}  ${tmp}" | sha256sum -c -
  sudo dpkg -i "${tmp}"
  rm -f "${tmp}"
}

install_archive_bin \
  "https://github.com/jesseduffield/lazygit/releases/download/v0.66.2/lazygit_0.66.2_linux_x86_64.tar.gz" \
  "158efc9ddfa1241fab5fb3526bb442c8dc6c4b1904d73213b9e4e93e03c52034" \
  "lazygit_0.66.2_linux_x86_64.tar.gz" \
  "lazygit" \
  "lazygit"

install_archive_bin \
  "https://github.com/joshmedeski/sesh/releases/download/v2.33.0/sesh_Linux_x86_64.tar.gz" \
  "e04fe35778eaf768a1fbbaa6e01d008ae22038c52b8d7834d38c9e7ad4822daa" \
  "sesh_Linux_x86_64.tar.gz" \
  "sesh" \
  "sesh"

install_archive_bin \
  "https://github.com/raine/tmux-snaglord/releases/download/v0.1.8/x86_64-unknown-linux-gnu.tar.gz" \
  "388099442aa34332af6cbb03290d1987f67e766ebe88b09cf2ba8316ecb491d6" \
  "x86_64-unknown-linux-gnu.tar.gz" \
  "tmux-snaglord" \
  "tmux-snaglord"

install_deb \
  "https://github.com/sxyazi/yazi/releases/download/v26.9.1/yazi-x86_64-unknown-linux-gnu.deb" \
  "331d6247c0a0fd549f14cd599a928a529b8b22eded1bb1f61a518376c7621b45" \
  "yazi"

install_deb \
  "https://github.com/o2sh/onefetch/releases/download/3.0.0/onefetch_amd64.deb" \
  "cba1e74cc8360acbade4e7678f67b02e77c5f428a340568359778a7f65233126" \
  "onefetch"

hash -r
./scripts/install-deps.sh --check
./bootstrap.sh

plugin_tmp="$(mktemp -d)"
export TMUX_TMPDIR="${plugin_tmp}"
"${HOME}/.config/tmux/plugins/tpm/bin/install_plugins"
tmux kill-server >/dev/null 2>&1 || true
unset TMUX_TMPDIR
rm -rf "${plugin_tmp}"

plugins=(
  tmux-sensible
  tmux-cowboy
  tmux-fzf
  extrakto
  tmux-fzf-links
  tmux-resurrect
  tmux-continuum
)
for plugin in "${plugins[@]}"; do
  if [ ! -d "${HOME}/.config/tmux/plugins/${plugin}" ]; then
    echo "TPM plugin missing: ${plugin}" >&2
    exit 1
  fi
done

echo "tmux-conf cloud install complete"
