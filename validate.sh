#!/usr/bin/env bash

set -euo pipefail

REPO_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TARGET_CONF="${HOME}/.tmux.conf"
THEME_FILE="${REPO_DIR}/themes/tokyo-night.conf"
TARGET_THEME_DIR="${HOME}/.config/tmux/themes"
TARGET_SCRIPT="${HOME}/.config/tmux/scripts/edit-scrollback.sh"
TARGET_COPY_PATH_SCRIPT="${HOME}/.config/tmux/scripts/copy-pane-path.sh"
TARGET_SHOW_CLIENTS_SCRIPT="${HOME}/.config/tmux/scripts/show-clients.sh"
TARGET_AI_SCRIPT="${HOME}/.config/tmux/scripts/ai-assist.sh"
TARGET_AI_PROMPT_SCRIPT="${HOME}/.config/tmux/scripts/ai-prompt.sh"
TARGET_AICHAT_TUI_SCRIPT="${HOME}/.config/tmux/scripts/aichat-tui.py"
TARGET_AGENT_PANE_SCRIPT="${HOME}/.config/tmux/scripts/launch-agent-pane.sh"
TARGET_FILE_PICKER_SCRIPT="${HOME}/.config/tmux/scripts/tmux-file-picker"
TARGET_INSERT_PATHS_SCRIPT="${HOME}/.config/tmux/scripts/tmux-insert-paths.sh"
TARGET_YAZI_PICKER_SCRIPT="${HOME}/.config/tmux/scripts/tmux-yazi-picker.sh"
TARGET_DEFER_TMUX_SCRIPT="${HOME}/.config/tmux/scripts/defer-tmux-command.sh"
TARGET_WHICH_KEY_POPUP_SCRIPT="${HOME}/.config/tmux/scripts/tmux-which-key-popup.sh"
TARGET_TMUX_MENU_SCRIPT="${HOME}/.config/tmux/scripts/tmux-menu.py"
TARGET_TMUX_OMNI_SCRIPT="${HOME}/.config/tmux/scripts/tmux-omni"
TARGET_PREFIX_HELP_SCRIPT="${HOME}/.config/tmux/scripts/tmux-prefix-help.py"
TARGET_TMUX_MENU_DIR="${HOME}/.config/tmux-menu"
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
expect_symlink_target "$TARGET_THEME_DIR" "$REPO_DIR/themes"
expect_contains "grep -F 'source-file' \"$REPO_DIR/.tmux.conf\"" '$HOME/.config/tmux/themes/tokyo-night.conf' "Tokyo Night theme source"
expect_symlink_target "$TARGET_SCRIPT" "$REPO_DIR/scripts/edit-scrollback.sh"
expect_symlink_target "$TARGET_COPY_PATH_SCRIPT" "$REPO_DIR/scripts/copy-pane-path.sh"
expect_symlink_target "$TARGET_SHOW_CLIENTS_SCRIPT" "$REPO_DIR/scripts/show-clients.sh"
expect_symlink_target "$TARGET_AI_SCRIPT" "$REPO_DIR/scripts/ai-assist.sh"
expect_symlink_target "$TARGET_AI_PROMPT_SCRIPT" "$REPO_DIR/scripts/ai-prompt.sh"
expect_symlink_target "$TARGET_AICHAT_TUI_SCRIPT" "$REPO_DIR/scripts/aichat-tui.py"
expect_symlink_target "$TARGET_AGENT_PANE_SCRIPT" "$REPO_DIR/scripts/launch-agent-pane.sh"
expect_symlink_target "$TARGET_FILE_PICKER_SCRIPT" "$REPO_DIR/scripts/tmux-file-picker"
expect_symlink_target "$TARGET_INSERT_PATHS_SCRIPT" "$REPO_DIR/scripts/tmux-insert-paths.sh"
expect_symlink_target "$TARGET_YAZI_PICKER_SCRIPT" "$REPO_DIR/scripts/tmux-yazi-picker.sh"
expect_symlink_target "$TARGET_DEFER_TMUX_SCRIPT" "$REPO_DIR/scripts/defer-tmux-command.sh"
expect_symlink_target "$TARGET_WHICH_KEY_POPUP_SCRIPT" "$REPO_DIR/scripts/tmux-which-key-popup.sh"
expect_symlink_target "$TARGET_TMUX_MENU_SCRIPT" "$REPO_DIR/scripts/tmux-menu.py"
[ -x "$TARGET_TMUX_OMNI_SCRIPT" ] || fail "tmux-omni executable missing at $TARGET_TMUX_OMNI_SCRIPT"
pass "tmux-omni executable present at $TARGET_TMUX_OMNI_SCRIPT"
expect_symlink_target "$TARGET_PREFIX_HELP_SCRIPT" "$REPO_DIR/scripts/tmux-prefix-help.py"
expect_symlink_target "$TARGET_TMUX_MENU_DIR" "$REPO_DIR/tmux-menu"
expect_symlink_target "$TARGET_SNAGLORD_DIR" "$REPO_DIR/tmux-snaglord"
expect_contains "grep -F 'prompt_lines = 3' \"$TARGET_SNAGLORD_DIR/config.toml\"" "prompt_lines = 3" "Snaglord multiline prompt height"
expect_contains "grep -F 'nerd_fonts = true' \"$TARGET_SNAGLORD_DIR/config.toml\"" "nerd_fonts = true" "Snaglord Nerd Font mode"

