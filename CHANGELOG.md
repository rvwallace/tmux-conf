# Changelog

This file documents all notable changes to the `tmux-conf` configuration.

The format follows [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

## [Unreleased]

### 2026-08-28

- Enhanced Linux dependency installer (`scripts/install-deps.sh`) with native Fedora/DNF package support for `onefetch` and `uv`, `bun-bin` detection for Terra/COPR repositories, automated COPR handling for `sesh`, `lazygit`, and `yazi`, and Cargo fallback for `tmux-snaglord`.

### 2026-08-25

- Added `tmux-omni --states` saved-state inspector (`prefix + Space` → `t` → `S` / `C`) listing Resurrect and Continuum snapshot history newest-first with age, latest indicator, session/window/pane counts, session names, and search across raw snapshot files. Press `<CR>` to make the selected snapshot `last` and restore it, or `<Ctrl-y>` to copy its raw preview.
- Reorganized AI agent menu entries (`prefix + Space` → `a` → `g`) to use `<Agent> - Open` and `<Agent> - Resume` naming pattern grouped by agent (Antigravity, Claude Code, Codex, Cursor Agent, fx, Pi).
- Added `CHANGELOG.md` following Keep a Changelog standard and updated `README.md` reference.

### 2026-08-24

- Added support for `fx` and `pi` AI coding agents in `scripts/launch-agent-pane.sh`, leader and palette menus (`tmux-menu/config.json`), and file picker path insertion (`scripts/tmux-insert-paths.sh`).

### 2026-08-18

- Switched inspector actions in `tmux-omni` to modifier keys (`Ctrl-a`, `Ctrl-d`, `Ctrl-y`, `Ctrl-i`) so alphanumeric characters can be typed freely into the search filter without triggering accidental item actions.

### 2026-08-16

- Added quote-aware command argument parsing in `tmux-omni` runner (`exec.go`), improved AI error output routing, handled send command errors gracefully, and fixed tree directory resolution.

### 2026-08-15

- Added auto-dismissing status and clipboard notifications (2.5s timeout) in `tmux-omni`.
- Added nested screen back-navigation (`Backspace` on empty search query or `Esc` returns to previous screen).
- Added Action Picker modal (`Ctrl-a` / `Ctrl-o`) for Environment and Options inspectors with quick keys (`v` Value, `n` Name, `e` Shell Export, `s` Tmux Set, `p` Prompt, `i` Insert).
- Added `tmux show-messages` log inspector (`tmux-omni --messages` / `prefix + Space` → `t` → `M`).
- Added universal target modifiers (`Ctrl-i` insert prompt, `Ctrl-w` new window, `Ctrl-v` horizontal split, `Ctrl-s` vertical split, `Ctrl-y` clipboard copy) across all menus and inspectors.
- Added prompt insertion without execution for shell review and manual editing.

### 2026-08-14

- Replaced Python/Textual AI assistant with native Go Bubble Tea implementation integrated directly into `tmux-omni` (`tmux-omni ai <mode>`).
- Added interactive Code Block Picker modal (`X` / `Ctrl-x`) in AI assistant to extract and copy/send any code block from responses into target panes.
- Added Model Switcher overlay (`m`) with runtime switching across models (Claude 3.5 Sonnet, GPT-4o, DeepSeek, Ollama).
- Added session transcript export to clean Markdown (`S` / `E`) and open in `$EDITOR`.
- Added Glamour terminal Markdown rendering, live Braille spinner animation, and bracketed paste auto-run feedback.
- Redesigned command generation and fix suggestion popups with compact top-down card layout.
- Added `tmux-menu/config.schema.json` for editor autocompletion and hover documentation.
- Added `CONFIG_GUIDE.md` detailing menu configuration, 7 execution targets, modifiers, and icons.
- Implemented native `tmux-omni --validate` configuration linter and validator.
- Removed legacy Python leader menu and prefix help scripts (`scripts/tmux-menu.py`, `scripts/tmux-prefix-help.py`).
- Implemented `tmux-omni` in Go + Bubble Tea: sub-10ms unified Which-Key leader menu (`prefix + Space`) and fuzzy Command Palette (`prefix + P`).
- Added Go to `Brewfile` and `scripts/install-deps.sh`, so `bootstrap.sh` compiles `tmux-omni`.

### 2026-08-13

- Added standalone Textual TUI app in `scripts/aichat-tui.py` for AI actions with multiline input, follow-up turns, and copy support.
- Added `scripts/launch-agent-pane.sh` for opening Antigravity, Claude Code, Codex, and Cursor Agent in 45% right-side panes rooted at active pane directory.

### 2026-08-10

- Fixed theme loading for existing installations by symlinking `themes/` to `~/.config/tmux/themes`.

### 2026-08-09

- Added Tokyo Night Storm status line theme in `themes/tokyo-night.conf` with dark base, layered surfaces, and Nerd Font icons.
- Reorganized `.tmux.conf` into modular sections for core behavior, terminal capabilities, window/pane defaults, bindings, context menus, theme loading, and plugins.
- Added tool location menus and directory inspection utilities to menu configuration.

### 2026-08-07

- Added `alberti42/tmux-fzf-links` plugin (`prefix + Ctrl-h`) for fuzzy-finding URLs and file links from pane scrollback.
- Added Yazi file picker integration (`scripts/tmux-yazi-picker.sh`, `scripts/tmux-insert-paths.sh`) as a navigable, multi-file chooser for the active pane.

### 2026-07-25

- Added initial which-key workflow and keybinding hints popup.

### 2026-07-22

- Set explicit `update-environment` list to prevent duplicate `TERM` and `TERM_PROGRAM` entries on repeated configuration reloads.
- Enabled terminal escape-sequence passthrough (`allow-passthrough on`) and client environment refresh for Yazi image/PDF graphics previews in Ghostty and Kitty.
- Configured mouse right-click context menus with persistent mode (`display-menu -O`) so mouse movement does not dismiss menus.
- Vendored and integrated `scripts/tmux-file-picker` (`prefix + Ctrl-f`) with `fd` and `fzf` for fuzzy selecting and inserting file paths into active panes.
- Added `tmux-snaglord` prompt parser configuration in `tmux-snaglord/config.toml` (`prefix + Ctrl-y`) for browsing command history as searchable blocks.

### 2026-07-21

- Improved AI assistant workflows, scrollback capture (200 lines default), and candidate command copy workflows.

### 2026-07-20

- Added `LICENSE` (MIT).
- Configured `detach-on-destroy off` to switch client to another active session when current session is destroyed.
- Standardized AI palette action naming with `AI:` prefix and searchable tags.
- Added pane-aware `aichat` actions (Ask, Diagnose Error, Suggest Fix, Summarize Pane, Explain).

### 2026-07-19

- Expanded status bar with Nerd Font icons, UTC/local clock segments, and pane navigation helpers.
- Enabled labeled top pane borders showing pane index, title, command, and directory basename (`pane-border-status top`).

### 2026-07-18

- Added `bootstrap.sh`, `scripts/install-deps.sh`, and `Brewfile` for cross-platform dependency management across macOS and Linux.
- Added workflow utilities: `copy-pane-path.sh`, `edit-scrollback.sh` (`prefix + Ctrl-e`), and `show-clients.sh`.
- Added `laktak/extrakto` plugin (`prefix + Tab`) for fuzzy extracting text, URLs, and paths from pane scrollback.
- Refined window tab styling with rounded caps and active window highlights.

### 2026-06-30

- Enabled CSI-u extended keys reporting (`extended-keys on`) for modified key combinations (e.g. `Shift-Enter`).

### 2026-06-28

- Removed redundant next-window keybinding.

### 2026-06-20

- Improved tmux status line styling with native formatting and color definitions.

### 2026-05-23

- Added `tmux-plugins/tmux-resurrect` and `tmux-plugins/tmux-continuum` for automated session saving and restoration.

### 2026-05-15

- Added `tmux-palette` plugin integration and initial palette actions.

### 2026-04-18

- Added `prefix + T` keybinding for `sesh` session picker with `fzf-tmux`.

### 2026-04-09

- Initial repository commit with core `.tmux.conf` configuration, `Ctrl-a` prefix, 100k history limit, and vi copy mode.
