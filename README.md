# tmux-conf

Minimal tmux configuration managed from this repo.

The entry point in this repo is symlinked to `~/.tmux.conf`. It loads repo-owned supporting configuration, including the active theme.

## Files

- [`./.tmux.conf`](./.tmux.conf): authoritative entry point, organized into core behavior, terminal capabilities, window and pane defaults, bindings, context menus, theme loading, plugins, and post-plugin overrides.
- [`./themes/tokyo-night.conf`](./themes/tokyo-night.conf): Tokyo Night status bar, pane borders, messages, copy mode, and which-key popup surface; exposed to tmux through the `~/.config/tmux/themes` symlink.
- [`./bootstrap.sh`](./bootstrap.sh): refreshes repo-managed symlinks and sets up the TPM path on a machine.
- [`./Brewfile`](./Brewfile): macOS runtime dependencies used by the configuration and palette utilities.
- [`./validate.sh`](./validate.sh): checks symlinks, TPM, tmux reloadability, and expected live settings.
- [`./CHEATSHEET.md`](./CHEATSHEET.md): commonly used key bindings and commands.
- [`./AGENTS.md`](./AGENTS.md): maintenance rules for future edits.
- [`./LICENSE`](./LICENSE): MIT license for using, modifying, and distributing this configuration.
- [`./scripts/edit-scrollback.sh`](./scripts/edit-scrollback.sh): captures pane output and opens it in `$VISUAL` or `$EDITOR`.
- [`./scripts/copy-pane-path.sh`](./scripts/copy-pane-path.sh): safely copies the active pane directory to the system clipboard or tmux buffer.
- [`./scripts/show-clients.sh`](./scripts/show-clients.sh): lists each attached tmux client and waits so the palette popup remains visible.
- [`./scripts/install-deps.sh`](./scripts/install-deps.sh): checks or installs runtime dependencies on macOS and Linux.
- [`./scripts/launch-agent-pane.sh`](./scripts/launch-agent-pane.sh): opens Antigravity or Codex in a right-side pane rooted at the active pane directory.
- [`./scripts/tmux-file-picker`](./scripts/tmux-file-picker): vendored upstream fuzzy file and directory picker that inserts selected paths into the active pane.
- [`./scripts/tmux-file-picker.UPSTREAM.md`](./scripts/tmux-file-picker.UPSTREAM.md): source commit and license attribution for the vendored picker.
- [`./scripts/tmux-yazi-picker.sh`](./scripts/tmux-yazi-picker.sh): launches Yazi as a multi-file chooser for the active pane.
- [`./scripts/tmux-insert-paths.sh`](./scripts/tmux-insert-paths.sh): shared agent-aware formatter and tmux-buffer inserter for both file pickers.
- [`./scripts/defer-tmux-command.sh`](./scripts/defer-tmux-command.sh): closes popups before starting another tmux popup, picker, prompt, or pane.
- [`./scripts/tmux-which-key-popup.sh`](./scripts/tmux-which-key-popup.sh): runs inspection-oriented helper popups without fragile nested shell quoting.
- [`./tmux-omni/`](./tmux-omni/): fast Go + Bubble Tea unified leader menu and command palette with LazyVim/NvChad styling, dynamic target modifiers, and error guard with clipboard copy (binary installed to `~/.config/tmux/scripts/tmux-omni`).
- [`./tmux-menu/config.json`](./tmux-menu/config.json): unified command and group registry, symlinked through `~/.config/tmux-menu`.
- [`./tmux-snaglord/config.toml`](./tmux-snaglord/config.toml): prompt parser config for stock `tmux-snaglord`, symlinked through `~/.config/tmux-snaglord`.

## Current Behavior

This config intentionally stays small:

