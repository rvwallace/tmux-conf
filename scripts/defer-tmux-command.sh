#!/usr/bin/env bash

set -euo pipefail

if [ "$#" -eq 0 ]; then
  printf 'usage: %s tmux-command [argument ...]\n' "$0" >&2
  exit 2
fi

# which-key itself runs in a tmux popup. Let that popup close before starting
# commands that may open another popup, prompt, picker, or pane.
(
  sleep 0.15
  tmux "$@"
) >/dev/null 2>&1 &
