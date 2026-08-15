#!/usr/bin/env -S uv run --script
# /// script
# requires-python = ">=3.10"
# dependencies = [
#     "textual>=0.80.0",
# ]
# ///
"""Textual TUI interface for tmux aichat actions."""

import asyncio
import os
import shutil
import subprocess
import sys
from pathlib import Path
from typing import Optional

from textual.app import App, ComposeResult
from textual.binding import Binding
from textual.containers import Container, Horizontal, Vertical, VerticalScroll
from textual.events import Key, MouseUp
from textual.message import Message
from textual.screen import ModalScreen
from textual.widgets import Markdown, Static, TextArea

MODE_TITLES = {
    "ask": "AI: Ask",
    "error": "AI: Diagnose Error",
    "fix": "AI: Suggest Fix",
    "summarize": "AI: Summarize Pane",
    "command": "AI: Generate Command",
    "explain": "AI: Explain",
    "explain-copy": "AI: Explain Last Copy",
}

INPUT_PLACEHOLDERS = {
    "ask": "Ask a question... (Shift+Enter for newline, /refresh for context)",
    "error": "Ask a follow-up question about this error... (/refresh for context)",
    "summarize": "Ask a follow-up question about this summary... (/refresh for context)",
    "explain": "Enter command, snippet, or topic to explain...",
    "explain-copy": "Ask a follow-up question about the copied snippet...",
    "command": "Describe the command to generate...",
    "fix": "Refine fix (e.g. 'use brew', 'different flag')... or press Enter/s to send",
}

SCROLLBACK_DEPTHS = ["100", "200", "500", "1000", "all"]

THEME_CSS = """
Screen {
    background: #1a1b26;
    color: #c0caf5;
    layout: vertical;
}

#header-container {
    dock: top;
    height: auto;
    background: #24283b;
    border-bottom: solid #414868;
    padding: 0 1;
}

#header-top-bar {
    height: 1;
    width: 100%;
}

#header-title {
    color: #7aa2f7;
    text-style: bold;
    width: 1fr;
}

#header-depth-badge {
    color: #7dcfff;
    text-style: bold;
    margin-right: 1;
    width: auto;
}

#header-badge {
    color: #bb9af7;
    text-style: bold;
    width: auto;
}

#header-path {
    height: 1;
    color: #7dcfff;
    text-overflow: ellipsis;
}

#transcript-container {
    height: 1fr;
    background: #1a1b26;
    padding: 0 1;
}

#transcript-container:focus {
    border: tall #7aa2f7;
}

.message-card {
    margin: 1 0;
    padding: 0 1;
    background: #24283b;
    border-left: solid #414868;
    height: auto;
}

.user-card {
    border-left: solid #7aa2f7;
}

.assistant-card {
    border-left: solid #bb9af7;
}

.system-card {
    border-left: solid #e0af68;
}

.error-card {
    border-left: solid #f7768e;
}

.success-card {
    border-left: solid #9ece6a;
}

.candidate-card {
    margin: 1 0;
    padding: 1;
    background: #24283b;
    border: tall #7aa2f7;
    height: auto;
}

.candidate-title {
    color: #7aa2f7;
    text-style: bold;
    margin-bottom: 0;
}

.candidate-command {
    color: #9ece6a;
    text-style: bold;
    background: #1a1b26;
    padding: 0 1;
    height: auto;
    margin-top: 1;
    border-left: solid #9ece6a;
}

.role-badge {
    height: 1;
    text-style: bold;
}

.user-role {
    color: #7aa2f7;
}

.assistant-role {
    color: #bb9af7;
}

.system-role {
    color: #e0af68;
}

.error-role {
    color: #f7768e;
}

.success-role {
    color: #9ece6a;
}

.message-text {
    color: #c0caf5;
    height: auto;
}

.error-text {
    color: #f7768e;
    height: auto;
}

.success-text {
    color: #9ece6a;
    height: auto;
}

.loading-text {
    color: #7dcfff;
    text-style: bold;
    height: auto;
}

.message-markdown {
    color: #c0caf5;
    height: auto;
    background: transparent;
    padding: 0;
    margin: 0;
}

#input-container {
    dock: bottom;
    height: auto;
    background: #1a1b26;
    padding: 0 1;
}

ChatTextArea {
    background: #24283b;
    color: #c0caf5;
    border: tall #414868;
    height: 4;
}

ChatTextArea:focus {
    border: tall #7aa2f7;
}

.compact-input ChatTextArea {
    height: 3;
}

#footer-container {
    dock: bottom;
    height: 1;
    background: #1f2335;
    color: #565f89;
    padding: 0 1;
}

/* Help Modal Dialog */
#help-dialog {
    padding: 1 2;
    background: #24283b;
    border: thick #7aa2f7;
    width: 74;
    height: auto;
    align: center middle;
}

#help-title {
    color: #7aa2f7;
    text-style: bold;
    text-align: center;
    margin-bottom: 1;
}

#help-content {
    height: auto;
    margin-bottom: 1;
}

.help-row {
    height: 1;
    margin: 0;
}

.help-key {
    color: #7dcfff;
    text-style: bold;
    width: 28;
}

.help-desc {
    color: #c0caf5;
    width: 1fr;
}

#help-footer {
    color: #565f89;
    text-align: center;
    text-style: italic;
}
"""


