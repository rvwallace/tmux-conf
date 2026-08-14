#!/usr/bin/env bash

set -euo pipefail

REPO_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TARGET_CONF="${HOME}/.tmux.conf"
SCRIPT_DIR="${HOME}/.config/tmux/scripts"
TARGET_THEME_DIR="${HOME}/.config/tmux/themes"
TARGET_SCROLLBACK_SCRIPT="${SCRIPT_DIR}/edit-scrollback.sh"
TARGET_COPY_PATH_SCRIPT="${SCRIPT_DIR}/copy-pane-path.sh"
TARGET_SHOW_CLIENTS_SCRIPT="${SCRIPT_DIR}/show-clients.sh"
TARGET_AI_SCRIPT="${SCRIPT_DIR}/ai-assist.sh"
TARGET_AI_PROMPT_SCRIPT="${SCRIPT_DIR}/ai-prompt.sh"
TARGET_AICHAT_TUI_SCRIPT="${SCRIPT_DIR}/aichat-tui.py"
TARGET_AGENT_PANE_SCRIPT="${SCRIPT_DIR}/launch-agent-pane.sh"
TARGET_FILE_PICKER_SCRIPT="${SCRIPT_DIR}/tmux-file-picker"
TARGET_INSERT_PATHS_SCRIPT="${SCRIPT_DIR}/tmux-insert-paths.sh"
TARGET_YAZI_PICKER_SCRIPT="${SCRIPT_DIR}/tmux-yazi-picker.sh"
TARGET_DEFER_TMUX_SCRIPT="${SCRIPT_DIR}/defer-tmux-command.sh"
TARGET_WHICH_KEY_POPUP_SCRIPT="${SCRIPT_DIR}/tmux-which-key-popup.sh"
TARGET_TMUX_MENU_SCRIPT="${SCRIPT_DIR}/tmux-menu.py"
TARGET_PREFIX_HELP_SCRIPT="${SCRIPT_DIR}/tmux-prefix-help.py"
TARGET_TMUX_MENU_DIR="${HOME}/.config/tmux-menu"
TARGET_SNAGLORD_DIR="${HOME}/.config/tmux-snaglord"
PLUGIN_DIR="${HOME}/.config/tmux/plugins"
TPM_DIR="${PLUGIN_DIR}/tpm"
TPM_REPO="https://github.com/tmux-plugins/tpm"

echo "Repo dir: ${REPO_DIR}"

if ! "${REPO_DIR}/scripts/install-deps.sh" --check; then
  if [ -t 0 ]; then
    read -r -p "Install available tmux-conf dependencies now? [y/N] " install_deps || true
    case "${install_deps:-}" in
      y | Y | yes | YES) "${REPO_DIR}/scripts/install-deps.sh" --install || true ;;
      *) echo "Skipped dependency installation." ;;
    esac
  else
    echo "Run ./scripts/install-deps.sh --install to install available dependencies."
  fi
fi

mkdir -p "${PLUGIN_DIR}"
mkdir -p "${SCRIPT_DIR}"

if [ -L "${TARGET_CONF}" ] || [ -e "${TARGET_CONF}" ]; then
  rm -f "${TARGET_CONF}"
fi
ln -s "${REPO_DIR}/.tmux.conf" "${TARGET_CONF}"
echo "Linked ${TARGET_CONF} -> ${REPO_DIR}/.tmux.conf"

if [ -L "${TARGET_THEME_DIR}" ]; then
  rm -f "${TARGET_THEME_DIR}"
elif [ -e "${TARGET_THEME_DIR}" ]; then
  backup="${TARGET_THEME_DIR}.backup.$(date +%Y%m%d-%H%M%S)"
  mv "${TARGET_THEME_DIR}" "${backup}"
  echo "Moved existing ${TARGET_THEME_DIR} to ${backup}"
fi
ln -s "${REPO_DIR}/themes" "${TARGET_THEME_DIR}"
echo "Linked ${TARGET_THEME_DIR} -> ${REPO_DIR}/themes"

if [ -L "${TARGET_SCROLLBACK_SCRIPT}" ] || [ -e "${TARGET_SCROLLBACK_SCRIPT}" ]; then
  rm -f "${TARGET_SCROLLBACK_SCRIPT}"
fi
ln -s "${REPO_DIR}/scripts/edit-scrollback.sh" "${TARGET_SCROLLBACK_SCRIPT}"
echo "Linked ${TARGET_SCROLLBACK_SCRIPT} -> ${REPO_DIR}/scripts/edit-scrollback.sh"

