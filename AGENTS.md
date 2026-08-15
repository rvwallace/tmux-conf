# tmux-conf Agents Guide

This repository owns the tmux config symlinked to `~/.tmux.conf`.

## Scope

- The authoritative entry point is [`./.tmux.conf`](./.tmux.conf).
- Repo-owned theme settings live in [`./themes/`](./themes/) and are sourced by the entry point.
- User-facing documentation lives in [`./README.md`](./README.md).
- TPM plugins are installed outside the repo in `~/.config/tmux/plugins/`.
- Repo-managed helper scripts are installed outside the repo in `~/.config/tmux/scripts/`.
- Repo-managed `tmux-menu` config is symlinked outside the repo to `~/.config/tmux-menu`.

## Working Rules

- Keep the setup minimal. Prefer native tmux options and bindings over adding plugins.
- Current plugin policy: only keep a plugin if native tmux would be materially worse or significantly more tedious.
- Before adding a plugin, document why native tmux is insufficient.
- Do not reintroduce `oh-my-tmux` or config layers that depend on `~/.tmux`, `~/.tmux.conf.local`, or generated wrapper files.

## Required Documentation Updates

- If you change [`./.tmux.conf`](./.tmux.conf) or a file in [`./themes/`](./themes/), review and update [`./README.md`](./README.md) in the same change.
- If you change key bindings or common workflows, review and update [`./CHEATSHEET.md`](./CHEATSHEET.md) in the same change.
- If you change setup flow or external paths, review and update [`./bootstrap.sh`](./bootstrap.sh) and [`./README.md`](./README.md) in the same change.
- If you change expected runtime behavior, review and update [`./validate.sh`](./validate.sh) in the same change.
- If you change menu commands or categories, review and update [`./tmux-menu/`](./tmux-menu), [`./README.md`](./README.md), [`./CHEATSHEET.md`](./CHEATSHEET.md), and [`./validate.sh`](./validate.sh) in the same change.
- Keep these sections accurate in `README.md`:
  - current plugins
  - key bindings added by this repo
  - install/reload instructions
  - file layout and external paths
- If behavior changes but the docs do not, the work is incomplete.

## Validation

- Reload config with: `tmux source-file ~/.tmux.conf`
- Verify active plugin bindings when relevant with: `tmux list-keys`
- Prefer validating against the running tmux server instead of only reading the file.

## Current Baseline

- Minimal standalone tmux config.
- Tokyo Night theme path: `themes/tokyo-night.conf`
- TPM path: `~/.config/tmux/plugins/`
- `tmux-menu` config path: `~/.config/tmux-menu`
- Unified Leader Menu & Command Palette: `tmux-omni` (Go + Bubble Tea TUI, binary at `~/.config/tmux/scripts/tmux-omni`)
- Active plugins:
  - `tmux-plugins/tmux-sensible`
  - `tmux-plugins/tmux-cowboy`
  - `sainnhe/tmux-fzf`
  - `laktak/extrakto`
  - `alberti42/tmux-fzf-links`
  - `tmux-plugins/tmux-resurrect`
  - `tmux-plugins/tmux-continuum`
- `tmux-continuum` should stay last in the TPM plugin list because it hooks through `status-right`.
- `tmux-resurrect` save files use `@resurrect-dir` if set; otherwise the plugin uses `~/.tmux/resurrect` when that directory exists, falling back to `${XDG_DATA_HOME:-~/.local/share}/tmux/resurrect`.