def find_ai_assist_script() -> Path:
    """Find the path to the ai-assist.sh backend script."""
    script_dir = Path(__file__).resolve().parent
    local_path = script_dir / "ai-assist.sh"
    if local_path.is_file():
        return local_path
    config_path = Path.home() / ".config" / "tmux" / "scripts" / "ai-assist.sh"
    if config_path.is_file():
        return config_path
    return local_path


def copy_text_to_clipboard(text: str) -> bool:
    """Copy text to system clipboard (pbcopy, wl-copy, xclip) and tmux buffer."""
    if not text:
        return False

    copied = False

    # 1. System clipboard
    if shutil.which("pbcopy"):
        try:
            proc = subprocess.Popen(["pbcopy"], stdin=subprocess.PIPE)
            proc.communicate(input=text.encode("utf-8"))
            if proc.returncode == 0:
                copied = True
        except Exception:
            pass
    elif shutil.which("wl-copy"):
        try:
            proc = subprocess.Popen(["wl-copy"], stdin=subprocess.PIPE)
            proc.communicate(input=text.encode("utf-8"))
            if proc.returncode == 0:
                copied = True
        except Exception:
            pass
    elif shutil.which("xclip"):
        try:
            proc = subprocess.Popen(["xclip", "-selection", "clipboard"], stdin=subprocess.PIPE)
            proc.communicate(input=text.encode("utf-8"))
            if proc.returncode == 0:
                copied = True
        except Exception:
            pass

    # 2. tmux buffer
    if shutil.which("tmux"):
        try:
            subprocess.run(["tmux", "set-buffer", "--", text], check=False)
            copied = True
        except Exception:
            pass

    return copied


def get_pane_info(pane_id: str) -> dict:
    """Fetch pane metadata (working directory, foreground command) from tmux."""
    info = {"path": "", "command": ""}
    if not shutil.which("tmux"):
        return info
    try:
        res = subprocess.run(
            ["tmux", "display-message", "-p", "-t", pane_id, "#{pane_current_path}\t#{pane_current_command}"],
            capture_output=True,
            text=True,
            check=False,
        )
        if res.returncode == 0:
            parts = res.stdout.strip().split("\t")
            if len(parts) >= 1:
                info["path"] = parts[0]
            if len(parts) >= 2:
                info["command"] = parts[1]
    except Exception:
        pass
    return info


def format_path(path: str) -> str:
    """Format an absolute path with ~ if it resides inside $HOME."""
    if not path:
        return "~"
    home = str(Path.home())
    if path == home:
        return "~"
    if path.startswith(home + "/"):
        return "~" + path[len(home):]
    return path


class ChatTextArea(TextArea):
    """Custom multiline TextArea supporting Enter to submit, Shift+Enter for newlines, and Tab navigation."""

    class Submitted(Message):
        """Event posted when the user submits the input content."""

        def __init__(self, text: str):
            super().__init__()
            self.text = text

    async def _on_key(self, event: Key) -> None:
        if event.key == "enter":
            event.prevent_default()
            event.stop()
            self.post_message(self.Submitted(self.text))
            return
        elif event.key in ("shift+enter", "alt+enter", "ctrl+j"):
            event.prevent_default()
            event.stop()
            self.insert("\n")
            return
        elif event.key in ("tab", "shift+tab"):
            event.prevent_default()
            event.stop()
            self.app.action_switch_focus()
            return
        await super()._on_key(event)


