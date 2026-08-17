package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSplitTmuxCommands(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "single command",
			input:    "select-pane -L",
			expected: []string{"select-pane -L"},
		},
		{
			name:     "chained unquoted commands",
			input:    "set-window-option synchronize-panes ; refresh-client -S",
			expected: []string{"set-window-option synchronize-panes", "refresh-client -S"},
		},
		{
			name:     "escaped semicolon",
			input:    `bind-key -n C-a send-prefix \; display-message "hello"`,
			expected: []string{`bind-key -n C-a send-prefix ; display-message "hello"`},
		},
		{
			name:     "real capture pane to file action",
			input:    `run-shell 'mkdir -p "$HOME/tmux-captures"; file="$HOME/tmux-captures/tmux-capture-$(date +%Y%m%d-%H%M%S).txt"; tmux capture-pane -pS - -t "#{pane_id}" > "$file"; tmux display-message "Captured pane to $file"'`,
			expected: []string{`run-shell 'mkdir -p "$HOME/tmux-captures"; file="$HOME/tmux-captures/tmux-capture-$(date +%Y%m%d-%H%M%S).txt"; tmux capture-pane -pS - -t "#{pane_id}" > "$file"; tmux display-message "Captured pane to $file"'`},
		},
		{
			name:     "single quotes with semicolon",
			input:    "display-message 'hello; world'",
			expected: []string{"display-message 'hello; world'"},
		},
		{
			name:     "double quotes with semicolon",
			input:    `display-message "hello; world"`,
			expected: []string{`display-message "hello; world"`},
		},
		{
			name:     "multiple quoted commands with unquoted separator",
			input:    `run-shell "echo 1; echo 2" ; display-message "finished"`,
			expected: []string{`run-shell "echo 1; echo 2"`, `display-message "finished"`},
		},
		{
			name:     "nested single in double quotes",
			input:    `set-option -g status on ; run-shell "echo '1;2'; echo 3" ; refresh-client`,
			expected: []string{"set-option -g status on", `run-shell "echo '1;2'; echo 3"`, "refresh-client"},
		},
		{
			name:     "escaped quote inside double quotes with semicolon",
			input:    `display-message "hello \"quoted; semicolon\" world" ; refresh-client`,
			expected: []string{`display-message "hello \"quoted; semicolon\" world"`, "refresh-client"},
		},
		{
			name:     "copy-mode binding with escaped semicolon",
			input:    `bind-key -T copy-mode-vi y send-keys -X copy-selection-and-cancel \; display-message "copied"`,
			expected: []string{`bind-key -T copy-mode-vi y send-keys -X copy-selection-and-cancel ; display-message "copied"`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SplitTmuxCommands(tt.input)
			if len(got) != len(tt.expected) {
				t.Fatalf("input: %s -> expected %d parts, got %d: %#v", tt.input, len(tt.expected), len(got), got)
			}
			for i := range got {
				if got[i] != tt.expected[i] {
					t.Errorf("part %d -> expected %s, got %s", i, tt.expected[i], got[i])
				}
			}
		})
	}
}

func TestBuildGuardedShellScript(t *testing.T) {
	script := BuildGuardedShellScript("ls -la", "List Directory", false)
	if !strings.Contains(script, "ls -la") {
		t.Errorf("expected script to contain command")
	}
	if !strings.Contains(script, "List Directory") {
		t.Errorf("expected script to contain title")
	}
	if !strings.Contains(script, "Dropping to debug shell") {
		t.Errorf("expected script to contain debug shell fallback")
	}

	persistScript := BuildGuardedShellScript("htop", "System Monitor", true)
	if !strings.Contains(persistScript, "exec \"${SHELL:-/bin/zsh}\"") {
		t.Errorf("expected persist script to exec user shell")
	}
}

func TestFormatForShell(t *testing.T) {
	tests := []struct {
		action         string
		originalTarget string
		expected       string
	}{
		{
			action:         "show-messages",
			originalTarget: "tmux",
			expected:       "tmux show-messages",
		},
		{
			action:         "tmux show-messages",
			originalTarget: "tmux",
			expected:       "tmux show-messages",
		},
		{
			action:         "set-window-option synchronize-panes ; refresh-client -S",
			originalTarget: "tmux",
			expected:       "tmux set-window-option synchronize-panes ; tmux refresh-client -S",
		},
		{
			action:         "lazygit",
			originalTarget: "popup",
			expected:       "lazygit",
		},
		{
			action:         `run-shell 'echo "hello; world"'`,
			originalTarget: "tmux",
			expected:       `tmux run-shell 'echo "hello; world"'`,
		},
	}

	for _, tt := range tests {
		got := FormatForShell(tt.action, tt.originalTarget)
		if got != tt.expected {
			t.Errorf("FormatForShell(%q, %q) = %q, expected %q", tt.action, tt.originalTarget, got, tt.expected)
		}
	}
}

