package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigLoadAndFlatten(t *testing.T) {
	testJSON := `{
		"title": "Test Menu",
		"items": [
			{
				"key": "p",
				"title": "Panes",
				"icon": "󰓦",
				"description": "pane actions",
				"items": [
					{
						"key": "h",
						"title": "Select Left",
						"icon": "󰁍",
						"description": "focus left",
						"action": "select-pane -L",
						"target": "tmux"
					},
					{
						"key": "v",
						"title": "Split Horizontal",
						"icon": "󰘬",
						"description": "split pane side",
						"action": "split-window -h",
						"target": "tmux"
					}
				]
			},
			{
				"key": "w",
				"title": "New Window",
				"icon": "󰐕",
				"description": "create window",
				"action": "new-window",
				"target": "tmux"
			}
		]
	}`

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "config.json")
	if err := os.WriteFile(tmpFile, []byte(testJSON), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	cfg, err := LoadConfig(tmpFile)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if cfg.Title != "Test Menu" {
		t.Errorf("expected title 'Test Menu', got '%s'", cfg.Title)
	}

	if len(cfg.Items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(cfg.Items))
	}

	flat := FlattenCommands(cfg.Items, nil, "")
	if len(flat) != 3 {
		t.Fatalf("expected 3 flat commands, got %d", len(flat))
	}

	// Verify Select Left command
	cmd0 := flat[0]
	if cmd0.Title != "Select Left" {
		t.Errorf("expected title 'Select Left', got '%s'", cmd0.Title)
	}
	if cmd0.Category != "Panes" {
		t.Errorf("expected category 'Panes', got '%s'", cmd0.Category)
	}
	if cmd0.KeySeq != "p h" {
		t.Errorf("expected KeySeq 'p h', got '%s'", cmd0.KeySeq)
	}

	// Verify Root level command
	cmd2 := flat[2]
	if cmd2.Title != "New Window" {
		t.Errorf("expected title 'New Window', got '%s'", cmd2.Title)
	}
	if cmd2.Category != "Root" {
		t.Errorf("expected category 'Root', got '%s'", cmd2.Category)
	}
	if cmd2.KeySeq != "w" {
		t.Errorf("expected KeySeq 'w', got '%s'", cmd2.KeySeq)
	}
}