class HelpScreen(ModalScreen):
    """Modal help screen showing keybindings for the current mode."""

    BINDINGS = [
        Binding("escape", "dismiss", "Close", show=False),
        Binding("q", "dismiss", "Close", show=False),
        Binding("question_mark", "dismiss", "Close", show=False),
        Binding("enter", "dismiss", "Close", show=False),
        Binding("space", "dismiss", "Close", show=False),
    ]

    def __init__(self, mode: str, scrollback_depth: str = "200"):
        super().__init__()
        self.mode = mode
        self.scrollback_depth = scrollback_depth

    def compose(self) -> ComposeResult:
        with Container(id="help-dialog"):
            yield Static(f"Help — {MODE_TITLES.get(self.mode, 'AI Assistant')}", id="help-title")
            with Vertical(id="help-content"):
                if self.mode in ("command", "fix"):
                    yield Horizontal(Static("Enter", classes="help-key"), Static("Send to pane (empty input) / Refine (typed text)", classes="help-desc"), classes="help-row")
                    yield Horizontal(Static("s / i", classes="help-key"), Static("Send & insert candidate command to pane", classes="help-desc"), classes="help-row")
                    yield Horizontal(Static("y / c", classes="help-key"), Static("Copy candidate command to clipboard", classes="help-desc"), classes="help-row")
                else:
                    yield Horizontal(Static("Enter", classes="help-key"), Static("Send prompt / follow-up question", classes="help-desc"), classes="help-row")
                    yield Horizontal(Static("Shift+Enter / Alt+Enter", classes="help-key"), Static("Insert newline (multiline)", classes="help-desc"), classes="help-row")
                    yield Horizontal(Static("Tab / Shift+Tab", classes="help-key"), Static("Switch focus (Input ↔ Transcript)", classes="help-desc"), classes="help-row")
                    yield Horizontal(Static("/refresh", classes="help-key"), Static("Refresh pane context in session", classes="help-desc"), classes="help-row")
                    yield Horizontal(Static("Ctrl+L", classes="help-key"), Static("Clear transcript display", classes="help-desc"), classes="help-row")
                if self.mode == "summarize":
                    yield Horizontal(Static("1 / 2 / 3 / 4 / 5", classes="help-key"), Static("Set depth (100 / 200 / 500 / 1k / all) & reload", classes="help-desc"), classes="help-row")
                    yield Horizontal(Static("d", classes="help-key"), Static("Cycle scrollback depth", classes="help-desc"), classes="help-row")
                    yield Horizontal(Static("r", classes="help-key"), Static("Reload summary", classes="help-desc"), classes="help-row")
                if self.mode not in ("command", "fix"):
                    yield Horizontal(Static("y / c / Mouse Select", classes="help-key"), Static("Copy selected text to clipboard", classes="help-desc"), classes="help-row")
                    yield Horizontal(Static("j / k or ↓ / ↑", classes="help-key"), Static("Scroll transcript", classes="help-desc"), classes="help-row")
                    yield Horizontal(Static("PgUp / PgDn", classes="help-key"), Static("Page up / page down", classes="help-desc"), classes="help-row")
                yield Horizontal(Static("Esc", classes="help-key"), Static("Cancel / close in all modes", classes="help-desc"), classes="help-row")
                yield Horizontal(Static("q", classes="help-key"), Static("Cancel / quit (when input is unfocused)", classes="help-desc"), classes="help-row")
                yield Horizontal(Static("?", classes="help-key"), Static("Toggle this help screen", classes="help-desc"), classes="help-row")
            yield Static("Press Esc, q, or Enter to close", id="help-footer")


