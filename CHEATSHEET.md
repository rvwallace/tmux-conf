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
- `prefix + Space`: open the mnemonic Which-Key leader menu (`tmux-omni`)
- `prefix + P` / `prefix + Ctrl-p`: open the fuzzy Command Palette (`tmux-omni --search`)
- `prefix + ?`: open live help for the prefix key table (`tmux-omni --keys`)
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

## Leader Menu & Command Palette (`tmux-omni`)

Triggered via:
- `prefix + Space`: Which-Key mnemonic tree
- `prefix + P` or `prefix + Ctrl-p`: Fuzzy Command Palette
- `prefix + ?`: Live prefix keybindings inspector

Inside the TUI:
- **Which-Key navigation**: press a shortcut key (e.g. `p` panes, `w` windows, `g` git, `a` AI, `t` tools, `P` plugins, `o` options), or use `↑`/`↓`/`Tab` to highlight.
- **Switch to Palette**: `/` or `Ctrl-p` jumps directly into Command Palette fuzzy search.
- **Switch to Leader**: `Esc` / `Backspace` (when input is empty) or `Ctrl-Space` / `Ctrl-l` instantly switches back to Which-Key Leader view.
- **Inspectors & Navigation**: Selecting an internal inspector (e.g. Environment, Options, Buffers, Keys, Messages) opens directly inside `tmux-omni`; pressing `<Backspace>` (on empty search), `<Esc>`, or `q` returns directly to the Leader menu or Palette screen. Status messages (e.g. copied to clipboard) auto-dismiss after 2.5s.
- **Environment Inspector**: Press `<CR>` to edit variable in tmux prompt, `y`/`c` (or `a`) to open the **Action Picker** modal, or use direct single-key shortcuts: `v` (copy value only), `n` (copy variable name), `e` (copy shell export).
- **Navigation & Close**: `Esc` / `Backspace` goes up one group level (or close at root), `q` quits.

### Execution Modifiers & Actions

| Key | Mode | Behavior |
| :--- | :--- | :--- |
| **`Enter`** | **Run** | Executes command in its target container. If it fails, pauses with error prompt and allows `[s]` Debug shell or `[any key]` Close. |
| **`Ctrl-i`** / `Alt-i` / `Tab` | **Insert into Prompt** | Populates `tmux <command>` directly into the active shell prompt without executing, allowing editing or manual review. |
| **`Ctrl-w`** / `Alt-w` / `Ctrl-t` | **New Window** | Opens command in a new window with a persistent shell. |
| **`Ctrl-v`** / `Alt-v` | **Horizontal Split** | Opens command in a side split pane with a persistent shell. |
| **`Ctrl-s`** / `Alt-s` | **Vertical Split** | Opens command in a bottom split pane with a persistent shell. |
| **`y`** / `c` | **Copy to Clipboard** | Copies the command to system clipboard. |

Quick editing and customization:
- `prefix + Space` → `o` → `m`: edit `config.json` in `$EDITOR`
- `prefix + Space` → `o` → `t`: edit `.tmux.conf` in `$EDITOR`
- `prefix + Space` → `t` → `M`: open Tmux Messages log inspector (`tmux-omni --messages`)
- Refer to [`CONFIG_GUIDE.md`](./CONFIG_GUIDE.md) and [`config.schema.json`](./tmux-menu/config.schema.json) for schema rules and target documentation.

## AI Assistant

These palette actions require `aichat` in `PATH` and are powered directly by `tmux-omni ai`. Open `prefix + P` / `prefix + Ctrl-p` and search for `ai` or `aichat` (or `prefix + Space` → `a`). Actions open the Go Bubble Tea interface directly in a 40% right-side pane. Pane-aware actions capture the current path, foreground command, and scrollback lines (default 200). All analysis and generation actions maintain conversation sessions and accept multi-turn follow-up questions, slash commands (`/git`, `/diff`, `/tree`, `/env`), and `/refresh` to update context. Summarize Pane supports switching depth via `1`–`5` (100, 200, 500, 1000, all), `d` (cycle), or `r` (reload). Generate Command and Suggest Fix allow sending generated commands or extracted code blocks directly into the active pane (`s`, `X`, or `Enter` on empty input) via bracketed paste for review before running. Explain Last Copy reads but does not delete the latest tmux paste buffer and rejects empty buffers or buffers larger than 32 KiB.

