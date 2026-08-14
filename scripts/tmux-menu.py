#!/usr/bin/env -S uv run --script
# /// script
# requires-python = ">=3.10"
# dependencies = [
#     "textual>=0.80.0",
# ]
# ///
"""Unified Leader Menu & Command Palette for tmux with Tokyo Night theme."""

import argparse
import json
import os
import re
import shlex
import shutil
import subprocess
import sys
from pathlib import Path
from typing import Any, Dict, List, Optional, Tuple

from textual.app import App, ComposeResult
from textual.binding import Binding
from textual.containers import Container, Grid, Horizontal, Vertical, VerticalScroll
from textual.events import Key
from textual.message import Message
from textual.reactive import reactive
from textual.screen import ModalScreen
from textual.widgets import Input, Label, ListItem, ListView, Static

THEME_CSS = """
Screen {
    background: #1a1b26;
    color: #c0caf5;
    layout: vertical;
}

#main-container {
    background: #1a1b26;
    height: 100%;
    width: 100%;
    layout: vertical;
}

#header-bar {
    dock: top;
    height: 3;
    background: #24283b;
    border-bottom: solid #414868;
    padding: 0 2;
    layout: horizontal;
    align-vertical: middle;
}

#breadcrumb-label {
    width: 1fr;
    color: #7aa2f7;
    text-style: bold;
}

#header-hint {
    width: auto;
    color: #565f89;
    text-style: bold;
}

#search-bar-container {
    dock: top;
    height: 3;
    background: #24283b;
    border-bottom: solid #414868;
    padding: 0 1;
    layout: horizontal;
    align-vertical: middle;
}

#search-input {
    width: 1fr;
    background: #1a1b26;
    color: #c0caf5;
    border: none;
    padding: 0 1;
}

#search-input:focus {
    border: none;
}

#search-badge {
    width: auto;
    color: #7dcfff;
    padding-left: 1;
    padding-right: 1;
    text-style: bold;
}

#grid-scroll {
    height: 1fr;
    padding: 1 2;
}

#menu-grid {
    layout: grid;
    grid-size: 3;
    grid-columns: 1fr 1fr 1fr;
    grid-rows: 3;
    grid-gutter: 1 2;
    padding: 0;
    height: auto;
}

.menu-card {
    height: 3;
    background: #24283b;
    border: solid #414868;
    padding: 0 1;
    layout: horizontal;
    align-vertical: middle;
}

.menu-card:hover {
    background: #283457;
    border: solid #7aa2f7;
}

.menu-card-key {
    width: auto;
    background: #3d59a1;
    color: #ffffff;
    text-style: bold;
    padding: 0 1;
    margin-right: 1;
}

.menu-card-icon {
    width: auto;
    color: #7dcfff;
    padding-right: 1;
}

.menu-card-text-container {
    width: 1fr;
    layout: vertical;
}

.menu-card-title {
    color: #c0caf5;
    text-style: bold;
}

.menu-card-desc {
    color: #7a88cf;
}

#palette-scroll {
    height: 1fr;
    padding: 0;
}

#palette-list {
    background: #1a1b26;
    height: 100%;
    padding: 0;
}

ListView > ListItem {
    background: #1a1b26;
    padding: 0 1;
    height: 2;
    layout: horizontal;
    align-vertical: middle;
    border-left: solid transparent;
}

ListView > ListItem:hover {
    background: #24283b;
}

ListView > ListItem.-highlight,
ListView:blur > ListItem.-highlight,
PaletteItemWidget.-highlight {
    background: #34446e;
    border-left: wide #7aa2f7;
}

.palette-item {
    background: transparent;
    width: 100%;
    height: 100%;
    layout: horizontal;
    align-vertical: middle;
}

.palette-icon {
    width: auto;
    color: #7dcfff;
    padding-right: 1;
}

.palette-category {
    width: auto;
    color: #bb9af7;
    text-style: bold;
    padding-right: 1;
}

.palette-title {
    width: 1fr;
    color: #c0caf5;
    text-style: bold;
}

ListView > ListItem.-highlight .palette-title,
PaletteItemWidget.-highlight .palette-title {
    color: #ffffff;
    text-style: bold;
}

.palette-keys {
    width: auto;
    background: #1f2335;
    color: #ff9e64;
    text-style: bold;
    padding: 0 1;
    margin-right: 1;
}

ListView > ListItem.-highlight .palette-keys,
PaletteItemWidget.-highlight .palette-keys {
    background: #7aa2f7;
    color: #1a1b26;
    text-style: bold;
}

.palette-desc {
    width: auto;
    color: #565f89;
}

ListView > ListItem.-highlight .palette-desc,
PaletteItemWidget.-highlight .palette-desc {
    color: #c0caf5;
}

#footer-bar {
    dock: bottom;
    height: 2;
    background: #24283b;
    border-top: solid #414868;
    padding: 0 2;
    layout: horizontal;
    align-vertical: middle;
}

#footer-legend {
    width: 1fr;
    color: #7aa2f7;
}

#footer-status {
    width: auto;
    color: #9ece6a;
    text-style: bold;
}

#error-modal {
    width: 80%;
    height: 60%;
    background: #1a1b26;
    border: double #f7768e;
    padding: 1 2;
    layout: vertical;
}

#error-title {
    color: #f7768e;
    text-style: bold;
    height: 1;
}

#error-details {
    height: 1fr;
    background: #24283b;
    border: solid #414868;
    padding: 1;
    margin: 1 0;
    color: #c0caf5;
}

#error-footer {
    height: 1;
    color: #7dcfff;
}
"""