class AiChatTuiApp(App):
    """Main Textual TUI Application for tmux aichat."""

    CSS = THEME_CSS

    BINDINGS = [
        Binding("tab", "switch_focus", "Focus", show=False),
        Binding("shift+tab", "switch_focus_reverse", "Focus", show=False),
        Binding("ctrl+l", "clear_transcript", "Clear", show=False),
        Binding("escape", "quit_app", "Quit", show=False),
        Binding("q", "quit_if_not_input", "Quit", show=False),
        Binding("y", "copy_selection_if_not_input", "Copy", show=False),
        Binding("c", "copy_selection_if_not_input", "Copy", show=False),
        Binding("s", "send_command_if_not_input", "Send Command", show=False),
        Binding("i", "send_command_if_not_input", "Send Command", show=False),
        Binding("d", "cycle_depth_if_not_input", "Cycle Depth", show=False),
        Binding("r", "reload_if_not_input", "Reload", show=False),
        Binding("1", "set_depth_1_if_not_input", "100 lines", show=False),
        Binding("2", "set_depth_2_if_not_input", "200 lines", show=False),
        Binding("3", "set_depth_3_if_not_input", "500 lines", show=False),
        Binding("4", "set_depth_4_if_not_input", "1000 lines", show=False),
        Binding("5", "set_depth_5_if_not_input", "all lines", show=False),
        Binding("question_mark", "toggle_help_if_not_input", "Help", show=False),
        Binding("j", "scroll_down", "Scroll Down", show=False),
        Binding("k", "scroll_up", "Scroll Up", show=False),
        Binding("down", "scroll_down", "Scroll Down", show=False),
        Binding("up", "scroll_up", "Scroll Up", show=False),
        Binding("pageup", "page_up", "Page Up", show=False),
        Binding("pagedown", "page_down", "Page Down", show=False),
        Binding("home", "scroll_home", "Top", show=False),
        Binding("end", "scroll_end", "Bottom", show=False),
    ]

    def __init__(self, mode: str, pane_id: str):
        super().__init__()
        self.mode = mode
        self.pane_id = pane_id
        self.session_id = f"tmux-ai-{mode}-{pane_id.lstrip('%')}-{os.getpid()}"
        self.turn_count = 0
        self.is_busy = False
        self.pane_info = get_pane_info(pane_id)
        self.latest_assistant_response: Optional[str] = None
        self.scrollback_depth = "200"
        self.current_candidate_command: Optional[str] = None
        self.original_prompt: Optional[str] = None

    def compose(self) -> ComposeResult:
        # Header
        with Container(id="header-container"):
            with Horizontal(id="header-top-bar"):
                yield Static(MODE_TITLES.get(self.mode, "AI Assistant"), id="header-title")
                if self.mode == "summarize":
                    yield Static(f"[{self.scrollback_depth} lines]", id="header-depth-badge")
                yield Static("aichat", id="header-badge")
            formatted_path = format_path(self.pane_info.get("path", ""))
            yield Static(formatted_path, id="header-path")

        # Transcript Area
        yield VerticalScroll(id="transcript-container", can_focus=True)

        # Input Area (all modes support interactive input / follow-ups / refinements)
        classes = "compact-input" if self.mode in ("command", "fix") else ""
        with Container(id="input-container", classes=classes):
            placeholder = INPUT_PLACEHOLDERS.get(self.mode, "Type your prompt...")
            yield ChatTextArea(id="user-input", placeholder=placeholder, show_line_numbers=False)

        # Footer
        yield Static(self._build_footer_markup(), id="footer-container")

    def _build_footer_markup(self) -> str:
        """Construct the footer key hints depending on the mode."""
        if self.mode in ("command", "fix"):
            if self.current_candidate_command:
                return "[bold #7dcfff]Enter[/bold #7dcfff] send / refine  [bold #414868]•[/bold #414868]  [bold #7dcfff]s[/bold #7dcfff] send  [bold #414868]•[/bold #414868]  [bold #7dcfff]y/c[/bold #7dcfff] copy  [bold #414868]•[/bold #414868]  [bold #7dcfff]Esc[/bold #7dcfff] cancel  [bold #414868]•[/bold #414868]  [bold #7dcfff]?[/bold #7dcfff] help"
            else:
                return "[bold #7dcfff]Enter[/bold #7dcfff] generate  [bold #414868]•[/bold #414868]  [bold #7dcfff]Esc[/bold #7dcfff] cancel  [bold #414868]•[/bold #414868]  [bold #7dcfff]?[/bold #7dcfff] help"
        elif self.mode == "summarize":
            return "[bold #7dcfff]1-5[/bold #7dcfff] depth (100-all)  [bold #414868]•[/bold #414868]  [bold #7dcfff]d[/bold #7dcfff] cycle  [bold #414868]•[/bold #414868]  [bold #7dcfff]Enter[/bold #7dcfff] follow-up  [bold #414868]•[/bold #414868]  [bold #7dcfff]Tab[/bold #7dcfff] focus  [bold #414868]•[/bold #414868]  [bold #7dcfff]y/c[/bold #7dcfff] copy  [bold #414868]•[/bold #414868]  [bold #7dcfff]Esc[/bold #7dcfff] quit  [bold #414868]•[/bold #414868]  [bold #7dcfff]?[/bold #7dcfff] help"
        else:
            return "[bold #7dcfff]Enter[/bold #7dcfff] send / follow-up  [bold #414868]•[/bold #414868]  [bold #7dcfff]Shift+Enter[/bold #7dcfff] newline  [bold #414868]•[/bold #414868]  [bold #7dcfff]/refresh[/bold #7dcfff] context  [bold #414868]•[/bold #414868]  [bold #7dcfff]Tab[/bold #7dcfff] focus  [bold #414868]•[/bold #414868]  [bold #7dcfff]y/c[/bold #7dcfff] copy  [bold #414868]•[/bold #414868]  [bold #7dcfff]Esc[/bold #7dcfff] quit  [bold #414868]•[/bold #414868]  [bold #7dcfff]?[/bold #7dcfff] help"

    async def on_mount(self) -> None:
        """Handle automatic startup tasks after UI mount."""
        # Focus input widget initially for modes that require initial user prompt
        if self.mode in ("ask", "explain", "command"):
            try:
                self.query_one("#user-input", ChatTextArea).focus()
            except Exception:
                pass

        # Check for required tools upfront
        if not shutil.which("aichat"):
            self._add_error("aichat is not in PATH. Please install aichat to use AI features.")
            return

        # Auto-run modes that require no initial user input
        if self.mode in ("error", "summarize", "explain-copy", "fix"):
            asyncio.create_task(self._run_backend_query())

    def on_mouse_up(self, event: MouseUp) -> None:
        """Automatically copy text to clipboard when mouse selection is released."""
        sel = self.screen.get_selected_text()
        if sel:
            copy_text_to_clipboard(sel)
            self.copy_to_clipboard(sel)
            self.notify("Copied selection to clipboard", timeout=1.5)

    async def on_chat_text_area_submitted(self, event: ChatTextArea.Submitted) -> None:
        """Handle user input submission from ChatTextArea."""
        if self.is_busy:
            return

        text = event.text.strip()

        # In command / fix mode: if candidate is ready and user presses enter on empty input, insert into pane!
        if self.mode in ("command", "fix") and not text and self.current_candidate_command:
            self._insert_candidate_and_exit()
            return

        if not text:
            return

        input_widget = self.query_one("#user-input", ChatTextArea)
        input_widget.text = ""

        # Run query / refinement
        asyncio.create_task(self._run_backend_query(text))

    def _insert_candidate_and_exit(self) -> None:
        """Insert candidate command into target pane via bracketed paste and close."""
        if not self.current_candidate_command:
            return
        ai_assist_path = find_ai_assist_script()
        env = os.environ.copy()
        env["TMUX_AI_ASSIST_NO_PAUSE"] = "1"
        try:
            subprocess.run(
                [str(ai_assist_path), "insert-command", self.pane_id, self.current_candidate_command],
                env=env,
                stdin=subprocess.DEVNULL,
                stdout=subprocess.PIPE,
                stderr=subprocess.PIPE,
                check=False,
                timeout=5,
            )
        except Exception:
            pass
        self.exit()

    async def _run_backend_query(self, prompt_text: str = "") -> None:
        """Execute ai-assist.sh asynchronously and update the UI."""
        if self.is_busy:
            return
        self.is_busy = True

        input_widget: Optional[ChatTextArea] = None
        try:
            input_widget = self.query_one("#user-input", ChatTextArea)
            input_widget.disabled = True
        except Exception:
            pass

        transcript = self.query_one("#transcript-container", VerticalScroll)

        # Append user message or refinement if applicable
        if prompt_text:
            if self.mode in ("command", "fix") and self.current_candidate_command:
                await transcript.mount(Container(
                    Static("Refinement", classes="role-badge user-role"),
                    Static(prompt_text, classes="message-text"),
                    classes="message-card user-card"
                ))
            else:
                if prompt_text == "/refresh":
                    await transcript.mount(Container(
                        Static("You", classes="role-badge user-role"),
                        Static("/refresh", classes="message-text"),
                        classes="message-card user-card"
                    ))
                else:
                    await transcript.mount(Container(
                        Static("You", classes="role-badge user-role"),
                        Static(prompt_text, classes="message-text"),
                        classes="message-card user-card"
                    ))

        # Mount loading indicator
        loading_label = "● Generating..."
        if self.mode == "error":
            loading_label = "● Diagnosing error..." if self.turn_count == 0 else "● Thinking..."
        elif self.mode == "summarize":
            if self.turn_count == 0:
                depth_desc = "all" if self.scrollback_depth == "all" else f"{self.scrollback_depth} lines"
                loading_label = f"● Summarizing pane ({depth_desc})..."
            else:
                loading_label = "● Thinking..."
        elif self.mode == "explain" or self.mode == "explain-copy":
            loading_label = "● Explaining..." if self.turn_count == 0 else "● Thinking..."
        elif self.mode == "fix":
            loading_label = "● Suggesting fix..." if not self.current_candidate_command else "● Refining fix..."
        elif self.mode == "command":
            loading_label = "● Generating command..." if not self.current_candidate_command else "● Refining command..."
        elif prompt_text == "/refresh":
            loading_label = "● Refreshing pane context..."

        loading_box = Container(
            Static(loading_label, classes="loading-text"),
            id="loading-indicator",
            classes="message-card"
        )
        await transcript.mount(loading_box)
        transcript.scroll_end(animate=False)

        # Prepare backend execution
        ai_assist_path = find_ai_assist_script()
        env = os.environ.copy()
        env["TMUX_AI_ASSIST_NO_PAUSE"] = "1"
        env["TMUX_AI_ASSIST_SCROLLBACK"] = str(self.scrollback_depth)

        args = [str(ai_assist_path), self.mode, self.pane_id]

        if self.mode in ("command", "fix"):
            env["TMUX_AI_ASSIST_PRINT_ONLY"] = "1"
            if self.current_candidate_command and prompt_text:
                # Refinement mode
                env["TMUX_AI_ASSIST_REFINE"] = "1"
                env["TMUX_AI_ASSIST_ORIGINAL_PROMPT"] = self.original_prompt or self.current_candidate_command
                env["TMUX_AI_ASSIST_PREVIOUS_COMMAND"] = self.current_candidate_command
                args.append(prompt_text)
            else:
                if prompt_text:
                    self.original_prompt = prompt_text
                    args.append(prompt_text)
        else:
            # Conversational analysis modes: ask, error, summarize, explain, explain-copy
            env["TMUX_AI_ASSIST_SESSION"] = self.session_id
            if prompt_text == "/refresh":
                env["TMUX_AI_ASSIST_REFRESH"] = "1"
                args.append(prompt_text)
            elif self.turn_count > 0:
                env["TMUX_AI_ASSIST_FOLLOW_UP"] = "1"
                if prompt_text:
                    args.append(prompt_text)
            elif prompt_text:
                args.append(prompt_text)

        # Execute subprocess
        try:
            proc = await asyncio.create_subprocess_exec(
                *args,
                env=env,
                stdout=asyncio.subprocess.PIPE,
                stderr=asyncio.subprocess.PIPE
            )
            stdout_bytes, stderr_bytes = await proc.communicate()
            stdout_str = stdout_bytes.decode("utf-8", errors="replace").strip()
            stderr_str = stderr_bytes.decode("utf-8", errors="replace").strip()
            returncode = proc.returncode
        except Exception as e:
            returncode = 1
            stdout_str = ""
            stderr_str = f"Failed to execute backend: {e}"

        # Remove loading indicator
        try:
            await loading_box.remove()
        except Exception:
            pass

        # Handle result
        if returncode == 0:
            if self.mode in ("command", "fix"):
                # Clean candidate command received
                self.current_candidate_command = stdout_str
                self.latest_assistant_response = stdout_str
                
                # Mount prominent candidate command box
                await transcript.mount(Container(
                    Static("Candidate Command (press Enter / s to send to pane, or type to refine):", classes="candidate-title"),
                    Static(stdout_str, classes="candidate-command"),
                    classes="candidate-card"
                ))
                
                # Update placeholder
                if input_widget is not None:
                    input_widget.placeholder = "Refine command (e.g. 'exclude node_modules')... or Enter/s to send"
                
                # Refresh footer
                try:
                    footer = self.query_one("#footer-container", Static)
                    footer.update(self._build_footer_markup())
                except Exception:
                    pass
                self.turn_count += 1
            elif prompt_text == "/refresh":
                response_text = stdout_str or "Refreshed context acknowledged."
                self.latest_assistant_response = response_text
                await transcript.mount(Container(
                    Static("System", classes="role-badge system-role"),
                    Static(f"Pane context refreshed.\n{response_text}", classes="message-text"),
                    classes="message-card system-card"
                ))
                self.turn_count += 1
            else:
                # Normal assistant response
                self.latest_assistant_response = stdout_str or ""
                await transcript.mount(Container(
                    Static("Assistant", classes="role-badge assistant-role"),
                    Markdown(stdout_str or "(No output received)", classes="message-markdown"),
                    classes="message-card assistant-card"
                ))
                self.turn_count += 1
        else:
            # Failure
            error_message = stderr_str or stdout_str or "Unknown error occurred"
            if error_message.startswith("tmux ai-assist: "):
                error_message = error_message[len("tmux ai-assist: "):]
            self._add_error(error_message)

        transcript.scroll_end(animate=False)

        # Re-enable input
        self.is_busy = False
        if input_widget is not None:
            input_widget.disabled = False
            if self.mode in ("ask", "explain", "command") or self.turn_count > 1:
                input_widget.focus()

    def _add_error(self, message: str) -> None:
        """Helper to append an error block to the transcript."""
        transcript = self.query_one("#transcript-container", VerticalScroll)
        transcript.mount(Container(
            Static("Error", classes="role-badge error-role"),
            Static(message, classes="error-text"),
            classes="message-card error-card"
        ))
        transcript.scroll_end(animate=False)

    def _update_depth_and_reload(self, depth: str) -> None:
        """Update scrollback depth for summarize mode and reload."""
        if self.mode != "summarize" or self.is_busy:
            return
        self.scrollback_depth = depth
        try:
            badge = self.query_one("#header-depth-badge", Static)
            badge.update(f"[{depth} lines]")
        except Exception:
            pass
        try:
            transcript = self.query_one("#transcript-container", VerticalScroll)
            transcript.remove_children()
        except Exception:
            pass
        self.turn_count = 0
        self.notify(f"Set scrollback to {depth} lines", timeout=1.2)
        asyncio.create_task(self._run_backend_query())

    # Action implementations for bindings
    def action_send_command_if_not_input(self) -> None:
        """Send candidate command to pane if available."""
        if self.mode in ("command", "fix") and self.current_candidate_command:
            self._insert_candidate_and_exit()

    def action_set_depth_1_if_not_input(self) -> None:
        if not isinstance(self.focused, TextArea):
            self._update_depth_and_reload("100")

    def action_set_depth_2_if_not_input(self) -> None:
        if not isinstance(self.focused, TextArea):
            self._update_depth_and_reload("200")

    def action_set_depth_3_if_not_input(self) -> None:
        if not isinstance(self.focused, TextArea):
            self._update_depth_and_reload("500")

    def action_set_depth_4_if_not_input(self) -> None:
        if not isinstance(self.focused, TextArea):
            self._update_depth_and_reload("1000")

    def action_set_depth_5_if_not_input(self) -> None:
        if not isinstance(self.focused, TextArea):
            self._update_depth_and_reload("all")

    def action_cycle_depth_if_not_input(self) -> None:
        if not isinstance(self.focused, TextArea) and self.mode == "summarize":
            try:
                curr_idx = SCROLLBACK_DEPTHS.index(self.scrollback_depth)
                next_depth = SCROLLBACK_DEPTHS[(curr_idx + 1) % len(SCROLLBACK_DEPTHS)]
            except ValueError:
                next_depth = "200"
            self._update_depth_and_reload(next_depth)

    def action_reload_if_not_input(self) -> None:
        if not isinstance(self.focused, TextArea) and self.mode == "summarize" and not self.is_busy:
            try:
                transcript = self.query_one("#transcript-container", VerticalScroll)
                transcript.remove_children()
            except Exception:
                pass
            self.turn_count = 0
            asyncio.create_task(self._run_backend_query())

    def action_switch_focus(self) -> None:
        try:
            inp = self.query_one("#user-input", ChatTextArea)
            scroll = self.query_one("#transcript-container", VerticalScroll)
            if self.focused == inp:
                scroll.focus()
            else:
                inp.focus()
        except Exception:
            pass

    def action_switch_focus_reverse(self) -> None:
        self.action_switch_focus()

    def action_clear_transcript(self) -> None:
        try:
            scroll = self.query_one("#transcript-container", VerticalScroll)
            scroll.remove_children()
            if self.mode in ("ask", "error", "summarize", "explain", "explain-copy"):
                scroll.mount(Container(
                    Static("System", classes="role-badge system-role"),
                    Static("Transcript display cleared. Active aichat session retained.", classes="message-text"),
                    classes="message-card system-card"
                ))
        except Exception:
            pass

    def action_copy_selection_if_not_input(self) -> None:
        """Copy selected text or fallback to candidate command / latest assistant response."""
        if isinstance(self.focused, TextArea):
            if self.focused.selected_text:
                copy_text_to_clipboard(self.focused.selected_text)
                self.copy_to_clipboard(self.focused.selected_text)
                self.notify("Copied selection to clipboard", timeout=1.5)
            return

        # 1. Check screen text selection
        sel = self.screen.get_selected_text()
        if sel:
            copy_text_to_clipboard(sel)
            self.copy_to_clipboard(sel)
            self.notify("Copied selection to clipboard", timeout=1.5)
            return

        # 2. Check candidate command in command/fix mode
        if self.mode in ("command", "fix") and self.current_candidate_command:
            copy_text_to_clipboard(self.current_candidate_command)
            self.copy_to_clipboard(self.current_candidate_command)
            self.notify("Copied command to clipboard", timeout=1.5)
            return

        # 3. Fallback to latest assistant response
        if self.latest_assistant_response:
            copy_text_to_clipboard(self.latest_assistant_response)
            self.copy_to_clipboard(self.latest_assistant_response)
            self.notify("Copied latest response to clipboard", timeout=1.5)

    def action_quit_app(self) -> None:
        if len(self.screen_stack) > 1:
            self.pop_screen()
            return
        self.exit()

    def action_quit_if_not_input(self) -> None:
        if isinstance(self.focused, TextArea):
            return
        if len(self.screen_stack) > 1:
            self.pop_screen()
            return
        self.exit()

    def action_toggle_help_if_not_input(self) -> None:
        if isinstance(self.focused, TextArea):
            return
        if isinstance(self.screen, HelpScreen):
            self.pop_screen()
        else:
            self.push_screen(HelpScreen(mode=self.mode, scrollback_depth=self.scrollback_depth))

    def action_scroll_down(self) -> None:
        if isinstance(self.focused, TextArea):
            return
        try:
            scroll = self.query_one("#transcript-container", VerticalScroll)
            scroll.scroll_relative(y=2, animate=False)
        except Exception:
            pass

    def action_scroll_up(self) -> None:
        if isinstance(self.focused, TextArea):
            return
        try:
            scroll = self.query_one("#transcript-container", VerticalScroll)
            scroll.scroll_relative(y=-2, animate=False)
        except Exception:
            pass

    def action_page_down(self) -> None:
        if isinstance(self.focused, TextArea):
            return
        try:
            scroll = self.query_one("#transcript-container", VerticalScroll)
            scroll.scroll_page_down(animate=False)
        except Exception:
            pass

    def action_page_up(self) -> None:
        if isinstance(self.focused, TextArea):
            return
        try:
            scroll = self.query_one("#transcript-container", VerticalScroll)
            scroll.scroll_page_up(animate=False)
        except Exception:
            pass

    def action_scroll_home(self) -> None:
        if isinstance(self.focused, TextArea):
            return
        try:
            scroll = self.query_one("#transcript-container", VerticalScroll)
            scroll.scroll_home(animate=False)
        except Exception:
            pass

    def action_scroll_end(self) -> None:
        if isinstance(self.focused, TextArea):
            return
        try:
            scroll = self.query_one("#transcript-container", VerticalScroll)
            scroll.scroll_end(animate=False)
        except Exception:
            pass


