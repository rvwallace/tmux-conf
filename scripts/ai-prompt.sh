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
  ai-prompt.sh command <pane_id>
  ai-prompt.sh explain <pane_id>
EOF
}

case "$mode" in
  ask|error|command|explain) ;;
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
    command)
      tmux display-popup -E -h '60%' -w '80%' "~/.config/tmux/scripts/ai-prompt.sh '$mode' '$pane_id' run"
      ;;
    ask|error|explain)
      pane_path="$(tmux display-message -p -t "$pane_id" '#{pane_current_path}')"
      tmux split-window -h -p 40 -t "$pane_id" -c "$pane_path" "~/.config/tmux/scripts/ai-prompt.sh '$mode' '$pane_id' run"
      ;;
  esac
  exit 0
fi

prompt_label() {
  case "$mode" in
    ask) printf 'Ask AI' ;;
    command) printf 'Command' ;;
    explain) printf 'Explain' ;;
  esac
}

case "$mode" in
  error)
    exec ~/.config/tmux/scripts/ai-assist.sh error "$pane_id"
    ;;
  ask)
    session="tmux-ask-${pane_id#%}-$$"
    prompt_prefix="$(prompt_label)"
    while true; do
      printf '%s: ' "$prompt_prefix"
      IFS= read -r prompt_text || exit 0
      [ -n "$prompt_text" ] || exit 0
      printf '\nGenerating...\n\n'
      TMUX_AI_ASSIST_NO_PAUSE=1 TMUX_AI_ASSIST_SESSION="$session" \
        ~/.config/tmux/scripts/ai-assist.sh ask "$pane_id" "$prompt_text"
      printf '\n'
      prompt_prefix='Follow-up (blank closes)'
    done
    ;;
  command|explain)
    label="$(prompt_label)"
    printf '%s: ' "$label"
    IFS= read -r prompt_text || exit 0
    [ -n "$prompt_text" ] || exit 0
    printf '\nGenerating...\n\n'
    exec ~/.config/tmux/scripts/ai-assist.sh "$mode" "$pane_id" "$prompt_text"
    ;;
esac
