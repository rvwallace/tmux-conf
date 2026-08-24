#!/usr/bin/env bash

set -euo pipefail

usage() {
  printf 'usage: %s {antigravity|codex|codex-resume|fx|fx-resume|pi|pi-continue|pi-resume} pane-id\n' "$0" >&2
  exit 2
}

fail() {
  tmux display-message "$1"
  exit 1
}

[ "$#" -eq 2 ] || usage

agent="$1"
pane_id="$2"

case "$agent" in
  antigravity)
    executable="agy"
    agent_command="agy"
    ;;
  codex)
    executable="codex"
    agent_command="codex"
    ;;
  codex-resume)
    executable="codex"
    agent_command="codex resume"
    ;;
  fx)
    executable="fx"
    agent_command="fx"
    ;;
  fx-resume)
    executable="fx"
    agent_command="fx session resume last"
    ;;
  pi)
    executable="pi"
    agent_command="pi"
    ;;
  pi-continue)
    executable="pi"
    agent_command="pi --continue"
    ;;
  pi-resume)
    executable="pi"
    agent_command="pi --resume"
    ;;
  *)
    usage
    ;;
esac

command -v "$executable" >/dev/null 2>&1 || fail "$executable is not in PATH"

pane_path="$(tmux display-message -p -t "$pane_id" '#{pane_current_path}')"
[ -n "$pane_path" ] || fail "Could not determine the pane directory"

tmux split-window -h -p 45 -t "$pane_id" -c "$pane_path" "$agent_command"
