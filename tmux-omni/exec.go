package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

// CopyToClipboard copies text to the system clipboard across macOS, Wayland, X11, or tmux buffer.
func CopyToClipboard(text string) bool {
	if _, err := exec.LookPath("pbcopy"); err == nil {
		cmd := exec.Command("pbcopy")
		cmd.Stdin = strings.NewReader(text)
		if err := cmd.Run(); err == nil {
			return true
		}
	}
	if _, err := exec.LookPath("wl-copy"); err == nil {
		cmd := exec.Command("wl-copy")
		cmd.Stdin = strings.NewReader(text)
		if err := cmd.Run(); err == nil {
			return true
		}
	}
	if _, err := exec.LookPath("xclip"); err == nil {
		cmd := exec.Command("xclip", "-selection", "clipboard")
		cmd.Stdin = strings.NewReader(text)
		if err := cmd.Run(); err == nil {
			return true
		}
	}
	if _, err := exec.LookPath("tmux"); err == nil {
		cmd := exec.Command("tmux", "set-buffer", "--", text)
		if err := cmd.Run(); err == nil {
			return true
		}
	}
	return false
}

// GetCurrentPaneID resolves the active tmux pane ID.
func GetCurrentPaneID(provided string) string {
	clean := strings.Trim(provided, "'\" \t\r\n")
	if strings.HasPrefix(clean, "%") {
		return clean
	}
	if _, err := exec.LookPath("tmux"); err == nil {
		out, err := exec.Command("tmux", "display-message", "-p", "#{pane_id}").Output()
		if err == nil {
			val := strings.TrimSpace(string(out))
			if strings.HasPrefix(val, "%") {
				return val
			}
		}
	}
	return "%0"
}

// GetPaneCWD retrieves the current working directory of the specified pane.
func GetPaneCWD(paneID string) string {
	if _, err := exec.LookPath("tmux"); err == nil && paneID != "" {
		out, err := exec.Command("tmux", "display-message", "-p", "-t", paneID, "#{pane_current_path}").Output()
		if err == nil {
			dir := strings.TrimSpace(string(out))
			if info, err := os.Stat(dir); err == nil && info.IsDir() {
				return dir
			}
		}
	}
	return ""
}

// QuoteShellSingle quotes a string safely for sh/bash/zsh single quotes.
func QuoteShellSingle(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "'\\''") + "'"
}

// BuildGuardedShellScript generates an interactive error-guard shell wrapper.
func BuildGuardedShellScript(cmdStr string, title string, persistShell bool) string {
	if title == "" {
		title = cmdStr
	}
	quotedTitle := QuoteShellSingle(title)

	if persistShell {
		return fmt.Sprintf("%s ; exec \"${SHELL:-/bin/zsh}\"", cmdStr)
	}

	return fmt.Sprintf(`
%s
_status=$?

if [ "$_status" -ne 0 ]; then
  printf '\n\033[1;31m✖ Command failed with exit code %%d: %%s\033[0m\n' "$_status" %s
  printf '\033[1;33m[s]\033[0m Debug shell   \033[1;37m[any key]\033[0m Close\n'
  read -r _key 2>/dev/null || _key=""
  case "$_key" in
    s|S)
      printf '\nDropping to debug shell...\n'
      exec "${SHELL:-/bin/zsh}"
      ;;
    *)
      ;;
  esac
fi
exit "$_status"
`, cmdStr, quotedTitle)
}

// SplitTmuxCommands splits a chained tmux command string by unescaped semicolons.
func SplitTmuxCommands(s string) []string {
	var parts []string
	var current strings.Builder
	escaped := false

	for i := 0; i < len(s); i++ {
		c := s[i]

		if escaped {
			if c == ';' {
				current.WriteByte(';')
			} else {
				current.WriteByte('\\')
				current.WriteByte(c)
			}
			escaped = false
			continue
		}

		if c == '\\' {
			escaped = true
			continue
		}

		if c == ';' {
			trimmed := strings.TrimSpace(current.String())
			if trimmed != "" {
				parts = append(parts, trimmed)
			}
			current.Reset()
			continue
		}

		current.WriteByte(c)
	}

	if escaped {
		current.WriteByte('\\')
	}

	trimmed := strings.TrimSpace(current.String())
	if trimmed != "" {
		parts = append(parts, trimmed)
	}

	return parts
}