func TestParseArgs(t *testing.T) {
	args := parseArgs(`display-popup -w "68%" -h "65%" 'echo "hello world"'`)
	if len(args) != 6 {
		t.Fatalf("expected 6 args, got %d: %#v", len(args), args)
	}
	if args[0] != "display-popup" || args[1] != "-w" || args[2] != "68%" || args[3] != "-h" || args[4] != "65%" || args[5] != "echo \"hello world\"" {
		t.Errorf("unexpected args parsed: %#v", args)
	}
}

func createFakeTmux(t *testing.T) (string, string) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "tmux_calls.log")
	script := fmt.Sprintf(`#!/bin/sh
echo "$@" >> %q
if [ "$TMUX_TEST_FAIL_SET_BUFFER" = "1" ] && [ "$1" = "set-buffer" ]; then
    echo "mock set-buffer error" >&2
    exit 1
fi
if [ "$TMUX_TEST_FAIL_PASTE_BUFFER" = "1" ] && [ "$1" = "paste-buffer" ]; then
    echo "mock paste-buffer error" >&2
    exit 1
fi
if [ "$1" = "display-message" ] && [ "$2" = "-p" ]; then
    if [ "$3" = "#{pane_id}" ]; then
        echo "%%1"
        exit 0
    fi
    if [ "$3" = "-t" ] && [ "$5" = "#{pane_current_path}" ]; then
        echo "/tmp"
        exit 0
    fi
fi
exit 0
`, logFile)
	tmuxPath := filepath.Join(tmpDir, "tmux")
	if err := os.WriteFile(tmuxPath, []byte(script), 0755); err != nil {
		t.Fatalf("failed to write fake tmux: %v", err)
	}
	t.Setenv("PATH", tmpDir+":"+os.Getenv("PATH"))
	return tmpDir, logFile
}

func TestRunTmuxTargetSend_Success(t *testing.T) {
	_, logFile := createFakeTmux(t)
	t.Setenv("TMUX_TEST_FAIL_SET_BUFFER", "0")
	t.Setenv("TMUX_TEST_FAIL_PASTE_BUFFER", "0")

	err := RunTmuxTarget("echo 'hello'", "send", "%1", "Send Command", false, "tmux")
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}

	data, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}
	logContent := string(data)
	if !strings.Contains(logContent, "set-buffer") {
		t.Errorf("log missing set-buffer call: %s", logContent)
	}
	if !strings.Contains(logContent, "paste-buffer -d -b") {
		t.Errorf("log missing paste-buffer call: %s", logContent)
	}
	if !strings.Contains(logContent, "Inserted into %1") {
		t.Errorf("log missing success display-message call: %s", logContent)
	}
}

func TestRunTmuxTargetSend_SetBufferError(t *testing.T) {
	_, logFile := createFakeTmux(t)
	t.Setenv("TMUX_TEST_FAIL_SET_BUFFER", "1")
	t.Setenv("TMUX_TEST_FAIL_PASTE_BUFFER", "0")

	err := RunTmuxTarget("echo 'hello'", "send", "%1", "Send Command", false, "tmux")
	if err == nil {
		t.Fatal("expected error when set-buffer fails, got nil")
	}
	if !strings.Contains(err.Error(), "set-buffer failed") {
		t.Errorf("expected 'set-buffer failed' in error, got: %v", err)
	}

	data, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}
	logContent := string(data)
	if strings.Contains(logContent, "Inserted into %1") {
		t.Errorf("should not display success when set-buffer fails: %s", logContent)
	}
	if !strings.Contains(logContent, "Failed to set buffer") {
		t.Errorf("log missing error display-message call: %s", logContent)
	}
}

func TestRunTmuxTargetSend_PasteBufferError(t *testing.T) {
	_, logFile := createFakeTmux(t)
	t.Setenv("TMUX_TEST_FAIL_SET_BUFFER", "0")
	t.Setenv("TMUX_TEST_FAIL_PASTE_BUFFER", "1")

	err := RunTmuxTarget("echo 'hello'", "send", "%1", "Send Command", false, "tmux")
	if err == nil {
		t.Fatal("expected error when paste-buffer fails, got nil")
	}
	if !strings.Contains(err.Error(), "paste-buffer failed") {
		t.Errorf("expected 'paste-buffer failed' in error, got: %v", err)
	}

	data, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}
	logContent := string(data)
	if strings.Contains(logContent, "Inserted into %1") {
		t.Errorf("should not display success when paste-buffer fails: %s", logContent)
	}
	if !strings.Contains(logContent, "delete-buffer -b") {
		t.Errorf("log missing buffer cleanup delete-buffer call on paste error: %s", logContent)
	}
	if !strings.Contains(logContent, "Failed to paste buffer") {
		t.Errorf("log missing error display-message call: %s", logContent)
	}
}