- prefix is `Ctrl-a`
- `default-terminal` is set to `tmux-256color`
- RGB terminal features are enabled for common terminals
- extended keyboard reporting is enabled for `xterm-256color` and `tmux-256color`, using CSI-u format so modified keys such as `Shift-Enter` can reach apps inside tmux
- terminal escape-sequence passthrough is enabled, and `TERM` plus `TERM_PROGRAM` are refreshed when clients attach, so graphics-aware applications such as Yazi can show image and PDF previews through tmux in terminals with compatible graphics support, including Ghostty and Kitty
- the client environment refresh list is set explicitly so repeated config reloads do not accumulate duplicate `TERM` and `TERM_PROGRAM` entries
- history limit is `100000`
- mouse mode starts enabled
- destroying the attached session switches the client to another session instead of detaching it
- command prompt and copy mode use vi keys
- windows and panes are numbered from `1`
- windows are renumbered after closes
- new windows and splits inherit the current pane path
- each pane has a labeled top border showing its pane number, pane title, active command, and working-directory basename, with the active pane highlighted in Tokyo Night cyan
- the top status bar uses a Tokyo Night Storm palette with a dark base, layered surfaces, and Nerd Font icons
- the session is a left-anchored purple segment with a pointed transition, so it is distinct from the window tabs
- transient tmux messages appear as compact yellow segments with a rounded right edge instead of full-width blocks
- pressing the prefix shows a rounded yellow `PREFIX` indicator in the right-side state group
- window tabs have rounded caps on both sides; the active window is highlighted in blue, inactive windows use a muted raised surface, and zoomed windows are marked
- the right side can show prefix, copy mode, marked-pane, mouse, and synchronized-pane state indicators
- the right side shows icon-labeled local date/time, UTC time, and hostname
- sessions, windows, panes, working directories, and captured pane output can be saved and restored with `tmux-resurrect`
- sessions are autosaved every 15 minutes while tmux is running and the status line is enabled
- saved sessions are restored automatically when a new tmux server starts
- command history can be browsed as searchable command/output blocks in a Snaglord popup
- files and directories can be fuzzy-selected in a popup and inserted at the active pane's cursor

## Key Bindings Added Here

These are the bindings this repo adds or overrides directly:

- `prefix + r`: reload `~/.tmux.conf`
- `prefix + m`: toggle mouse mode
- `prefix + M`: mark or unmark the current pane for later swap/join operations; the `MARK` indicator updates immediately
- `prefix + Ctrl-e`: capture the current pane plus scrollback into `$VISUAL` or `$EDITOR` in a new tmux window
- `prefix + Ctrl-f`: fuzzy-find files under the current pane directory and insert one or more selected paths
- `prefix + Ctrl-y`: open the Snaglord command/output browser for the current pane
- `prefix + Ctrl-s`: open a popup attached to the reusable `Scratch-Terminal` session
- `prefix + S`: save the current tmux state with `tmux-resurrect`
- `prefix + Ctrl-r`: restore the most recent saved tmux state with `tmux-resurrect`
- `prefix + T`: open a `sesh` session picker in an `fzf-tmux` popup and connect to the selected entry
- `prefix + Space`: open the mnemonic which-key leader menu (`tmux-omni`)
- `prefix + P` / `prefix + Ctrl-p`: open the fuzzy command palette (`tmux-omni --search`)
- `prefix + ?`: open live, human-readable help generated from the active prefix key table
- `prefix + prefix`: send literal `Ctrl-a` to the pane
- `prefix + c`: new window in current pane path
- `prefix + n`: next window, explicitly restored for long-lived tmux servers
- `prefix + "`: vertical split in current pane path
- `prefix + %`: horizontal split in current pane path
- `prefix + -`: vertical split in current pane path
- `prefix + |`: horizontal split in current pane path

Inside the `prefix + T` picker:

- `Ctrl-a`: show all `sesh` entries
- `Ctrl-t`: show tmux sessions only
- `Ctrl-g`: show config sessions only
- `Ctrl-x`: show zoxide sessions only
- `Ctrl-f`: run a directory search (`fd`) and show matches
- `Ctrl-d`: kill the selected tmux session, then refresh

This config explicitly clears several bindings before plugins initialize, preventing stale definitions in long-lived servers:

- `prefix + Space`
- `prefix + ?`
- `prefix + F`

Right-click context menus for panes, window tabs, and the session block open on button release with tmux's persistent-menu mode, so mouse movement does not dismiss them. They remain available until an item is chosen by mouse or shortcut, or the menu is explicitly dismissed.

## Plugins

Plugins are managed with TPM from:

- `~/.config/tmux/plugins/tpm`

Current plugins:

- `tmux-plugins/tmux-sensible`
- `tmux-plugins/tmux-cowboy`
- `sainnhe/tmux-fzf`
- `laktak/extrakto`
- `alberti42/tmux-fzf-links`
- `tmux-plugins/tmux-resurrect`
- `tmux-plugins/tmux-continuum`

`tmux-omni` provides both the leader key menu (`prefix + Space`) and the command palette (`prefix + P` / `prefix + Ctrl-p`), replacing external which-key and palette plugins with a single repository-owned Go+Bubble Tea tool with sub-10ms startup.

Snaglord is a standalone CLI launched by tmux, not a TPM plugin. The repo keeps its prompt parser configuration under `./tmux-snaglord/`; it recognizes this Starship prompt's final `➜` or `✖` line and treats the prompt as three terminal lines so preceding decoration does not leak into adjacent command output. Snaglord's stock command view supports `c` for command-only copy, `y` or `Enter` for output-only copy, and `Y` for the final prompt line plus output.

`tmux-file-picker` is also a standalone CLI rather than a TPM plugin. Its upstream script is vendored under `./scripts/` at the commit recorded in `tmux-file-picker.UPSTREAM.md`, with a small local integration patch. It uses `fd` and `fzf` to select files or directories. Yazi is available alongside it from both TMUX Tools menus as a navigable, multi-file chooser; its fd (`s`) and ripgrep (`S`) virtual search results are normalized back to filesystem paths. Both pickers pass their selections through a temporary named tmux buffer and delete that buffer after pasting; selected paths are inserted without being executed. When a pane runs Claude, Gemini, Codex, Cursor `agent`, or `cursor-agent`, paths are prefixed with `@`; otherwise they are shell-escaped.

Notes:

- `tmux-sensible` provides sane defaults and a few standard bindings.
- `tmux-cowboy` provides `prefix + *` to send `SIGKILL` to the foreground process in the current pane.
- `tmux-fzf` provides a fuzzy menu for tmux sessions, windows, panes, buffers, and more.
- `extrakto` provides `prefix + Tab` to fuzzy-find text, paths, URLs, and lines from pane history, then insert or copy a selection. Native copy mode can perform manual selection and search, but extracting structured items from large pane histories is materially more tedious.
- `tmux-omni` reads repo-managed JSON from `~/.config/tmux-menu/config.json` and supports both Which-Key tree navigation and fuzzy Command Palette search.
- `tmux-fzf-links` provides `prefix + Ctrl-h` to fuzzy-find links, paths, and other supported addresses in pane history. Opening a selected file starts Neovim in a new tmux window at the detected line.
- `tmux-resurrect` saves and restores tmux sessions, layouts, working directories, and captured pane output.
- `tmux-continuum` autosaves every 15 minutes and restores the most recent saved state when a new tmux server starts.
- `tmux-continuum` must stay last in `.tmux.conf` because it hooks through `status-right`.

`tmux-continuum` does not start tmux at OS login in this config. Start tmux after a reboot, then let continuum restore automatically or use `prefix + Ctrl-r`.

Saved tmux state lives in the `tmux-resurrect` save directory. This config does not set `@resurrect-dir`, so the plugin uses its default lookup:

- `~/.tmux/resurrect` if that directory already exists
- otherwise `${XDG_DATA_HOME:-~/.local/share}/tmux/resurrect`

On this machine, saves are currently under `~/.local/share/tmux/resurrect`, with `last` pointing at the most recent save file.

## tmux-omni (Leader Key & Command Palette)

