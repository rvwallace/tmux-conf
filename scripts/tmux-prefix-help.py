#!/usr/bin/env -S uv run --script
# /// script
# requires-python = ">=3.10"
# dependencies = [
#     "textual>=0.80.0",
# ]
# ///
"""Interactive Tokyo Night Prefix Keybindings Browser for tmux."""

import argparse
import re
import shlex
import shutil
import subprocess
import sys
from dataclasses import dataclass
from typing import List, Optional

from textual.app import App, ComposeResult
from textual.binding import Binding
from textual.containers import Container, Horizontal, Vertical
from textual.events import Key
from textual.reactive import reactive
from textual.widgets import Input, Label, ListItem, ListView


THEME_CSS = """
Screen {
    background: transparent;
}

#root-container {
    width: 100%;
    height: 100%;
    background: #1a1b26;
    border: solid #414868;
}

#header-bar {
    dock: top;
    height: 2;
    background: #24283b;
    border-bottom: solid #414868;
    padding: 0 2;
    layout: horizontal;
    align-vertical: middle;
}

#header-title {
    width: 1fr;
    color: #7aa2f7;
    text-style: bold;
}

#header-count {
    width: auto;
    color: #7dcfff;
    text-style: bold;
}

#search-bar {
    dock: top;
    height: 3;
    background: #24283b;
    border-bottom: solid #414868;
    padding: 0 1;
    layout: horizontal;
    align-vertical: middle;
}

#search-icon {
    width: auto;
    color: #7dcfff;
    padding-right: 1;
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

#list-scroll {
    height: 1fr;
    padding: 0;
}

#binding-list {
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
BindingItemWidget.-highlight {
    background: #34446e;
    border-left: wide #7aa2f7;
}

.binding-key {
    width: auto;
    min-width: 16;
    background: #1f2335;
    color: #7dcfff;
    text-style: bold;
    padding: 0 1;
    margin-right: 1;
}

ListView > ListItem.-highlight .binding-key,
BindingItemWidget.-highlight .binding-key {
    background: #7aa2f7;
    color: #1a1b26;
}

.binding-category {
    width: auto;
    min-width: 12;
    color: #bb9af7;
    text-style: bold;
    padding-right: 1;
}

.binding-desc {
    width: 1fr;
    color: #c0caf5;
    text-style: bold;
}

ListView > ListItem.-highlight .binding-desc,
BindingItemWidget.-highlight .binding-desc {
    color: #ffffff;
}

.binding-cmd {
    width: auto;
    color: #565f89;
}

ListView > ListItem.-highlight .binding-cmd,
BindingItemWidget.-highlight .binding-cmd {
    color: #a9b1d6;
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
    text-style: bold;
}

#footer-hint {
    width: auto;
    color: #565f89;
}
"""


@dataclass
class KeyBinding:
    key: str
    combo: str
    desc: str
    cmd: str
    category: str

    @property
    def searchable_text(self) -> str:
        return f"{self.key} {self.combo} {self.desc} {self.cmd} {self.category}".lower()


def copy_to_clipboard(text: str) -> None:
    """Copy text to system clipboard and tmux buffer."""
    if shutil.which("pbcopy"):
        try:
            p = subprocess.Popen(["pbcopy"], stdin=subprocess.PIPE)
            p.communicate(text.encode("utf-8"))
        except Exception:
            pass
    elif shutil.which("wl-copy"):
        try:
            p = subprocess.Popen(["wl-copy"], stdin=subprocess.PIPE)
            p.communicate(text.encode("utf-8"))
        except Exception:
            pass
    elif shutil.which("xclip"):
        try:
            p = subprocess.Popen(["xclip", "-selection", "clipboard"], stdin=subprocess.PIPE)
            p.communicate(text.encode("utf-8"))
        except Exception:
            pass

    if shutil.which("tmux"):
        try:
            subprocess.run(["tmux", "set-buffer", text], check=False)
        except Exception:
            pass


def categorize_binding(key: str, desc: str, cmd: str) -> str:
    """Categorize keybinding based on key and action."""
    d = desc.lower()
    c = cmd.lower()

    if "leader" in d or "palette" in d or "menu" in d or key in ("Space", "P", "?"):
        return "Leader"
    if "pane" in d or "split" in d or "layout" in d or "zoom" in d or key in ("v", "s", "z", "x", "-", "|", "%", "\"", "o", ";", "q"):
        return "Panes"
    if "window" in d or key in ("c", "n", "p", "w", "&", ",", ".", "l", "f") or key.isdigit():
        return "Windows"
    if "session" in d or "client" in d or "detach" in d or key in ("d", "s", "$", "D", "L", "(", ")"):
        return "Sessions"
    if "buffer" in d or "copy" in d or "paste" in d or key in ("[", "]", "#", "="):
        return "Buffers"
    if "plugin" in d or "extrakto" in c or "snaglord" in c or "cowboy" in c or "fzf" in c or "yazi" in c or "file-picker" in c:
        return "Tools"
    return "System"


