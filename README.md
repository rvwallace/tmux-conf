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
- [`./tmux-palette/`](./tmux-palette): user config for `tmux-palette`, symlinked to `~/.config/tmux-palette`.

## Current Behavior

This config intentionally stays small:

- prefix is `Ctrl-a`
- `default-terminal` is set to `tmux-256color`
- RGB terminal features are enabled for common terminals
- extended keyboard reporting is enabled for `xterm-256color` and `tmux-256color` so modified keys such as `Shift-Enter` can reach apps inside tmux
- history limit is `100000`
- mouse mode starts enabled
- command prompt and copy mode use vi keys
- windows and panes are numbered from `1`
- windows are renumbered after closes
- new windows and splits inherit the current pane path
- the status bar sits at the top with a distinct dark background
- the active window is highlighted in yellow and inactive windows are muted
- the right side can show `PREFIX`, `mouse`, and `sync` state indicators when active
- the right side shows both local time and UTC time for log-reading/reference

## Key Bindings Added Here

These are the bindings this repo adds or overrides directly:

- `prefix + r`: reload `~/.tmux.conf`
- `prefix + m`: toggle mouse mode
- `prefix + Ctrl-e`: capture the current pane plus scrollback into `$VISUAL` or `$EDITOR` in a new tmux window
- `prefix + Ctrl-s`: open a popup attached to the reusable `Scratch-Terminal` session
- `prefix + T`: open a `sesh` session picker in an `fzf-tmux` popup and connect to the selected entry
- `prefix + prefix`: send literal `Ctrl-a` to the pane
- `prefix + c`: new window in current pane path
- `prefix + "`: vertical split in current pane path
- `prefix + %`: horizontal split in current pane path

Inside the `prefix + T` picker:

- `Ctrl-a`: show all `sesh` entries
- `Ctrl-t`: show tmux sessions only
- `Ctrl-g`: show config sessions only
- `Ctrl-x`: show zoxide sessions only
- `Ctrl-f`: run a directory search (`fd`) and show matches
- `Ctrl-d`: kill the selected tmux session, then refresh

This config also explicitly removes stale bindings from previously used plugins:

- `prefix + Tab`
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
- `eduwass/tmux-palette`

Notes:

- `tmux-sensible` provides sane defaults and a few standard bindings.
- `tmux-cowboy` provides `prefix + *` to send `SIGKILL` to the foreground process in the current pane.
- `tmux-fzf` provides a fuzzy menu for tmux sessions, windows, panes, buffers, and more.
- `tmux-palette` provides the `Ctrl-p` command palette and reads repo-managed JSON from `~/.config/tmux-palette`.

## tmux-palette

The palette config is tracked in this repo at:

```text
./tmux-palette/
```

Bootstrap links it to:

```text
~/.config/tmux-palette
```

The main palette includes a `TMUX Plugins` category with individual palettes for `tmux-fzf`, `TPM`, `tmux-cowboy`, and `tmux-sensible`. It also adds tmux utility commands such as capture pane to file, synchronize panes, clear scrollback, show messages, and the scratch popup.

## Reloading

From inside tmux:

```sh
tmux source-file ~/.tmux.conf
```

Or use:

```text
prefix + r
```

## Bootstrap On A New Machine

Run:

```sh
./bootstrap.sh
```

This script:

- recreates the `~/.tmux.conf` symlink to this repo
- recreates the `~/.config/tmux/scripts/edit-scrollback.sh` symlink to this repo
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
- `~/.config/tmux-palette` symlink target
- TPM installation path
- `tmux source-file ~/.tmux.conf`
- expected live tmux options, terminal features, and key bindings

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
