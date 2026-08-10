#!/usr/bin/env bash

set -euo pipefail

REPO_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TARGET_CONF="${HOME}/.tmux.conf"
THEME_FILE="${REPO_DIR}/themes/tokyo-night.conf"
TARGET_SCRIPT="${HOME}/.config/tmux/scripts/edit-scrollback.sh"
TARGET_COPY_PATH_SCRIPT="${HOME}/.config/tmux/scripts/copy-pane-path.sh"
TARGET_SHOW_CLIENTS_SCRIPT="${HOME}/.config/tmux/scripts/show-clients.sh"
TARGET_AI_SCRIPT="${HOME}/.config/tmux/scripts/ai-assist.sh"
TARGET_AI_PROMPT_SCRIPT="${HOME}/.config/tmux/scripts/ai-prompt.sh"
TARGET_FILE_PICKER_SCRIPT="${HOME}/.config/tmux/scripts/tmux-file-picker"
TARGET_INSERT_PATHS_SCRIPT="${HOME}/.config/tmux/scripts/tmux-insert-paths.sh"
TARGET_YAZI_PICKER_SCRIPT="${HOME}/.config/tmux/scripts/tmux-yazi-picker.sh"
TARGET_DEFER_TMUX_SCRIPT="${HOME}/.config/tmux/scripts/defer-tmux-command.sh"
TARGET_WHICH_KEY_POPUP_SCRIPT="${HOME}/.config/tmux/scripts/tmux-which-key-popup.sh"
TARGET_PALETTE_DIR="${HOME}/.config/tmux-palette"
TARGET_WHICH_KEY_DIR="${HOME}/.config/tmux-which-key"
TARGET_SNAGLORD_DIR="${HOME}/.config/tmux-snaglord"
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
[ -f "$THEME_FILE" ] || fail "Tokyo Night theme file missing"
pass "Tokyo Night theme file present"
expect_contains "grep -F 'source-file -F \"#{d:current_file}/themes/tokyo-night.conf\"' \"$REPO_DIR/.tmux.conf\"" "themes/tokyo-night.conf" "Tokyo Night theme source"
expect_symlink_target "$TARGET_SCRIPT" "$REPO_DIR/scripts/edit-scrollback.sh"
expect_symlink_target "$TARGET_COPY_PATH_SCRIPT" "$REPO_DIR/scripts/copy-pane-path.sh"
expect_symlink_target "$TARGET_SHOW_CLIENTS_SCRIPT" "$REPO_DIR/scripts/show-clients.sh"
expect_symlink_target "$TARGET_AI_SCRIPT" "$REPO_DIR/scripts/ai-assist.sh"
expect_symlink_target "$TARGET_AI_PROMPT_SCRIPT" "$REPO_DIR/scripts/ai-prompt.sh"
expect_symlink_target "$TARGET_FILE_PICKER_SCRIPT" "$REPO_DIR/scripts/tmux-file-picker"
expect_symlink_target "$TARGET_INSERT_PATHS_SCRIPT" "$REPO_DIR/scripts/tmux-insert-paths.sh"
expect_symlink_target "$TARGET_YAZI_PICKER_SCRIPT" "$REPO_DIR/scripts/tmux-yazi-picker.sh"
expect_symlink_target "$TARGET_DEFER_TMUX_SCRIPT" "$REPO_DIR/scripts/defer-tmux-command.sh"
expect_symlink_target "$TARGET_WHICH_KEY_POPUP_SCRIPT" "$REPO_DIR/scripts/tmux-which-key-popup.sh"
expect_symlink_target "$TARGET_PALETTE_DIR" "$REPO_DIR/tmux-palette"
expect_symlink_target "$TARGET_WHICH_KEY_DIR" "$REPO_DIR/tmux-which-key"
expect_symlink_target "$TARGET_SNAGLORD_DIR" "$REPO_DIR/tmux-snaglord"
expect_contains "grep -F 'prompt_lines = 3' \"$TARGET_SNAGLORD_DIR/config.toml\"" "prompt_lines = 3" "Snaglord multiline prompt height"
expect_contains "grep -F 'nerd_fonts = true' \"$TARGET_SNAGLORD_DIR/config.toml\"" "nerd_fonts = true" "Snaglord Nerd Font mode"

