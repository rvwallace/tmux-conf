#!/usr/bin/env bash

set -euo pipefail

if [ "$#" -lt 3 ]; then
  printf 'usage: %s PANE_ID PANE_PID PATH...\n' "$0" >&2
  exit 2
fi

pane_id="$1"
pane_pid="$2"
shift 2

is_agent_process() {
  local pid="$1"
  local command_name args

  command_name="$(ps -p "$pid" -o comm= 2>/dev/null | awk '{ print $1 }')"
  command_name="${command_name##*/}"
  args="$(ps -p "$pid" -o args= 2>/dev/null || true)"

  case "$command_name" in
    claude | codex | cursor-agent | agent | fx | pi) return 0 ;;
    node | bun)
      [[ "$args" =~ (^|[[:space:]/])(gemini|pi)([[:space:]]|$) ]] && return 0
      ;;
  esac

  return 1
}

pane_runs_agent() {
  local -a pending=("$pane_pid")
  local current child

  while [ "${#pending[@]}" -gt 0 ]; do
    current="${pending[0]}"
    pending=("${pending[@]:1}")

    if [ "$current" != "$pane_pid" ] && is_agent_process "$current"; then
      return 0
    fi

    while IFS= read -r child; do
      [ -n "$child" ] && pending+=("$child")
    done < <(pgrep -P "$current" 2>/dev/null || true)
  done

  return 1
}

output=""
if pane_runs_agent; then
  printf -v output '@%s ' "$@"
else
  for path in "$@"; do
    printf -v escaped_path '%q' "$path"
    output+="${escaped_path} "
  done
fi

buffer_name="tmux-path-picker-${pane_id#%}-$$"
tmux set-buffer -b "$buffer_name" -- "$output"
tmux paste-buffer -d -b "$buffer_name" -t "$pane_id"
