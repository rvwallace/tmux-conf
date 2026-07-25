# tmux-conf

Minimal tmux configuration managed from this repo.

The file in this repo is symlinked to `~/.tmux.conf`, so editing [`./.tmux.conf`](./.tmux.conf) changes the live config source.

## Files

- [`./.tmux.conf`](./.tmux.conf): source of truth for tmux configuration.
- [`./bootstrap.sh`](./bootstrap.sh): sets up the `~/.tmux.conf` symlink and TPM path on a machine.
- [`./Brewfile`](./Brewfile): macOS runtime dependencies used by the configuration and palette utilities.
- [`./validate.sh`](./validate.sh): checks symlinks, TPM, tmux reloadability, and expected live settings.
- [`./CHEATSHEET.md`](./CHEATSHEET.md): commonly used key bindings and commands.
- [`./AGENTS.md`](./AGENTS.md): maintenance rules for future edits.
- [`./LICENSE`](./LICENSE): MIT license for using, modifying, and distributing this configuration.
- [`./scripts/edit-scrollback.sh`](./scripts/edit-scrollback.sh): captures pane output and opens it in `$VISUAL` or `$EDITOR`.
- [`./scripts/copy-pane-path.sh`](./scripts/copy-pane-path.sh): safely copies the active pane directory to the system clipboard or tmux buffer.
- [`./scripts/show-clients.sh`](./scripts/show-clients.sh): lists each attached tmux client and waits so the palette popup remains visible.
- [`./scripts/install-deps.sh`](./scripts/install-deps.sh): checks or installs runtime dependencies on macOS and Linux.
- [`./scripts/ai-assist.sh`](./scripts/ai-assist.sh): pane-aware `aichat` assistant for shell Q&A, recent-error diagnosis, command suggestions, and explanations.
- [`./scripts/ai-prompt.sh`](./scripts/ai-prompt.sh): tmux prompt launcher used by AI key bindings and palette actions.
- [`./scripts/tmux-file-picker`](./scripts/tmux-file-picker): vendored upstream fuzzy file and directory picker that inserts selected paths into the active pane.
- [`./scripts/tmux-file-picker.UPSTREAM.md`](./scripts/tmux-file-picker.UPSTREAM.md): source commit and license attribution for the vendored picker.
- [`./scripts/defer-tmux-command.sh`](./scripts/defer-tmux-command.sh): closes the which-key popup before starting another tmux popup, picker, prompt, or pane.
- [`./scripts/tmux-which-key-popup.sh`](./scripts/tmux-which-key-popup.sh): runs inspection-oriented which-key popups without fragile nested shell quoting.
- [`./tmux-palette/`](./tmux-palette): user config for `tmux-palette`, symlinked to `~/.config/tmux-palette`.
- [`./tmux-which-key/config.json`](./tmux-which-key/config.json): mnemonic which-key hierarchy, symlinked through `~/.config/tmux-which-key`.
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
- each pane has a labeled top border showing its pane number, pane title, active command, and working-directory basename, with the active pane highlighted in cyan
- the status bar sits at the top with a distinct dark background and Nerd Font icons
- transient tmux messages appear as compact yellow segments with a pointed right edge instead of full-width blocks
- pressing the prefix shows a yellow `PREFIX` indicator in the right-side state group
- window tabs have angled caps on both sides; the active window is highlighted in yellow, inactive windows are muted, and zoomed windows are marked
- the right side can show prefix, copy mode, marked-pane, mouse, and synchronized-pane state indicators
- the right side shows local date/time, UTC time, and hostname
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
- `prefix + Space`: open the mnemonic which-key menu
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
- `eduwass/tmux-palette`
- `Nucc/tmux-which-key`
- `tmux-plugins/tmux-resurrect`
- `tmux-plugins/tmux-continuum`

Snaglord is a standalone CLI launched by tmux, not a TPM plugin. The repo keeps its prompt parser configuration under `./tmux-snaglord/`; it recognizes this Starship prompt's final `➜` or `✖` line and treats the prompt as three terminal lines so preceding decoration does not leak into adjacent command output. Snaglord's stock command view supports `c` for command-only copy, `y` or `Enter` for output-only copy, and `Y` for the final prompt line plus output.

`tmux-file-picker` is also a standalone CLI rather than a TPM plugin. Its upstream script is vendored unchanged under `./scripts/` at the commit recorded in `tmux-file-picker.UPSTREAM.md`. It uses `fd` and `fzf` to select files or directories and inserts the paths without executing them. The direct binding searches the current pane directory; its palette also provides current-directory and Zoxide-ranked recent-directory workflows. When the foreground process looks like Claude, Gemini, or Codex, upstream prefixes inserted paths with `@`; otherwise it shell-escapes them.

