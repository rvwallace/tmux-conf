package main

import (
	"strings"
	"testing"
)

func TestSplitTmuxCommands(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{
			input:    "select-pane -L",
			expected: []string{"select-pane -L"},
		},
		{
			input:    "set-window-option synchronize-panes ; refresh-client -S",
			expected: []string{"set-window-option synchronize-panes", "refresh-client -S"},
		},
		{
			input:    `bind-key -n C-a send-prefix \; display-message "hello"`,
			expected: []string{`bind-key -n C-a send-prefix ; display-message "hello"`},
		},
	}

	for _, tt := range tests {
		got := SplitTmuxCommands(tt.input)
		if len(got) != len(tt.expected) {
			t.Errorf("input: %s -> expected %d parts, got %d: %#v", tt.input, len(tt.expected), len(got), got)
			continue
		}
		for i := range got {
			if got[i] != tt.expected[i] {
				t.Errorf("input: %s part %d -> expected %s, got %s", tt.input, i, tt.expected[i], got[i])
			}
		}
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

func TestParseArgs(t *testing.T) {
	args := parseArgs(`display-popup -w "68%" -h "65%" 'echo "hello world"'`)
	if len(args) != 6 {
		t.Fatalf("expected 6 args, got %d: %#v", len(args), args)
	}
	if args[0] != "display-popup" || args[1] != "-w" || args[2] != "68%" || args[3] != "-h" || args[4] != "65%" || args[5] != "echo \"hello world\"" {
		t.Errorf("unexpected args parsed: %#v", args)
	}
}
