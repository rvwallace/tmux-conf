#!/usr/bin/env bash
# Unit tests for copy-pane-path.sh clipboard order and helper detachment.
set -euo pipefail

REPO_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SCRIPT="${REPO_DIR}/scripts/copy-pane-path.sh"
TMP="$(mktemp -d)"
trap 'rm -rf "${TMP}"' EXIT

fail() { printf 'FAIL: %s\n' "$*" >&2; exit 1; }
pass() { printf 'PASS: %s\n' "$*"; }

mkdir -p "${TMP}/bin" "${TMP}/path"
# Keep PATH narrow so real clipboard tools are invisible, but leave
# enough utilities for shebangs and the fake helpers.
for cmd in bash env setsid ps readlink cat tr mkdir grep sed; do
  src="$(command -v "${cmd}")" || fail "missing helper: ${cmd}"
  ln -s "${src}" "${TMP}/bin/${cmd}"
done

cat > "${TMP}/path/tmux" << 'EOF'
#!/usr/bin/env bash
if [ "${1-}" = "display-message" ]; then
  if [ "${2-}" = "-p" ] && [ "${3-}" = "#{pane_id}" ]; then
    printf '%s\n' '%1'
    exit 0
  fi
  if [ "${2-}" = "-p" ] && [ "${3-}" = "-t" ] && [ "${5-}" = "#{pane_current_path}" ]; then
    printf '%s\n' '/tmp/pane-path-probe'
    exit 0
  fi
  printf '%s\n' "$*" > "${FAKE_LOG_DIR}/tmux-message"
  exit 0
fi
if [ "${1-}" = "set-buffer" ]; then
  printf '%s\n' "$*" > "${FAKE_LOG_DIR}/tmux-buffer"
  exit 0
fi
printf 'unexpected tmux: %s\n' "$*" >&2
exit 1
EOF
chmod +x "${TMP}/path/tmux"

install_tool() {
  local name="$1"
  cat > "${TMP}/path/${name}" << EOF
#!/usr/bin/env bash
log="${TMP}/logs/${name}.log"
mkdir -p "${TMP}/logs"
sid="\$(ps -o sid= -p "\$\$" | tr -d ' ')"
ppid_sid="\$(ps -o sid= -p "\$PPID" | tr -d ' ')"
stdout_link="\$(readlink /proc/\$\$/fd/1 2>/dev/null || true)"
stderr_link="\$(readlink /proc/\$\$/fd/2 2>/dev/null || true)"
stdin="\$(cat)"
{
  echo "sid=\${sid}"
  echo "ppid_sid=\${ppid_sid}"
  echo "stdout=\${stdout_link}"
  echo "stderr=\${stderr_link}"
  echo "text=\${stdin}"
} > "\${log}"
exit 0
EOF
  chmod +x "${TMP}/path/${name}"
}

run_script() {
  local case="$1"
  shift
  rm -rf "${TMP}/logs"
  mkdir -p "${TMP}/logs"
  FAKE_LOG_DIR="${TMP}/logs" PATH="${TMP}/path:${TMP}/bin" "$@" \
    bash "${SCRIPT}" >/dev/null 2>&1 || fail "${case}: script exited $?"
}

expect_only() {
  local case="$1" tool="$2"
  [ -f "${TMP}/logs/${tool}.log" ] || fail "${case}: ${tool} was not invoked"
  local text
  text="$(grep '^text=' "${TMP}/logs/${tool}.log" | sed 's/^text=//')"
  [ "${text}" = "/tmp/pane-path-probe" ] || fail "${case}: ${tool} got '${text}'"
  local other
  for other in pbcopy clip.exe wl-copy xclip xsel; do
    if [ "${other}" != "${tool}" ] && [ -e "${TMP}/path/${other}" ] && [ -f "${TMP}/logs/${other}.log" ]; then
      fail "${case}: ${other} was also invoked"
    fi
  done
  grep -F "Copied pane directory: /tmp/pane-path-probe" "${TMP}/logs/tmux-message" >/dev/null \
    || fail "${case}: missing success message"
}

expect_detached() {
  local case="$1" tool="$2"
  local sid ppid_sid stdout_link stderr_link
  sid="$(grep '^sid=' "${TMP}/logs/${tool}.log" | sed 's/^sid=//')"
  ppid_sid="$(grep '^ppid_sid=' "${TMP}/logs/${tool}.log" | sed 's/^ppid_sid=//')"
  stdout_link="$(grep '^stdout=' "${TMP}/logs/${tool}.log" | sed 's/^stdout=//')"
  stderr_link="$(grep '^stderr=' "${TMP}/logs/${tool}.log" | sed 's/^stderr=//')"
  [ "${stdout_link}" = "/dev/null" ] || fail "${case}: stdout was '${stdout_link}', expected /dev/null"
  [ "${stderr_link}" = "/dev/null" ] || fail "${case}: stderr was '${stderr_link}', expected /dev/null"
  [ -n "${sid}" ] && [ "${sid}" != "${ppid_sid}" ] || fail "${case}: helper sid '${sid}' was not a new session (parent sid '${ppid_sid}')"
}

# pbcopy wins over clip.exe
install_tool pbcopy
install_tool clip.exe
run_script "pbcopy preference" 
expect_only "pbcopy preference" pbcopy
expect_detached "pbcopy preference" pbcopy
pass "pbcopy still wins when clip.exe exists"

# clip.exe beats wl-copy
rm -f "${TMP}/path/pbcopy"
install_tool clip.exe
install_tool wl-copy
install_tool xsel
run_script "clip.exe preference"
expect_only "clip.exe preference" clip.exe
expect_detached "clip.exe preference" clip.exe
pass "clip.exe is preferred over wl-copy"

# wl-copy is used and detached when no earlier candidate exists
rm -f "${TMP}/path/clip.exe" "${TMP}/path/xsel"
install_tool wl-copy
run_script "wl-copy detach"
expect_only "wl-copy detach" wl-copy
expect_detached "wl-copy detach" wl-copy
pass "wl-copy runs in a new session with stdio discarded"

# Without setsid (macOS path), stdout/stderr are still discarded
rm -f "${TMP}/bin/setsid"
run_script "no-setsid redirect" 
expect_only "no-setsid redirect" wl-copy
stdout_link="$(grep '^stdout=' "${TMP}/logs/wl-copy.log" | sed 's/^stdout=//')"
stderr_link="$(grep '^stderr=' "${TMP}/logs/wl-copy.log" | sed 's/^stderr=//')"
[ "${stdout_link}" = "/dev/null" ] || fail "no-setsid: stdout was '${stdout_link}'"
[ "${stderr_link}" = "/dev/null" ] || fail "no-setsid: stderr was '${stderr_link}'"
pass "without setsid, stdout and stderr are still discarded"

# tmux buffer fallback when no clipboard tool is present
rm -f "${TMP}/path/wl-copy"
rm -rf "${TMP}/logs"
mkdir -p "${TMP}/logs"
FAKE_LOG_DIR="${TMP}/logs" PATH="${TMP}/path:${TMP}/bin" \
  bash "${SCRIPT}" >/dev/null 2>&1 || fail "buffer fallback: script exited $?"
grep -F "set-buffer -- /tmp/pane-path-probe" "${TMP}/logs/tmux-buffer" >/dev/null \
  || fail "buffer fallback: set-buffer was not used"
grep -F "no system clipboard tool found" "${TMP}/logs/tmux-message" >/dev/null \
  || fail "buffer fallback: missing fallback message"
pass "falls back to a tmux buffer when no clipboard tool is found"

printf 'All copy-pane-path checks passed.\n'
