# tmux Cheat Sheet

This sheet documents the most useful keys and commands for the current config.

The status bar uses (with a Nerd Font configured in the outer terminal):

- blue session block with a terminal icon on the left
- window tabs with angled caps on both sides, a yellow active highlight, muted inactive windows, and a zoom marker
- dark background so it stays visually separate from pane content
- top position
- right-side state flags for prefix, copy mode, mouse, and synchronized panes
- active command and working-directory basename
- local date/time, UTC time, and hostname
- the status line must remain enabled because `tmux-continuum` uses it for autosave

Terminal input behavior:

- modified keys are forwarded with tmux extended key reporting enabled in CSI-u format
- this is intended to let apps inside tmux distinguish keys such as `Shift-Enter` from plain `Enter` when the outer terminal supports extended keys

Pane borders:

- every pane has a top border label showing its pane number and active command
- the active pane border is cyan; inactive pane borders are muted
- borders use single-line characters

Assume the tmux prefix is the default:

```text
Ctrl-a
```

## Repo-Specific Bindings

- `prefix + r`: reload `~/.tmux.conf`
- `prefix + m`: toggle mouse mode
- `prefix + Ctrl-e`: open current pane output plus scrollback in `$VISUAL` or `$EDITOR`
- `prefix + Ctrl-s`: open the reusable `Scratch-Terminal` popup session
- `prefix + S`: save tmux state now with `tmux-resurrect`
- `prefix + Ctrl-r`: restore the most recent saved tmux state with `tmux-resurrect`
- `prefix + T`: open the `sesh` session picker popup
- `prefix + prefix`: send literal `Ctrl-a` to the current pane
- `prefix + c`: create a new window in the current pane path
- `prefix + n`: next window
- `prefix + "`: split vertically in the current pane path
- `prefix + %`: split horizontally in the current pane path
- `prefix + -`: split vertically in the current pane path
- `prefix + |`: split horizontally in the current pane path

Inside the `prefix + T` sesh picker:

- `Ctrl-a`: all entries
- `Ctrl-t`: tmux entries
- `Ctrl-g`: config entries
- `Ctrl-x`: zoxide entries
- `Ctrl-f`: find directories with `fd`
- `Ctrl-d`: kill selected tmux session and refresh

## Plugin Bindings

- `Ctrl-p`: open `tmux-palette`
- `prefix + *`: kill the foreground process in the current pane with `SIGKILL` via `tmux-cowboy`
- `prefix + F`: open `tmux-fzf`
- `prefix + Tab`: open Extrakto to fuzzy-find text from pane history
- `prefix + I`: install plugins with TPM
- `prefix + U`: update plugins with TPM
- `prefix + Alt-u`: remove plugins no longer declared in `.tmux.conf`
- `prefix + S`: save sessions, windows, panes, working directories, and captured pane output with `tmux-resurrect`
- `prefix + Ctrl-r`: restore saved tmux state with `tmux-resurrect`

Inside Extrakto:

- type to fuzzy-find text, paths, URLs, or complete lines
- `Tab`: insert the selected item into the current pane
- `Enter`: copy the selected item to the macOS clipboard
- `Ctrl-f`: change the extraction filter
- `Ctrl-g`: change which pane history is searched
- `Ctrl-l`: show Extrakto help

## Reboot Recovery

- Before a planned reboot, run `prefix + S` to save immediately.
- After reboot, start tmux. `tmux-continuum` restores the most recent saved state when the new tmux server starts.
- If automatic restore does not run or you need to restore explicitly, use `prefix + Ctrl-r`.
- Autosave runs every 15 minutes while tmux is running and the status line is enabled.
- Captured pane output is restored as a recovery net, but important command output should still be saved with a log file, `tee`, or another durable artifact.
- Saved state is currently under `~/.local/share/tmux/resurrect`; `last` points at the most recent save. If `~/.tmux/resurrect` exists or `@resurrect-dir` is set, Resurrect can use that path instead.

## tmux-palette

Open with `Ctrl-p`.

Repo-managed additions:

- `Panes`: capture pane to file, copy pane directory, synchronize panes, clear scrollback
- `Git`: open Lazygit or show a Onefetch repository overview in the current pane directory
- `System`: open `btop`, show messages, scratch popup
- `TMUX Plugins`: individual palettes for `tmux-fzf`, Extrakto, `TPM`, `tmux-cowboy`, `tmux-sensible`, `tmux-resurrect`, and `tmux-continuum`

Useful plugin palette actions:

- `Extrakto`: default extraction, focused word/line/path/URL extraction, plugin help
- `tmux-resurrect`: save state, restore state, list saved states from the active save directory, show latest save, show options
- `tmux-continuum`: run continuum save, run continuum restore, show status, show options, list autosave files

## Useful tmux-sensible Bindings

- `prefix + Ctrl-p`: previous window
- `prefix + Ctrl-n`: next window
- `prefix + R`: reload `~/.tmux.conf` using `tmux-sensible`
- `prefix + b`: last window

## Useful Built-In tmux Bindings

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
