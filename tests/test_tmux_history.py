import importlib.util
import os
from pathlib import Path
import subprocess
import tempfile
import unittest
from unittest.mock import patch
from types import SimpleNamespace


spec = importlib.util.spec_from_file_location(
    "history", Path(__file__).parents[1] / "scripts/tmux-history.py"
)
history = importlib.util.module_from_spec(spec)
spec.loader.exec_module(history)


class HistoryTests(unittest.TestCase):
    def test_prompt_decoration_and_empty_final_prompt(self):
        blocks = history.parse_blocks(
            "older output\n\n~/repo\nCMD> echo hello\nhello\n\n~/repo\n"
            "CMD> false\n\n~/repo\nCMD> "
        )
        self.assertEqual([b["command"] for b in blocks], ["echo hello", "false"])
        self.assertEqual([b["output"] for b in blocks], ["hello", ""])

    def test_custom_prompt_and_incomplete_block(self):
        blocks = history.parse_blocks("$ pwd\n/home/me", r"^\$ ", 1)
        self.assertEqual(blocks[0]["output"], "/home/me")
        self.assertEqual(history.parse_blocks("no commands"), [])

    def test_all_copy_actions_and_newest_first(self):
        blocks = history.parse_blocks("CMD> first\nold\nCMD> second\nnew", prompt_lines=1)
        for key, expected in [
            ("enter", "new"), ("ctrl-y", "second"), ("ctrl-b", "CMD> second\nnew")
        ]:
            with self.subTest(key=key), tempfile.TemporaryDirectory() as directory:
                folder = Path(directory)
                def pick(command, **kwargs):
                    self.assertEqual(kwargs["input"].splitlines(), ["1\tsecond", "0\tfirst"])
                    self.assertEqual((folder / "1").read_text(), "CMD> second\nnew")
                    self.assertNotIn("FZF_DEFAULT_OPTS", kwargs["env"])
                    self.assertNotIn("FZF_DEFAULT_OPTS_FILE", kwargs["env"])
                    return subprocess.CompletedProcess(command, 0, key + "\n1\tsecond\n")
                with patch.object(history.shutil, "which", return_value="fzf"), \
                     patch.object(history.subprocess, "run", side_effect=pick), \
                     patch.object(history, "copy_text") as copy, \
                     patch.dict(os.environ, {"FZF_DEFAULT_OPTS": "--multi", "FZF_DEFAULT_OPTS_FILE": "bad"}):
                    history.run_picker(blocks, folder)
                    copy.assert_called_once_with(expected)

    def test_cancel_does_not_copy(self):
        with tempfile.TemporaryDirectory() as directory, \
             patch.object(history.shutil, "which", return_value="fzf"), \
             patch.object(history.subprocess, "run", return_value=subprocess.CompletedProcess([], 130)), \
             patch.object(history, "copy_text") as copy:
            history.run_picker([dict(command="echo hi", output="hi", prompt_line="CMD> echo hi")], Path(directory))
            copy.assert_not_called()

    def test_capture_targets_originating_pane_and_joins_wrapping(self):
        with patch.object(history.subprocess, "run", return_value=subprocess.CompletedProcess([], 0, "capture")) as run:
            self.assertEqual(history.capture_text(SimpleNamespace(input=None, pane="%42")), "capture")
            self.assertEqual(run.call_args.args[0], ["tmux", "capture-pane", "-J", "-p", "-t", "%42", "-S", "-"])

    def test_missing_pane_is_an_error(self):
        with self.assertRaises(ValueError):
            history.capture_text(SimpleNamespace(input=None, pane=None))

    def test_clipboard_platforms(self):
        for program, env, command in [
            ("pbcopy", {}, ["pbcopy"]),
            ("wl-copy", {"WAYLAND_DISPLAY": "wayland-0"}, ["wl-copy"]),
            ("xclip", {"DISPLAY": ":0"}, ["xclip", "-selection", "clipboard"]),
            ("xsel", {"DISPLAY": ":0"}, ["xsel", "--clipboard", "--input"]),
        ]:
            with self.subTest(program=program), patch.dict(os.environ, env, clear=True), \
                 patch.object(history.shutil, "which", side_effect=lambda name: name == program), \
                 patch.object(history.subprocess, "run", return_value=subprocess.CompletedProcess([], 0)) as run:
                history.copy_text("copied")
                run.assert_called_once_with(command, input="copied", text=True)

    def test_failed_clipboard_falls_back_to_buffer(self):
        with patch.dict(os.environ, {}, clear=True), \
             patch.object(history.shutil, "which", return_value="pbcopy"), \
             patch.object(history.subprocess, "run", side_effect=[
                 subprocess.CompletedProcess([], 1), subprocess.CompletedProcess([], 0),
                 subprocess.CompletedProcess([], 0)
             ]) as run:
            history.copy_text("literal $(command);\noutput")
            self.assertEqual(run.call_args_list[1].args[0], ["tmux", "load-buffer", "-"])
            self.assertEqual(run.call_args_list[1].kwargs["input"], "literal $(command);\noutput")


if __name__ == "__main__":
    unittest.main()