"$REPO_DIR/scripts/install-deps.sh" --check || fail "runtime dependencies missing"
pass "runtime dependencies present"
expect_contains "grep -F 'brew \"uv\"' \"$REPO_DIR/Brewfile\"" "brew \"uv\"" "uv declared in Brewfile"
expect_contains "grep -F 'brew \"go\"' \"$REPO_DIR/Brewfile\"" "brew \"go\"" "go declared in Brewfile"

[ -f "$TARGET_TMUX_MENU_DIR/config.json" ] || fail "tmux-menu config.json missing"
jq empty "$TARGET_TMUX_MENU_DIR/config.json" || fail "tmux-menu config is invalid JSON"
pass "tmux-menu config is valid JSON"

jq -e '
  def valid_item:
    (type == "object")
    and (.key | type == "string" and length > 0)
    and (.title | type == "string" and length > 0)
    and if has("items") then
      (.items | type == "array" and length > 0)
      and (([.items[].key] | length) == ([.items[].key] | unique | length))
      and all(.items[]; valid_item)
    else
      (.action | type == "string" and length > 0)
    end;
  (.items | type == "array" and length > 0)
  and (([.items[].key] | length) == ([.items[].key] | unique | length))
  and all(.items[]; valid_item)
' "$TARGET_TMUX_MENU_DIR/config.json" >/dev/null || fail "tmux-menu config structure or per-group key uniqueness failed"
pass "tmux-menu config structure and uniqueness"

expect_contains "grep -F '#!/usr/bin/env -S uv run --script' \"$REPO_DIR/scripts/tmux-menu.py\"" "uv run --script" "tmux-menu shebang metadata"
expect_contains "grep -F 'textual' \"$REPO_DIR/scripts/tmux-menu.py\"" "textual" "tmux-menu Textual dependency"
expect_contains "grep -F 'build_guarded_shell_script' \"$REPO_DIR/scripts/tmux-menu.py\"" "build_guarded_shell_script" "tmux-menu error guard"
expect_contains "grep -F 'copy_to_clipboard' \"$REPO_DIR/scripts/tmux-menu.py\"" "copy_to_clipboard" "tmux-menu clipboard copy support"

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
expect_contains "grep -F \"set -g @plugin 'alberti42/tmux-fzf-links'\" \"$REPO_DIR/.tmux.conf\"" "alberti42/tmux-fzf-links" "tmux-fzf-links plugin declaration"
expect_tmux_option "tmux show -g @fzf-links-editor-open-cmd" "@fzf-links-editor-open-cmd \"tmux new-window -n 'nvim' /opt/homebrew/bin/nvim +%line '%file'\""
expect_tmux_option "tmux show -g @resurrect-capture-pane-contents" "@resurrect-capture-pane-contents on"
expect_tmux_option "tmux show -g @continuum-restore" "@continuum-restore on"
expect_tmux_option "tmux show -g @continuum-save-interval" "@continuum-save-interval 15"
expect_tmux_option "tmux show -g status" "status on"
expect_contains "tmux show -g default-command" "default-command" "default-command present"
if tmux show -g default-command | grep -E '&&|\|\|' >/dev/null; then
  fail "default-command contains shell conditionals that can interfere with restore"
fi
pass "default-command restore-compatible"

expect_contains "tmux list-keys -T prefix -N" "Open leader key menu" "leader key menu binding description"
expect_contains "tmux list-keys -T prefix | grep 'bind-key.*Space'" "scripts/tmux-omni" "leader key menu Space binding"
expect_contains "tmux list-keys -T prefix | grep 'bind-key.*C-p'" "scripts/tmux-omni" "command palette C-p prefix binding"
expect_contains "tmux list-keys -T prefix | grep 'bind-key.*P'" "scripts/tmux-omni" "command palette P prefix binding"

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

