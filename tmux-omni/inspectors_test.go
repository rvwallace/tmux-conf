package main

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestIsInspectorCommand(t *testing.T) {
	tests := []struct {
		action   string
		expected string
		isInsp   bool
	}{
		{"~/.config/tmux/scripts/tmux-omni --env", "env", true},
		{"~/.config/tmux/scripts/tmux-omni --options", "options", true},
		{"~/.config/tmux/scripts/tmux-omni --buffers", "buffers", true},
		{"~/.config/tmux/scripts/tmux-omni --messages", "messages", true},
		{"~/.config/tmux/scripts/tmux-omni --commands", "commands", true},
		{"~/.config/tmux/scripts/tmux-omni --keys", "keys", true},
		{"~/.config/tmux/scripts/tmux-omni --clients", "clients", true},
		{"tmux-omni -k", "keys", true},
		{"tmux-omni -E", "env", true},
		{"tmux-omni -O", "options", true},
		{"tmux-omni -B", "buffers", true},
		{"tmux-omni -M", "messages", true},
		{"tmux-omni -C", "commands", true},
		// Non-inspectors
		{"lazygit", "", false},
		{"btop", "", false},
		{"tmux-snaglord", "", false},
		{"~/.config/tmux/scripts/tmux-omni ai ask \"#{pane_id}\"", "", false},
		{"~/.config/tmux/scripts/tmux-omni --search", "", false},
		{"onefetch ; printf '\\nPress Enter to close' ; read -r _", "", false},
	}

	for _, tt := range tests {
		gotType, gotOk := IsInspectorCommand(tt.action)
		if gotOk != tt.isInsp {
			t.Errorf("IsInspectorCommand(%q) ok = %v, want %v", tt.action, gotOk, tt.isInsp)
		}
		if gotType != tt.expected {
			t.Errorf("IsInspectorCommand(%q) type = %q, want %q", tt.action, gotType, tt.expected)
		}
	}
}

func TestCreateInspector(t *testing.T) {
	types := []string{"keys", "commands", "options", "env", "buffers", "messages", "clients"}
	for _, itype := range types {
		insp := CreateInspector(itype)
		if insp.Title == "" {
			t.Errorf("CreateInspector(%q) returned empty title", itype)
		}
		if insp.TextInput.Prompt == "" {
			t.Errorf("CreateInspector(%q) text input uninitialized", itype)
		}
	}
}

func TestInspectStatusMsgTimeout(t *testing.T) {
	insp := NewInspectModel("Test", "󰘳", "Col1", "Col2", "Col3", "Col4", nil, 10, 10, 10)
	if insp.StatusMsg != "" {
		t.Errorf("Initial StatusMsg should be empty, got %q", insp.StatusMsg)
	}

	cmd := insp.SetStatus("Copied to clipboard: test")
	if cmd == nil {
		t.Fatalf("SetStatus should return a tick command")
	}
	if insp.StatusMsg != "Copied to clipboard: test" {
		t.Errorf("StatusMsg not set, got %q", insp.StatusMsg)
	}
	firstID := insp.StatusMsgID

	// Simulate older timeout arriving after a newer status was set
	insp.SetStatus("Newer message")
	secondID := insp.StatusMsgID

	app := InspectorAppModel{Model: insp}

	// Stale message ID should NOT clear the newer status
	updatedModel, _ := app.Update(ClearInspectStatusMsg{ID: firstID})
	app = updatedModel.(InspectorAppModel)
	if app.Model.StatusMsg != "Newer message" {
		t.Errorf("Stale ClearInspectStatusMsg cleared newer status, got %q", app.Model.StatusMsg)
	}

	// Current message ID SHOULD clear the status
	updatedModel, _ = app.Update(ClearInspectStatusMsg{ID: secondID})
	app = updatedModel.(InspectorAppModel)
	if app.Model.StatusMsg != "" {
		t.Errorf("Matching ClearInspectStatusMsg failed to clear status, got %q", app.Model.StatusMsg)
	}
}

