#!/usr/bin/env bash

set -euo pipefail

pane_id="${1-}"
if [ -z "${pane_id}" ]; then
  pane_id="$(tmux display-message -p '#{pane_id}')"
fi

pane_path="$(tmux display-message -p -t "${pane_id}" '#{pane_current_path}')"

if command -v pbcopy >/dev/null 2>&1; then
  printf '%s' "${pane_path}" | pbcopy
elif command -v wl-copy >/dev/null 2>&1; then
  printf '%s' "${pane_path}" | wl-copy
elif command -v xclip >/dev/null 2>&1; then
  printf '%s' "${pane_path}" | xclip -selection clipboard
else
  tmux set-buffer -- "${pane_path}"
  tmux display-message "Copied pane directory to tmux buffer (no system clipboard tool found)"
  exit 0
fi

tmux display-message "Copied pane directory: ${pane_path}"
