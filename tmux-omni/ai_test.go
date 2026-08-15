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

func TestAICompactMode(t *testing.T) {
	mCmd := NewAIModel(AIModeCommand, "%1")
	if !mCmd.isCompactMode() {
		t.Errorf("expected command mode to be compact")
	}

	mFix := NewAIModel(AIModeFix, "%1")
	if !mFix.isCompactMode() {
		t.Errorf("expected fix mode to be compact")
	}

	mAsk := NewAIModel(AIModeAsk, "%1")
	if mAsk.isCompactMode() {
		t.Errorf("expected ask mode not to be compact")
	}
}

func TestAICardsRendering(t *testing.T) {
	m := NewAIModel(AIModeAsk, "%1")

	userCard := m.renderUserCard("How do I list files?")
	if !strings.Contains(userCard, "You") || !strings.Contains(userCard, "How do I list files?") {
		t.Errorf("userCard missing expected text: %s", userCard)
	}

	asstCard := m.renderAssistantCard("Use `ls -la`")
	if !strings.Contains(asstCard, "Assistant") || !strings.Contains(asstCard, "Use `ls -la`") {
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
