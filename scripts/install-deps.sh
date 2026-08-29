#!/usr/bin/env bash

set -euo pipefail

export PATH="${HOME}/.cargo/bin:${HOME}/.local/bin:${HOME}/go/bin:${PATH}"

mode="${1:---check}"
case "${mode}" in
  --check | --install) ;;
  *)
    printf 'usage: %s [--check|--install]\n' "$0" >&2
    exit 2
    ;;
esac

repo_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
commands=(bash btop bun fd fzf fzf-tmux git go jq lazygit nvim onefetch python3 sesh tmux tmux-snaglord tree uv yazi zoxide)
missing=()

find_missing() {
  missing=()
  local command_name
  for command_name in "${commands[@]}"; do
    command -v "${command_name}" >/dev/null 2>&1 || missing+=("${command_name}")
  done
}

print_missing() {
  if [ "${#missing[@]}" -eq 0 ]; then
    printf 'Dependencies satisfied.\n'
    return 0
  fi
  printf 'Missing commands: %s\n' "${missing[*]}"
  return 1
}

install_macos() {
  if ! command -v brew >/dev/null 2>&1; then
    printf 'Homebrew is required: https://brew.sh\n' >&2
    return 1
  fi
  if ! brew bundle -h >/dev/null 2>&1; then
    printf 'Homebrew Bundle is required. Run: brew tap Homebrew/bundle\n' >&2
    return 1
  fi
  brew bundle install --file="${repo_dir}/Brewfile"
}

linux_package_for() {
  local manager="$1"
  local command_name="$2"
  case "${manager}:${command_name}" in
    apt:fd | dnf:fd) printf 'fd-find' ;;
    apt:nvim | dnf:nvim | pacman:nvim) printf 'neovim' ;;
    apt:python3 | dnf:python3) printf 'python3' ;;
    pacman:python3) printf 'python' ;;
    apt:fzf-tmux | dnf:fzf-tmux | pacman:fzf-tmux) printf 'fzf' ;;
    apt:go) printf 'golang-go' ;;
    dnf:go) printf 'golang' ;;
    pacman:go) printf 'go' ;;
    dnf:bun)
      if dnf -C list --available bun-bin >/dev/null 2>&1 || dnf -C list --installed bun-bin >/dev/null 2>&1; then
        printf 'bun-bin'
      else
        printf 'bun'
      fi
      ;;
    apt:bash | apt:btop | apt:fzf | apt:git | apt:jq | apt:tmux | apt:tree | apt:zoxide) printf '%s' "${command_name}" ;;
    dnf:bash | dnf:btop | dnf:fzf | dnf:git | dnf:jq | dnf:onefetch | dnf:tmux | dnf:tree | dnf:uv | dnf:zoxide) printf '%s' "${command_name}" ;;
    pacman:bash | pacman:btop | pacman:bun | pacman:fd | pacman:fzf | pacman:git | pacman:jq | pacman:lazygit | pacman:onefetch | pacman:tmux | pacman:tree | pacman:uv | pacman:yazi | pacman:zoxide) printf '%s' "${command_name}" ;;
    *) return 1 ;;
  esac
}

