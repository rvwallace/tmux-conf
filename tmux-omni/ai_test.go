package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewAIModel(t *testing.T) {
	modes := []string{
		AIModeAsk,
		AIModeError,
		AIModeFix,
		AIModeSummarize,
		AIModeCommand,
		AIModeExplain,
		AIModeExplainCopy,
	}

	for _, mode := range modes {
		m := NewAIModel(mode, "%1")
		if m.Mode != mode {
			t.Errorf("expected Mode %q, got %q", mode, m.Mode)
		}
		if m.PaneID != "%1" {
			t.Errorf("expected PaneID %%1, got %q", m.PaneID)
		}
		if !strings.HasPrefix(m.SessionID, "tmux-ai-") {
			t.Errorf("expected SessionID prefix 'tmux-ai-', got %q", m.SessionID)
		}
		if m.DepthIndex != 1 {
			t.Errorf("expected default DepthIndex 1, got %d", m.DepthIndex)
		}
	}
}

func TestAICardsRendering(t *testing.T) {
	m := NewAIModel(AIModeAsk, "%1")

	userCard := m.renderUserCard("How do I list files?")
	if !strings.Contains(userCard, "You") || !strings.Contains(userCard, "How do I list files?") {
		t.Errorf("userCard missing expected text: %s", userCard)
	}

	asstCard := m.renderAssistantCard("Use `ls -la`")
	if !strings.Contains(asstCard, "Assistant") || !strings.Contains(asstCard, "ls -la") {
		t.Errorf("asstCard missing expected text: %s", asstCard)
	}

	errCard := m.renderErrorCard("Command failed")
	if !strings.Contains(errCard, "Error") || !strings.Contains(errCard, "Command failed") {
		t.Errorf("errCard missing expected text: %s", errCard)
	}

	refineCard := m.renderRefinementCard("use brew instead")
	if !strings.Contains(refineCard, "Refinement") || !strings.Contains(refineCard, "use brew instead") {
		t.Errorf("refineCard missing expected text: %s", refineCard)
	}
}

func TestAIGetLatestCopyableText(t *testing.T) {
	m := NewAIModel(AIModeAsk, "%1")
	m.Messages = []ChatMessage{
		{Role: "user", Content: "Hello", Timestamp: time.Now()},
		{Role: "assistant", Content: "Hi there!", Timestamp: time.Now()},
	}

	if m.getLatestCopyableText() != "Hi there!" {
		t.Errorf("expected 'Hi there!', got %q", m.getLatestCopyableText())
	}

	mCmd := NewAIModel(AIModeCommand, "%1")
	mCmd.CandidateCommand = "git status"
	if mCmd.getLatestCopyableText() != "git status" {
		t.Errorf("expected 'git status', got %q", mCmd.getLatestCopyableText())
	}
}

func TestAIHelpModal(t *testing.T) {
	m := NewAIModel(AIModeAsk, "%1")
	m.Width = 80
	m.Height = 24

	rendered := m.renderHelpModal("bg")
	if !strings.Contains(rendered, "AI Assistant Shortcuts") {
		t.Errorf("help modal missing title: %s", rendered)
	}
	if !strings.Contains(rendered, "<Enter>") || !strings.Contains(rendered, "<Tab>") {
		t.Errorf("help modal missing key hints: %s", rendered)
	}
}

func TestExtractCodeBlocks(t *testing.T) {
	markdown := "Here is a step to run:\n\n```bash\ngit status\n```\n\nAnd next:\n```\ngit log -n 5\n```\n"
	blocks := extractCodeBlocks(markdown)
	if len(blocks) != 2 {
		t.Fatalf("expected 2 code blocks, got %d: %v", len(blocks), blocks)
	}
	if blocks[0].Content != "git status" || blocks[0].Language != "bash" {
		t.Errorf("expected 'git status' (bash), got %+v", blocks[0])
	}
	if blocks[1].Content != "git log -n 5" || blocks[1].Language != "code" {
		t.Errorf("expected 'git log -n 5' (code), got %+v", blocks[1])
	}
}

func TestBlockPickerModal(t *testing.T) {
	m := NewAIModel(AIModeAsk, "%1")
	m.Width = 80
	m.Height = 24
	m.CodeBlocks = []CodeBlock{
		{Index: 1, Language: "bash", Content: "echo 1", Preview: "echo 1"},
		{Index: 2, Language: "python", Content: "print(2)", Preview: "print(2)"},
	}

	rendered := m.renderBlockPickerModal("bg")
	if !strings.Contains(rendered, "Select Code Block") {
		t.Errorf("missing title in block picker: %s", rendered)
	}
	if !strings.Contains(rendered, "[bash]") || !strings.Contains(rendered, "[python]") {
		t.Errorf("missing language tags in block picker: %s", rendered)
	}
}

func TestGatherSlashContext(t *testing.T) {
	prompt := "Check this /env and /refresh"
	cleaned, extra := gatherSlashContext(prompt, "/tmp")
	if strings.Contains(cleaned, "/env") {
		t.Errorf("expected /env stripped from prompt, got %q", cleaned)
	}
	if !strings.Contains(extra, "Environment Info") {
		t.Errorf("expected extra context to contain Environment Info, got %q", extra)
	}
}

