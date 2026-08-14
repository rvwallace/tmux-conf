# tmux Cheat Sheet

This sheet documents the most useful keys and commands for the current config.

The status bar uses (with a Nerd Font configured in the outer terminal):

The visual settings are maintained in `themes/tokyo-night.conf` and loaded by
the main `.tmux.conf` entry point.

- left-anchored purple session segment with a terminal icon and pointed transition
- rounded window tabs with a blue active highlight, muted raised inactive tabs, and a zoom marker
- Tokyo Night Storm background and surfaces so the bar stays visually separate from pane content
- transient messages use a compact yellow segment with a rounded right edge
- top position
- right-side state flags for prefix, copy mode, marked panes, mouse, and synchronized panes
- icon-labeled local date/time, UTC time, and hostname
- the status line must remain enabled because `tmux-continuum` uses it for autosave

Terminal input behavior:

- modified keys are forwarded with tmux extended key reporting enabled in CSI-u format
- this is intended to let apps inside tmux distinguish keys such as `Shift-Enter` from plain `Enter` when the outer terminal supports extended keys

Pane borders:

- every pane has a top border label showing its pane number, pane title, active command, and working-directory basename
- the active pane border is Tokyo Night cyan; inactive pane borders are muted
- borders use single-line characters

Assume the tmux prefix is the default:

```text
Ctrl-a
```

## Repo-Specific Bindings

- `prefix + r`: reload `~/.tmux.conf`
- `prefix + m`: toggle mouse mode
- `prefix + M`: mark or unmark the current pane; `MARK` refreshes immediately
- `prefix + Ctrl-e`: open current pane output plus scrollback in `$VISUAL` or `$EDITOR`
- `prefix + Ctrl-f`: fuzzy-find files under the current pane directory and insert selected paths
- `prefix + Ctrl-y`: open the Snaglord command/output browser
- `prefix + Ctrl-s`: open the reusable `Scratch-Terminal` popup session
- `prefix + S`: save tmux state now with `tmux-resurrect`
- `prefix + Ctrl-r`: restore the most recent saved tmux state with `tmux-resurrect`
- `prefix + T`: open the `sesh` session picker popup
- `prefix + Space`: open the mnemonic Which-Key leader menu (`tmux-menu.py`)
- `Ctrl-p` / `prefix + P`: open the fuzzy Command Palette (`tmux-menu.py --search`)
- `prefix + ?`: open live help for the prefix key table
- `prefix + prefix`: send literal `Ctrl-a` to the current pane
- `prefix + c`: create a new window in the current pane path
- `prefix + n`: next window
- `prefix + "`: split vertically in the current pane path
- `prefix + %`: split horizontally in the current pane path
- `prefix + -`: split vertically in the current pane path
- `prefix + |`: split horizontally in the current pane path

Mouse context menus:

- right-click a pane, window tab, or session block and release to open its menu
- moving the mouse does not dismiss an open context menu
- choose an item by clicking it or pressing its displayed shortcut
- click away or press Escape to dismiss the menu

Inside the `prefix + T` sesh picker:

- `Ctrl-a`: all entries
- `Ctrl-t`: tmux entries
- `Ctrl-g`: config entries
- `Ctrl-x`: zoxide entries
- `Ctrl-f`: find directories with `fd`
- `Ctrl-d`: kill selected tmux session and refresh

## Leader Menu & Command Palette (`tmux-menu`)

Triggered via:
- `prefix + Space`: Which-Key mnemonic tree
- `Ctrl-p` or `prefix + P`: Fuzzy Command Palette

Inside the TUI:
- **Which-Key navigation**: press a shortcut key (e.g. `p` panes, `w` windows, `g` git, `a` AI, `t` tools, `P` plugins, `o` options).
- `/`: jump directly into Command Palette fuzzy search.
- `Esc` / `Backspace`: go up one group level (or close at root).
- `q`: quit.

### Execution Modifiers & Error Recovery

| Key | Mode | Behavior |
| :--- | :--- | :--- |
| **`Enter`** | **Run** | Executes command in its target container. Closes on exit if successful (`exit 0`), or keeps shell open if `"persist_shell": true` is configured in `config.json`. **If it fails, pauses with error prompt and allows `[s]` Debug shell or `[any key]` Close.** |
| **`Alt-v`** / `Ctrl-v` | **Horizontal Split** | Opens command in a side split pane. |
| **`Alt-s`** / `Ctrl-s` | **Vertical Split** | Opens command in a bottom split pane. |
| **`Alt-w`** / `Ctrl-t` | **New Window** | Opens command in a new window. |
| **`Alt-i`** / `Ctrl-y` | **Insert in Shell** | Inserts command text into active pane prompt without executing. |