print_linux_guidance() {
  local command_name
  printf '\nCommands still unavailable should be installed from their official sources:\n'
  for command_name in "${missing[@]}"; do
    case "${command_name}" in
      bun)
        if command -v dnf >/dev/null 2>&1; then
          printf '  bun:           sudo dnf install bun-bin (or curl -fsSL https://bun.sh/install | bash)\n'
        else
          printf '  bun:           curl -fsSL https://bun.sh/install | bash (or https://bun.sh/docs/installation)\n'
        fi
        ;;
      go) printf '  go:            https://go.dev/doc/install\n' ;;
      sesh)
        if command -v dnf >/dev/null 2>&1; then
          printf '  sesh:          sudo dnf copr enable buckaroogeek/Tmux_sesh && sudo dnf install tmux-sesh (or https://github.com/joshmedeski/sesh#installation)\n'
        else
          printf '  sesh:          go install github.com/joshmedeski/sesh/v2@latest (or https://github.com/joshmedeski/sesh#installation)\n'
        fi
        ;;
      lazygit)
        if command -v dnf >/dev/null 2>&1; then
          printf '  lazygit:       sudo dnf copr enable atim/lazygit && sudo dnf install lazygit (or https://github.com/jesseduffield/lazygit#installation)\n'
        else
          printf '  lazygit:       go install github.com/jesseduffield/lazygit@latest (or https://github.com/jesseduffield/lazygit#installation)\n'
        fi
        ;;
      onefetch)
        if command -v dnf >/dev/null 2>&1; then
          printf '  onefetch:      sudo dnf install onefetch (or https://github.com/o2sh/onefetch#installation)\n'
        elif command -v cargo >/dev/null 2>&1; then
          printf '  onefetch:      cargo install onefetch (or https://github.com/o2sh/onefetch#installation)\n'
        else
          printf '  onefetch:      https://github.com/o2sh/onefetch#installation\n'
        fi
        ;;
      tmux-snaglord) printf '  tmux-snaglord: cargo install --locked tmux-snaglord\n' ;;
      uv)
        if command -v dnf >/dev/null 2>&1; then
          printf '  uv:            sudo dnf install uv (or curl -LsSf https://astral.sh/uv/install.sh | sh)\n'
        else
          printf '  uv:            curl -LsSf https://astral.sh/uv/install.sh | sh (or https://docs.astral.sh/uv/getting-started/installation/)\n'
        fi
        ;;
      yazi)
        if command -v dnf >/dev/null 2>&1; then
          printf '  yazi:          sudo dnf copr enable atim/yazi && sudo dnf install yazi (or https://yazi-rs.github.io/docs/installation)\n'
        else
          printf '  yazi:          https://yazi-rs.github.io/docs/installation\n'
        fi
        ;;
      fd) printf '  fd:            distro package may install the binary as fd-find; add an fd symlink if needed\n' ;;
      *) printf '  %s\n' "${command_name}" ;;
    esac
  done
  printf 'Clipboard support is optional: install wl-clipboard (Wayland) or xclip (X11).\n'
}

setup_fedora_repos() {
  if ! rpm -q rpmfusion-free-release >/dev/null 2>&1 || ! rpm -q rpmfusion-nonfree-release >/dev/null 2>&1; then
    printf 'Enabling RPM Fusion (free and nonfree)...\n'
    local fedora_ver
    fedora_ver="$(rpm -E %fedora 2>/dev/null || true)"
    if [ -n "${fedora_ver}" ] && [ "${fedora_ver}" != "%fedora" ]; then
      sudo dnf install -y \
        "https://mirrors.rpmfusion.org/free/fedora/rpmfusion-free-release-${fedora_ver}.noarch.rpm" \
        "https://mirrors.rpmfusion.org/nonfree/fedora/rpmfusion-nonfree-release-${fedora_ver}.noarch.rpm" || true
    fi
  fi

  if ! rpm -q terra-release >/dev/null 2>&1 && [ ! -f /etc/yum.repos.d/terra.repo ]; then
    printf 'Enabling Terra repository...\n'
    sudo dnf install -y --nogpgcheck --repofrompath 'terra,https://repos.fyralabs.com/terra$releasever' terra-release 2>/dev/null || \
      sudo dnf config-manager addrepo --from-repofile=https://terra.fyralabs.com/terra.repo 2>/dev/null || true
  fi
}