The unified leader key menu and command palette is implemented in Go with Bubble Tea and Lipgloss ([`./tmux-omni/`](./tmux-omni/)), providing instantaneous (<10ms) startup and Tokyo Night styling inspired by LazyVim and NvChad.

The tracked configuration at `./tmux-menu/config.json` is symlinked to:

```text
~/.config/tmux-menu/config.json
```

### Modes & Navigation

- **Which-Key Mode (`prefix + Space`)**: Mnemonic tree navigation. Root groups are `p` panes, `w` windows, `s` sessions, `b` buffers, `c` clients, `g` Git, `a` AI, `t` tools, `P` plugins, and `o` options. Press a key to drill into a submenu or execute an action. Press `/` to switch into Command Palette search mode, `Esc` or `Backspace` to navigate up one level, and `q` or `Esc` at the root to close.
- **Command Palette Mode (`prefix + P` or `prefix + Ctrl-p`)**: Fuzzy search across all actions, categories, descriptions, and key shortcuts. As you type, results update in real time. Press `Enter` to run the selected action, `Esc` (clears input or returns to Which-Key Leader mode), `Backspace` on empty input, or `Ctrl-Space` / `Ctrl-l` to switch back to Leader mode.

### Execution Targets & Lifecycle

When selecting any command, the execution behavior can be customized dynamically:

- **`<Enter>`**: Runs the command in its target container (popup, split, or window). If the command succeeds (`exit 0`), the container closes cleanly upon exit (unless `"persist_shell": true` is configured on that item in `config.json`, which keeps an active `$SHELL` open). If the command fails (`exit != 0`), the output remains visible on screen with an interactive error prompt:
  - `[s]`: Drop into an interactive `$SHELL` in that container to debug the issue.
  - `[any key]`: Dismiss the container.
- **`Alt-v` / `Ctrl-v`**: Run selected command in a Horizontal (side) split.
- **`Alt-s` / `Ctrl-s`**: Run selected command in a Vertical (bottom) split.
- **`Alt-w` / `Ctrl-t`**: Run selected command in a New Window.
- **`Alt-i` / `Ctrl-y`**: Insert command text directly into active pane prompt without executing.

### Configuration & Schema Tooling

- **Configuration Guide**: See [`./CONFIG_GUIDE.md`](./CONFIG_GUIDE.md) for full documentation on menu structure, targets, custom scripts, and icon reference.
- **JSON Schema**: [`./tmux-menu/config.schema.json`](./tmux-menu/config.schema.json) provides autocompletion, hover docs, and type validation in Neovim, VSCode, and Cursor.
- **Linter & Validation**: Run `tmux-omni --validate` or `./validate.sh` to check for missing fields, invalid targets, or duplicate keybindings.
- **Quick Edit**: Use `prefix + Space` → `o` → `m` to edit `config.json` directly in `$EDITOR`.

### AI Assistant Subsystem

`tmux-omni` includes a native, interactive AI assistant powered by `aichat` (requires `aichat` in `PATH`). AI actions open in a 40% horizontal split pane with Tokyo Night styling and Vim modal editing:

- **Modal Navigation**: `Tab` toggles between Input (Insert mode) and Transcript (Normal mode); `Esc` exits Insert mode; `i`/`a` enters Insert mode.
- **Normal Mode**:
  - `j`/`k`, `Ctrl-d`/`Ctrl-u`, `g`/`G`: Viewport navigation
  - `y`/`c`: Copy full response or candidate command to clipboard and tmux buffer
  - `x`: Quick cycle-copy code blocks (step 1, 2, 3...)
  - `X` / `Ctrl-x`: Open interactive Code Block Picker modal to choose and copy/send any code block
  - `s`: Insert candidate command or selected code block into target pane via bracketed paste
  - `m`: Model switcher overlay (`claude-3-5-sonnet`, `gpt-4o`, `deepseek-chat`, `ollama`, etc.)
  - `S` / `E`: Export session transcript to clean Markdown and open in `$EDITOR` in a new tmux window
  - `1`–`5` / `d` / `r`: Scrollback depth control (100, 200, 500, 1000, all) and context reload