if [ -L "${TARGET_COPY_PATH_SCRIPT}" ] || [ -e "${TARGET_COPY_PATH_SCRIPT}" ]; then
  rm -f "${TARGET_COPY_PATH_SCRIPT}"
fi
ln -s "${REPO_DIR}/scripts/copy-pane-path.sh" "${TARGET_COPY_PATH_SCRIPT}"
echo "Linked ${TARGET_COPY_PATH_SCRIPT} -> ${REPO_DIR}/scripts/copy-pane-path.sh"

if [ -L "${TARGET_SHOW_CLIENTS_SCRIPT}" ] || [ -e "${TARGET_SHOW_CLIENTS_SCRIPT}" ]; then
  rm -f "${TARGET_SHOW_CLIENTS_SCRIPT}"
fi
ln -s "${REPO_DIR}/scripts/show-clients.sh" "${TARGET_SHOW_CLIENTS_SCRIPT}"
echo "Linked ${TARGET_SHOW_CLIENTS_SCRIPT} -> ${REPO_DIR}/scripts/show-clients.sh"

if [ -L "${TARGET_AI_SCRIPT}" ] || [ -e "${TARGET_AI_SCRIPT}" ]; then
  rm -f "${TARGET_AI_SCRIPT}"
fi
ln -s "${REPO_DIR}/scripts/ai-assist.sh" "${TARGET_AI_SCRIPT}"
echo "Linked ${TARGET_AI_SCRIPT} -> ${REPO_DIR}/scripts/ai-assist.sh"

if [ -L "${TARGET_AI_PROMPT_SCRIPT}" ] || [ -e "${TARGET_AI_PROMPT_SCRIPT}" ]; then
  rm -f "${TARGET_AI_PROMPT_SCRIPT}"
fi
ln -s "${REPO_DIR}/scripts/ai-prompt.sh" "${TARGET_AI_PROMPT_SCRIPT}"
echo "Linked ${TARGET_AI_PROMPT_SCRIPT} -> ${REPO_DIR}/scripts/ai-prompt.sh"

if [ -L "${TARGET_AICHAT_TUI_SCRIPT}" ] || [ -e "${TARGET_AICHAT_TUI_SCRIPT}" ]; then
  rm -f "${TARGET_AICHAT_TUI_SCRIPT}"
fi
ln -s "${REPO_DIR}/scripts/aichat-tui.py" "${TARGET_AICHAT_TUI_SCRIPT}"
echo "Linked ${TARGET_AICHAT_TUI_SCRIPT} -> ${REPO_DIR}/scripts/aichat-tui.py"

if [ -L "${TARGET_AGENT_PANE_SCRIPT}" ] || [ -e "${TARGET_AGENT_PANE_SCRIPT}" ]; then
  rm -f "${TARGET_AGENT_PANE_SCRIPT}"
fi
ln -s "${REPO_DIR}/scripts/launch-agent-pane.sh" "${TARGET_AGENT_PANE_SCRIPT}"
echo "Linked ${TARGET_AGENT_PANE_SCRIPT} -> ${REPO_DIR}/scripts/launch-agent-pane.sh"

if [ -L "${TARGET_FILE_PICKER_SCRIPT}" ] || [ -e "${TARGET_FILE_PICKER_SCRIPT}" ]; then
  rm -f "${TARGET_FILE_PICKER_SCRIPT}"
fi
ln -s "${REPO_DIR}/scripts/tmux-file-picker" "${TARGET_FILE_PICKER_SCRIPT}"
echo "Linked ${TARGET_FILE_PICKER_SCRIPT} -> ${REPO_DIR}/scripts/tmux-file-picker"

if [ -L "${TARGET_INSERT_PATHS_SCRIPT}" ] || [ -e "${TARGET_INSERT_PATHS_SCRIPT}" ]; then
  rm -f "${TARGET_INSERT_PATHS_SCRIPT}"
fi
ln -s "${REPO_DIR}/scripts/tmux-insert-paths.sh" "${TARGET_INSERT_PATHS_SCRIPT}"
echo "Linked ${TARGET_INSERT_PATHS_SCRIPT} -> ${REPO_DIR}/scripts/tmux-insert-paths.sh"

if [ -L "${TARGET_YAZI_PICKER_SCRIPT}" ] || [ -e "${TARGET_YAZI_PICKER_SCRIPT}" ]; then
  rm -f "${TARGET_YAZI_PICKER_SCRIPT}"
