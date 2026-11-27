#!/usr/bin/env python3
"""Browse command/output blocks from tmux scrollback or a saved capture."""
# Adapted from rex-scrollback-tool; see tmux-history.LICENSE.

import argparse
import json
import os
from pathlib import Path
import re
import shlex
import shutil
import subprocess
import sys
import tempfile


def parse_blocks(text, prompt=r"^CMD>\s?", prompt_lines=3):
    """Parse prompt-marked commands and the output that follows each command."""
    lines = text.splitlines()
    pattern = re.compile(prompt)
    marks = [(i, pattern.match(line)) for i, line in enumerate(lines)]
    marks = [(i, match) for i, match in marks if match]
    blocks = []
    for n, (start, match) in enumerate(marks):
        command = lines[start][match.end():]
        if not command.strip():
            continue
        end = len(lines)
        if n + 1 < len(marks):
            end = max(start + 1, marks[n + 1][0] - (prompt_lines - 1))
        output = "\n".join(lines[start + 1:end]).rstrip("\n")
        blocks.append({"command": command, "output": output, "prompt_line": lines[start]})
    return blocks


def payload(block, kind):
    if kind == "command":
        return block["command"]
    if kind == "output":
        return block["output"]
    return block["prompt_line"] + ("\n" + block["output"] if block["output"] else "")


def clean_label(text):
    return "".join(c if c.isprintable() else " " for c in text)


def copy_text(text):
    """Use the local clipboard when available; otherwise use a tmux buffer."""
    # clip.exe is listed ahead of wl-copy so WSL (where WSLg sets
    # WAYLAND_DISPLAY) uses the fast Windows interop path. pbcopy stays
    # first, so macOS behavior is unchanged. clip.exe is absent on macOS
    # and native Linux, so those paths also stay the same.
    candidates = [
        (["pbcopy"], True),
        (["clip.exe"], True),
        (["wl-copy"], bool(os.environ.get("WAYLAND_DISPLAY"))),
        (["xclip", "-selection", "clipboard"], bool(os.environ.get("DISPLAY"))),
        (["xsel", "--clipboard", "--input"], bool(os.environ.get("DISPLAY"))),
    ]
    for command, available in candidates:
        if available and shutil.which(command[0]):
            # A clipboard owner can outlive the launcher. Keep it off the popup PTY.
            result = subprocess.run(
                command, input=text, text=True, stdout=subprocess.DEVNULL,
                stderr=subprocess.DEVNULL, start_new_session=True,
            )
            if result.returncode == 0:
                return
    subprocess.run(["tmux", "load-buffer", "-"], input=text, text=True, check=True)
    subprocess.run(["tmux", "display-message", "Copied to tmux buffer (system clipboard unavailable)"], check=True)


def run_picker(blocks, folder):
    if not shutil.which("fzf"):
        raise FileNotFoundError("fzf is required; install it and make sure it is on PATH")
    rows = []
    for index, block in reversed(list(enumerate(blocks))):
        preview_text = "".join(
            c if c.isprintable() or c in "\n\t" else " " for c in payload(block, "both")
        )
        (folder / str(index)).write_text(preview_text, encoding="utf-8")
        rows.append(f"{index}\t{clean_label(block['command'])}")
    preview = f"/bin/cat {shlex.quote(str(folder))}/{{1}}"
    env = os.environ.copy()
    env.pop("FZF_DEFAULT_OPTS", None)
    env.pop("FZF_DEFAULT_OPTS_FILE", None)
    result = subprocess.run(
        [
            "fzf", "--no-sort", "--layout=reverse", "--delimiter=\t", "--with-nth=2..",
            "--prompt=Commands> ",
            "--header=Enter: output | Ctrl-y: command | Ctrl-b: both | Esc: cancel",
            "--expect=enter,ctrl-y,ctrl-b", "--preview", preview,
            "--preview-window=right,60%,wrap",
            "--bind=ctrl-d:preview-half-page-down,ctrl-u:preview-half-page-up",
        ],
        input="\n".join(rows), text=True, stdout=subprocess.PIPE, env=env,
    )
    if result.returncode in (1, 130):
        return
    result.check_returncode()
    lines = result.stdout.splitlines()
    if len(lines) < 2:
        return
    key, selection = lines[:2]
    index = int(selection.split("\t", 1)[0])
    kind = {"enter": "output", "ctrl-y": "command", "ctrl-b": "both"}[key]
    copy_text(payload(blocks[index], kind))


def capture_text(args):
    if args.input:
        return args.input.read_text(encoding="utf-8")
    if not args.pane:
        raise ValueError("--pane is required unless --input is used")
    return subprocess.run(
        ["tmux", "capture-pane", "-J", "-p", "-t", args.pane, "-S", "-"],
        text=True, stdout=subprocess.PIPE, check=True,
    ).stdout


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("mode", choices=["browse", "preview"], nargs="?", default="browse")
    parser.add_argument("--pane", help="Source tmux pane ID, supplied before opening the popup")
    parser.add_argument("--input", type=Path, help="Read a saved text capture instead of querying tmux")
    parser.add_argument("--prompt", default=r"^CMD>\s?", help="Regex that identifies a command line")
    parser.add_argument("--prompt-lines", type=int, default=3,
                        help="Prompt lines to remove before the next command (default: 3)")
    args = parser.parse_args()
    if args.prompt_lines < 1:
        parser.error("--prompt-lines must be at least 1")
    text = capture_text(args)
    if args.mode == "preview":
        print(json.dumps(parse_blocks(text, args.prompt, args.prompt_lines), indent=2))
        return
    with tempfile.TemporaryDirectory(prefix="tmux-history-") as directory:
        folder = Path(directory)
        blocks = parse_blocks(text, args.prompt, args.prompt_lines)
        if not blocks:
            print("No command blocks found. Check the CMD> marker, --prompt, and --prompt-lines.")
            if not args.input and sys.stdin.isatty():
                try:
                    input("Press Enter to close this popup.")
                except (EOFError, KeyboardInterrupt):
                    pass
            return
        run_picker(blocks, folder)


if __name__ == "__main__":
    try:
        main()
    except (subprocess.CalledProcessError, OSError, ValueError, re.error) as error:
        print(f"tmux history: {error}", file=sys.stderr)
        sys.exit(1)
