#!/usr/bin/env bash

set -euo pipefail

REPO_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TARGET_CONF="${HOME}/.tmux.conf"
TARGET_SCRIPT="${HOME}/.config/tmux/scripts/edit-scrollback.sh"
TARGET_COPY_PATH_SCRIPT="${HOME}/.config/tmux/scripts/copy-pane-path.sh"
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
expect_symlink_target "$TARGET_COPY_PATH_SCRIPT" "$REPO_DIR/scripts/copy-pane-path.sh"
expect_symlink_target "$TARGET_PALETTE_DIR" "$REPO_DIR/tmux-palette"

"$REPO_DIR/scripts/install-deps.sh" --check || fail "runtime dependencies missing"
pass "runtime dependencies present"

[ -f "$TARGET_PALETTE_DIR/commands.json" ] || fail "tmux-palette commands.json missing"
pass "tmux-palette commands.json present"

[ -f "$TARGET_PALETTE_DIR/palettes/plugin-tmux-fzf.json" ] || fail "tmux-fzf palette missing"
pass "tmux-fzf palette present"

[ -f "$TARGET_PALETTE_DIR/palettes/plugin-extrakto.json" ] || fail "Extrakto palette missing"
pass "Extrakto palette present"

[ -f "$TARGET_PALETTE_DIR/palettes/plugin-tmux-resurrect.json" ] || fail "tmux-resurrect palette missing"
pass "tmux-resurrect palette present"

[ -f "$TARGET_PALETTE_DIR/palettes/plugin-tmux-continuum.json" ] || fail "tmux-continuum palette missing"
pass "tmux-continuum palette present"

[ -d "$TPM_DIR/.git" ] || fail "TPM is not installed at $TPM_DIR"
pass "TPM installed at $TPM_DIR"

tmux source-file "$REPO_DIR/.tmux.conf"
pass "tmux source-file succeeded"

expect_tmux_option "tmux show -g prefix" "prefix C-a"
expect_tmux_option "tmux show -g base-index" "base-index 1"
expect_tmux_option "tmux show-window -g pane-base-index" "pane-base-index 1"
expect_tmux_option "tmux show -g status-position" "status-position top"
expect_tmux_option "tmux show -g status-interval" "status-interval 5"
expect_tmux_option "tmux show -g extended-keys" "extended-keys on"
expect_tmux_option "tmux show -g extended-keys-format" "extended-keys-format csi-u"
expect_tmux_option "tmux show-window -g pane-border-status" "pane-border-status top"
expect_tmux_option "tmux show-window -g pane-border-lines" "pane-border-lines single"
expect_contains "tmux show-window -g pane-border-format" "#{pane_current_command}" "pane border active command"

expect_contains "tmux show -g window-status-format" "#[fg=#252a31,bg=#1f2329]" "inactive window left-pointing cap"
expect_contains "tmux show -g window-status-format" "#[fg=#252a31,bg=#1f2329]" "inactive window right-pointing cap"
expect_contains "tmux show -g window-status-current-format" "#[fg=#f4d35e,bg=#1f2329,nobold]" "active window left-pointing cap"
expect_contains "tmux show -g window-status-current-format" "#[fg=#f4d35e,bg=#1f2329,nobold]" "active window right-pointing cap"
expect_contains "tmux show -g status-right" "#{pane_current_command}" "status-right active command"
expect_contains "tmux show -g status-right" "#{b:pane_current_path}" "status-right working directory"
expect_contains "tmux show -g status-right" "%Y-%m-%d %H:%M" "status-right local clock"
expect_contains "tmux show -g status-right" "TZ=UTC date +'%%H:%%M'" "status-right UTC clock"
expect_contains "tmux show -g status-right" "#{?pane_in_mode" "status-right copy-mode indicator"
expect_contains "tmux show -g terminal-features" "xterm-256color:RGB:extkeys" "xterm extended keys terminal feature"
expect_contains "tmux show -g terminal-features" "tmux-256color:RGB:extkeys" "nested tmux extended keys terminal feature"
expect_contains "grep -F \"set -g @plugin 'tmux-plugins/tmux-resurrect'\" \"$REPO_DIR/.tmux.conf\"" "tmux-plugins/tmux-resurrect" "tmux-resurrect plugin declaration"
expect_contains "grep -F \"set -g @plugin 'tmux-plugins/tmux-continuum'\" \"$REPO_DIR/.tmux.conf\"" "tmux-plugins/tmux-continuum" "tmux-continuum plugin declaration"
expect_contains "grep -F \"set -g @plugin 'laktak/extrakto'\" \"$REPO_DIR/.tmux.conf\"" "laktak/extrakto" "extrakto plugin declaration"
expect_tmux_option "tmux show -g @resurrect-capture-pane-contents" "@resurrect-capture-pane-contents on"
expect_tmux_option "tmux show -g @continuum-restore" "@continuum-restore on"
expect_tmux_option "tmux show -g @continuum-save-interval" "@continuum-save-interval 15"
expect_tmux_option "tmux show -g status" "status on"
expect_contains "tmux show -g default-command" "default-command" "default-command present"
if tmux show -g default-command | grep -E '&&|\|\|' >/dev/null; then
  fail "default-command contains shell conditionals that can interfere with restore"