fi
ln -s "${REPO_DIR}/scripts/tmux-yazi-picker.sh" "${TARGET_YAZI_PICKER_SCRIPT}"
echo "Linked ${TARGET_YAZI_PICKER_SCRIPT} -> ${REPO_DIR}/scripts/tmux-yazi-picker.sh"

if [ -L "${TARGET_DEFER_TMUX_SCRIPT}" ] || [ -e "${TARGET_DEFER_TMUX_SCRIPT}" ]; then
  rm -f "${TARGET_DEFER_TMUX_SCRIPT}"
fi
ln -s "${REPO_DIR}/scripts/defer-tmux-command.sh" "${TARGET_DEFER_TMUX_SCRIPT}"
echo "Linked ${TARGET_DEFER_TMUX_SCRIPT} -> ${REPO_DIR}/scripts/defer-tmux-command.sh"

if [ -L "${TARGET_WHICH_KEY_POPUP_SCRIPT}" ] || [ -e "${TARGET_WHICH_KEY_POPUP_SCRIPT}" ]; then
  rm -f "${TARGET_WHICH_KEY_POPUP_SCRIPT}"
fi
ln -s "${REPO_DIR}/scripts/tmux-which-key-popup.sh" "${TARGET_WHICH_KEY_POPUP_SCRIPT}"
echo "Linked ${TARGET_WHICH_KEY_POPUP_SCRIPT} -> ${REPO_DIR}/scripts/tmux-which-key-popup.sh"

if [ -L "${TARGET_TMUX_MENU_SCRIPT}" ] || [ -e "${TARGET_TMUX_MENU_SCRIPT}" ]; then
  rm -f "${TARGET_TMUX_MENU_SCRIPT}"
fi
ln -s "${REPO_DIR}/scripts/tmux-menu.py" "${TARGET_TMUX_MENU_SCRIPT}"
echo "Linked ${TARGET_TMUX_MENU_SCRIPT} -> ${REPO_DIR}/scripts/tmux-menu.py"

if [ -L "${TARGET_PREFIX_HELP_SCRIPT}" ] || [ -e "${TARGET_PREFIX_HELP_SCRIPT}" ]; then
  rm -f "${TARGET_PREFIX_HELP_SCRIPT}"
fi
ln -s "${REPO_DIR}/scripts/tmux-prefix-help.py" "${TARGET_PREFIX_HELP_SCRIPT}"
echo "Linked ${TARGET_PREFIX_HELP_SCRIPT} -> ${REPO_DIR}/scripts/tmux-prefix-help.py"

if [ -L "${TARGET_TMUX_MENU_DIR}" ]; then
  rm -f "${TARGET_TMUX_MENU_DIR}"
elif [ -e "${TARGET_TMUX_MENU_DIR}" ]; then
  backup="${TARGET_TMUX_MENU_DIR}.backup.$(date +%Y%m%d-%H%M%S)"
  mv "${TARGET_TMUX_MENU_DIR}" "${backup}"
  echo "Moved existing ${TARGET_TMUX_MENU_DIR} to ${backup}"
fi
ln -s "${REPO_DIR}/tmux-menu" "${TARGET_TMUX_MENU_DIR}"
echo "Linked ${TARGET_TMUX_MENU_DIR} -> ${REPO_DIR}/tmux-menu"

# Clean legacy symlinks if present
rm -f "${HOME}/.config/tmux-palette" "${HOME}/.config/tmux-which-key" 2>/dev/null || true

if [ -L "${TARGET_SNAGLORD_DIR}" ]; then
  rm -f "${TARGET_SNAGLORD_DIR}"
elif [ -e "${TARGET_SNAGLORD_DIR}" ]; then
  backup="${TARGET_SNAGLORD_DIR}.backup.$(date +%Y%m%d-%H%M%S)"
  mv "${TARGET_SNAGLORD_DIR}" "${backup}"
  echo "Moved existing ${TARGET_SNAGLORD_DIR} to ${backup}"
fi
ln -s "${REPO_DIR}/tmux-snaglord" "${TARGET_SNAGLORD_DIR}"
echo "Linked ${TARGET_SNAGLORD_DIR} -> ${REPO_DIR}/tmux-snaglord"

if [ ! -d "${TPM_DIR}/.git" ]; then
  git clone "${TPM_REPO}" "${TPM_DIR}"
  echo "Cloned TPM into ${TPM_DIR}"
else
  echo "TPM already present at ${TPM_DIR}"
fi

cat <<'EOF'

Next steps:
  1. Start tmux or reload it with: tmux source-file ~/.tmux.conf
  2. Install/update declared plugins with: prefix + I

EOF
