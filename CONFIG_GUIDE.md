# tmux-omni Configuration Guide

`tmux-omni` is configured via a single JSON file located at:

```text
~/.config/tmux-menu/config.json
```

This file is tracked in this repository at [`./tmux-menu/config.json`](./tmux-menu/config.json) and symlinked by [`./bootstrap.sh`](./bootstrap.sh).

---

## JSON Schema & Editor Autocompletion

The repository provides a formal JSON Schema at [`./tmux-menu/config.schema.json`](./tmux-menu/config.schema.json).

To enable instant autocompletion, hover documentation, and syntax linting in Neovim, VSCode, or Cursor, ensure the `$schema` property is present at the top of your `config.json`:

```json
{
  "$schema": "./config.schema.json",
  "title": "Tmux Menu",
  "items": [
    ...
  ]
}
```

---

## Structure Overview

A menu configuration consists of a root `title` and a list of `items`. Each item can either be a **Submenu** (group containing child `items`) or a **Leaf Action** (executable command).

```json
{
  "key": "p",
  "title": "Panes",
  "icon": "󰓦",
  "description": "splits, layout, zoom, marked panes",
  "items": [
    {
      "key": "v",
      "title": "Split Horizontal",
      "icon": "󰘬",
      "description": "split pane side by side",
      "action": "split-window -h -c '#{pane_current_path}' -t #{pane_id}",
      "target": "tmux"
    }
  ]
}
```

---

## Item Properties

