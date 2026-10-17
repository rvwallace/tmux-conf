#!/usr/bin/env bash
set -euo pipefail
unset TMUX TMUX_PANE
if tmux has-session -t dev >/dev/null 2>&1; then
  exit 0
fi
tmux new-session -d -s dev -c /workspace