def copy_to_clipboard(text: str) -> bool:
    """Copy text to system clipboard across macOS, Wayland, X11, or tmux buffer."""
    if shutil.which("pbcopy"):
        try:
            subprocess.run(["pbcopy"], input=text.encode("utf-8"), check=True)
            return True
        except Exception:
            pass
    elif shutil.which("wl-copy"):
        try:
            subprocess.run(["wl-copy"], input=text.encode("utf-8"), check=True)
            return True
        except Exception:
            pass
    elif shutil.which("xclip"):
        try:
            subprocess.run(["xclip", "-selection", "clipboard"], input=text.encode("utf-8"), check=True)
            return True
        except Exception:
            pass

    # Fallback to tmux buffer
    if shutil.which("tmux"):
        try:
            subprocess.run(["tmux", "set-buffer", "--", text], check=True)
            return True
        except Exception:
            pass
    return False


def get_default_config_path() -> Path:
    """Find the menu configuration file."""
    # 1. ~/.config/tmux-menu/config.json
    xdg_path = Path.home() / ".config" / "tmux-menu" / "config.json"
    if xdg_path.exists():
        return xdg_path

    # 2. Alongside script in repo: ../tmux-menu/config.json
    repo_path = Path(__file__).resolve().parent.parent / "tmux-menu" / "config.json"
    if repo_path.exists():
        return repo_path

    # 3. ~/.config/tmux/menu.json
    alt_path = Path.home() / ".config" / "tmux" / "menu.json"
    if alt_path.exists():
        return alt_path

    # 4. Fallback to repo path
    return repo_path


def load_config(config_path: Path) -> Dict[str, Any]:
    """Load configuration from JSON file."""
    if not config_path.exists():
        return {
            "title": "Tmux Menu",
            "items": [],
        }
    with open(config_path, "r", encoding="utf-8") as f:
        return json.load(f)


class FlatCommand:
    """Represents a leaf command in the flattened palette index."""

    def __init__(
        self,
        title: str,
        icon: str,
        category: str,
        description: str,
        key_seq: str,
        action: str,
        target: str,
        popup_width: str = "80%",
        popup_height: str = "80%",
        persist_shell: bool = False,
        raw: Optional[Dict[str, Any]] = None,
    ):
        self.title = title
        self.icon = icon
        self.category = category
        self.description = description
        self.key_seq = key_seq
        self.action = action
        self.target = target or "tmux"
        self.popup_width = popup_width
        self.popup_height = popup_height
        self.persist_shell = persist_shell
        self.raw = raw or {}

    @property
    def searchable_text(self) -> str:
        return f"{self.title} {self.category} {self.description} {self.key_seq}".lower()


