#!/usr/bin/env bash

set -euo pipefail

REPO_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TARGET_CONF="${HOME}/.tmux.conf"
TARGET_SCRIPT="${HOME}/.config/tmux/scripts/edit-scrollback.sh"
TARGET_PALETTE_DIR="${HOME}/.config/tmux-palette"
TPM_DIR="${HOME}/.config/tmux/plugins/tpm"

fail() {
  printf 'FAIL: %s\n' "$*" >&2
  exit 1
}

pass() {
  printf 'PASS: %s\n' "$*"
}

expect_symlink_target() {
  local path="$1"
  local expected="$2"
  local actual

  [ -L "$path" ] || fail "$path is not a symlink"
  actual="$(readlink "$path")"
  [ "$actual" = "$expected" ] || fail "$path points to $actual, expected $expected"
  pass "$path -> $expected"
}

expect_tmux_option() {
  local cmd="$1"
  local expected="$2"
  local actual

  actual="$(eval "$cmd")"
  [ "$actual" = "$expected" ] || fail "expected '$expected' from [$cmd], got '$actual'"
  pass "$expected"
}

expect_contains() {
  local cmd="$1"
  local pattern="$2"
  local label="$3"

  eval "$cmd" | grep -F "$pattern" >/dev/null || fail "$label missing pattern: $pattern"
  pass "$label"
}

expect_symlink_target "$TARGET_CONF" "$REPO_DIR/.tmux.conf"
expect_symlink_target "$TARGET_SCRIPT" "$REPO_DIR/scripts/edit-scrollback.sh"
expect_symlink_target "$TARGET_PALETTE_DIR" "$REPO_DIR/tmux-palette"

[ -f "$TARGET_PALETTE_DIR/commands.json" ] || fail "tmux-palette commands.json missing"
pass "tmux-palette commands.json present"

[ -f "$TARGET_PALETTE_DIR/palettes/plugin-tmux-fzf.json" ] || fail "tmux-fzf palette missing"
pass "tmux-fzf palette present"

[ -d "$TPM_DIR/.git" ] || fail "TPM is not installed at $TPM_DIR"
pass "TPM installed at $TPM_DIR"

tmux source-file "$REPO_DIR/.tmux.conf"
pass "tmux source-file succeeded"

expect_tmux_option "tmux show -g prefix" "prefix C-a"
expect_tmux_option "tmux show -g base-index" "base-index 1"
expect_tmux_option "tmux show-window -g pane-base-index" "pane-base-index 1"
expect_tmux_option "tmux show -g status-position" "status-position top"
expect_tmux_option "tmux show -g extended-keys" "extended-keys on"

expect_contains "tmux show -g status-right" "local #(LC_ALL=C date +'%%Y-%%m-%%d %%H:%%M')" "status-right local clock"
expect_contains "tmux show -g status-right" "UTC #(LC_ALL=C TZ=UTC date +'%%Y-%%m-%%d %%H:%%M')" "status-right UTC clock"
expect_contains "tmux show -g terminal-features" "xterm-256color:RGB:extkeys" "xterm extended keys terminal feature"
expect_contains "tmux show -g terminal-features" "tmux-256color:RGB:extkeys" "nested tmux extended keys terminal feature"
expect_contains "tmux list-keys" "bind-key    -T prefix       C-s" "scratch popup binding"
expect_contains "tmux list-keys" "bind-key    -T prefix       C-e" "scrollback editor binding"
expect_contains "tmux list-keys -T root C-p" "tmux-palette/bin/tmux-palette.sh" "tmux-palette root binding"

printf '\nValidation completed successfully.\n'