func TestAppModelInspectorTransitions(t *testing.T) {
	cfg := &Config{
		Title: "Tmux Menu",
		Items: []MenuItem{
			{
				Key:   "o",
				Title: "Options",
				Items: []MenuItem{
					{
						Key:    "e",
						Title:  "Show Environment",
						Action: "~/.config/tmux/scripts/tmux-omni --env",
						Target: "popup",
					},
				},
			},
		},
	}
	flat := FlattenCommands(cfg.Items, nil, "")

	// 1. Test WhichKey -> Inspector -> WhichKey via Backspace on empty query
	app := InitialModel(cfg, flat, false, "%0", "")
	if app.Mode != ModeWhichKey {
		t.Fatalf("Expected ModeWhichKey, got %v", app.Mode)
	}

	// Navigate into 'o'
	m, _ := app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'o'}})
	app = m.(AppModel)

	// Press 'e' (Environment inspector)
	m, _ = app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	app = m.(AppModel)

	if app.Mode != ModeInspector {
		t.Fatalf("Expected ModeInspector after selecting environment, got %v", app.Mode)
	}
	if app.Inspector.Title != "Environment" {
		t.Fatalf("Expected Environment inspector, got %q", app.Inspector.Title)
	}

	// Press Backspace when search input is empty -> should return to ModeWhichKey
	m, _ = app.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	app = m.(AppModel)

	if app.Mode != ModeWhichKey {
		t.Fatalf("Expected Backspace on empty search to return to ModeWhichKey, got %v", app.Mode)
	}

	// 2. Test Palette -> Inspector -> Palette via Esc
	app = InitialModel(cfg, flat, true, "%0", "")
	if app.Mode != ModePalette {
		t.Fatalf("Expected ModePalette, got %v", app.Mode)
	}

	// Select the environment command with enter
	m, _ = app.Update(tea.KeyMsg{Type: tea.KeyEnter})
	app = m.(AppModel)

	if app.Mode != ModeInspector {
		t.Fatalf("Expected ModeInspector from Palette, got %v", app.Mode)
	}

	// Press Esc with empty search -> should return to ModePalette
	m, _ = app.Update(tea.KeyMsg{Type: tea.KeyEsc})
	app = m.(AppModel)

	if app.Mode != ModePalette {
		t.Fatalf("Expected Esc on empty search to return to ModePalette, got %v", app.Mode)
	}
}

func TestStandaloneInspectorExit(t *testing.T) {
	app := InitialModel(nil, nil, false, "%0", "keys")
	if app.Mode != ModeInspector {
		t.Fatalf("Expected ModeInspector, got %v", app.Mode)
	}
	if !app.IsStandaloneInspector {
		t.Fatalf("Expected IsStandaloneInspector = true")
	}

	// Pressing esc on standalone inspector should quit
	_, cmd := app.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if cmd == nil {
		t.Fatalf("Expected tea.Quit command on Esc in standalone inspector")
	}

	// Pressing backspace on empty search in standalone inspector should quit
	_, cmd = app.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	if cmd == nil {
		t.Fatalf("Expected tea.Quit command on Backspace in standalone inspector")
	}
}

func TestGetItemActionOptions(t *testing.T) {
	// 1. Global environment variable
	globalItem := InspectItem{
		Col1:      "API_KEY",
		Col2:      "Global",
		Col3:      "secret-token-123",
		RawCopy:   "export API_KEY=\"secret-token-123\"",
		ActionCmd: "command-prompt -I 'set-environment -g API_KEY secret-token-123'",
	}
	opts := GetItemActionOptions(globalItem, "Environment")
	if len(opts) < 5 {
		t.Fatalf("Expected at least 5 action options for Global env, got %d", len(opts))
	}

	keys := make(map[string]string)
	for _, opt := range opts {
		keys[opt.Key] = opt.Payload
	}

	if keys["v"] != "secret-token-123" {
		t.Errorf("Expected 'v' to copy value 'secret-token-123', got %q", keys["v"])
	}
	if keys["n"] != "API_KEY" {
		t.Errorf("Expected 'n' to copy name 'API_KEY', got %q", keys["n"])
	}
	if keys["e"] != "export API_KEY=\"secret-token-123\"" {
		t.Errorf("Expected 'e' to copy export, got %q", keys["e"])
	}
	if keys["s"] != "set-environment -g API_KEY \"secret-token-123\"" {
		t.Errorf("Expected 's' to copy tmux set, got %q", keys["s"])
	}

	// 2. Unset environment variable
	unsetItem := InspectItem{
		Col1:      "OLD_VAR",
		Col2:      "Unset",
		Col3:      "(not set in global env)",
		RawCopy:   "unset OLD_VAR",
		ActionCmd: "command-prompt -I 'set-environment -g OLD_VAR '",
	}
	unsetOpts := GetItemActionOptions(unsetItem, "Environment")
	if len(unsetOpts) < 3 {
		t.Fatalf("Expected at least 3 action options for Unset env, got %d", len(unsetOpts))
	}
}

