#!/usr/bin/env bash

set -euo pipefail

pane_id="${1-}"
if [ -z "${pane_id}" ]; then
  pane_id="$(tmux display-message -p '#{pane_id}')"
fi

pane_path="$(tmux display-message -p -t "${pane_id}" '#{pane_current_path}')"
tmp_dir="${TMPDIR:-/tmp}"
tmp_file="$(mktemp "${tmp_dir%/}/tmux-scrollback.XXXXXX")"

# Capture the entire pane history as plain text and join wrapped lines.
tmux capture-pane -J -p -t "${pane_id}" -S - > "${tmp_file}"

# Trim trailing blank lines so the editor opens on meaningful content.
awk '
  { lines[NR] = $0 }
  $0 ~ /[^[:space:]]/ { last = NR }
  END {
    for (i = 1; i <= last; i++) print lines[i]
  }
' "${tmp_file}" > "${tmp_file}.trimmed"
mv "${tmp_file}.trimmed" "${tmp_file}"

tmux new-window -c "${pane_path}" -n "scrollback" \
  "sh -lc '\"\${VISUAL:-\${EDITOR:-nvim}}\" \"${tmp_file}\"; rm -f \"${tmp_file}\"'"
