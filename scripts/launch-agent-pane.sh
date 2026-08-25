#!/usr/bin/env bash

set -euo pipefail

usage() {
  printf 'usage: %s {antigravity|antigravity-resume|claude|codex|codex-resume|cursor-agent|fx|fx-resume|pi|pi-resume} pane-id\n' "$0" >&2
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
  antigravity-resume)
    executable="agy"
    agent_command="agy --continue"
    ;;
  claude)
    executable="claude"
    agent_command="claude"
    ;;
  codex)
    executable="codex"
    agent_command="codex"
    ;;
  codex-resume)
    executable="codex"
    agent_command="codex resume"
    ;;
  cursor-agent)
    executable="cursor-agent"
    agent_command="cursor-agent"
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
  pi-resume)
    executable="pi"
    agent_command="pi --continue"
    ;;
  *)
    usage
    ;;
esac

command -v "$executable" >/dev/null 2>&1 || fail "$executable is not in PATH"

pane_path="$(tmux display-message -p -t "$pane_id" '#{pane_current_path}')"
[ -n "$pane_path" ] || fail "Could not determine the pane directory"

tmux split-window -h -p 45 -t "$pane_id" -c "$pane_path" "$agent_command"