"$REPO_DIR/scripts/install-deps.sh" --check || fail "runtime dependencies missing"
pass "runtime dependencies present"

[ -f "$TARGET_PALETTE_DIR/commands.json" ] || fail "tmux-palette commands.json missing"
pass "tmux-palette commands.json present"

[ -f "$TARGET_PALETTE_DIR/palettes/plugin-tmux-fzf.json" ] || fail "tmux-fzf palette missing"
pass "tmux-fzf palette present"

[ -f "$TARGET_PALETTE_DIR/palettes/plugin-extrakto.json" ] || fail "Extrakto palette missing"
pass "Extrakto palette present"

[ -f "$TARGET_PALETTE_DIR/palettes/tmux-file-picker.json" ] || fail "tmux-file-picker palette missing"
pass "tmux-file-picker palette present"

[ -f "$TARGET_PALETTE_DIR/palettes/lazygit.json" ] || fail "Lazygit palette missing"
pass "Lazygit palette present"

[ -f "$TARGET_PALETTE_DIR/palettes/system-monitor.json" ] || fail "system monitor palette missing"
pass "system monitor palette present"

[ -f "$TARGET_PALETTE_DIR/palettes/marked-panes.json" ] || fail "marked panes palette missing"
pass "marked panes palette present"

[ -f "$TARGET_PALETTE_DIR/palettes/layouts.json" ] || fail "layouts palette missing"
pass "layouts palette present"

[ -f "$TARGET_PALETTE_DIR/palettes/clients.json" ] || fail "clients palette missing"
pass "clients palette present"

[ -f "$TARGET_PALETTE_DIR/palettes/plugin-tmux-resurrect.json" ] || fail "tmux-resurrect palette missing"
pass "tmux-resurrect palette present"

[ -f "$TARGET_PALETTE_DIR/palettes/plugin-tmux-continuum.json" ] || fail "tmux-continuum palette missing"
pass "tmux-continuum palette present"

[ -f "$TARGET_WHICH_KEY_DIR/config.json" ] || fail "tmux which-key config missing"
jq empty "$TARGET_WHICH_KEY_DIR/config.json" || fail "tmux which-key config is invalid JSON"
pass "tmux which-key config is valid JSON"

jq -e '
  def valid:
    (type == "object")
    and (.key | type == "string" and length > 0)
    and (.description | type == "string" and length > 0)
    and (.type | IN("group", "action", "tmux", "script", "popup"))
    and if .type == "group" then
      (.items | type == "array" and length > 0)
      and (([.items[].key] | length) == ([.items[].key] | unique | length))
      and all(.items[]; valid)
    else
      (.command | type == "string" and length > 0)
    end;
  (.items | type == "array" and length > 0)
  and (([.items[].key] | length) == ([.items[].key] | unique | length))
  and all(.items[]; valid)
' "$TARGET_WHICH_KEY_DIR/config.json" >/dev/null || fail "tmux which-key config schema or per-menu key uniqueness failed"
pass "tmux which-key config structure and keys"

if jq -r '.. | objects | select(.type == "popup") | .command' "$TARGET_WHICH_KEY_DIR/config.json" | grep -F "'" >/dev/null; then
  fail "tmux which-key popup commands must not contain single quotes"
fi
pass "tmux which-key popup commands avoid upstream single-quote bug"