def build_flat_commands(items: List[Dict[str, Any]], breadcrumbs: Optional[List[str]] = None, key_prefix: str = "") -> List[FlatCommand]:
    """Recursively flatten hierarchical menu into searchable leaf commands."""
    if breadcrumbs is None:
        breadcrumbs = []

    results: List[FlatCommand] = []

    for item in items:
        key = item.get("key", "")
        title = item.get("title", "")
        icon = item.get("icon", "󰘳")
        desc = item.get("description", "")
        current_seq = f"{key_prefix} {key}".strip() if key else key_prefix

        if "items" in item and item["items"]:
            # Submenu group
            sub_crumbs = breadcrumbs + [title]
            results.extend(build_flat_commands(item["items"], sub_crumbs, current_seq))
        elif "action" in item:
            category = " > ".join(breadcrumbs) if breadcrumbs else "Root"
            results.append(
                FlatCommand(
                    title=title,
                    icon=icon,
                    category=category,
                    description=desc,
                    key_seq=current_seq,
                    action=item.get("action", ""),
                    target=item.get("target", "tmux"),
                    popup_width=item.get("popup_width", "80%"),
                    popup_height=item.get("popup_height", "80%"),
                    persist_shell=bool(item.get("persist_shell") or item.get("shell")),
                    raw=item,
                )
            )

    return results


class ErrorModal(ModalScreen[None]):
    """Modal screen displaying an execution error with copy support."""

    def __init__(self, title: str, error_text: str):
        super().__init__()
        self.modal_title = title
        self.error_text = error_text

    def compose(self) -> ComposeResult:
        with Container(id="error-modal"):
            yield Label(f"✖ {self.modal_title}", id="error-title", markup=False)
            yield Static(self.error_text, id="error-details")
            yield Label("[y/c] Copy Error to Clipboard   [Esc/q/Enter] Dismiss", id="error-footer", markup=False)

    def on_key(self, event: Key) -> None:
        if event.key in ("y", "c"):
            copy_to_clipboard(self.error_text)
            self.notify("Copied error to clipboard!", severity="information")
        elif event.key in ("escape", "q", "enter"):
            self.dismiss()


class PaletteItemWidget(ListItem):
    """List item widget for the command palette."""

    def __init__(self, cmd: FlatCommand, initial_highlight: bool = False):
        classes = "-highlight" if initial_highlight else ""
        super().__init__(classes=classes)
        self.cmd = cmd

    def compose(self) -> ComposeResult:
        with Horizontal(classes="palette-item"):
            yield Label(self.cmd.icon, classes="palette-icon", markup=False)
            if self.cmd.category:
                yield Label(f"[{self.cmd.category}]", classes="palette-category", markup=False)
            yield Label(self.cmd.title, classes="palette-title", markup=False)
            if self.cmd.key_seq:
                yield Label(f"<{self.cmd.key_seq}>", classes="palette-keys", markup=False)
            yield Label(self.cmd.description, classes="palette-desc", markup=False)


