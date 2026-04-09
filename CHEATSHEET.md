# tmux Cheat Sheet

This sheet documents the most useful keys and commands for the current config.

The status bar uses:

- blue session block on the left
- yellow active window highlight
- muted inactive windows
- dark background so it stays visually separate from pane content
- top position
- right-side state flags for prefix, mouse, and synchronized panes
- both local time and UTC time are shown on the right

Terminal input behavior:

- modified keys are forwarded with tmux extended key reporting enabled
- this is intended to let apps inside tmux distinguish keys such as `Shift-Enter` from plain `Enter` when the outer terminal supports extended keys

Assume the tmux prefix is the default:

```text
Ctrl-a
```

## Repo-Specific Bindings

- `prefix + r`: reload `~/.tmux.conf`
- `prefix + m`: toggle mouse mode
- `prefix + Ctrl-e`: open current pane output plus scrollback in `$VISUAL` or `$EDITOR`
- `prefix + Ctrl-s`: open the reusable `Scratch-Terminal` popup session
- `prefix + prefix`: send literal `Ctrl-a` to the current pane
- `prefix + c`: create a new window in the current pane path
- `prefix + "`: split vertically in the current pane path
- `prefix + %`: split horizontally in the current pane path

## Plugin Bindings

- `prefix + *`: kill the foreground process in the current pane with `SIGKILL` via `tmux-cowboy`
- `prefix + F`: open `tmux-fzf`
- `prefix + I`: install plugins with TPM
- `prefix + U`: update plugins with TPM
- `prefix + Alt-u`: remove plugins no longer declared in `.tmux.conf`

## Useful tmux-sensible Bindings

- `prefix + Ctrl-p`: previous window
- `prefix + Ctrl-n`: next window
- `prefix + R`: reload `~/.tmux.conf` using `tmux-sensible`
- `prefix + b`: last window

## Useful Built-In tmux Bindings

- `prefix + n`: next window
- `prefix + p`: previous window
- `prefix + 1..9`: jump to window by index
- window and pane numbering starts at `1`
- `prefix + ,`: rename current window
- `prefix + $`: rename current session
- `prefix + z`: zoom/unzoom the current pane
- `prefix + x`: kill current pane with confirmation
- `prefix + &`: kill current window with confirmation
- `prefix + [`: enter copy mode
- `prefix + ]`: paste most recent tmux buffer
- `prefix + d`: detach from tmux
- `prefix + ?`: list key bindings
- `prefix + :`: open the tmux command prompt

## Copy Mode

This config uses vi keys in copy mode.

- `prefix + [`: enter copy mode
- `h`, `j`, `k`, `l`: move
- `w`, `b`, `e`: move by word
- `0`, `^`, `$`: move within line
- `/`: search forward
- `?`: search backward
- `Space`: begin selection
- `Enter`: copy selection and leave copy mode
- `q`: quit copy mode

## Common Commands

Run these in a shell:

```sh
tmux source-file ~/.tmux.conf
tmux list-keys
tmux list-sessions
tmux attach
tmux new -s work
```

Run these from the tmux command prompt (`prefix + :`):

```text
new-window -c "#{pane_current_path}"
split-window -h -c "#{pane_current_path}"
split-window -v -c "#{pane_current_path}"
display-message "hello"
```
