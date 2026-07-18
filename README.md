# tmux-conf

Minimal tmux configuration managed from this repo.

The file in this repo is symlinked to `~/.tmux.conf`, so editing [`./.tmux.conf`](./.tmux.conf) changes the live config source.

## Files

- [`./.tmux.conf`](./.tmux.conf): source of truth for tmux configuration.
- [`./bootstrap.sh`](./bootstrap.sh): sets up the `~/.tmux.conf` symlink and TPM path on a machine.
- [`./validate.sh`](./validate.sh): checks symlinks, TPM, tmux reloadability, and expected live settings.
- [`./CHEATSHEET.md`](./CHEATSHEET.md): commonly used key bindings and commands.
- [`./AGENTS.md`](./AGENTS.md): maintenance rules for future edits.
- [`./scripts/edit-scrollback.sh`](./scripts/edit-scrollback.sh): captures pane output and opens it in `$VISUAL` or `$EDITOR`.
- [`./scripts/copy-pane-path.sh`](./scripts/copy-pane-path.sh): safely copies the active pane directory to the system clipboard or tmux buffer.
- [`./tmux-palette/`](./tmux-palette): user config for `tmux-palette`, symlinked to `~/.config/tmux-palette`.

## Current Behavior

This config intentionally stays small:

- prefix is `Ctrl-a`
- `default-terminal` is set to `tmux-256color`
- RGB terminal features are enabled for common terminals
- extended keyboard reporting is enabled for `xterm-256color` and `tmux-256color`, using CSI-u format so modified keys such as `Shift-Enter` can reach apps inside tmux
- history limit is `100000`
- mouse mode starts enabled
- command prompt and copy mode use vi keys
- windows and panes are numbered from `1`
- windows are renumbered after closes
- new windows and splits inherit the current pane path
- each pane has a labeled top border showing its pane number and active command, with the active pane highlighted in cyan
- the status bar sits at the top with a distinct dark background and Nerd Font icons
- window tabs have angled caps on both sides; the active window is highlighted in yellow, inactive windows are muted, and zoomed windows are marked
- the right side can show `PREFIX`, copy mode, mouse, and synchronized-pane state indicators
- the right side shows the active command, working-directory basename, local date/time, UTC time, and hostname
- sessions, windows, panes, working directories, and captured pane output can be saved and restored with `tmux-resurrect`
- sessions are autosaved every 15 minutes while tmux is running and the status line is enabled
- saved sessions are restored automatically when a new tmux server starts

## Key Bindings Added Here

These are the bindings this repo adds or overrides directly:

- `prefix + r`: reload `~/.tmux.conf`
- `prefix + m`: toggle mouse mode
- `prefix + Ctrl-e`: capture the current pane plus scrollback into `$VISUAL` or `$EDITOR` in a new tmux window
- `prefix + Ctrl-s`: open a popup attached to the reusable `Scratch-Terminal` session
- `prefix + S`: save the current tmux state with `tmux-resurrect`
- `prefix + Ctrl-r`: restore the most recent saved tmux state with `tmux-resurrect`
- `prefix + T`: open a `sesh` session picker in an `fzf-tmux` popup and connect to the selected entry
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

## Plugins

Plugins are managed with TPM from:

- `~/.config/tmux/plugins/tpm`

Current plugins:

- `tmux-plugins/tmux-sensible`
- `tmux-plugins/tmux-cowboy`
- `sainnhe/tmux-fzf`
- `laktak/extrakto`
- `eduwass/tmux-palette`
- `tmux-plugins/tmux-resurrect`
- `tmux-plugins/tmux-continuum`

Notes:

- `tmux-sensible` provides sane defaults and a few standard bindings.
- `tmux-cowboy` provides `prefix + *` to send `SIGKILL` to the foreground process in the current pane.
- `tmux-fzf` provides a fuzzy menu for tmux sessions, windows, panes, buffers, and more.
- `extrakto` provides `prefix + Tab` to fuzzy-find text, paths, URLs, and lines from pane history, then insert or copy a selection. Native copy mode can perform manual selection and search, but extracting structured items from large pane histories is materially more tedious.
- `tmux-palette` provides the `Ctrl-p` command palette and reads repo-managed JSON from `~/.config/tmux-palette`.
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

The main palette includes a `TMUX Plugins` category with individual palettes for `tmux-fzf`, Extrakto, TPM, `tmux-cowboy`, `tmux-sensible`, `tmux-resurrect`, and `tmux-continuum`. It also adds tmux utility commands such as capture pane to file, copy pane directory, synchronize panes, clear scrollback, Lazygit, Onefetch, and `btop` popups, show messages, and the scratch popup.

The Git and system-monitor palette actions require `lazygit`, `onefetch`, and `btop` to be installed and available on `PATH`. Lazygit and Onefetch run in the current pane directory. The copy-directory action uses `pbcopy`, `wl-copy`, or `xclip` when available, falling back to the tmux buffer.

The Extrakto palette can start default, word, line, path, or URL extraction and open the plugin help. The `tmux-resurrect` palette can save and restore tmux state, list saved snapshots from the active save directory, show the latest save target, and inspect resurrect options. The `tmux-continuum` palette can trigger save/restore scripts, show continuum status, inspect autosave settings, and list autosave files. It intentionally does not include boot-start commands.

## Reloading

From inside tmux:

```sh
tmux source-file ~/.tmux.conf
```

Or use:

```text
prefix + r
```

After adding or updating plugins, install them with:

```text
prefix + I
```

Before a planned reboot, use `prefix + S` to save immediately instead of waiting for the next autosave interval.

The status icons require a [Nerd Font](https://www.nerdfonts.com/) in the outer terminal. The status line remains usable if those glyphs are unavailable, but the icons will render as missing-character boxes.

## Bootstrap On A New Machine

Run:

```sh
./bootstrap.sh
```

This script:

- recreates the `~/.tmux.conf` symlink to this repo
- recreates the `~/.config/tmux/scripts/edit-scrollback.sh` symlink to this repo
- recreates the `~/.config/tmux/scripts/copy-pane-path.sh` symlink to this repo
- recreates the `~/.config/tmux-palette` symlink to this repo
- ensures `~/.config/tmux/plugins/` exists
- clones TPM into `~/.config/tmux/plugins/tpm` if needed

## Validation

Run:

```sh
./validate.sh
```

It checks:

- `~/.tmux.conf` symlink target
- `~/.config/tmux/scripts/edit-scrollback.sh` symlink target
- `~/.config/tmux/scripts/copy-pane-path.sh` symlink target
- `~/.config/tmux-palette` symlink target
- TPM installation path
- `tmux source-file ~/.tmux.conf`
- expected live tmux options, terminal features, plugin declarations, and key bindings

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