class SearchInput(Input):
    """Search input field with palette list navigation and target modifier shortcuts."""

    BINDINGS = [
        # Navigation
        Binding("tab", "cursor_down", "Next Item", show=False, priority=True),
        Binding("shift+tab", "cursor_up", "Previous Item", show=False, priority=True),
        Binding("down", "cursor_down", "Next Item", show=False, priority=True),
        Binding("up", "cursor_up", "Previous Item", show=False, priority=True),
        Binding("ctrl+n", "cursor_down", "Next Item", show=False, priority=True),
        Binding("ctrl+p", "cursor_up", "Previous Item", show=False, priority=True),
        Binding("ctrl+j", "cursor_down", "Next Item", show=False, priority=True),
        Binding("ctrl+k", "cursor_up", "Previous Item", show=False, priority=True),

        # Target modifiers (supports both Alt and Ctrl combinations)
        Binding("alt+v", "run_split_h", "Side Split", show=False, priority=True),
        Binding("ctrl+v", "run_split_h", "Side Split", show=False, priority=True),
        Binding("alt+s", "run_split_v", "Bottom Split", show=False, priority=True),
        Binding("ctrl+s", "run_split_v", "Bottom Split", show=False, priority=True),
        Binding("alt+w", "run_new_window", "New Window", show=False, priority=True),
        Binding("ctrl+w", "run_new_window", "New Window", show=False, priority=True),
        Binding("ctrl+t", "run_new_window", "New Window", show=False, priority=True),
        Binding("alt+i", "run_send_keys", "Send to Pane", show=False, priority=True),
        Binding("ctrl+i", "run_send_keys", "Send to Pane", show=False, priority=True),
        Binding("ctrl+y", "run_send_keys", "Send to Pane", show=False, priority=True),
    ]

    async def _on_key(self, event: Key) -> None:
        if event.key in ("square_root", "√") or event.character == "√":
            self.action_run_split_h()
            event.prevent_default()
            event.stop()
            return
        if event.key in ("ß", "ssharp") or event.character == "ß":
            self.action_run_split_v()
            event.prevent_default()
            event.stop()
            return
        if event.key in ("n_ary_summation", "∑") or event.character == "∑":
            self.action_run_new_window()
            event.prevent_default()
            event.stop()
            return
        if event.key in ("circumflex_accent", "ˆ", "^") or event.character in ("ˆ", "^"):
            self.action_run_send_keys()
            event.prevent_default()
            event.stop()
            return
        await super()._on_key(event)

    def action_cursor_down(self) -> None:
        lv = self.app.query_one("#palette-list", ListView)
        if lv.children:
            if lv.index is None or lv.index < 0:
                lv.index = 0
            elif lv.index >= len(lv.children) - 1:
                lv.index = 0
            else:
                lv.action_cursor_down()

    def action_cursor_up(self) -> None:
        lv = self.app.query_one("#palette-list", ListView)
        if lv.children:
            if lv.index is None or lv.index <= 0:
                lv.index = len(lv.children) - 1
            else:
                lv.action_cursor_up()

    def action_run_split_h(self) -> None:
        self.app.action_run_split_h()

    def action_run_split_v(self) -> None:
        self.app.action_run_split_v()

    def action_run_new_window(self) -> None:
        self.app.action_run_new_window()

    def action_run_send_keys(self) -> None:
        self.app.action_run_send_keys()


