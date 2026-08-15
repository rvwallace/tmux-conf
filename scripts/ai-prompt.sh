#!/usr/bin/env bash

set -euo pipefail

mode="${1-}"
pane_id="${2-}"
phase="${3-}"

usage() {
  cat <<'EOF'
Usage:
  ai-prompt.sh ask <pane_id>
  ai-prompt.sh error <pane_id>
  ai-prompt.sh fix <pane_id>
  ai-prompt.sh summarize <pane_id>
  ai-prompt.sh command <pane_id>
  ai-prompt.sh explain <pane_id>
  ai-prompt.sh explain-copy <pane_id>
EOF
}

case "$mode" in
  ask|error|fix|summarize|command|explain|explain-copy) ;;
  -h|--help|help) usage; exit 0 ;;
  *) usage >&2; exit 2 ;;
esac

[ -n "$pane_id" ] || { printf 'tmux ai-prompt: missing pane_id\n' >&2; exit 1; }

tmux display-message -p -t "$pane_id" '#{pane_id}' >/dev/null 2>&1 || {
  printf 'tmux ai-prompt: pane not found: %s\n' "$pane_id" >&2
  exit 1
}

if [ "$phase" != "run" ]; then
  case "$mode" in
    command|fix)
      popup_title=' AI Command '
      [ "$mode" = "fix" ] && popup_title=' AI Suggest Fix '
      tmux display-popup -E -h '65%' -w '70%' -t "$pane_id" -T "$popup_title" "~/.config/tmux/scripts/ai-prompt.sh '$mode' '$pane_id' run"
      ;;
    ask|error|summarize|explain|explain-copy)
      pane_path="$(tmux display-message -p -t "$pane_id" '#{pane_current_path}')"
      tmux split-window -h -p 40 -t "$pane_id" -c "$pane_path" "~/.config/tmux/scripts/ai-prompt.sh '$mode' '$pane_id' run"
      ;;
  esac
  exit 0
fi

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
if [ -x "${script_dir}/aichat-tui.py" ]; then
  exec "${script_dir}/aichat-tui.py" "$mode" "$pane_id"
elif [ -x "${HOME}/.config/tmux/scripts/aichat-tui.py" ]; then
  exec "${HOME}/.config/tmux/scripts/aichat-tui.py" "$mode" "$pane_id"
else
  exec aichat-tui.py "$mode" "$pane_id"
fi
