#!/usr/bin/env bash

set -euo pipefail

pane_id="${1-}"
if [ -z "${pane_id}" ]; then
  pane_id="$(tmux display-message -p '#{pane_id}')"
fi

pane_path="$(tmux display-message -p -t "${pane_id}" '#{pane_current_path}')"

# Run the clipboard helper off the caller's PTY. setsid -w starts a new
# session and still waits for the helper to accept the data. A forked
# wl-copy owner then cannot keep a tmux popup open. macOS has no setsid;
# pbcopy does not leave a daemon on the popup PTY, but stdout/stderr are
# still discarded. Candidate order is unchanged: pbcopy, clip.exe, wl-copy.
copy_to_clipboard() {
  if command -v setsid >/dev/null 2>&1; then
    printf '%s' "${pane_path}" | setsid -w "$@" >/dev/null 2>&1
  else
    printf '%s' "${pane_path}" | "$@" >/dev/null 2>&1
  fi
}

if command -v pbcopy >/dev/null 2>&1; then
  copy_to_clipboard pbcopy
elif command -v clip.exe >/dev/null 2>&1; then
  copy_to_clipboard clip.exe
elif command -v wl-copy >/dev/null 2>&1; then
  copy_to_clipboard wl-copy
elif command -v xclip >/dev/null 2>&1; then
  copy_to_clipboard xclip -selection clipboard
elif command -v xsel >/dev/null 2>&1; then
  copy_to_clipboard xsel --clipboard --input
else
  tmux set-buffer -- "${pane_path}"
  tmux display-message "Copied pane directory to tmux buffer (no system clipboard tool found)"
  exit 0
fi

tmux display-message "Copied pane directory: ${pane_path}"