- **Insert Mode**:
  - `Enter`: Submit prompt (or send candidate command if input empty); `Shift+Enter`: multiline newline
  - `Ctrl-x`: Open Code Block Picker modal
  - `↑`/`↓` (`Ctrl-p`/`Ctrl-n`): Persistent prompt history (`~/.local/share/tmux/ai_history`)
  - Slash Commands: `/git`, `/diff`, `/tree`, `/env`, `/refresh` to inject live context
## Reloading

After pulling repository updates, rerun bootstrap before reloading tmux:

```sh
./bootstrap.sh
tmux source-file ~/.tmux.conf
```

Bootstrap refreshes the repo-managed symlinks used by the configuration. This
step is required for existing installations when an update adds a managed path,
including the `~/.config/tmux/themes` path.

For local configuration edits that do not change setup or external paths, reload
from inside tmux:

```sh
tmux source-file ~/.tmux.conf
```

Or use:

```text
prefix + r
```

Most changes apply after a reload. Changes involving terminal detection or the
client environment, including Yazi image-preview support, require starting a
new tmux server. Save any sessions first, then run:

```sh
tmux kill-server
tmux
```

For Yazi image and PDF previews through tmux, use a terminal with compatible
graphics passthrough. Ghostty and Kitty are verified with their Kitty graphics
adapters. Tested stable and nightly WezTerm builds briefly rendered and then
cleared IIP previews during tmux redraws, so WezTerm is not recommended for
this workflow.

After adding or updating plugins in `.tmux.conf`, manage them with TPM:

- **Install new plugins**: `prefix + I` (capital `I`)
- **Update plugins**: `prefix + U` (capital `U`)
- **Remove / clean uninstalled plugins**: `prefix + Alt-u` (or `~/.config/tmux/plugins/tpm/bin/clean_plugins`)

Before a planned reboot, use `prefix + S` to save immediately instead of waiting for the next autosave interval.

The `aichat` actions in the AI category require the `aichat` CLI in `PATH` and are powered directly by the native Go Bubble Tea TUI (`tmux-omni ai <mode>`). Open the palette with `prefix + P` / `prefix + Ctrl-p` and search for `ai` or `aichat` (or `prefix + Space` → `a`). Every AI title starts with `AI:`, and each `aichat` helper description names `aichat`. Pane-aware actions capture the current path, foreground command, and recent scrollback (200 lines by default). All AI actions (Ask, Diagnose Error, Suggest Fix, Summarize Pane, Generate Command, Explain, Explain Last Copy) open right-side tmux panes with persistent conversation sessions, accepting multi-turn follow-up questions, slash commands (`/git`, `/diff`, `/tree`, `/env`), and `/refresh` to update context, with full Vim modal keyboard navigation (`Tab`, `j`/`k`, `y`/`c`, `x`/`b` code block extraction, `X`/`s` command insertion, `m` model switcher, `S`/`E` transcript export, `q`, `Esc`, `?`). Summarize Pane supports switching depth directly in the TUI (`1`–`5` for 100/200/500/1000/all lines, `d` to cycle, `r` to reload). Output can be copied via `y`/`c` keys to both system clipboard and tmux buffer. Explain Last Copy reads the newest tmux paste buffer without deleting it and rejects empty buffers or buffers larger than 32 KiB. In Generate Command and Suggest Fix, candidate commands or extracted code blocks can be inserted directly into the original pane via bracketed paste (`s`, `X`, or `Enter` on empty input) for review and are never executed automatically. Prompt history is persisted across sessions in `~/.local/share/tmux/ai_history` and navigable via `↑`/`↓` (`Ctrl-p`/`Ctrl-n`).

The AI category also opens Antigravity and Codex as 45%-wide right-side panes in the active pane directory. Antigravity always starts a new session. Codex can start a new session or open its directory-filtered resume picker. These actions require `agy` or `codex` in `PATH`. They do not add dedicated tmux key bindings. Focus an agent pane and use `prefix + !` to move the running pane into its own window.