// RunTmuxTarget executes the selected action with modifiers and error guarding.
func RunTmuxTarget(
	action string,
	target string,
	paneID string,
	title string,
	persistShell bool,
	originalTarget string,
) error {
	cleanPaneID := GetCurrentPaneID(paneID)
	cwd := GetPaneCWD(cleanPaneID)

	if cwd != "" {
		_ = os.Chdir(cwd)
	}

	// Template replacement
	expandedAction := strings.ReplaceAll(action, "#{pane_id}", cleanPaneID)
	if cwd != "" {
		expandedAction = strings.ReplaceAll(expandedAction, "#{pane_current_path}", cwd)
	}

	shellBin := os.Getenv("SHELL")
	if shellBin == "" {
		shellBin = "/bin/zsh"
	}

	// 1. Direct tmux commands
	if target == "tmux" {
		parts := SplitTmuxCommands(expandedAction)
		for _, part := range parts {
			args := parseArgs(part)
			cmd := exec.Command("tmux", args...)
			cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
			if err := cmd.Start(); err != nil {
				_ = exec.Command("tmux", "display-message", fmt.Sprintf("✖ Execution failed: %v", err)).Run()
			}
		}
		return nil
	}

	// 2. Send keys to target pane
	if target == "send_keys" {
		textToSend := expandedAction
		if originalTarget == "tmux" {
			subCmds := SplitTmuxCommands(textToSend)
			var formatted []string
			for _, sc := range subCmds {
				if !strings.HasPrefix(sc, "tmux ") {
					formatted = append(formatted, "tmux "+sc)
				} else {
					formatted = append(formatted, sc)
				}
			}
			textToSend = strings.Join(formatted, " ; ")
		}

		var args []string
		args = append(args, "send-keys")
		if cleanPaneID != "" {
			args = append(args, "-t", cleanPaneID)
		}
		args = append(args, textToSend)

		cmd := exec.Command("tmux", args...)
		cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
		return cmd.Start()
	}

	// 3. Shell wrapper commands (splits, windows, popups)
	guardedScript := BuildGuardedShellScript(expandedAction, title, persistShell)

	switch target {
	case "split_h":
		args := []string{"split-window", "-h"}
		if cleanPaneID != "" {
			args = append(args, "-t", cleanPaneID)
		}
		if cwd != "" {
			args = append(args, "-c", cwd)
		}
		args = append(args, shellBin, "-lc", guardedScript)
		cmd := exec.Command("tmux", args...)
		cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
		return cmd.Start()

	case "split_v":
		args := []string{"split-window", "-v"}
		if cleanPaneID != "" {
			args = append(args, "-t", cleanPaneID)
		}
		if cwd != "" {
			args = append(args, "-c", cwd)
		}
		args = append(args, shellBin, "-lc", guardedScript)
		cmd := exec.Command("tmux", args...)
		cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
		return cmd.Start()

	case "window":
		winTitle := title
		if winTitle == "" {
			winTitle = "shell"
		}
		args := []string{"new-window", "-n", winTitle}
		if cwd != "" {
			args = append(args, "-c", cwd)
		}
		args = append(args, shellBin, "-lc", guardedScript)
		cmd := exec.Command("tmux", args...)
		cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
		return cmd.Start()

	case "popup":
		// Replace current popup process
		return syscall.Exec(shellBin, []string{shellBin, "-lc", guardedScript}, os.Environ())

	default:
		return syscall.Exec(shellBin, []string{shellBin, "-lc", guardedScript}, os.Environ())
	}
}

// parseArgs parses a command string into slice of arguments respecting quotes.
func parseArgs(cmdStr string) []string {
	var args []string
	var current strings.Builder
	inSingle := false
	inDouble := false
	escaped := false

	for i := 0; i < len(cmdStr); i++ {
		c := cmdStr[i]

		if escaped {
			current.WriteByte(c)
			escaped = false
			continue
		}

		if c == '\\' {
			escaped = true
			continue
		}

		if c == '\'' && !inDouble {
			inSingle = !inSingle
			continue
		}

		if c == '"' && !inSingle {
			inDouble = !inDouble
			continue
		}

		if (c == ' ' || c == '\t') && !inSingle && !inDouble {
			if current.Len() > 0 {
				args = append(args, current.String())
				current.Reset()
			}
			continue
		}

		current.WriteByte(c)
	}

	if current.Len() > 0 {
		args = append(args, current.String())
	}

	return args
}