Inside the AI interface (Vim modal):
- **Navigation**: `Tab` toggles focus between Input (Insert mode) and Transcript (Normal mode); `Esc` exits Insert mode (or closes when empty); `i`/`a` enters Insert mode.
- **Transcript (Normal Mode)**:
  - `j`/`k` or `↑`/`↓`: scroll transcript line by line; `Ctrl-d`/`Ctrl-u`: half-page scroll; `g`/`G`: jump to top/bottom
  - `y`/`c`: copy full response or candidate command to clipboard and tmux buffer
  - `x`: quick cycle-copy code blocks from response (step 1, 2, 3...)
  - `X` / `Ctrl-x`: open interactive Code Block Picker menu overlay to select and copy/send any code block
  - `s`: insert candidate command or selected code block into target pane
  - `m`: open interactive AI Model picker overlay (`claude-3-5-sonnet`, `gpt-4o`, `deepseek-chat`, `ollama`, etc.)
  - `S` / `E`: export session transcript to clean Markdown and open in `$EDITOR` (`nvim`) in a new tmux window
  - `1`–`5` / `d` / `r`: change scrollback depth or reload context
  - `?`: toggle help dialog; `q` / `Esc`: quit
- **Input (Insert Mode)**:
  - `Enter`: submit prompt (or send candidate command if input is empty); `Shift+Enter`: multiline newline
  - `Ctrl-x`: open interactive Code Block Picker menu overlay
  - `↑` / `Ctrl-p`: navigate backwards in persistent prompt history (`~/.local/share/tmux/ai_history`)
  - `↓` / `Ctrl-n`: navigate forwards in persistent prompt history
  - Slash Commands: type `/git` (injects git status & log), `/diff` (injects uncommitted diff), `/tree` (injects directory tree), `/env` (injects environment variables), or `/refresh` (reloads scrollback)

AI Actions:
- `AI: Ask`: ask a question about the current pane
- `AI: Diagnose Error`: diagnose recent pane output
- `AI: Suggest Fix`: suggest one corrective command for the latest visible failure
- `AI: Summarize Pane`: summarize recent output; customize depth with `1`–`5`
- `AI: Generate Command`: create a command from a description
- `AI: Explain`: explain a command or concept
- `AI: Explain Last Copy`: explain the latest tmux buffer content
- `AI: Open Antigravity Agent`: split pane and launch `agy`
- `AI: Open Codex Agent`: split pane and launch `codex`
- `AI: Resume Codex Agent`: split pane and resume the latest `codex` session

## Snaglord (Output Browser)

- `prefix + Ctrl-y`: open the Snaglord popup for the active pane
- `c`: copy command text only
- `y` or `Enter`: copy command output only
- `Y`: copy final prompt line and command output
- `j`/`k`, `Ctrl-n`/`Ctrl-p`, `↑`/`↓`: navigate entries
- `/`: filter entries
- `q` or `Esc`: close Snaglord

The stock Snaglord CLI intentionally does not include preceding multiline Starship decoration in a command block.

## File Picker

Open files under the current pane directory with `prefix + Ctrl-f`. For the
picker menus, use `prefix + P` → `File Picker` or `prefix + Space` → `t`;
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

- `prefix + P` / `prefix + Ctrl-p`: open command palette (`tmux-omni --search`)
- `prefix + Space`: open Which-Key leader menu (`tmux-omni`)
- `prefix + ?`: open live prefix keybindings browser (`tmux-omni --keys`)
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