## AI Assistant

These palette actions require `aichat` and `uv` in `PATH`. Open `Ctrl-p` and search for `ai` or `aichat` (or `prefix + Space` → `a`). Actions open a Textual interface inside a right-side pane or compact popup. Pane-aware actions capture the current path, foreground command, and scrollback lines (default 200). All analysis actions (Ask, Diagnose Error, Summarize Pane, Explain, Explain Last Copy) maintain conversation sessions and accept multi-turn follow-up questions and `/refresh` to update context. Summarize Pane also supports switching depth via `1`–`5` (100, 200, 500, 1000, all), `d` (cycle), or `r` (reload). Generate Command and Suggest Fix display the candidate command in a preview card where you can send it (`Enter` on empty input or `s`), refine it by typing adjustments, copy it (`y`/`c`), or cancel (`Esc`). Commands are inserted via bracketed paste and never run automatically; empty, multiline, or fenced output is rejected. Explain Last Copy reads but does not delete the latest tmux paste buffer and rejects empty buffers or buffers larger than 32 KiB.

Inside the Textual interface: `Enter` sends prompts (or `Shift+Enter` for multiline), `Tab` toggles focus between input and transcript, `j`/`k` or `↑`/`↓` scrolls transcript, `y`/`c` or mouse drag selection copies text to clipboard and tmux buffer, `Ctrl+L` clears transcript display, `q` quits when input is not active, `Esc` closes, and `?` toggles help.

- `AI: Ask`: ask a question about the current pane
- `AI: Diagnose Error`: diagnose recent pane output
- `AI: Suggest Fix`: suggest one corrective command for the latest visible failure
- `AI: Summarize Pane`: summarize recent pane output and its current state
- `AI: Generate Command`: suggest one shell command into the prompt
- `AI: Explain`: explain a command or snippet
- `AI: Explain Last Copy`: explain the newest tmux copy buffer

Agent actions open a 45%-wide right-side pane in the current pane directory. They require the named CLI in `PATH` and do not have dedicated tmux key bindings.

- `AI: Open Antigravity Agent`: start a new `agy` session
- `AI: Open Codex Agent`: start a new `codex` session
- `AI: Resume Codex Agent`: open the directory-filtered `codex resume` picker
- `prefix + !`: move the focused agent pane into its own window without restarting it

## Snaglord Command Blocks

Open the current pane's searchable command/output history with `prefix + Ctrl-y`.

- `j` / `k`: move between commands
- `/`: search command history
- `c`: copy only the command text, without prompt artifacts
- `y` or `Enter`: copy only command output
- `Y`: copy the final prompt line and command output
- `Space`: add or remove a command from the multi-selection scratchpad
- `Tab`: switch between commands, paths, and JSON views
- `q`: close Snaglord

The stock Snaglord CLI intentionally does not include preceding multiline Starship decoration in a command block.

## File Picker

Open files under the current pane directory with `prefix + Ctrl-f`. For the
picker menus, use `Ctrl-p` → `File Picker` or `prefix + Space` → `t`;
choose `y` in which-key for navigable Yazi browsing or `f`, `d`, `r`, and `R`
for the fzf file, directory, and Zoxide-ranked recent-directory workflows.

- fzf: type to filter; `Tab` toggles selections; `Enter` accepts
- Yazi: navigate normally; `Space` toggles selections; `Enter` accepts
- Yazi `s` (fd) and `S` (ripgrep) search results are converted back to filesystem paths
- accepted paths are pasted through a temporary tmux buffer at the active pane's cursor without executing them
- `Esc`: cancel without inserting anything
- inside Claude, Gemini, Codex, Cursor `agent`, or `cursor-agent`, selected paths are prefixed with `@`
- in other foreground programs, selected paths are shell-escaped

## Plugin Bindings

- `Ctrl-p` / `prefix + P`: open command palette (`tmux-menu.py --search`)
- `prefix + Space`: open Which-Key leader menu (`tmux-menu.py`)
- `prefix + ?`: open live prefix keybindings browser (`tmux-prefix-help.py`)
- `prefix + *`: kill the foreground process in the current pane with `SIGKILL` via `tmux-cowboy`
- `prefix + Ctrl-h`: fuzzy-find links, paths, and supported addresses in pane history with `tmux-fzf-links`
- `prefix + Tab`: open Extrakto to fuzzy-find text from pane history
- `prefix + I`: install plugins with TPM
- `prefix + U`: update plugins with TPM
- `prefix + Alt-u`: remove / uninstall plugins no longer declared in `.tmux.conf` with TPM
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
# After pulling repository updates, refresh managed paths before reloading.
./bootstrap.sh
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