| Property | Type | Required | Description |
| :--- | :--- | :---: | :--- |
| **`key`** | `string` | **Yes** | Hotkey character pressed in Which-Key mode (e.g. `"p"`, `"w"`, `"?"`, `":"`, `"C-s"`). Must be unique among sibling items. |
| **`title`** | `string` | **Yes** | Display title in Which-Key columns and Command Palette search results. |
| **`icon`** | `string` | No | Nerd Font glyph displayed next to the title (e.g. `"󰓦"`, `"󰖲"`, `"󰋜"`). |
| **`description`** | `string` | **Yes** | Explanatory note displayed under Which-Key breadcrumbs and in the Palette table. |
| **`action`** | `string` | If leaf | Shell command or native tmux command to run when selected. |
| **`target`** | `enum` | No (default `"tmux"`) | Execution container. See [Execution Targets](#execution-targets) below. |
| **`popup_width`** | `string` | No (default `"80%"`) | Width of popup window when `target` is `"popup"` or `"popup-shell"`. |
| **`popup_height`** | `string` | No (default `"70%"`) | Height of popup window when `target` is `"popup"` or `"popup-shell"`. |
| **`persist_shell`** | `boolean`| No (default `false`) | If `true`, keeps shell open even after command exits with code 0. |
| **`items`** | `array` | If group | List of child menu items for nested submenus. |

---

## Execution Targets

`tmux-omni` supports 7 execution targets:

### 1. `tmux` (Native Tmux Commands)
Runs native tmux server commands directly (e.g., splitting, window switching, setting options).
```json
{
  "key": "z",
  "title": "Toggle Zoom",
  "icon": "󰊓",
  "description": "zoom or unzoom current pane",
  "action": "resize-pane -Z -t #{pane_id}",
  "target": "tmux"
}
```

### 2. `popup` (Interactive Floating TUI / CLI)
Opens an interactive CLI tool in a floating tmux popup. Automatically guarded with error trapping: if the command fails, it pauses with an error dialog giving you the option to drop into a debug shell `[s]` or close `[any key]`.
```json
{
  "key": "g",
  "title": "Lazygit",
  "icon": "󰊢",
  "description": "terminal Git UI (lazygit)",
  "action": "lazygit",
  "target": "popup",
  "popup_width": "90%",
  "popup_height": "90%"
}
```

### 3. `popup-shell` (Run in Popup & Keep Open)
Runs a script or command in a popup and keeps the interactive shell alive after it finishes so you can inspect output.
```json
{
  "key": "b",
  "title": "Build Project",
  "icon": "󰑕",
  "description": "run build script and keep shell active",
  "action": "make build",
  "target": "popup-shell",
  "popup_width": "80%",
  "popup_height": "75%"
}
```

### 4. `send` (Insert Text into Active Pane Prompt)
Types the command text directly into the active pane's prompt using bracketed paste without executing it.
```json
{
  "key": "d",
  "title": "Docker Status",
  "icon": "󰡨",
  "description": "insert docker ps command into prompt",
  "action": "docker ps --format 'table {{.Names}}\t{{.Status}}\t{{.Ports}}'",
  "target": "send"
}
```

### 5. `split-v` (Vertical Bottom Split)
Spawns a new split pane at the bottom of the current window and runs the command.
```json
{
  "key": "l",
  "title": "Server Logs",
  "icon": "󰈞",
  "description": "tail logs in bottom pane",
  "action": "tail -f /var/log/system.log",
  "target": "split-v"
}
```

### 6. `split-h` (Horizontal Right-Side Split)
Spawns a new split pane on the right side of the current window and runs the command.
```json
{
  "key": "a",
  "title": "AI Assistant Side Pane",
  "icon": "󰚩",
  "description": "open assistant alongside current work",
  "action": "agy",
  "target": "split-h"
}
```

### 7. `window` (New Dedicated Window)
Opens a brand new tmux window and runs the command.
```json
{
  "key": "m",
  "title": "System Monitor Window",
  "icon": "󰍛",
  "description": "open btop in new window",
  "action": "btop",
  "target": "window"
}
```

---

## Dynamic Execution Modifiers in Command Palette

When browsing items in the Command Palette (`prefix + P`), you do not need to edit `config.json` to change how a command is launched. You can override its target on the fly using modifier keys:

* **`<CR>`**: Default target defined in `config.json`
* **`<Alt-v>` / `<Ctrl-v>`**: Override and launch in **Horizontal Side Split**
* **`<Alt-s>` / `<Ctrl-s>`**: Override and launch in **Vertical Bottom Split**
* **`<Alt-w>` / `<Ctrl-t>`**: Override and launch in **New Dedicated Window**
* **`<Alt-i>` / `<Ctrl-y>`**: Override and **Insert in Pane Shell**

---

## Recommended Tokyo Night Nerd Font Icons

| Category | Icon Glyph | Usage Example |
| :--- | :---: | :--- |
| **Panes & Splits** | `󰓦`, `󰘬`, `󰊓` | Select pane, Split pane, Toggle zoom |
| **Windows** | `󰖲`, `󰐕`, `󰁔`, `󰁍` | Windows, New window, Next/Prev window |
| **Sessions** | `󰋜`, `󰙅`, `󰍉` | Session root, Session tree, Choose session |
| **Buffers & History** | `󰅍`, `󰆒`, `󰆏` | Buffers root, Paste buffer, Copy mode |
| **Clients & Displays** | `󰒍`, `󰍛` | Clients root, System monitor |
| **Git & Version Control**| `󰊢`, `󰘬` | Lazygit, Onefetch repository stats |
| **AI & Assistants** | `󰧑`, `󰅚`, `󰁨`, `󰚩` | AI Q&A, Diagnose, Fix, Agent pane |
| **Tools & Utilities** | `󰆍`, `󰉋`, `󰈞` | Tools root, File picker, Search |
| **Configuration** | `󰘳`, `󰑕`, `󰋗` | Options root, Reload config, Help |
| **Destructive Actions** | `󰅖` | Kill pane, Kill window, Kill session |

---

## Adding Custom Commands & Menus

### Example: Adding a "Docker" Submenu

To add a new top-level group for Docker under key `d`:

```json
{
  "key": "d",
  "title": "Docker",
  "icon": "󰡨",
  "description": "manage containers, images, and compose",
  "items": [
    {
      "key": "p",
      "title": "List Running Containers",
      "icon": "󰡨",
      "description": "interactive lazydocker TUI",
      "action": "lazydocker",
      "target": "popup",
      "popup_width": "90%",
      "popup_height": "90%"
    },
    {
      "key": "l",
      "title": "Tail Container Logs",
      "icon": "󰈞",
      "description": "follow docker compose logs in bottom split",
      "action": "docker compose logs -f",
      "target": "split-v"
    }
  ]
}
```

### Example: Adding a Single Shortcut at Root

To add a quick shortcut to run tests at the root level under key `T`:

```json
{
  "key": "T",
  "title": "Run Test Suite",
  "icon": "󰙨",
  "description": "run make test in a floating popup",
  "action": "make test",
  "target": "popup",
  "popup_width": "80%",
  "popup_height": "75%"
}
```

---

## Validating Your Configuration

To verify your configuration syntax and assertions at any time:

1. **Quick In-Editor Check:** VSCode and Neovim will highlight any schema violations if `$schema` is set.
2. **Repository Validation Script:**
   ```bash
   ./validate.sh
   ```