expect_contains "grep -F '\"title\": \"Synchronize Panes\"' \"$REPO_DIR/tmux-menu/config.json\"" "Synchronize Panes" "synchronize panes menu entry"
expect_contains "grep -F '\"title\": \"Marked Panes\"' \"$REPO_DIR/tmux-menu/config.json\"" "Marked Panes" "marked panes menu group"
expect_contains "grep -F 'swap-pane -s {marked}' \"$REPO_DIR/tmux-menu/config.json\"" "swap-pane" "swap marked pane command"
expect_contains "grep -F 'join-pane -s {marked}' \"$REPO_DIR/tmux-menu/config.json\"" "join-pane" "join marked pane command"
expect_contains "grep -F '\"title\": \"Layouts\"' \"$REPO_DIR/tmux-menu/config.json\"" "Layouts" "layouts menu group"
expect_contains "grep -F 'select-layout' \"$REPO_DIR/tmux-menu/config.json\" | grep -F 'tiled'" "tiled" "tiled layout command"
expect_contains "grep -F '\"title\": \"Clients\"' \"$REPO_DIR/tmux-menu/config.json\"" "Clients" "clients menu group"
expect_contains "grep -F 'tmux-omni --clients' \"$REPO_DIR/tmux-menu/config.json\"" "tmux-omni --clients" "attached clients helper command"
expect_contains "grep -F 'detach-client -a' \"$REPO_DIR/tmux-menu/config.json\"" "detach-client -a" "detach other clients command"
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
expect_contains "tmux list-keys -T prefix | grep 'bind-key.*C-y'" "tmux-snaglord" "Snaglord popup command"
expect_contains "tmux list-keys -T prefix -N" "Open links with fuzzy finder (tmux-fzf-links plugin)" "tmux-fzf-links prefix binding"
expect_contains "grep -F '\"title\": \"Resurrect\"' \"$REPO_DIR/tmux-menu/config.json\"" "Resurrect" "tmux-resurrect menu group"
expect_contains "grep -F '\"title\": \"Continuum\"' \"$REPO_DIR/tmux-menu/config.json\"" "Continuum" "tmux-continuum menu group"
expect_contains "grep -F '\"title\": \"Extrakto\"' \"$REPO_DIR/tmux-menu/config.json\"" "Extrakto" "Extrakto menu group"
expect_contains "grep -F '\"title\": \"Snaglord\"' \"$REPO_DIR/tmux-menu/config.json\"" "Snaglord" "Snaglord menu entry"
expect_contains "grep -F 'tmux-snaglord' \"$REPO_DIR/tmux-menu/config.json\"" "tmux-snaglord" "Snaglord menu command"
expect_contains "grep -F '\"title\": \"File Picker\"' \"$REPO_DIR/tmux-menu/config.json\"" "File Picker" "tmux-file-picker menu group"
expect_contains "grep -F '\"title\": \"Lazygit\"' \"$REPO_DIR/tmux-menu/config.json\"" "Lazygit" "Lazygit menu entry"
expect_contains "grep -F '\"title\": \"System Monitor (btop)\"' \"$REPO_DIR/tmux-menu/config.json\"" "System Monitor (btop)" "system monitor menu entry"
expect_contains "grep -F 'scripts/tmux-file-picker --zoxide --dir-only' \"$REPO_DIR/tmux-menu/config.json\"" "scripts/tmux-file-picker --zoxide --dir-only" "tmux-file-picker recent directory command"
expect_contains "grep -F 'scripts/tmux-yazi-picker.sh' \"$REPO_DIR/tmux-menu/config.json\"" "scripts/tmux-yazi-picker.sh" "Yazi menu command"
expect_contains "grep -F 'while IFS= read -r path ||' \"$REPO_DIR/scripts/tmux-yazi-picker.sh\"" "while IFS= read -r path ||" "Yazi chooser final path without trailing newline"
expect_contains "grep -F 'path=\"/\${path#search://*//}\"' \"$REPO_DIR/scripts/tmux-yazi-picker.sh\"" "path=\"/\${path#search://*//}\"" "Yazi virtual search URL normalization"
expect_contains "grep -F 'tmux-insert-paths.sh' \"$REPO_DIR/scripts/tmux-file-picker\"" "tmux-insert-paths.sh" "shared picker insertion helper"
expect_contains "grep -F 'cursor-agent | agent' \"$REPO_DIR/scripts/tmux-insert-paths.sh\"" "cursor-agent | agent" "Cursor CLI agent detection"
expect_contains "grep -F 'tmux set-buffer -b' \"$REPO_DIR/scripts/tmux-insert-paths.sh\"" "tmux set-buffer -b" "picker temporary tmux buffer"
expect_contains "grep -F 'tmux paste-buffer -d -b' \"$REPO_DIR/scripts/tmux-insert-paths.sh\"" "tmux paste-buffer -d -b" "picker paste with buffer cleanup"
expect_contains "grep -F 'extrakto/scripts/open.sh' \"$REPO_DIR/tmux-menu/config.json\"" "extrakto/scripts/open.sh" "Extrakto extraction command"
expect_contains "grep -F 'copy-pane-path.sh' \"$REPO_DIR/tmux-menu/config.json\"" "copy-pane-path.sh" "copy pane directory menu command"
expect_contains "grep -F 'onefetch' \"$REPO_DIR/tmux-menu/config.json\"" "onefetch" "Onefetch popup menu command"
expect_contains "grep -F '\"title\": \"AI Assistants\"' \"$REPO_DIR/tmux-menu/config.json\"" "AI Assistants" "AI menu group"
expect_contains "grep -F '\"title\": \"AI: Ask\"' \"$REPO_DIR/tmux-menu/config.json\"" "AI: Ask" "AI ask menu entry"
expect_contains "grep -F '\"title\": \"AI: Suggest Fix\"' \"$REPO_DIR/tmux-menu/config.json\"" "AI: Suggest Fix" "AI suggest-fix menu entry"
expect_contains "grep -F '\"title\": \"AI: Summarize Pane\"' \"$REPO_DIR/tmux-menu/config.json\"" "AI: Summarize Pane" "AI summarize-pane menu entry"
expect_contains "grep -F '\"title\": \"AI: Explain Last Copy\"' \"$REPO_DIR/tmux-menu/config.json\"" "AI: Explain Last Copy" "AI explain-last-copy menu entry"
expect_contains "grep -F '\"title\": \"AI: Open Antigravity Agent\"' \"$REPO_DIR/tmux-menu/config.json\"" "AI: Open Antigravity Agent" "Antigravity agent menu entry"
expect_contains "grep -F '\"title\": \"AI: Open Codex Agent\"' \"$REPO_DIR/tmux-menu/config.json\"" "AI: Open Codex Agent" "Codex agent menu entry"
expect_contains "grep -F '\"title\": \"AI: Resume Codex Agent\"' \"$REPO_DIR/tmux-menu/config.json\"" "AI: Resume Codex Agent" "Codex resume menu entry"
expect_contains "grep -F '\"title\": \"AI Agents\"' \"$REPO_DIR/tmux-menu/config.json\"" "AI Agents" "AI agents menu group"
expect_contains "grep -F 'launch-agent-pane.sh antigravity' \"$REPO_DIR/tmux-menu/config.json\"" "launch-agent-pane.sh antigravity" "Antigravity menu command"
expect_contains "grep -F 'launch-agent-pane.sh codex-resume' \"$REPO_DIR/tmux-menu/config.json\"" "launch-agent-pane.sh codex-resume" "Codex resume menu command"
expect_contains "grep -F 'tmux split-window -h -p 45' \"$REPO_DIR/scripts/launch-agent-pane.sh\"" "tmux split-window -h -p 45" "agent right-side pane size"
expect_contains "grep -F 'tmux display-message -p -t' \"$REPO_DIR/scripts/launch-agent-pane.sh\"" "tmux display-message -p -t" "agent pane source directory lookup"
expect_contains "grep -F 'tmux show-buffer' \"$REPO_DIR/scripts/ai-assist.sh\"" "show-buffer" "AI latest-copy buffer read"
expect_contains "grep -F 'exceeds 32 KiB' \"$REPO_DIR/scripts/ai-assist.sh\"" "exceeds 32 KiB" "AI latest-copy size limit"
expect_contains "grep -F 'tmux-resurrect/scripts/save.sh' \"$REPO_DIR/tmux-menu/config.json\"" "tmux-resurrect/scripts/save.sh" "tmux-resurrect save command"
expect_contains "grep -F 'tmux-continuum/scripts/continuum_save.sh' \"$REPO_DIR/tmux-menu/config.json\"" "tmux-continuum/scripts/continuum_save.sh" "tmux-continuum save command"

printf '\nValidation completed successfully.\n'