fi
pass "default-command restore-compatible"
expect_contains "tmux list-keys" "bind-key    -T prefix       C-s" "scratch popup binding"
expect_contains "tmux list-keys -T prefix" "extrakto/scripts/open.sh" "extrakto binding"
expect_contains "tmux list-keys -T prefix" "bind-key    -T prefix n       next-window" "next window binding"
expect_contains "tmux list-keys -T prefix" "split-window -v -c \"#{pane_current_path}\"" "dash vertical split binding"
expect_contains "tmux list-keys" "bind-key    -T prefix       |                         split-window -h -c \"#{pane_current_path}\"" "pipe horizontal split binding"
expect_contains "tmux list-keys -T prefix" "tmux-resurrect/scripts/save.sh" "tmux-resurrect save binding"
expect_contains "tmux list-keys -T prefix" "tmux-resurrect/scripts/restore.sh" "tmux-resurrect restore binding"
expect_contains "tmux list-keys" "bind-key    -T prefix       C-e" "scrollback editor binding"
expect_contains "tmux list-keys -T root" "tmux-palette/bin/tmux-palette.sh" "tmux-palette root binding"
expect_contains "grep -F '\"palette\": \"plugin-tmux-resurrect\"' \"$REPO_DIR/tmux-palette/commands.json\"" "plugin-tmux-resurrect" "tmux-resurrect palette root entry"
expect_contains "grep -F '\"palette\": \"plugin-tmux-continuum\"' \"$REPO_DIR/tmux-palette/commands.json\"" "plugin-tmux-continuum" "tmux-continuum palette root entry"
expect_contains "grep -F '\"palette\": \"plugin-extrakto\"' \"$REPO_DIR/tmux-palette/commands.json\"" "plugin-extrakto" "Extrakto palette root entry"
expect_contains "grep -F 'extrakto/scripts/open.sh' \"$REPO_DIR/tmux-palette/palettes/plugin-extrakto.json\"" "extrakto/scripts/open.sh" "Extrakto palette extraction command"
expect_contains "grep -F 'copy-pane-path.sh' \"$REPO_DIR/tmux-palette/commands.json\"" "copy-pane-path.sh" "copy pane directory palette command"
expect_contains "grep -F \"display-popup -E -d '#{pane_current_path}'\" \"$REPO_DIR/tmux-palette/commands.json\"" "lazygit" "lazygit popup palette command"
expect_contains "grep -F 'onefetch' \"$REPO_DIR/tmux-palette/commands.json\"" "onefetch" "Onefetch popup palette command"
expect_contains "grep -F \"display-popup -E -w '90%' -h '85%' btop\" \"$REPO_DIR/tmux-palette/commands.json\"" "btop" "system monitor palette command"
expect_contains "grep -F 'tmux-resurrect/scripts/save.sh' \"$REPO_DIR/tmux-palette/palettes/plugin-tmux-resurrect.json\"" "tmux-resurrect/scripts/save.sh" "tmux-resurrect palette save command"
expect_contains "grep -F 'tmux-continuum/scripts/continuum_save.sh' \"$REPO_DIR/tmux-palette/palettes/plugin-tmux-continuum.json\"" "tmux-continuum/scripts/continuum_save.sh" "tmux-continuum palette save command"

printf '\nValidation completed successfully.\n'
