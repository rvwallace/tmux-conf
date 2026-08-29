package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"
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
	if _, err := exec.LookPath("xsel"); err == nil {
		cmd := exec.Command("xsel", "--clipboard", "--input")
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

// FormatForShell prepares an action string for safe shell execution.
// If originalTarget is "tmux", it splits chained commands and ensures each has a "tmux " prefix.
func FormatForShell(action string, originalTarget string) string {
	if originalTarget == "tmux" {
		subCmds := SplitTmuxCommands(action)
		var formatted []string
		for _, sc := range subCmds {
			trimmed := strings.TrimSpace(sc)
			if trimmed == "" {
				continue
			}
			if !strings.HasPrefix(trimmed, "tmux ") && trimmed != "tmux" {
				formatted = append(formatted, "tmux "+trimmed)
			} else {
				formatted = append(formatted, trimmed)
			}
		}
		return strings.Join(formatted, " ; ")
	}
	return action
}

// BuildGuardedShellScript generates an interactive error-guard shell wrapper.
func BuildGuardedShellScript(cmdStr string, title string, persistShell bool) string {
	if title == "" {
		title = cmdStr
	}
	quotedTitle := QuoteShellSingle(title)

	if persistShell {
		return fmt.Sprintf(`
%s
_status=$?
if [ "$_status" -ne 0 ]; then
  printf '\n\033[1;31m✖ Command failed with exit code %%d: %%s\033[0m\n' "$_status" %s
fi
exec "${SHELL:-/bin/zsh}"
`, cmdStr, quotedTitle)
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

// SplitTmuxCommands splits a chained tmux command string by unescaped semicolons outside of quotes.
func SplitTmuxCommands(s string) []string {
	var parts []string
	var current strings.Builder
	inSingle := false
	inDouble := false
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

		if c == '\'' && !inDouble {
			inSingle = !inSingle
			current.WriteByte('\'')
			continue
		}

		if c == '"' && !inSingle {
			inDouble = !inDouble
			current.WriteByte('"')
			continue
		}

		if c == ';' && !inSingle && !inDouble {
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
		if expandedAction == "show-messages" {
			time.Sleep(60 * time.Millisecond)
			homeDir, _ := os.UserHomeDir()
			omniBin := filepath.Join(homeDir, ".config/tmux/scripts/tmux-omni")
			cmd := exec.Command("tmux", "display-popup", "-E", "-w", "80%", "-h", "80%", omniBin, "--messages")
			cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
			return cmd.Start()
		}

		time.Sleep(60 * time.Millisecond)
		parts := SplitTmuxCommands(expandedAction)
		for _, part := range parts {
			args := parseArgs(part)
			cmd := exec.Command("tmux", args...)
			cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
			if err := cmd.Run(); err != nil {
				_ = exec.Command("tmux", "display-message", fmt.Sprintf("✖ Execution failed: %v", err)).Run()
				return err
			}
		}
		if !strings.Contains(expandedAction, "display-message") {
			_ = exec.Command("tmux", "display-message", fmt.Sprintf("✔ Executed: %s", expandedAction)).Run()
		}
		return nil
	}

	// 2. Send keys to target pane
	if target == "send" || target == "send_keys" || target == "send-keys" {
		textToSend := FormatForShell(expandedAction, originalTarget)

		// Brief delay for the popup to close and target pane terminal to settle
		time.Sleep(60 * time.Millisecond)

		// Use a temporary tmux paste buffer for atomic paste to avoid dropped characters from pty redraw
		bufName := fmt.Sprintf("omni-send-%d", time.Now().UnixNano())
		if out, err := exec.Command("tmux", "set-buffer", "-b", bufName, "--", textToSend).CombinedOutput(); err != nil {
			errMsg := fmt.Sprintf("✖ Failed to set buffer: %s (%v)", strings.TrimSpace(string(out)), err)
			_ = exec.Command("tmux", "display-message", errMsg).Run()
			return fmt.Errorf("set-buffer failed: %s (%w)", strings.TrimSpace(string(out)), err)
		}

		var pasteArgs []string
		pasteArgs = append(pasteArgs, "paste-buffer", "-d", "-b", bufName)
		if cleanPaneID != "" {
			pasteArgs = append(pasteArgs, "-t", cleanPaneID)
		}
		if out, err := exec.Command("tmux", pasteArgs...).CombinedOutput(); err != nil {
			_ = exec.Command("tmux", "delete-buffer", "-b", bufName).Run()
			errMsg := fmt.Sprintf("✖ Failed to paste buffer into %s: %s (%v)", cleanPaneID, strings.TrimSpace(string(out)), err)
			_ = exec.Command("tmux", "display-message", errMsg).Run()
			return fmt.Errorf("paste-buffer failed: %s (%w)", strings.TrimSpace(string(out)), err)
		}

		_ = exec.Command("tmux", "display-message", fmt.Sprintf("󰍡 Inserted into %s: %s", cleanPaneID, textToSend)).Run()
		return nil
	}

	// 3. Shell wrapper commands (splits, windows, popups)
	formattedAction := FormatForShell(expandedAction, originalTarget)
	guardedScript := BuildGuardedShellScript(formattedAction, title, persistShell)

	time.Sleep(60 * time.Millisecond)

	switch target {
	case "split-h", "split_h":
		args := []string{"split-window", "-h", "-p", "40"}
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

	case "split-v", "split_v":
		args := []string{"split-window", "-v", "-p", "40"}
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

	case "popup", "popup-shell", "popup_shell":
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