def load_prefix_bindings() -> List[KeyBinding]:
    """Parse active prefix keybindings from running tmux server."""
    if not shutil.which("tmux"):
        return []

    # Detect prefix key
    prefix_key = "C-a"
    try:
        raw_prefix = subprocess.check_output(["tmux", "show-option", "-gqv", "prefix"], text=True).strip()
        if raw_prefix:
            prefix_key = raw_prefix
    except Exception:
        pass

    # Read notes
    notes = {}
    try:
        notes_out = subprocess.check_output(["tmux", "list-keys", "-T", "prefix", "-N"], text=True)
        for line in notes_out.splitlines():
            line = line.strip()
            if not line:
                continue
            m = re.match(r"^(\S+\s+\S+)\s+(.+)$", line)
            if m:
                combo, desc = m.groups()
                k = combo.split()[-1]
                notes[k] = (combo, desc.strip())
    except Exception:
        pass

    # Read raw bindings
    bindings: List[KeyBinding] = []
    seen_keys = set()
    try:
        raw_out = subprocess.check_output(["tmux", "list-keys", "-T", "prefix"], text=True)
        for line in raw_out.splitlines():
            line = line.strip()
            if not line:
                continue
            tokens = line.split()
            if "prefix" in tokens:
                idx = tokens.index("prefix") + 1
                if idx < len(tokens) and tokens[idx] == "-N":
                    idx += 2
                if idx < len(tokens):
                    key = tokens[idx]
                    cmd = " ".join(tokens[idx + 1:])
                    combo, desc = notes.get(key, (f"{prefix_key} {key}", cmd))
                    if key not in seen_keys:
                        seen_keys.add(key)
                        cat = categorize_binding(key, desc, cmd)
                        bindings.append(KeyBinding(key=key, combo=combo, desc=desc, cmd=cmd, category=cat))
    except Exception:
        pass

    # Sort logically: Leader, Panes, Windows, Sessions, Buffers, Tools, System
    cat_order = {"Leader": 0, "Panes": 1, "Windows": 2, "Sessions": 3, "Buffers": 4, "Tools": 5, "System": 6}
    bindings.sort(key=lambda b: (cat_order.get(b.category, 99), b.key.lower()))
    return bindings


class BindingItemWidget(ListItem):
    """List item widget for a keybinding entry."""

    def __init__(self, binding: KeyBinding, initial_highlight: bool = False):
        classes = "-highlight" if initial_highlight else ""
        super().__init__(classes=classes)
        self.binding = binding

    def compose(self) -> ComposeResult:
        with Horizontal():
            yield Label(f"<{self.binding.combo}>", classes="binding-key", markup=False)
            yield Label(f"[{self.binding.category}]", classes="binding-category", markup=False)
            yield Label(self.binding.desc, classes="binding-desc", markup=False)
            # Truncate cmd preview if long
            cmd_preview = self.binding.cmd
            if len(cmd_preview) > 36:
                cmd_preview = cmd_preview[:33] + "..."
            yield Label(cmd_preview, classes="binding-cmd", markup=False)


class SearchInput(Input):
    """Search input field with list navigation shortcuts."""

    BINDINGS = [
        Binding("tab", "cursor_down", "Next", show=False, priority=True),
        Binding("shift+tab", "cursor_up", "Previous", show=False, priority=True),
        Binding("down", "cursor_down", "Next", show=False, priority=True),
        Binding("up", "cursor_up", "Previous", show=False, priority=True),
        Binding("ctrl+n", "cursor_down", "Next", show=False, priority=True),
        Binding("ctrl+p", "cursor_up", "Previous", show=False, priority=True),
        Binding("ctrl+j", "cursor_down", "Next", show=False, priority=True),
        Binding("ctrl+k", "cursor_up", "Previous", show=False, priority=True),
    ]

    def action_cursor_down(self) -> None:
        lv = self.app.query_one("#binding-list", ListView)
        if lv.children:
            if lv.index is None or lv.index < 0:
                lv.index = 0
            elif lv.index >= len(lv.children) - 1:
                lv.index = 0
            else:
                lv.action_cursor_down()

    def action_cursor_up(self) -> None:
        lv = self.app.query_one("#binding-list", ListView)
        if lv.children:
            if lv.index is None or lv.index <= 0:
                lv.index = len(lv.children) - 1
            else:
                lv.action_cursor_up()