install_linux() {
  local manager
  local -a install_command packages
  if command -v pacman >/dev/null 2>&1; then
    manager="pacman"
    install_command=(sudo pacman -S --needed)
  elif command -v apt >/dev/null 2>&1; then
    manager="apt"
    install_command=(sudo apt install)
    printf 'Updating apt package index...\n'
    sudo apt update || true
  elif command -v dnf >/dev/null 2>&1; then
    manager="dnf"
    install_command=(sudo dnf install)
    setup_fedora_repos
    printf 'Ensuring DNF metadata cache is populated...\n'
    dnf makecache || true
  else
    printf 'No supported Linux package manager found (pacman, apt, or dnf).\n' >&2
    print_linux_guidance
    return 1
  fi

  packages=()
  local command_name package_name
  for command_name in "${missing[@]}"; do
    if package_name="$(linux_package_for "${manager}" "${command_name}")"; then
      if [[ ! " ${packages[*]} " =~ " ${package_name} " ]]; then
        packages+=("${package_name}")
      fi
    fi
  done
  if [ "${#packages[@]}" -gt 0 ]; then
    printf 'Installing with %s: %s\n' "${manager}" "${packages[*]}"
    "${install_command[@]}" "${packages[@]}"
  fi

  if [ "${manager}" = "dnf" ]; then
    if [[ " ${missing[*]} " =~ " sesh " ]] && ! command -v sesh >/dev/null 2>&1; then
      printf 'Attempting sesh installation via COPR...\n'
      sudo dnf copr enable -y buckaroogeek/Tmux_sesh && sudo dnf install -y tmux-sesh || true
    fi
    if [[ " ${missing[*]} " =~ " lazygit " ]] && ! command -v lazygit >/dev/null 2>&1; then
      printf 'Attempting lazygit installation via COPR...\n'
      sudo dnf copr enable -y atim/lazygit && sudo dnf install -y lazygit || true
    fi
    if [[ " ${missing[*]} " =~ " yazi " ]] && ! command -v yazi >/dev/null 2>&1; then
      printf 'Attempting yazi installation via COPR...\n'
      sudo dnf copr enable -y atim/yazi && sudo dnf install -y yazi || true
    fi
  fi

  if [[ " ${missing[*]} " =~ " tmux-snaglord " ]] && ! command -v tmux-snaglord >/dev/null 2>&1; then
    if command -v cargo >/dev/null 2>&1; then
      printf 'Installing tmux-snaglord with cargo...\n'
      cargo install --locked tmux-snaglord || true
    fi
  fi

  if [[ " ${missing[*]} " =~ " sesh " ]] && ! command -v sesh >/dev/null 2>&1; then
    if command -v go >/dev/null 2>&1; then
      printf 'Installing sesh with go...\n'
      go install github.com/joshmedeski/sesh/v2@latest || true
    fi
  fi

  if [[ " ${missing[*]} " =~ " lazygit " ]] && ! command -v lazygit >/dev/null 2>&1; then
    if command -v go >/dev/null 2>&1; then
      printf 'Installing lazygit with go...\n'
      go install github.com/jesseduffield/lazygit@latest || true
    fi
  fi

  if [[ " ${missing[*]} " =~ " uv " ]] && ! command -v uv >/dev/null 2>&1; then
    if command -v curl >/dev/null 2>&1; then
      printf 'Installing uv with official installer...\n'
      curl -LsSf https://astral.sh/uv/install.sh | sh || true
    fi
  fi

  if [[ " ${missing[*]} " =~ " bun " ]] && ! command -v bun >/dev/null 2>&1; then
    if command -v curl >/dev/null 2>&1; then
      printf 'Installing bun with official installer...\n'
      curl -fsSL https://bun.sh/install | bash || true
    fi
  fi

  if ! command -v fd >/dev/null 2>&1 && command -v fdfind >/dev/null 2>&1; then
    mkdir -p "${HOME}/.local/bin"
    ln -sf "$(command -v fdfind)" "${HOME}/.local/bin/fd"
    export PATH="${HOME}/.local/bin:${PATH}"
    printf 'Linked %s -> %s\n' "${HOME}/.local/bin/fd" "$(command -v fdfind)"
  fi

  find_missing
  if ! print_missing; then
    print_linux_guidance
    return 1
  fi
}

find_missing
if [ "${mode}" = "--check" ]; then
  print_missing
  exit $?
fi

case "$(uname -s)" in
  Darwin) install_macos ;;
  Linux) install_linux ;;
  *)
    printf 'Unsupported platform: %s\n' "$(uname -s)" >&2
    exit 1
    ;;
esac

find_missing
print_missing