class TmuxMenuApp(App[None]):
    """Textual app for Which-Key & Command Palette."""

    CSS = THEME_CSS
    TITLE = "Tmux Menu"

    BINDINGS = [
        Binding("alt+v", "run_split_h", "Side Split", show=False),
        Binding("ctrl+v", "run_split_h", "Side Split", show=False),
        Binding("alt+s", "run_split_v", "Bottom Split", show=False),
        Binding("ctrl+s", "run_split_v", "Bottom Split", show=False),
        Binding("alt+w", "run_new_window", "New Window", show=False),
        Binding("ctrl+w", "run_new_window", "New Window", show=False),
        Binding("ctrl+t", "run_new_window", "New Window", show=False),
        Binding("alt+i", "run_send_keys", "Send to Pane", show=False),
        Binding("ctrl+i", "run_send_keys", "Send to Pane", show=False),
        Binding("ctrl+y", "run_send_keys", "Send to Pane", show=False),
    ]

    mode = reactive("which-key")  # "which-key" or "palette"
    search_query = reactive("")

    def __init__(self, pane_id: str, start_search: bool = False, config_path: Optional[Path] = None):
        super().__init__()
        self.pane_id = pane_id
        self.start_search = start_search
        self.config_path = config_path or get_default_config_path()
        self.config = load_config(self.config_path)

        # Navigation stack for Which-Key: [(title, icon, items)]
        root_items = self.config.get("items", [])
        self.nav_stack: List[Tuple[str, str, List[Dict[str, Any]]]] = [
            (self.config.get("title", "Tmux Leader"), "󰍡", root_items)
        ]

        # Flat index for Palette
        self.flat_commands = build_flat_commands(root_items)
        self.filtered_commands: List[FlatCommand] = list(self.flat_commands)

        # Action to execute after exiting TUI
        self.pending_execution: Optional[Tuple[str, str, str, str, str, bool]] = None

    def compose(self) -> ComposeResult:
        with Container(id="root-container"):
            # Header Bar for Which-Key
            with Horizontal(id="header-bar"):
                yield Label("󰍡 Tmux Leader", id="breadcrumb-label", markup=False)
                yield Label("[/] Search   [Esc] Back   [q] Quit", id="header-hint", markup=False)

            # Search Bar for Command Palette
            with Horizontal(id="search-bar-container"):
                yield Label("󰍉", classes="palette-icon", markup=False)
                yield SearchInput(placeholder="Type to filter commands, categories, keys...", id="search-input")
                yield Label(f"{len(self.flat_commands)} commands", id="search-badge", markup=False)

            # Which-Key Grid Container
            with VerticalScroll(id="grid-scroll"):
                yield Grid(id="menu-grid")

            # Palette List Container
            with Vertical(id="palette-scroll"):
                yield ListView(id="palette-list")

            # Footer Bar
            with Horizontal(id="footer-bar"):
                yield Label(
                    "[Enter] Run  [Alt+v] Side  [Alt+s] Split  [Alt+w] Window  [Alt+i] Send",
                    id="footer-legend",
                    markup=False,
                )
                yield Label("Tokyo Night", id="footer-status", markup=False)

    def on_mount(self) -> None:
        if self.start_search:
            self.set_mode("palette")
        else:
            self.set_mode("which-key")

    def set_mode(self, mode: str) -> None:
        """Switch between Which-Key mode and Command Palette mode."""
        self.mode = mode
        header_bar = self.query_one("#header-bar")
        search_bar = self.query_one("#search-bar-container")
        grid_scroll = self.query_one("#grid-scroll")
        palette_scroll = self.query_one("#palette-scroll")
        search_input = self.query_one("#search-input", SearchInput)

        if mode == "palette":
            header_bar.display = False
            search_bar.display = True
            grid_scroll.display = False
            palette_scroll.display = True
            search_input.disabled = False
            self.refresh_palette_results()
            search_input.focus()
        else:
            header_bar.display = True
            search_bar.display = False
            grid_scroll.display = True
            palette_scroll.display = False
            search_input.disabled = True
            self.set_focus(None)
            self.refresh_which_key_grid()

    def refresh_which_key_grid(self) -> None:
        """Render the grid for the current navigation stack level."""
        current_title, current_icon, current_items = self.nav_stack[-1]

        # Update breadcrumb label
        breadcrumbs = "  >  ".join(f"{icon} {title}" for title, icon, _ in self.nav_stack)
        self.query_one("#breadcrumb-label", Label).update(breadcrumbs)

        grid = self.query_one("#menu-grid", Grid)
        grid.remove_children()

        cards = []
        for item in current_items:
            key = item.get("key", "")
            title = item.get("title", "")
            icon = item.get("icon", "󰘳")
            desc = item.get("description", "")
            is_group = "items" in item and item["items"]
            subtext = f"+{len(item['items'])} items" if is_group else desc

            card = Container(
                Label(f"[{key}]", classes="menu-card-key", markup=False),
                Label(icon, classes="menu-card-icon", markup=False),
                Vertical(
                    Label(title, classes="menu-card-title", markup=False),
                    Label(subtext, classes="menu-card-desc", markup=False),
                    classes="menu-card-text-container",
                ),
                classes="menu-card",
            )
            cards.append(card)

        grid.mount_all(cards)

    def refresh_palette_results(self) -> None:
        """Filter and render palette items based on current search query."""
        query = self.search_query.strip().lower()
        if not query:
            self.filtered_commands = list(self.flat_commands)
        else:
            words = query.split()
            self.filtered_commands = [
                cmd for cmd in self.flat_commands if all(w in cmd.searchable_text for w in words)
            ]

        self.query_one("#search-badge", Label).update(f"{len(self.filtered_commands)} commands")

        palette_list = self.query_one("#palette-list", ListView)
        palette_list.clear()

        widgets = []
        for idx, cmd in enumerate(self.filtered_commands):
            w = PaletteItemWidget(cmd, initial_highlight=(idx == 0))
            if idx == 0:
                w.highlighted = True
            widgets.append(w)

        for w in widgets:
            palette_list.append(w)

        self.call_after_refresh(self._sync_palette_index)

    def _sync_palette_index(self) -> None:
        palette_list = self.query_one("#palette-list", ListView)
        if palette_list.children:
            palette_list.index = 0

    def on_input_changed(self, event: Input.Changed) -> None:
        if event.input.id == "search-input":
            self.search_query = event.value
            self.refresh_palette_results()

    def on_input_submitted(self, event: Input.Submitted) -> None:
        if event.input.id == "search-input":
            self.execute_selected_palette_item(persist_shell=False)

    def on_list_view_selected(self, event: ListView.Selected) -> None:
        if isinstance(event.item, PaletteItemWidget):
            self.execute_command(event.item.cmd, persist_shell=False)

    def on_key(self, event: Key) -> None:
        if self.mode == "which-key":
            self.handle_which_key_keypress(event)
        elif self.mode == "palette":
            self.handle_palette_keypress(event)

    def handle_which_key_keypress(self, event: Key) -> None:
        """Handle single-key mnemonic navigation in Which-Key mode."""
        if event.key in ("slash", "ctrl+p"):
            self.set_mode("palette")
            event.prevent_default()
            return

        if event.key in ("escape", "backspace"):
            self.action_handle_back()
            event.prevent_default()
            return

        if event.key == "q":
            self.action_handle_quit()
            event.prevent_default()
            return

        # Look up key in current items
        _, _, current_items = self.nav_stack[-1]
        for item in current_items:
            item_key = item.get("key", "")
            if event.character == item_key or event.key == item_key:
                if "items" in item and item["items"]:
                    # Drill down into submenu
                    self.nav_stack.append((item.get("title", "Submenu"), item.get("icon", "󰘳"), item["items"]))
                    self.refresh_which_key_grid()
                elif "action" in item:
                    # Leaf action
                    cmd = FlatCommand(
                        title=item.get("title", ""),
                        icon=item.get("icon", "󰘳"),
                        category="",
                        description=item.get("description", ""),
                        key_seq=item_key,
                        action=item.get("action", ""),
                        target=item.get("target", "tmux"),
                        popup_width=item.get("popup_width", "80%"),
                        popup_height=item.get("popup_height", "80%"),
                        persist_shell=bool(item.get("persist_shell") or item.get("shell")),
                    )
                    self.execute_command(cmd, target_override=None)
                event.prevent_default()
                return

    def handle_palette_keypress(self, event: Key) -> None:
        """Handle navigation and selection in Command Palette mode."""
        palette_list = self.query_one("#palette-list", ListView)

        if event.key in ("down", "ctrl+n", "tab"):
            palette_list.action_cursor_down()
            event.prevent_default()
            return

        if event.key in ("up", "ctrl+p", "shift+tab"):
            palette_list.action_cursor_up()
            event.prevent_default()
            return

        if event.key == "escape":
            search_input = self.query_one("#search-input", Input)
            if search_input.value:
                search_input.value = ""
            elif not self.start_search:
                self.set_mode("which-key")
            else:
                self.exit()
            event.prevent_default()
            return

        if event.key == "enter":
            self.execute_selected_palette_item()
            event.prevent_default()
            return

    def get_selected_palette_command(self) -> Optional[FlatCommand]:
        """Get currently highlighted command in palette ListView."""
        palette_list = self.query_one("#palette-list", ListView)
        if palette_list.highlighted_child is not None and isinstance(palette_list.highlighted_child, PaletteItemWidget):
            return palette_list.highlighted_child.cmd
        if self.filtered_commands:
            return self.filtered_commands[0]
        return None

    def execute_selected_palette_item(self, target_override: Optional[str] = None) -> None:
        cmd = self.get_selected_palette_command()
        if cmd:
            self.execute_command(cmd, target_override=target_override)

    def execute_command(self, cmd: FlatCommand, target_override: Optional[str] = None) -> None:
        """Queue command execution and exit TUI."""
        target = target_override or cmd.target
        self.pending_execution = (cmd.action, target, cmd.popup_width, cmd.popup_height, cmd.title, cmd.persist_shell, cmd.target)
        self.exit()

    def action_handle_back(self) -> None:
        if len(self.nav_stack) > 1:
            self.nav_stack.pop()
            self.refresh_which_key_grid()
        else:
            self.exit()

    def action_handle_quit(self) -> None:
        self.exit()

    def action_toggle_search(self) -> None:
        if self.mode == "which-key":
            self.set_mode("palette")
        else:
            self.set_mode("which-key")

    def action_run_split_h(self) -> None:
        if self.mode == "palette":
            self.execute_selected_palette_item(target_override="split_h")

    def action_run_split_v(self) -> None:
        if self.mode == "palette":
            self.execute_selected_palette_item(target_override="split_v")

    def action_run_new_window(self) -> None:
        if self.mode == "palette":
            self.execute_selected_palette_item(target_override="window")

    def action_run_send_keys(self) -> None:
        if self.mode == "palette":
            self.execute_selected_palette_item(target_override="send_keys")