func TestEnvironmentInspectorFilteringAndShortcuts(t *testing.T) {
	items := []InspectItem{
		{
			Col1:      "ANTHROPIC_API_KEY",
			Col2:      "Global",
			Col3:      "sk-ant-12345",
			RawCopy:   "export ANTHROPIC_API_KEY=\"sk-ant-12345\"",
			ActionCmd: "command-prompt -I 'set-environment -g ANTHROPIC_API_KEY sk-ant-12345'",
			SearchText: "anthropic_api_key global sk-ant-12345",
		},
		{
			Col1:      "CLOUDFLARE_API_TOKEN",
			Col2:      "Global",
			Col3:      "token-xyz-789",
			RawCopy:   "export CLOUDFLARE_API_TOKEN=\"token-xyz-789\"",
			ActionCmd: "command-prompt -I 'set-environment -g CLOUDFLARE_API_TOKEN token-xyz-789'",
			SearchText: "cloudflare_api_token global token-xyz-789",
		},
	}
	insp := NewInspectModel("Environment", "󰈞", "Variable", "Scope", "Value", "", items, 28, 10, 0)
	insp.AllowCopy = true
	insp.AllowExecute = true
	insp.ExecuteLabel = "Edit"

	app := InspectorAppModel{Model: insp}

	// 1. Typing 'a' filters for anthropic and does NOT open action modal
	m, _ := app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	app = m.(InspectorAppModel)
	if app.Model.ShowActionPicker {
		t.Fatalf("Typing 'a' should not trigger Action Picker modal")
	}
	if app.Model.TextInput.Value() != "a" {
		t.Fatalf("Expected textinput value 'a', got %q", app.Model.TextInput.Value())
	}
	if len(app.Model.Filtered) != 2 {
		t.Fatalf("Expected 2 matches for 'a', got %d", len(app.Model.Filtered))
	}

	// Type 'n' -> "an"
	m, _ = app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	app = m.(InspectorAppModel)
	if app.Model.TextInput.Value() != "an" {
		t.Fatalf("Expected textinput value 'an', got %q", app.Model.TextInput.Value())
	}
	if len(app.Model.Filtered) != 1 || app.Model.Filtered[0].Col1 != "ANTHROPIC_API_KEY" {
		t.Fatalf("Expected only ANTHROPIC_API_KEY to match 'an'")
	}

	// 2. Ctrl+y copies the selected item's export to clipboard
	m, cmd := app.Update(tea.KeyMsg{Type: tea.KeyCtrlY})
	app = m.(InspectorAppModel)
	if cmd == nil {
		t.Errorf("Expected status timer cmd on Ctrl+y")
	}
	if !strings.HasPrefix(app.Model.StatusMsg, "Copied to clipboard: export ANTHROPIC_API_KEY=") {
		t.Errorf("Expected export copy message, got %q", app.Model.StatusMsg)
	}
}

func TestActionPickerModal(t *testing.T) {
	items := []InspectItem{
		{
			Col1:      "PORT",
			Col2:      "Global",
			Col3:      "8080",
			RawCopy:   "export PORT=\"8080\"",
			ActionCmd: "command-prompt -I 'set-environment -g PORT 8080'",
			SearchText: "port global 8080",
		},
	}
	insp := NewInspectModel("Environment", "󰈞", "Variable", "Scope", "Value", "", items, 28, 10, 0)
	insp.AllowCopy = true
	insp.AllowExecute = true

	app := InspectorAppModel{Model: insp}

	// Press Ctrl+a (or Ctrl+o) to open Action Picker modal
	m, _ := app.Update(tea.KeyMsg{Type: tea.KeyCtrlA})
	app = m.(InspectorAppModel)
	if !app.Model.ShowActionPicker {
		t.Fatalf("Expected ShowActionPicker = true after pressing Ctrl+a")
	}
	if len(app.Model.ActionOptions) == 0 {
		t.Fatalf("Expected ActionOptions to be populated")
	}

	// Verify modal view renders
	view := app.Model.RenderActionPickerModal()
	if view == "" {
		t.Errorf("RenderActionPickerModal returned empty string")
	}

	// Inside modal: press 'v' to select Value action
	m, cmd := app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'v'}})
	app = m.(InspectorAppModel)
	if app.Model.ShowActionPicker {
		t.Errorf("Expected ShowActionPicker = false after choosing option")
	}
	if cmd == nil {
		t.Errorf("Expected status timer cmd after executing action")
	}
	if app.Model.StatusMsg != "Copied copy value only: 8080" {
		t.Errorf("Expected 'Copied copy value only: 8080', got %q", app.Model.StatusMsg)
	}

	// Reopen with Ctrl+a and test Esc to cancel
	m, _ = app.Update(tea.KeyMsg{Type: tea.KeyCtrlA})
	app = m.(InspectorAppModel)
	if !app.Model.ShowActionPicker {
		t.Fatalf("Expected ShowActionPicker = true after pressing Ctrl+a")
	}

	m, _ = app.Update(tea.KeyMsg{Type: tea.KeyEsc})
	app = m.(InspectorAppModel)
	if app.Model.ShowActionPicker {
		t.Errorf("Expected Esc to dismiss action picker modal")
	}
}

