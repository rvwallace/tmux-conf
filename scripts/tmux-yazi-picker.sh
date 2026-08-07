#!/usr/bin/env bash

set -euo pipefail

if [ -z "${TMUX-}" ]; then
  printf 'Error: This script must be run inside a tmux session.\n' >&2
  exit 1
fi

if ! command -v yazi >/dev/null 2>&1; then
  printf "Error: Required command 'yazi' not found. Please install it.\n" >&2
  exit 1
fi

pane_id="$(tmux display-message -p '#{pane_id}')"
pane_dir="$(tmux display-message -p '#{pane_current_path}')"
pane_pid="$(tmux display-message -p '#{pane_pid}')"
chooser_file="$(mktemp -t tmux-yazi-picker.XXXXXX)"
trap 'rm -f "$chooser_file"' EXIT

(
  cd "$pane_dir"
  yazi . --chooser-file="$chooser_file"
)

selected_paths=()
while IFS= read -r path || [ -n "$path" ]; do
  [ -n "$path" ] || continue
  if [[ "$path" == search://*//* ]]; then
    path="/${path#search://*//}"
  fi
  case "$path" in
    "$pane_dir") selected_paths+=(".") ;;
    "$pane_dir"/*) selected_paths+=("${path#"$pane_dir"/}") ;;
    *) selected_paths+=("$path") ;;
  esac
done < "$chooser_file"

[ "${#selected_paths[@]}" -gt 0 ] || exit 0

"$(dirname "${BASH_SOURCE[0]}")/tmux-insert-paths.sh" \
  "$pane_id" "$pane_pid" "${selected_paths[@]}"