def run_tmux_target(
    action: str,
    target: str,
    pane_id: str,
    popup_w: str = "80%",
    popup_h: str = "80%",
    title: str = "",
    persist_shell: bool = False,
    original_target: str = "tmux",
) -> None:
    """Execute action in tmux with error guards and persistence modes."""
    # Sanitize pane ID (strip any quotes passed from shell invocations)
    clean_pane_id = (pane_id or "").strip("'\" \t\r\n")
    if (not clean_pane_id or not clean_pane_id.startswith("%")) and shutil.which("tmux"):
        try:
            clean_pane_id = subprocess.check_output(["tmux", "display-message", "-p", "#{pane_id}"], text=True).strip()
        except Exception:
            clean_pane_id = "%0"

    # Resolve pane current path
    cwd = ""
    if clean_pane_id:
        try:
            cwd = subprocess.check_output(
                ["tmux", "display-message", "-p", "-t", clean_pane_id, "#{pane_current_path}"],
                text=True,
            ).strip()
        except Exception:
            pass

    # Change working directory if valid
    if cwd and os.path.isdir(cwd):
        try:
            os.chdir(cwd)
        except Exception:
            pass

    # Resolve tmux template variables if present
    expanded_action = action.replace("#{pane_id}", clean_pane_id)
    if cwd:
        expanded_action = expanded_action.replace("#{pane_current_path}", cwd)

    shell_bin = os.environ.get("SHELL", "/bin/zsh")

    # If it's a direct tmux command
    if target == "tmux":
        # Handle chained commands separated by ';'
        parts = [p.strip() for p in re.split(r'(?<!\\);', expanded_action) if p.strip()]
        for part in parts:
            part = part.replace(r'\;', ';')
            try:
                subprocess.Popen(["tmux"] + shlex.split(part), start_new_session=True)
            except Exception as e:
                subprocess.run(["tmux", "display-message", f"✖ Execution failed: {e}"])
        return

    # If it's send_keys to current pane
    if target == "send_keys":
        text_to_send = expanded_action
        if original_target == "tmux":
            # Split chained commands by ';' and prefix with tmux for shell execution
            sub_cmds = [p.strip() for p in re.split(r'(?<!\\);', text_to_send) if p.strip()]
            formatted_cmds = []
            for sc in sub_cmds:
                if not sc.startswith("tmux "):
                    formatted_cmds.append(f"tmux {sc}")
                else:
                    formatted_cmds.append(sc)
            text_to_send = " ; ".join(formatted_cmds)

        cmd = ["tmux", "send-keys"]
        if clean_pane_id:
            cmd.extend(["-t", clean_pane_id])
        cmd.append(text_to_send)
        subprocess.Popen(cmd, start_new_session=True)
        return

    # Build shell command wrapper with Error Guard & Persistent Shell support
    guarded_script = build_guarded_shell_script(expanded_action, title, persist_shell)

    if target == "split_h":
        cmd = ["tmux", "split-window", "-h"]
        if clean_pane_id:
            cmd.extend(["-t", clean_pane_id])
        if cwd:
            cmd.extend(["-c", cwd])
        cmd.extend([shell_bin, "-lc", guarded_script])
        subprocess.Popen(cmd, start_new_session=True)
        return
    elif target == "split_v":
        cmd = ["tmux", "split-window", "-v"]
        if clean_pane_id:
            cmd.extend(["-t", clean_pane_id])
        if cwd:
            cmd.extend(["-c", cwd])
        cmd.extend([shell_bin, "-lc", guarded_script])
        subprocess.Popen(cmd, start_new_session=True)
        return
    elif target == "window":
        window_title = title or "shell"
        cmd = ["tmux", "new-window", "-n", window_title]
        if cwd:
            cmd.extend(["-c", cwd])
        cmd.extend([shell_bin, "-lc", guarded_script])
        subprocess.Popen(cmd, start_new_session=True)
        return
    elif target == "popup":
        # tmux-menu runs inside a popup; replace current process with target app
        os.execvp(shell_bin, [shell_bin, "-lc", guarded_script])
    else:
        # Fallback to direct execution
        os.execvp(shell_bin, [shell_bin, "-lc", guarded_script])


