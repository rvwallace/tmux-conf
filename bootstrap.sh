#!/usr/bin/env bash

set -euo pipefail

REPO_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TARGET_CONF="${HOME}/.tmux.conf"
SCRIPT_DIR="${HOME}/.config/tmux/scripts"
TARGET_SCROLLBACK_SCRIPT="${SCRIPT_DIR}/edit-scrollback.sh"
TARGET_COPY_PATH_SCRIPT="${SCRIPT_DIR}/copy-pane-path.sh"
TARGET_SHOW_CLIENTS_SCRIPT="${SCRIPT_DIR}/show-clients.sh"
TARGET_PALETTE_DIR="${HOME}/.config/tmux-palette"
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

if [ -L "${TARGET_PALETTE_DIR}" ]; then
  rm -f "${TARGET_PALETTE_DIR}"
elif [ -e "${TARGET_PALETTE_DIR}" ]; then
  backup="${TARGET_PALETTE_DIR}.backup.$(date +%Y%m%d-%H%M%S)"
  mv "${TARGET_PALETTE_DIR}" "${backup}"
  echo "Moved existing ${TARGET_PALETTE_DIR} to ${backup}"
fi
ln -s "${REPO_DIR}/tmux-palette" "${TARGET_PALETTE_DIR}"
echo "Linked ${TARGET_PALETTE_DIR} -> ${REPO_DIR}/tmux-palette"

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