class PrefixHelpApp(App[None]):
    """Textual app for interactive prefix keybindings browser."""

    CSS = THEME_CSS
    TITLE = "Prefix Keybindings"

    BINDINGS = [
        Binding("escape", "quit_app", "Close", show=False),
        Binding("q", "quit_app", "Close", show=False),
        Binding("y", "copy_selected", "Copy", show=False),
        Binding("c", "copy_selected", "Copy", show=False),
    ]

    search_query = reactive("")

    def __init__(self) -> None:
        super().__init__()
        self.bindings_data = load_prefix_bindings()
        self.filtered_bindings: List[KeyBinding] = list(self.bindings_data)
        self.pending_exec: Optional[str] = None

    def compose(self) -> ComposeResult:
        with Container(id="root-container"):
            with Horizontal(id="header-bar"):
                yield Label("󰋗 Active Prefix Keybindings", id="header-title", markup=False)
                yield Label(f"{len(self.bindings_data)} bindings", id="header-count", markup=False)

            with Horizontal(id="search-bar"):
                yield Label("󰍉", id="search-icon", markup=False)
                yield SearchInput(placeholder="Type to filter by key, description, category, or command...", id="search-input")

            with Vertical(id="list-scroll"):
                yield ListView(id="binding-list")

            with Horizontal(id="footer-bar"):
                yield Label("[Enter] Execute  [y/c] Copy Command  [Tab/↑/↓] Navigate  [Esc/q] Close", id="footer-legend", markup=False)
                yield Label("Tokyo Night", id="footer-hint", markup=False)

    def on_mount(self) -> None:
        self.refresh_list()
        self.query_one("#search-input", SearchInput).focus()

    def refresh_list(self) -> None:
        query = self.search_query.strip().lower()
        if not query:
            self.filtered_bindings = list(self.bindings_data)
        else:
            words = query.split()
            self.filtered_bindings = [
                b for b in self.bindings_data if all(w in b.searchable_text for w in words)
            ]

        self.query_one("#header-count", Label).update(f"{len(self.filtered_bindings)} / {len(self.bindings_data)} bindings")

        bl = self.query_one("#binding-list", ListView)
        bl.clear()

        for idx, b in enumerate(self.filtered_bindings):
            w = BindingItemWidget(b, initial_highlight=(idx == 0))
            if idx == 0:
                w.highlighted = True
            bl.append(w)

        self.call_after_refresh(self._sync_index)

    def _sync_index(self) -> None:
        bl = self.query_one("#binding-list", ListView)
        if bl.children:
            bl.index = 0

    def on_input_changed(self, event: Input.Changed) -> None:
        if event.input.id == "search-input":
            self.search_query = event.value
            self.refresh_list()

    def get_selected_binding(self) -> Optional[KeyBinding]:
        bl = self.query_one("#binding-list", ListView)
        if bl.highlighted_child is not None and isinstance(bl.highlighted_child, BindingItemWidget):
            return bl.highlighted_child.binding
        if self.filtered_bindings:
            return self.filtered_bindings[0]
        return None

    def on_input_submitted(self, event: Input.Submitted) -> None:
        self.execute_selected()

    def on_list_view_selected(self, event: ListView.Selected) -> None:
        if isinstance(event.item, BindingItemWidget):
            self.pending_exec = event.item.binding.cmd
            self.exit()

    def execute_selected(self) -> None:
        b = self.get_selected_binding()
        if b:
            self.pending_exec = b.cmd
            self.exit()

    def action_copy_selected(self) -> None:
        b = self.get_selected_binding()
        if b:
            copy_to_clipboard(f"{b.combo}: {b.desc} ({b.cmd})")
            self.notify(f"Copied {b.combo} to clipboard!", severity="information")

    def action_quit_app(self) -> None:
        self.exit()


def main() -> None:
    app = PrefixHelpApp()
    app.run()

    if app.pending_exec and shutil.which("tmux"):
        try:
            cmd = app.pending_exec.strip()
            # Map plain listing commands to their interactive tmux equivalents
            if cmd == "list-buffers":
                cmd = "choose-buffer -Z"
            elif cmd == "list-sessions":
                cmd = "choose-session"
            elif cmd == "list-windows":
                cmd = "choose-tree -w"

            # Defer execution slightly to let popup teardown complete and discard key events
            script = f"sleep 0.15 ; tmux {cmd}"
            subprocess.Popen(["sh", "-c", script], start_new_session=True)
        except Exception:
            pass


if __name__ == "__main__":
    main()
