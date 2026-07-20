#!/usr/bin/env bash

set -euo pipefail

tmux list-clients -F '#{client_name}  session=#{session_name}  size=#{client_width}x#{client_height}  terminal=#{client_termname}  readonly=#{client_readonly}'
printf '\nPress Enter to close\n'
read -r _
