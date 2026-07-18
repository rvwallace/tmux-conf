#!/usr/bin/env bash

set -euo pipefail

mode="${1:---check}"
case "${mode}" in
  --check | --install) ;;
  *)
    printf 'usage: %s [--check|--install]\n' "$0" >&2
    exit 2
    ;;
esac

repo_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
commands=(bash btop bun fd fzf fzf-tmux git lazygit nvim onefetch python3 sesh tmux zoxide)
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
    apt:bash | apt:btop | apt:fzf | apt:git | apt:tmux | apt:zoxide) printf '%s' "${command_name}" ;;
    dnf:bash | dnf:btop | dnf:fzf | dnf:git | dnf:tmux | dnf:zoxide) printf '%s' "${command_name}" ;;
    pacman:bash | pacman:btop | pacman:bun | pacman:fd | pacman:fzf | pacman:git | pacman:lazygit | pacman:onefetch | pacman:tmux | pacman:zoxide) printf '%s' "${command_name}" ;;
    *) return 1 ;;
  esac
}

print_linux_guidance() {
  local command_name
  printf '\nCommands still unavailable should be installed from their official sources:\n'
  for command_name in "${missing[@]}"; do
    case "${command_name}" in
      bun) printf '  bun:      https://bun.sh/docs/installation\n' ;;
      sesh) printf '  sesh:     https://github.com/joshmedeski/sesh#installation\n' ;;
      lazygit) printf '  lazygit:  https://github.com/jesseduffield/lazygit#installation\n' ;;
      onefetch) printf '  onefetch: https://github.com/o2sh/onefetch#installation\n' ;;
      fd) printf '  fd:       distro package may install the binary as fd-find; add an fd symlink if needed\n' ;;
      *) printf '  %s\n' "${command_name}" ;;
    esac
  done
  printf 'Clipboard support is optional: install wl-clipboard (Wayland) or xclip (X11).\n'
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
  elif command -v dnf >/dev/null 2>&1; then
    manager="dnf"
    install_command=(sudo dnf install)
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