def print_usage(file=sys.stderr) -> None:
    """Print CLI usage information."""
    print(
        "Usage:\n"
        "  aichat-tui.py ask <pane_id>\n"
        "  aichat-tui.py error <pane_id>\n"
        "  aichat-tui.py fix <pane_id>\n"
        "  aichat-tui.py summarize <pane_id>\n"
        "  aichat-tui.py command <pane_id>\n"
        "  aichat-tui.py explain <pane_id>\n"
        "  aichat-tui.py explain-copy <pane_id>",
        file=file,
    )


def main() -> None:
    """CLI entrypoint."""
    args = sys.argv[1:]
    if not args:
        print_usage(sys.stderr)
        sys.exit(2)

    mode = args[0]
    if mode in ("-h", "--help", "help"):
        print_usage(sys.stdout)
        sys.exit(0)

    if mode not in MODE_TITLES:
        print_usage(sys.stderr)
        sys.exit(2)

    if len(args) < 2 or not args[1]:
        print("tmux aichat-tui: missing pane_id", file=sys.stderr)
        sys.exit(1)

    pane_id = args[1]

    # Check that tmux is available
    if not shutil.which("tmux"):
        print("tmux aichat-tui: tmux is not in PATH", file=sys.stderr)
        sys.exit(1)

    # Validate that the pane exists
    res = subprocess.run(["tmux", "display-message", "-p", "-t", pane_id, "#{pane_id}"], capture_output=True, text=True)
    if res.returncode != 0 or not res.stdout.strip():
        print(f"tmux aichat-tui: pane not found: {pane_id}", file=sys.stderr)
        sys.exit(1)

    app = AiChatTuiApp(mode=mode, pane_id=pane_id)
    app.run()


if __name__ == "__main__":
    main()