def build_guarded_shell_script(cmd: str, title: str, persist_shell: bool) -> str:
    """Generate shell script with exit code trap, clipboard copy on error, and shell recovery."""
    escaped_title = shlex.quote(title or cmd)
    if persist_shell:
        return f"{cmd} ; exec \"${{SHELL:-/bin/zsh}}\""

    # Ephemeral mode with Interactive Error Guard
    return f"""
{cmd}
_status=$?

if [ "$_status" -ne 0 ]; then
  printf '\\n\\033[1;31m✖ Command failed with exit code %d: %s\\033[0m\\n' "$_status" {escaped_title}
  printf '\\033[1;33m[s]\\033[0m Debug shell   \\033[1;37m[any key]\\033[0m Close\\n'
  read -r _key 2>/dev/null || _key=""
  case "$_key" in
    s|S)
      printf '\\nDropping to debug shell...\\n'
      exec "${{SHELL:-/bin/zsh}}"
      ;;
    *)
      ;;
  esac
fi
exit "$_status"
"""


def main() -> None:
    parser = argparse.ArgumentParser(description="Tmux Leader Menu & Command Palette")
    parser.add_argument("pane_id", nargs="?", default="", help="Target tmux pane ID")
    parser.add_argument("-s", "--search", action="store_true", help="Start directly in Command Palette search mode")
    parser.add_argument("--config", type=Path, default=None, help="Path to config.json")
    args = parser.parse_args()

    pane_id = args.pane_id.strip("'\" \t\r\n") if args.pane_id else ""
    if (not pane_id or not pane_id.startswith("%")) and shutil.which("tmux"):
        try:
            pane_id = subprocess.check_output(["tmux", "display-message", "-p", "#{pane_id}"], text=True).strip()
        except Exception:
            pane_id = "%0"

    app = TmuxMenuApp(pane_id=pane_id, start_search=args.search, config_path=args.config)
    app.run()

    # Execute pending action cleanly after popup exit
    if app.pending_execution:
        action, target, popup_w, popup_h, title, persist_shell, original_target = app.pending_execution
        run_tmux_target(
            action=action,
            target=target,
            pane_id=pane_id,
            popup_w=popup_w,
            popup_h=popup_h,
            title=title,
            persist_shell=persist_shell,
            original_target=original_target,
        )


if __name__ == "__main__":
    main()
