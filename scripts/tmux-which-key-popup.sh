#!/usr/bin/env bash

set -euo pipefail

mode="${1-}"

pause() {
  printf '\nPress Enter to close\n'
  read -r _
}

resurrect_dir() {
  local dir
  dir="$(tmux show-option -gqv @resurrect-dir)"
  if [ -z "$dir" ]; then
    if [ -d "$HOME/.tmux/resurrect" ]; then
      dir="$HOME/.tmux/resurrect"
    else
      dir="${XDG_DATA_HOME:-$HOME/.local/share}/tmux/resurrect"
    fi
  fi
  dir="${dir/#\~/$HOME}"
  dir="${dir//\$HOME/$HOME}"
  dir="${dir//\$HOSTNAME/$(hostname)}"
  printf '%s\n' "$dir"
}

show_resurrect_files() {
  local dir
  dir="$(resurrect_dir)"
  if [ -d "$dir" ]; then
    ls -lah "$dir"
  else
    printf 'No resurrect directory yet: %s\n' "$dir"
  fi
  pause
}

case "$mode" in
  repository-overview)
    onefetch
    pause
    ;;
  extrakto-help)
    exec less "$HOME/.config/tmux/plugins/extrakto/HELP.md"
    ;;
  effective-options)
    printf 'server options\n'
    for opt in escape-time display-time; do
      tmux show-options -s "$opt"
    done
    printf '\nglobal options\n'
    for opt in history-limit status-interval default-terminal status-keys focus-events; do
      tmux show-options -g "$opt"
    done
    printf '\nwindow options\n'
    for opt in aggressive-resize mode-keys; do
      tmux show-options -gw "$opt"
    done
    pause
    ;;
  resurrect-files | continuum-files)
    show_resurrect_files
    ;;
  resurrect-latest)
    dir="$(resurrect_dir)"
    last="$dir/last"
    if [ -L "$last" ] || [ -f "$last" ]; then
      ls -lah "$last"
      printf '\n'
      readlink "$last" 2>/dev/null || true
    else
      printf 'No latest resurrect save found: %s\n' "$last"
    fi
    pause
    ;;
  resurrect-options)
    for opt in \
      @resurrect-save \
      @resurrect-restore \
      @resurrect-capture-pane-contents \
      @resurrect-dir \
      @resurrect-processes \
      @resurrect-strategy-vim \
      @resurrect-strategy-nvim; do
      printf '%s ' "$opt"
      tmux show-options -gqv "$opt"
    done
    pause
    ;;
  continuum-status)
    "$HOME/.config/tmux/plugins/tmux-continuum/scripts/continuum_status.sh"
    pause
    ;;
  continuum-options)
    for opt in \
      @continuum-save-interval \
      @continuum-restore \
      @continuum-boot \
      @continuum-boot-options \
      @continuum-systemd-start-cmd; do
      printf '%s ' "$opt"
      tmux show-options -gqv "$opt"
    done
    tmux show-options -g status
    pause
    ;;
  *)
    printf 'usage: %s {repository-overview|extrakto-help|effective-options|resurrect-files|resurrect-latest|resurrect-options|continuum-status|continuum-options|continuum-files}\n' "$0" >&2
    exit 2
    ;;
esac