func TestGatherSlashContextTree(t *testing.T) {
	tmpDir := t.TempDir()
	file1 := filepath.Join(tmpDir, "file_alpha.txt")
	if err := os.WriteFile(file1, []byte("alpha"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}
	subDir := filepath.Join(tmpDir, "nested_dir")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("failed to create nested dir: %v", err)
	}
	file2 := filepath.Join(subDir, "file_beta.txt")
	if err := os.WriteFile(file2, []byte("beta"), 0644); err != nil {
		t.Fatalf("failed to create nested file: %v", err)
	}

	cleaned, extra := gatherSlashContext("Explain the tree structure /tree please", tmpDir)
	if strings.Contains(cleaned, "/tree") {
		t.Errorf("expected /tree removed from prompt, got %q", cleaned)
	}
	if !strings.Contains(extra, "Directory Tree (Depth 2)") {
		t.Errorf("expected extraContext to contain Directory Tree header, got %q", extra)
	}
	if !strings.Contains(extra, "file_alpha.txt") || !strings.Contains(extra, "nested_dir") {
		t.Errorf("expected tree to list files in tmpDir %s, got %q", tmpDir, extra)
	}

	// Test with empty panePath
	cleanedEmpty, _ := gatherSlashContext("Check /tree", "")
	if strings.Contains(cleanedEmpty, "/tree") {
		t.Errorf("expected /tree removed with empty panePath, got %q", cleanedEmpty)
	}

	// Test with nonexistent directory
	cleanedNonExistent, extraNonExistent := gatherSlashContext("Check /tree", "/path/to/definitely/nonexistent/directory")
	if strings.Contains(cleanedNonExistent, "/tree") {
		t.Errorf("expected /tree removed with nonexistent panePath, got %q", cleanedNonExistent)
	}
	if strings.Contains(extraNonExistent, "Directory Tree (Depth 2)") {
		t.Errorf("expected no directory tree for nonexistent path, got %q", extraNonExistent)
	}
}

func TestModelPickerModal(t *testing.T) {
	m := NewAIModel(AIModeAsk, "%1")
	m.Width = 80
	m.Height = 24
	m.AvailableModels = []string{"default", "claude-3-5-sonnet", "gpt-4o"}

	rendered := m.renderModelPickerModal("bg")
	if !strings.Contains(rendered, "Select AI Model") {
		t.Errorf("missing title in model picker: %s", rendered)
	}
	if !strings.Contains(rendered, "claude-3-5-sonnet") {
		t.Errorf("missing model name in model picker: %s", rendered)
	}
}