Notes:

- `tmux-sensible` provides sane defaults and a few standard bindings.
- `tmux-cowboy` provides `prefix + *` to send `SIGKILL` to the foreground process in the current pane.
- `tmux-fzf` provides a fuzzy menu for tmux sessions, windows, panes, buffers, and more.
- `extrakto` provides `prefix + Tab` to fuzzy-find text, paths, URLs, and lines from pane history, then insert or copy a selection. Native copy mode can perform manual selection and search, but extracting structured items from large pane histories is materially more tedious.
- `tmux-palette` provides the `Ctrl-p` command palette and reads repo-managed JSON from `~/.config/tmux-palette`.
- `tmux-which-key` provides the `prefix + Space` mnemonic menu and reads repo-managed JSON from `~/.config/tmux-which-key`.
- `tmux-resurrect` saves and restores tmux sessions, layouts, working directories, and captured pane output.
- `tmux-continuum` autosaves every 15 minutes and restores the most recent saved state when a new tmux server starts.
- `tmux-continuum` must stay last in `.tmux.conf` because it hooks through `status-right`.

`tmux-continuum` does not start tmux at OS login in this config. Start tmux after a reboot, then let continuum restore automatically or use `prefix + Ctrl-r`.

Saved tmux state lives in the `tmux-resurrect` save directory. This config does not set `@resurrect-dir`, so the plugin uses its default lookup:

- `~/.tmux/resurrect` if that directory already exists
- otherwise `${XDG_DATA_HOME:-~/.local/share}/tmux/resurrect`

On this machine, saves are currently under `~/.local/share/tmux/resurrect`, with `last` pointing at the most recent save file.

## tmux-palette

The palette config is tracked in this repo at:

```text
./tmux-palette/
```

Bootstrap links it to:

```text
~/.config/tmux-palette
```

The main palette includes an `AI` category for pane-aware `aichat` helpers, a `TMUX Tools` category for Snaglord and the file picker, and a `TMUX Plugins` category with individual palettes for `tmux-fzf`, Extrakto, TPM, `tmux-cowboy`, `tmux-sensible`, `tmux-resurrect`, and `tmux-continuum`. It also adds tmux utility commands such as capture pane to file, copy pane directory, synchronize panes with an immediate `SYNC` status refresh, marked-pane operations, window layouts and rotation, attached-client management, clear scrollback, Lazygit, Onefetch, and `btop` popups, show messages, and the scratch popup.

The marked-pane palette can toggle or clear a mark, swap the current pane with the marked pane, or join the marked pane into the current window. The layouts palette exposes tiled, even, and main-pane layouts plus clockwise/counterclockwise rotation. The clients palette uses the repo-managed `show-clients.sh` helper to inspect each attached client, or can detach every client except the one invoking the action after confirmation.

The Git and system-monitor palette actions require `lazygit`, `onefetch`, and `btop` to be installed and available on `PATH`. Lazygit and Onefetch run in the current pane directory. The copy-directory action uses `pbcopy`, `wl-copy`, or `xclip` when available, falling back to the tmux buffer.

The File Picker palette can insert files or directories from the current pane directory, or first select one or more recent directories ranked by Zoxide using visit frequency and recency. `Tab` marks multiple entries and `Enter` inserts them into the pane without running the resulting input. File previews use `bat` when available and fall back to `cat`; directory previews use the managed `tree` dependency.

The Extrakto palette can start default, word, line, path, or URL extraction and open the plugin help. The `tmux-resurrect` palette can save and restore tmux state, list saved snapshots from the active save directory, show the latest save target, and inspect resurrect options. The `tmux-continuum` palette can trigger save/restore scripts, show continuum status, inspect autosave settings, and list autosave files. It intentionally does not include boot-start commands.

## tmux which-key

Open which-key with `prefix + Space`. It is a mnemonic companion to the fuzzy
`Ctrl-p` palette: use which-key to learn and traverse short key paths, and use
the palette when searching by command name.

The tracked config at `./tmux-which-key/config.json` is symlinked to:

```text
~/.config/tmux-which-key
```

Root groups are `p` panes, `w` windows, `s` sessions, `b` buffers, `c` clients,
`g` Git, `a` AI, `t` tools, `P` plugins, and `o` options. `r` reloads the
configuration, `:` opens the tmux command prompt, and `?` shows live prefix-key
help. Press a displayed key to enter a group or run an action. Escape or
Backspace returns one level; Escape at the root closes the popup.

