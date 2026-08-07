# tmux-file-picker upstream

`tmux-file-picker` is vendored from:

- Repository: <https://github.com/raine/tmux-file-picker>
- Commit: `d1561a75aebfb50e5ad38facac684252014a44a0`
- License declared by upstream: MIT

This repo carries a small integration patch: selected paths are handed to
`tmux-insert-paths.sh` so they use a temporary tmux paste buffer and share
agent-aware formatting with the Yazi picker.

When updating it, reapply that integration after replacing
`scripts/tmux-file-picker`, update the commit above, and run `./validate.sh`.
