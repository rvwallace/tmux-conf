# tmux-conf Agents Guide

This repository owns the tmux config symlinked to `~/.tmux.conf`.

## Scope

- The source of truth is [`./.tmux.conf`](./.tmux.conf).
- User-facing documentation lives in [`./README.md`](./README.md).
- TPM plugins are installed outside the repo in `~/.config/tmux/plugins/`.
- Repo-managed helper scripts are installed outside the repo in `~/.config/tmux/scripts/`.

## Working Rules

- Keep the setup minimal. Prefer native tmux options and bindings over adding plugins.
- Current plugin policy: only keep a plugin if native tmux would be materially worse or significantly more tedious.
- Before adding a plugin, document why native tmux is insufficient.
- Do not reintroduce `oh-my-tmux` or config layers that depend on `~/.tmux`, `~/.tmux.conf.local`, or generated wrapper files.

## Required Documentation Updates

- If you change [`./.tmux.conf`](./.tmux.conf), review and update [`./README.md`](./README.md) in the same change.
- If you change key bindings or common workflows, review and update [`./CHEATSHEET.md`](./CHEATSHEET.md) in the same change.
- If you change setup flow or external paths, review and update [`./bootstrap.sh`](./bootstrap.sh) and [`./README.md`](./README.md) in the same change.
- If you change expected runtime behavior, review and update [`./validate.sh`](./validate.sh) in the same change.
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
- TPM path: `~/.config/tmux/plugins/`
- Active plugins:
  - `tmux-plugins/tmux-sensible`
  - `tmux-plugins/tmux-cowboy`