The custom hierarchy mirrors every repo-managed `tmux-palette` action, including
marked panes, layouts, file-picker modes, AI helpers, and plugin maintenance and
inspection commands. The palette remains installed at `Ctrl-p` so the two
workflows can be evaluated side by side. Direct which-key tool popups use the
plugin's standard popup size rather than the per-action palette sizes.

Which-key requires `jq`; it is managed alongside the other dependencies in the
`Brewfile` and `scripts/install-deps.sh`.

## Reloading

From inside tmux:

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

After adding or updating plugins, install them with:

```text
prefix + I
```

Before a planned reboot, use `prefix + S` to save immediately instead of waiting for the next autosave interval.

The AI palette actions require the `aichat` CLI in `PATH`. Open the palette with `Ctrl-p` and search for `ai` or `aichat`; every AI title starts with `AI:` and every description includes `aichat`. Pane-aware actions capture the current path, foreground command, and the most recent 200 lines of scrollback. Ask opens a right-side follow-up loop where blank Enter closes the pane; only its first turn sends pane context, and entering `/refresh` sends the latest context. Diagnose, Summarize Pane, Explain, and Explain Last Copy open right-side tmux panes so copy-mode works. Explain Last Copy reads the newest tmux paste buffer without deleting it and rejects empty buffers or buffers larger than 32 KiB. Generate Command and Suggest Fix use compact, titled popups and insert a single suggestion into the original pane with bracketed paste for review. Empty, multiline, or fenced command output is rejected and commands are never executed automatically.

The status icons require a [Nerd Font](https://www.nerdfonts.com/) in the outer terminal. The status line remains usable if those glyphs are unavailable, but the icons will render as missing-character boxes.

## Bootstrap On A New Machine

Run:

```sh
./bootstrap.sh
```

This script:

- checks required commands and offers to install available dependencies
- uses `brew bundle` with [`./Brewfile`](./Brewfile) on macOS
- uses `pacman`, `apt`, or `dnf` for packages available on Linux and prints official guidance for unresolved tools
- recreates the `~/.tmux.conf` symlink to this repo
- recreates the `~/.config/tmux/scripts/edit-scrollback.sh` symlink to this repo
- recreates the `~/.config/tmux/scripts/copy-pane-path.sh` symlink to this repo
- recreates the `~/.config/tmux/scripts/show-clients.sh` symlink to this repo
- recreates the `~/.config/tmux/scripts/ai-assist.sh` symlink to this repo
- recreates the `~/.config/tmux/scripts/ai-prompt.sh` symlink to this repo
- recreates the `~/.config/tmux/scripts/tmux-file-picker` symlink to this repo
- recreates the `~/.config/tmux/scripts/defer-tmux-command.sh` symlink to this repo
- recreates the `~/.config/tmux/scripts/tmux-which-key-popup.sh` symlink to this repo
- recreates the `~/.config/tmux-palette` symlink to this repo
- recreates the `~/.config/tmux-which-key` symlink to this repo
- recreates the `~/.config/tmux-snaglord` symlink to this repo
- ensures `~/.config/tmux/plugins/` exists
- clones TPM into `~/.config/tmux/plugins/tpm` if needed

Dependency checks can also be run directly:

```sh
./scripts/install-deps.sh --check
./scripts/install-deps.sh --install
```

The managed command set is `tmux`, `tmux-snaglord`, Git, Bash, `jq`, `fzf`/`fzf-tmux`, `fd`, Sesh, Zoxide, `tree`, Bun, Python 3, Neovim, Lazygit, `btop`, and Onefetch. Homebrew installs Snaglord from its official tap; Linux setup prints the official Cargo install command when it is unavailable. On Linux, clipboard integration optionally uses `wl-copy` or `xclip`; without either, copy-path actions fall back to the tmux buffer. Nerd Font installation remains manual because font deployment is desktop-environment specific.

## Validation

Run:

```sh
./validate.sh
```

It checks:

- `~/.tmux.conf` symlink target
- `~/.config/tmux/scripts/edit-scrollback.sh` symlink target
- `~/.config/tmux/scripts/copy-pane-path.sh` symlink target
- `~/.config/tmux/scripts/show-clients.sh` symlink target
- `~/.config/tmux/scripts/tmux-file-picker` symlink target
- `~/.config/tmux-palette` symlink target
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