palette_titles="$(
  jq -r '.[] | select(.action.palette == null) | .title' "$TARGET_PALETTE_DIR/commands.json"
  for palette_file in "$TARGET_PALETTE_DIR"/palettes/*.json; do
    jq -r '.items[].title' "$palette_file"
  done
)"
which_key_descriptions="$(jq -r '.. | objects | select(.type != "group") | .description' "$TARGET_WHICH_KEY_DIR/config.json")"
while IFS= read -r palette_title; do
  grep -Fx "$palette_title" <<<"$which_key_descriptions" >/dev/null ||
    fail "tmux which-key mirror missing palette action: $palette_title"
done <<<"$palette_titles"
pass "tmux which-key mirrors every custom palette action"

[ -d "$TPM_DIR/.git" ] || fail "TPM is not installed at $TPM_DIR"
pass "TPM installed at $TPM_DIR"

tmux source-file "$REPO_DIR/.tmux.conf"
pass "tmux source-file succeeded"

expect_tmux_option "tmux show -g prefix" "prefix C-a"
expect_tmux_option "tmux show -g base-index" "base-index 1"
expect_tmux_option "tmux show -g detach-on-destroy" "detach-on-destroy off"
expect_tmux_option "tmux show-window -g pane-base-index" "pane-base-index 1"
expect_tmux_option "tmux show -g status-position" "status-position top"
expect_tmux_option "tmux show -g status-interval" "status-interval 5"
expect_tmux_option "tmux show -g message-style" 'message-style "fg=#c0caf5,bg=#1a1b26"'
expect_contains "tmux show -g message-format" "#{message}" "message format text"
expect_contains "tmux show -g message-format" "" "message format rounded right cap"
expect_contains "tmux show -g message-format" "fill=#e0af68" "command prompt background fill"
expect_tmux_option "tmux show -g extended-keys" "extended-keys on"
expect_tmux_option "tmux show -g extended-keys-format" "extended-keys-format csi-u"
expect_tmux_option "tmux show -g allow-passthrough" "allow-passthrough on"
expect_contains "tmux show -g update-environment" " TERM" "TERM client environment refresh"
expect_contains "tmux show -g update-environment" " TERM_PROGRAM" "TERM_PROGRAM client environment refresh"
expect_tmux_option "tmux show-window -g pane-border-status" "pane-border-status top"
expect_tmux_option "tmux show-window -g pane-border-lines" "pane-border-lines single"
expect_contains "tmux show-window -g pane-border-format" "#{?pane_active,#[fg=#7dcfff]#[bold],#[fg=#565f89]}" "pane border active style"
expect_contains "tmux show-window -g pane-border-format" "#T" "pane border title"
expect_contains "tmux show-window -g pane-border-format" "#{pane_current_command}" "pane border active command"
expect_contains "tmux show-window -g pane-border-format" "#{b:pane_current_path}" "pane border working directory"

expect_tmux_option "tmux show -g status-style" 'status-style "fg=#c0caf5,bg=#1a1b26"'
expect_contains "tmux show -g status-left" "#[fg=#1a1b26,bg=#bb9af7,bold]" "purple session segment"
expect_contains "tmux show -g status-left" "#[fg=#bb9af7,bg=#1a1b26,nobold]" "session pointed transition"
expect_contains "tmux show -g window-status-format" "#[fg=#24283b,bg=#1a1b26]" "inactive window rounded left cap"
expect_contains "tmux show -g window-status-format" "#[fg=#24283b,bg=#1a1b26]" "inactive window rounded right cap"
expect_contains "tmux show -g window-status-current-format" "#[fg=#7aa2f7,bg=#1a1b26,nobold]" "active window rounded left cap"
expect_contains "tmux show -g window-status-current-format" "#[fg=#7aa2f7,bg=#1a1b26,nobold]" "active window rounded right cap"
expect_contains "tmux show -g status-right" "%Y-%m-%d %H:%M" "status-right local clock"
expect_contains "tmux show -g status-right" "TZ=UTC date +'%%H:%%M'" "status-right UTC clock"
expect_contains "tmux show -g status-right" "#{?pane_in_mode" "status-right copy-mode indicator"
expect_contains "tmux show -g status-right" "#{?pane_marked_set" "status-right marked-pane indicator"
expect_contains "tmux show -g status-right" "#{?pane_synchronized" "status-right synchronized-pane indicator"
expect_contains "tmux show -g status-right" "#{?client_prefix" "status-right prefix indicator"
expect_contains "tmux show -g status-right" "#[fg=#7dcfff]󰂋" "status-right hostname icon"
expect_contains "tmux show -g status-right" '#{?#{||:#{client_prefix},#{||:#{pane_in_mode},#{||:#{pane_marked_set},#{||:#{pane_synchronized},#{mouse}}}}}' "status-right conditional state separator"
expect_contains "tmux display-message -p -F '#{E:status-left}'" "󰆍" "status-left rendered session block"
expect_contains "tmux show -g terminal-features" "xterm-256color:RGB:extkeys" "xterm extended keys terminal feature"
expect_contains "tmux show -g terminal-features" "tmux-256color:RGB:extkeys" "nested tmux extended keys terminal feature"
expect_contains "grep -F \"set -g @plugin 'tmux-plugins/tmux-resurrect'\" \"$REPO_DIR/.tmux.conf\"" "tmux-plugins/tmux-resurrect" "tmux-resurrect plugin declaration"
expect_contains "grep -F \"set -g @plugin 'tmux-plugins/tmux-continuum'\" \"$REPO_DIR/.tmux.conf\"" "tmux-plugins/tmux-continuum" "tmux-continuum plugin declaration"
expect_contains "grep -F \"set -g @plugin 'laktak/extrakto'\" \"$REPO_DIR/.tmux.conf\"" "laktak/extrakto" "extrakto plugin declaration"
expect_contains "grep -F \"set -g @plugin 'Nucc/tmux-which-key'\" \"$REPO_DIR/.tmux.conf\"" "Nucc/tmux-which-key" "tmux which-key plugin declaration"
expect_contains "grep -F \"set -g @plugin 'alberti42/tmux-fzf-links'\" \"$REPO_DIR/.tmux.conf\"" "alberti42/tmux-fzf-links" "tmux-fzf-links plugin declaration"
expect_tmux_option "tmux show -g @fzf-links-editor-open-cmd" "@fzf-links-editor-open-cmd \"tmux new-window -n 'nvim' /opt/homebrew/bin/nvim +%line '%file'\""
expect_tmux_option "tmux show -g @which-key-config" "@which-key-config $HOME/.config/tmux-which-key/config.json"
expect_tmux_option "tmux show -g @which-key-trigger" "@which-key-trigger Space"
expect_tmux_option "tmux show -g @which-key-popup-bg" '@which-key-popup-bg "#1a1b26"'
expect_tmux_option "tmux show -g @which-key-popup-fg" '@which-key-popup-fg "#565f89"'
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
expect_contains "tmux list-keys -T prefix -N" "Toggle marked pane" "marked pane toggle binding"
expect_contains "tmux list-keys -T prefix | grep 'bind-key.* M '" "refresh-client -S" "marked pane status refresh"
expect_contains "tmux list-keys -T prefix -N" "Show prefix key help" "prefix help binding"
expect_contains "tmux list-keys -T root" "MouseUp3Pane" "pane context menu on mouse release"
expect_contains "tmux list-keys -T root" "MouseUp3Status" "window context menu on mouse release"
expect_contains "tmux list-keys -T root" "MouseUp3StatusLeft" "session context menu on mouse release"
expect_contains "tmux list-keys -T root | grep 'MouseUp3Pane'" "display-menu -O" "persistent pane context menu"
expect_contains "tmux list-keys -T root | grep 'MouseUp3Status '" "display-menu -O" "persistent window context menu"
expect_contains "tmux list-keys -T root | grep 'MouseUp3StatusLeft'" "display-menu -O" "persistent session context menu"
expect_contains "grep -F '\"title\": \"Synchronize Panes\"' \"$REPO_DIR/tmux-palette/commands.json\"" "Synchronize Panes" "synchronize panes palette entry"
expect_contains "grep -F 'synchronize-panes \\\\; refresh-client -S' \"$REPO_DIR/tmux-palette/commands.json\"" "refresh-client -S" "synchronize panes status refresh"
expect_contains "grep -F '\"palette\": \"marked-panes\"' \"$REPO_DIR/tmux-palette/commands.json\"" "marked-panes" "marked panes palette root entry"
expect_contains "grep -F 'swap-pane -s {marked}' \"$REPO_DIR/tmux-palette/palettes/marked-panes.json\"" "swap-pane" "swap marked pane command"
expect_contains "grep -F 'join-pane -s {marked}' \"$REPO_DIR/tmux-palette/palettes/marked-panes.json\"" "join-pane" "join marked pane command"
expect_contains "grep -F '\"palette\": \"layouts\"' \"$REPO_DIR/tmux-palette/commands.json\"" "layouts" "layouts palette root entry"
expect_contains "grep -F 'select-layout tiled' \"$REPO_DIR/tmux-palette/palettes/layouts.json\"" "select-layout tiled" "tiled layout command"
expect_contains "grep -F '\"palette\": \"clients\"' \"$REPO_DIR/tmux-palette/commands.json\"" "clients" "clients palette root entry"
expect_contains "grep -F 'show-clients.sh' \"$REPO_DIR/tmux-palette/palettes/clients.json\"" "show-clients.sh" "attached clients helper command"
expect_contains "grep -F '#{client_name}' \"$REPO_DIR/scripts/show-clients.sh\"" "#{client_name}" "attached clients per-client format"
expect_contains "grep -F 'detach-client -a' \"$REPO_DIR/tmux-palette/palettes/clients.json\"" "detach-client -a" "detach other clients command"
expect_contains "tmux list-keys -T prefix" "extrakto/scripts/open.sh" "extrakto binding"
expect_contains "tmux list-keys -T prefix" "bind-key    -T prefix n       next-window" "next window binding"
expect_contains "tmux list-keys -T prefix" "split-window -v -c \"#{pane_current_path}\"" "dash vertical split binding"
expect_contains "tmux list-keys" "bind-key    -T prefix       |                         split-window -h -c \"#{pane_current_path}\"" "pipe horizontal split binding"
expect_contains "tmux list-keys -T prefix" "tmux-resurrect/scripts/save.sh" "tmux-resurrect save binding"
expect_contains "tmux list-keys -T prefix" "tmux-resurrect/scripts/restore.sh" "tmux-resurrect restore binding"
expect_contains "tmux list-keys" "bind-key    -T prefix       C-e" "scrollback editor binding"
expect_contains "tmux list-keys -T prefix -N" "Insert file paths" "tmux-file-picker popup binding"
expect_contains "tmux list-keys -T prefix" "scripts/tmux-file-picker" "tmux-file-picker popup command"
expect_contains "tmux list-keys -T prefix -N" "Open Snaglord command-output browser" "Snaglord popup binding"
expect_contains "tmux list-keys -T prefix" "display-popup -E -h \"75%\" -w \"70%\" tmux-snaglord" "Snaglord popup command"
expect_contains "tmux list-keys -T root" "tmux-palette/bin/tmux-palette.sh" "tmux-palette root binding"
expect_contains "tmux list-keys -T prefix | grep 'bind-key.*Space'" "tmux-which-key/scripts/which-key.sh" "tmux which-key prefix Space binding"
expect_contains "tmux list-keys -T prefix -N" "Open links with fuzzy finder (tmux-fzf-links plugin)" "tmux-fzf-links prefix binding"
expect_contains "grep -F '\"palette\": \"plugin-tmux-resurrect\"' \"$REPO_DIR/tmux-palette/commands.json\"" "plugin-tmux-resurrect" "tmux-resurrect palette root entry"
expect_contains "grep -F '\"palette\": \"plugin-tmux-continuum\"' \"$REPO_DIR/tmux-palette/commands.json\"" "plugin-tmux-continuum" "tmux-continuum palette root entry"
expect_contains "grep -F '\"palette\": \"plugin-extrakto\"' \"$REPO_DIR/tmux-palette/commands.json\"" "plugin-extrakto" "Extrakto palette root entry"
expect_contains "grep -F '\"title\": \"Snaglord\"' \"$REPO_DIR/tmux-palette/commands.json\"" "Snaglord" "Snaglord palette entry"
expect_contains "grep -F 'tmux-snaglord' \"$REPO_DIR/tmux-palette/commands.json\"" "tmux-snaglord" "Snaglord palette command"
expect_contains "grep -F '\"palette\": \"tmux-file-picker\"' \"$REPO_DIR/tmux-palette/commands.json\"" "tmux-file-picker" "tmux-file-picker palette root entry"
expect_contains "grep -F '\"palette\": \"lazygit\"' \"$REPO_DIR/tmux-palette/commands.json\"" "lazygit" "Lazygit palette root entry"
expect_contains "grep -F '\"palette\": \"system-monitor\"' \"$REPO_DIR/tmux-palette/commands.json\"" "system-monitor" "system monitor palette root entry"
expect_contains "grep -F 'scripts/tmux-file-picker --zoxide --dir-only' \"$REPO_DIR/tmux-palette/palettes/tmux-file-picker.json\"" "scripts/tmux-file-picker --zoxide --dir-only" "tmux-file-picker recent directory command"
expect_contains "grep -F 'scripts/tmux-yazi-picker.sh' \"$REPO_DIR/tmux-palette/palettes/tmux-file-picker.json\"" "scripts/tmux-yazi-picker.sh" "Yazi palette command"
expect_contains "grep -F '\"description\": \"Yazi Picker\"' \"$REPO_DIR/tmux-which-key/config.json\"" "Yazi Picker" "Yazi which-key entry"
expect_contains "grep -F 'while IFS= read -r path ||' \"$REPO_DIR/scripts/tmux-yazi-picker.sh\"" "while IFS= read -r path ||" "Yazi chooser final path without trailing newline"
expect_contains "grep -F 'path=\"/\${path#search://*//}\"' \"$REPO_DIR/scripts/tmux-yazi-picker.sh\"" "path=\"/\${path#search://*//}\"" "Yazi virtual search URL normalization"
expect_contains "grep -F 'tmux-insert-paths.sh' \"$REPO_DIR/scripts/tmux-file-picker\"" "tmux-insert-paths.sh" "shared picker insertion helper"
expect_contains "grep -F 'cursor-agent | agent' \"$REPO_DIR/scripts/tmux-insert-paths.sh\"" "cursor-agent | agent" "Cursor CLI agent detection"
expect_contains "grep -F 'tmux set-buffer -b' \"$REPO_DIR/scripts/tmux-insert-paths.sh\"" "tmux set-buffer -b" "picker temporary tmux buffer"
expect_contains "grep -F 'tmux paste-buffer -d -b' \"$REPO_DIR/scripts/tmux-insert-paths.sh\"" "tmux paste-buffer -d -b" "picker paste with buffer cleanup"
expect_contains "grep -F 'extrakto/scripts/open.sh' \"$REPO_DIR/tmux-palette/palettes/plugin-extrakto.json\"" "extrakto/scripts/open.sh" "Extrakto palette extraction command"
expect_contains "grep -F 'copy-pane-path.sh' \"$REPO_DIR/tmux-palette/commands.json\"" "copy-pane-path.sh" "copy pane directory palette command"
expect_contains "grep -F \"display-popup -E -d '#{pane_current_path}' -w '90%' -h '90%' lazygit\" \"$REPO_DIR/tmux-palette/palettes/lazygit.json\"" "lazygit" "Lazygit popup command"
expect_contains "grep -F \"split-window -h -p 50 -c '#{pane_current_path}' lazygit\" \"$REPO_DIR/tmux-palette/palettes/lazygit.json\"" "split-window" "Lazygit side pane command"
expect_contains "grep -F \"new-window -n lazygit -c '#{pane_current_path}' lazygit\" \"$REPO_DIR/tmux-palette/palettes/lazygit.json\"" "new-window" "Lazygit new window command"
expect_contains "grep -F 'onefetch' \"$REPO_DIR/tmux-palette/commands.json\"" "onefetch" "Onefetch popup palette command"
expect_contains "grep -F \"display-popup -E -w '90%' -h '85%' btop\" \"$REPO_DIR/tmux-palette/palettes/system-monitor.json\"" "btop" "system monitor popup command"
expect_contains "grep -F \"split-window -h -p 50 -c '#{pane_current_path}' btop\" \"$REPO_DIR/tmux-palette/palettes/system-monitor.json\"" "split-window" "system monitor side pane command"
expect_contains "grep -F \"new-window -n btop -c '#{pane_current_path}' btop\" \"$REPO_DIR/tmux-palette/palettes/system-monitor.json\"" "new-window" "system monitor new window command"
expect_contains "grep -F '\"category\": \"AI\"' \"$REPO_DIR/tmux-palette/commands.json\"" "AI" "AI palette entries"
expect_contains "grep -F '\"title\": \"AI:' \"$REPO_DIR/tmux-palette/commands.json\"" "AI:" "AI searchable title prefix"
expect_contains "grep -F 'aichat' \"$REPO_DIR/tmux-palette/commands.json\"" "aichat" "AI searchable descriptions"
expect_contains "grep -F \"display-popup -E -h '25%' -w '70%' -T\" \"$REPO_DIR/scripts/ai-prompt.sh\"" "display-popup" "compact AI command popup"
expect_contains "grep -F \"popup_title=' AI Suggest Fix '\" \"$REPO_DIR/scripts/ai-prompt.sh\"" "AI Suggest Fix" "AI suggest-fix popup title"
expect_contains "grep -F 'tmux capture-pane -J -p -S -200' \"$REPO_DIR/scripts/ai-assist.sh\"" "capture-pane" "bounded AI pane context"
expect_contains "grep -F 'tmux paste-buffer -dp' \"$REPO_DIR/scripts/ai-assist.sh\"" "paste-buffer -dp" "AI command bracketed paste with buffer cleanup"
expect_contains "grep -F 'multiline or fenced output' \"$REPO_DIR/scripts/ai-assist.sh\"" "multiline or fenced output" "unsafe AI command rejection"
expect_contains "grep -F 'TMUX_AI_ASSIST_FOLLOW_UP' \"$REPO_DIR/scripts/ai-prompt.sh\"" "TMUX_AI_ASSIST_FOLLOW_UP" "AI follow-up context control"
expect_contains "grep -F 'TMUX_AI_ASSIST_REFRESH=1' \"$REPO_DIR/scripts/ai-prompt.sh\"" "TMUX_AI_ASSIST_REFRESH=1" "AI conversation context refresh"
expect_contains "grep -F '\"title\": \"AI: Suggest Fix\"' \"$REPO_DIR/tmux-palette/commands.json\"" "AI: Suggest Fix" "AI suggest-fix palette entry"
expect_contains "grep -F '\"title\": \"AI: Summarize Pane\"' \"$REPO_DIR/tmux-palette/commands.json\"" "AI: Summarize Pane" "AI summarize-pane palette entry"
expect_contains "grep -F '\"title\": \"AI: Explain Last Copy\"' \"$REPO_DIR/tmux-palette/commands.json\"" "AI: Explain Last Copy" "AI explain-last-copy palette entry"
expect_contains "grep -F 'tmux show-buffer' \"$REPO_DIR/scripts/ai-assist.sh\"" "show-buffer" "AI latest-copy buffer read"
expect_contains "grep -F 'exceeds 32 KiB' \"$REPO_DIR/scripts/ai-assist.sh\"" "exceeds 32 KiB" "AI latest-copy size limit"
expect_contains "grep -F 'tmux-resurrect/scripts/save.sh' \"$REPO_DIR/tmux-palette/palettes/plugin-tmux-resurrect.json\"" "tmux-resurrect/scripts/save.sh" "tmux-resurrect palette save command"
expect_contains "grep -F 'tmux-continuum/scripts/continuum_save.sh' \"$REPO_DIR/tmux-palette/palettes/plugin-tmux-continuum.json\"" "tmux-continuum/scripts/continuum_save.sh" "tmux-continuum palette save command"

printf '\nValidation completed successfully.\n'