func TestBuildAIPrompt(t *testing.T) {
	prompt, err := buildAIPrompt(AIModeAsk, "how to list files", "/tmp", "zsh", "ls output", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(prompt, "how to list files") || !strings.Contains(prompt, "BEGIN PANE CONTEXT") {
		t.Errorf("prompt missing expected content: %s", prompt)
	}

	cmdPrompt, err := buildAIPrompt(AIModeCommand, "list git commits", "/tmp", "zsh", "", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(cmdPrompt, "Generate exactly one command") || !strings.Contains(cmdPrompt, "list git commits") {
		t.Errorf("command prompt missing expected content: %s", cmdPrompt)
	}

	sumPrompt, err := buildAIPrompt(AIModeSummarize, "", "/tmp", "zsh", "build succeeded", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(sumPrompt, "Summarize the recent pane output") {
		t.Errorf("summarize prompt missing expected content: %s", sumPrompt)
	}
}

func TestAIMissingAIChat(t *testing.T) {
	oldCheck := checkLookPath
	defer func() { checkLookPath = oldCheck }()

	checkLookPath = func(file string) (string, error) {
		return "", exec.ErrNotFound
	}

	m := NewAIModel(AIModeError, "%1")
	if !m.IsBusy {
		t.Errorf("expected auto-run mode to start as busy")
	}

	cmd := m.Init()
	if cmd == nil {
		t.Fatalf("expected non-nil cmd from Init when aichat is missing")
	}

	msg := cmd()
	resMsg, ok := msg.(aiResultMsg)
	if !ok {
		t.Fatalf("expected aiResultMsg, got: %T", msg)
	}
	if resMsg.err == nil || !strings.Contains(resMsg.err.Error(), "aichat is not in PATH") {
		t.Fatalf("expected missing aichat error, got: %v", resMsg.err)
	}

	newModel, _ := m.Update(msg)
	aiM, ok := newModel.(AIModel)
	if !ok {
		t.Fatalf("expected AIModel from Update, got: %T", newModel)
	}

	if aiM.IsBusy {
		t.Errorf("expected IsBusy to be false after error received, got true")
	}
	if len(aiM.Messages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(aiM.Messages))
	}
	if aiM.Messages[0].Role != "error" {
		t.Errorf("expected message role 'error', got %q", aiM.Messages[0].Role)
	}
	if !strings.Contains(aiM.Messages[0].Content, "aichat is not in PATH") {
		t.Errorf("expected message content to mention 'aichat is not in PATH', got %q", aiM.Messages[0].Content)
	}

	view := aiM.Viewport.View()
	if !strings.Contains(view, "aichat is not in PATH") {
		t.Errorf("expected viewport to render error card, got: %s", view)
	}
}

func TestAIModel_InsertCandidateCommand(t *testing.T) {
	_, logFile := createFakeTmux(t)

	m := NewAIModel(AIModeCommand, "%1")

	// 1. Empty command
	errEmpty := m.insertCandidateCommand("")
	if errEmpty == nil {
		t.Errorf("expected error on empty command")
	}

	// 2. Multiline command
	errMultiline := m.insertCandidateCommand("echo 1\necho 2")
	if errMultiline == nil {
		t.Errorf("expected error on multiline command")
	}
	if !strings.Contains(m.ToastMsg, "Cannot insert multiline") {
		t.Errorf("expected toast for multiline, got: %q", m.ToastMsg)
	}

	// 3. Fenced command
	errFenced := m.insertCandidateCommand("```bash\necho 1\n```")
	if errFenced == nil {
		t.Errorf("expected error on fenced command")
	}

	// 4. Set-buffer failure
	t.Setenv("TMUX_TEST_FAIL_SET_BUFFER", "1")
	t.Setenv("TMUX_TEST_FAIL_PASTE_BUFFER", "0")
	errSetBuf := m.insertCandidateCommand("echo 1")
	if errSetBuf == nil {
		t.Errorf("expected error when set-buffer fails")
	}
	if !strings.Contains(m.ToastMsg, "Failed to set buffer") {
		t.Errorf("expected toast for set-buffer failure, got: %q", m.ToastMsg)
	}

	// 5. Paste-buffer failure
	t.Setenv("TMUX_TEST_FAIL_SET_BUFFER", "0")
	t.Setenv("TMUX_TEST_FAIL_PASTE_BUFFER", "1")
	errPasteBuf := m.insertCandidateCommand("echo 1")
	if errPasteBuf == nil {
		t.Errorf("expected error when paste-buffer fails")
	}
	if !strings.Contains(m.ToastMsg, "Failed to paste to pane") {
		t.Errorf("expected toast for paste-buffer failure, got: %q", m.ToastMsg)
	}

	data, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}
	logContent := string(data)
	if !strings.Contains(logContent, "delete-buffer -b") {
		t.Errorf("expected buffer cleanup delete-buffer call on paste error, log: %s", logContent)
	}

	// 6. Success
	t.Setenv("TMUX_TEST_FAIL_SET_BUFFER", "0")
	t.Setenv("TMUX_TEST_FAIL_PASTE_BUFFER", "0")
	errSuccess := m.insertCandidateCommand("echo 1")
	if errSuccess != nil {
		t.Errorf("expected success, got: %v", errSuccess)
	}
}

func TestAIModel_UpdateSendErrorPreventsQuit(t *testing.T) {
	createFakeTmux(t)
	t.Setenv("TMUX_TEST_FAIL_SET_BUFFER", "1")
	t.Setenv("TMUX_TEST_FAIL_PASTE_BUFFER", "0")

	// Normal mode 's' key with candidate command
	m := NewAIModel(AIModeCommand, "%1")
	m.CandidateCommand = "echo 1"
	m.FocusOnInput = false

	newModel, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	aiM := newModel.(AIModel)
	if cmd == nil {
		t.Errorf("expected toast cmd when insert fails")
	}
	if !strings.Contains(aiM.ToastMsg, "Failed to set buffer") {
		t.Errorf("expected toast msg for failed insert, got: %q", aiM.ToastMsg)
	}

	// Insert mode 'enter' key with empty input and candidate command
	m2 := NewAIModel(AIModeCommand, "%1")
	m2.CandidateCommand = "echo 1"
	m2.FocusOnInput = true
	m2.Input.SetValue("")

	newModel2, cmd2 := m2.Update(tea.KeyMsg{Type: tea.KeyEnter})
	aiM2 := newModel2.(AIModel)
	if cmd2 == nil {
		t.Errorf("expected toast cmd when enter insert fails")
	}
	if !strings.Contains(aiM2.ToastMsg, "Failed to set buffer") {
		t.Errorf("expected toast msg for failed enter insert, got: %q", aiM2.ToastMsg)
	}

	// Block picker overlay 's' key
	m3 := NewAIModel(AIModeAsk, "%1")
	m3.ShowBlockPicker = true
	m3.CodeBlocks = []CodeBlock{{Index: 1, Content: "echo 1"}}
	m3.BlockPickerCursor = 0

	newModel3, cmd3 := m3.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	aiM3 := newModel3.(AIModel)
	if cmd3 == nil {
		t.Errorf("expected toast cmd when block picker insert fails")
	}
	if aiM3.ShowBlockPicker {
		t.Errorf("expected block picker to close after action")
	}
	if !strings.Contains(aiM3.ToastMsg, "Failed to set buffer") {
		t.Errorf("expected toast msg for failed block picker insert, got: %q", aiM3.ToastMsg)
	}
}