The status icons require a [Nerd Font](https://www.nerdfonts.com/) in the outer terminal. The status line remains usable if those glyphs are unavailable, but the icons will render as missing-character boxes.

## Bootstrap On A New Machine Or After Updating

Run:

```sh
./bootstrap.sh
```

This script:

- checks required commands and offers to install available dependencies
- uses `brew bundle` with [`./Brewfile`](./Brewfile) on macOS
- uses `pacman`, `apt`, or `dnf` for packages available on Linux and prints official guidance for unresolved tools
- recreates the `~/.tmux.conf` symlink to this repo
- recreates the `~/.config/tmux/themes` symlink to this repo's `themes/` directory
- recreates the `~/.config/tmux/scripts/edit-scrollback.sh` symlink to this repo
- recreates the `~/.config/tmux/scripts/copy-pane-path.sh` symlink to this repo
- recreates the `~/.config/tmux/scripts/show-clients.sh` symlink to this repo
- recreates the `~/.config/tmux/scripts/launch-agent-pane.sh` symlink to this repo
- recreates the `~/.config/tmux/scripts/tmux-file-picker` symlink to this repo
- recreates the `~/.config/tmux/scripts/tmux-insert-paths.sh` symlink to this repo
- recreates the `~/.config/tmux/scripts/tmux-yazi-picker.sh` symlink to this repo
- recreates the `~/.config/tmux/scripts/tmux-which-key-popup.sh` symlink to this repo
- compiles and installs `tmux-omni` to `~/.config/tmux/scripts/tmux-omni`
- recreates the `~/.config/tmux-menu` symlink to this repo
- recreates the `~/.config/tmux-snaglord` symlink to this repo
- ensures `~/.config/tmux/plugins/` exists
- clones TPM into `~/.config/tmux/plugins/tpm` if needed

Dependency checks can also be run directly:

```sh
./scripts/install-deps.sh --check
./scripts/install-deps.sh --install
```

The managed command set is `tmux`, `tmux-snaglord`, Git, Bash, `jq`, `fzf`/`fzf-tmux`, `fd`, Sesh, Zoxide, `tree`, `uv`, Yazi, Bun, Python 3, Neovim, Lazygit, `btop`, and Onefetch. Homebrew installs Snaglord from its official tap; Linux setup prints official installation guidance for unresolved tools. On Linux, clipboard integration optionally uses `wl-copy` or `xclip`; without either, copy-path actions fall back to the tmux buffer. Nerd Font installation remains manual because font deployment is desktop-environment specific.

## Validation

Run:

```sh
./validate.sh
```

It checks:

- `~/.tmux.conf` symlink target
- `~/.config/tmux/themes` symlink target and Tokyo Night theme source
- `~/.config/tmux/scripts/edit-scrollback.sh` symlink target
- `~/.config/tmux/scripts/copy-pane-path.sh` symlink target
- `~/.config/tmux/scripts/show-clients.sh` symlink target
- `~/.config/tmux/scripts/tmux-file-picker` symlink target
- `~/.config/tmux/scripts/tmux-insert-paths.sh` and `tmux-yazi-picker.sh` symlink targets
- `~/.config/tmux-menu` symlink target
- TPM installation path
- `tmux source-file ~/.tmux.conf`
- expected live tmux options, terminal features, plugin declarations, and key bindings
- runtime command availability through `scripts/install-deps.sh`

## Portability

The config is intended to work on both macOS and Linux.

- plugin and helper-script paths are installed under `~/.config/tmux/`
- the scrollback helper avoids GNU-only `mktemp` usage
- the scrollback helper uses `$VISUAL`, then `$EDITOR`, then `nvim`

## Installing Or Updating Plugins

TPM bindings:

- `prefix + I`: install plugins
- `prefix + U`: update plugins
- `prefix + Alt-u`: clean removed plugins

## Maintenance

When `.tmux.conf` changes, update this README and [`./CHEATSHEET.md`](./CHEATSHEET.md) in the same change so they stay aligned with actual behavior.
