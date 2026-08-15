package main

import (
	"strings"
	"testing"
	"time"
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
