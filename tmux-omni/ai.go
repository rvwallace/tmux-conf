package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Mode constants
const (
	AIModeAsk         = "ask"
	AIModeError       = "error"
	AIModeFix         = "fix"
	AIModeSummarize   = "summarize"
	AIModeCommand     = "command"
	AIModeExplain     = "explain"
	AIModeExplainCopy = "explain-copy"
)

var aiModeTitles = map[string]string{
	AIModeAsk:         "AI: Ask",
	AIModeError:       "AI: Diagnose Error",
	AIModeFix:         "AI: Suggest Fix",
	AIModeSummarize:   "AI: Summarize Pane",
	AIModeCommand:     "AI: Generate Command",
	AIModeExplain:     "AI: Explain",
	AIModeExplainCopy: "AI: Explain Last Copy",
}

var aiInputPlaceholders = map[string]string{
	AIModeAsk:         "Ask a question... (Shift+Enter for newline, /refresh for context)",
	AIModeError:       "Ask a follow-up question about this error... (/refresh for context)",
	AIModeSummarize:   "Ask a follow-up question about this summary... (/refresh for context)",
	AIModeExplain:     "Enter command, snippet, or topic to explain...",
	AIModeExplainCopy: "Ask a follow-up question about the copied snippet...",
	AIModeCommand:     "Describe the command to generate...",
	AIModeFix:         "Refine fix (e.g. 'use brew', 'different flag')... or press Enter/s to send",
}

var scrollbackDepths = []string{"100", "200", "500", "1000", "all"}

// ChatMessage represents a single message in the transcript.
type ChatMessage struct {
	Role      string // "user", "assistant", "system", "error", "refinement"
	Content   string
	Timestamp time.Time
}

// AIModel is the Bubble Tea model for AI interaction.
type AIModel struct {
	Mode             string
	PaneID           string
	SessionID        string
	PanePath         string
	PaneCommand      string
	DepthIndex       int // index into scrollbackDepths (default 1 -> "200")
	Messages         []ChatMessage
	CandidateCommand string
	OriginalPrompt   string
	TurnCount        int
	IsBusy           bool
	FocusOnInput     bool
	ShowHelp         bool
	ToastMsg         string
	ToastTimer       int

	Width    int
	Height   int
	Viewport viewport.Model
	Input    textarea.Model
	CmdInput textinput.Model // compact single-line input for command/fix
	Spinner  spinner.Model
}

type aiResultMsg struct {
	content string
	err     error
}

type aiTickMsg time.Time

func NewAIModel(mode, paneID string) AIModel {
	if mode == "" {
		mode = AIModeAsk
	}

	sessionID := fmt.Sprintf("tmux-ai-%s-%s-%d", mode, strings.TrimPrefix(paneID, "%"), os.Getpid())

	// Fetch pane info
	panePath := ""
	paneCmd := ""
	if paneID != "" {
		out, err := exec.Command("tmux", "display-message", "-p", "-t", paneID, "#{pane_current_path}").Output()
		if err == nil {
			panePath = strings.TrimSpace(string(out))
		}
		outCmd, err := exec.Command("tmux", "display-message", "-p", "-t", paneID, "#{pane_current_command}").Output()
		if err == nil {
			paneCmd = strings.TrimSpace(string(outCmd))
		}
	}

	vp := viewport.New(80, 20)

	ta := textarea.New()
	placeholder := aiInputPlaceholders[mode]
	if placeholder == "" {
		placeholder = "Type your prompt..."
	}
	ta.Placeholder = placeholder
	ta.ShowLineNumbers = false
	ta.SetHeight(3)
	ta.FocusedStyle.Base = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(ColorAccentBlue)
	ta.BlurredStyle.Base = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(ColorFgMuted)

	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.Prompt = "❯ "
	ti.PromptStyle = lipgloss.NewStyle().Foreground(ColorAccentCyan).Bold(true)
	ti.TextStyle = lipgloss.NewStyle().Foreground(ColorFgDefault)

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(ColorAccentCyan).Bold(true)

	isCompact := (mode == AIModeCommand || mode == AIModeFix)

	m := AIModel{
		Mode:         mode,
		PaneID:       paneID,
		SessionID:    sessionID,
		PanePath:     panePath,
		PaneCommand:  paneCmd,
		DepthIndex:   1, // "200"
		Messages:     make([]ChatMessage, 0),
		FocusOnInput: true,
		Width:        80,
		Height:       24,
		Viewport:     vp,
		Input:        ta,
		CmdInput:     ti,
		Spinner:      s,
	}

	if isCompact {
		m.CmdInput.Focus()
	} else {
		m.Input.Focus()
	}

	return m
}

func (m AIModel) Init() tea.Cmd {
	var cmds []tea.Cmd
	cmds = append(cmds, m.Spinner.Tick)

	// Check if aichat is in PATH
	if _, err := exec.LookPath("aichat"); err != nil {
		m.Messages = append(m.Messages, ChatMessage{
			Role:      "error",
			Content:   "aichat is not in PATH. Please install aichat to use AI features (e.g. brew install aichat).",
			Timestamp: time.Now(),
		})
		return tea.Batch(cmds...)
	}

	// Auto-run queries for modes that don't need initial prompt
	if m.Mode == AIModeError || m.Mode == AIModeSummarize || m.Mode == AIModeExplainCopy || m.Mode == AIModeFix {
		cmds = append(cmds, m.runQueryCmd(""))
	}

	return tea.Batch(cmds...)
}

func (m AIModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		m.updateLayout()
		m.updateViewportContent()

	case spinner.TickMsg:
		if m.IsBusy {
			var cmd tea.Cmd
			m.Spinner, cmd = m.Spinner.Update(msg)
			cmds = append(cmds, cmd)
		}

	case aiResultMsg:
		m.IsBusy = false
		if msg.err != nil {
			m.Messages = append(m.Messages, ChatMessage{
				Role:      "error",
				Content:   msg.err.Error(),
				Timestamp: time.Now(),
			})
		} else {
			content := strings.TrimSpace(msg.content)
			if m.Mode == AIModeCommand || m.Mode == AIModeFix {
				m.CandidateCommand = content
				m.Messages = append(m.Messages, ChatMessage{
					Role:      "assistant",
					Content:   content,
					Timestamp: time.Now(),
				})
			} else {
				m.Messages = append(m.Messages, ChatMessage{
					Role:      "assistant",
					Content:   content,
					Timestamp: time.Now(),
				})
			}
			m.TurnCount++
		}
		m.updateViewportContent()
		m.Viewport.GotoBottom()

	case tea.KeyMsg:
		keyStr := msg.String()

		// Help modal toggle
		if m.ShowHelp {
			if keyStr == "?" || keyStr == "esc" || keyStr == "q" || keyStr == "enter" {
				m.ShowHelp = false
				return m, nil
			}
			return m, nil
		}

		if keyStr == "ctrl+c" {
			return m, tea.Quit
		}

		// When input is NOT focused, handle global navigation
		if !m.FocusOnInput {
			switch keyStr {
			case "q", "esc":
				return m, tea.Quit
			case "tab", "i", "a":
				m.FocusOnInput = true
				if m.isCompactMode() {
					m.CmdInput.Focus()
				} else {
					m.Input.Focus()
				}
				return m, nil
			case "y", "c":
				// Copy latest response or candidate
				textToCopy := m.getLatestCopyableText()
				if textToCopy != "" {
					CopyToClipboard(textToCopy)
					m.ToastMsg = "✓ Copied to clipboard"
					return m, m.toastCmd()
				}
			case "s":
				// Send candidate to pane
				if (m.Mode == AIModeCommand || m.Mode == AIModeFix) && m.CandidateCommand != "" {
					m.insertCandidateCommand(m.CandidateCommand)
					return m, tea.Quit
				}
			case "1", "2", "3", "4", "5":
				if m.Mode == AIModeSummarize {
					idx := int(keyStr[0] - '1')
					if idx >= 0 && idx < len(scrollbackDepths) && idx != m.DepthIndex {
						m.DepthIndex = idx
						return m, m.runQueryCmd("")
					}
				}
			case "d":
				if m.Mode == AIModeSummarize {
					m.DepthIndex = (m.DepthIndex + 1) % len(scrollbackDepths)
					return m, m.runQueryCmd("")
				}
			case "r":
				return m, m.runQueryCmd("")
			case "?":
				m.ShowHelp = true
				return m, nil
			case "j", "down":
				m.Viewport.LineDown(1)
				return m, nil
			case "k", "up":
				m.Viewport.LineUp(1)
				return m, nil
			case "ctrl+d", "pgdown":
				m.Viewport.HalfViewDown()
				return m, nil
			case "ctrl+u", "pgup":
				m.Viewport.HalfViewUp()
				return m, nil
			case "g", "home":
				m.Viewport.GotoTop()
				return m, nil
			case "G", "end":
				m.Viewport.GotoBottom()
				return m, nil
			}
		} else {
			// Input IS focused
			switch keyStr {
			case "esc":
				if m.isCompactMode() {
					return m, tea.Quit
				}
				// Switch focus to viewport
				m.FocusOnInput = false
				m.Input.Blur()
				return m, nil

			case "tab":
				m.FocusOnInput = false
				if m.isCompactMode() {
					m.CmdInput.Blur()
				} else {
					m.Input.Blur()
				}
				return m, nil

			case "enter":
				if m.isCompactMode() {
					text := strings.TrimSpace(m.CmdInput.Value())
					if text == "" && m.CandidateCommand != "" {
						// Send candidate command
						m.insertCandidateCommand(m.CandidateCommand)
						return m, tea.Quit
					}
					if text != "" {
						m.CmdInput.SetValue("")
						return m, m.runQueryCmd(text)
					}
					return m, nil
				} else {
					text := strings.TrimSpace(m.Input.Value())
					if text != "" {
						m.Input.SetValue("")
						return m, m.runQueryCmd(text)
					}
					return m, nil
				}
			}
		}
	}

	// Update active input widget or viewport
	if m.FocusOnInput {
		if m.isCompactMode() {
			var cmd tea.Cmd
			m.CmdInput, cmd = m.CmdInput.Update(msg)
			cmds = append(cmds, cmd)
		} else {
			var cmd tea.Cmd
			m.Input, cmd = m.Input.Update(msg)
			cmds = append(cmds, cmd)
		}
	} else {
		var cmd tea.Cmd
		m.Viewport, cmd = m.Viewport.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m *AIModel) isCompactMode() bool {
	return m.Mode == AIModeCommand || m.Mode == AIModeFix
}

func (m *AIModel) updateLayout() {
	headerHeight := 3
	footerHeight := 1
	inputHeight := 4
	if m.isCompactMode() {
		inputHeight = 3
	}

	availHeight := m.Height - headerHeight - footerHeight - inputHeight
	if availHeight < 4 {
		availHeight = 4
	}

	m.Viewport.Width = m.Width - 4
	m.Viewport.Height = availHeight

	m.Input.SetWidth(m.Width - 4)
	m.CmdInput.Width = m.Width - 8
}

func (m *AIModel) updateViewportContent() {
	var b strings.Builder

	for _, msg := range m.Messages {
		switch msg.Role {
		case "user":
			b.WriteString(m.renderUserCard(msg.Content))
		case "assistant":
			b.WriteString(m.renderAssistantCard(msg.Content))
		case "error":
			b.WriteString(m.renderErrorCard(msg.Content))
		case "refinement":
			b.WriteString(m.renderRefinementCard(msg.Content))
		}
		b.WriteString("\n")
	}

	if m.IsBusy {
		loadingLabel := " Generating response..."
		if m.Mode == AIModeError {
			loadingLabel = " Diagnosing error from context..."
		} else if m.Mode == AIModeSummarize {
			loadingLabel = fmt.Sprintf(" Summarizing pane (%s)...", scrollbackDepths[m.DepthIndex])
		} else if m.Mode == AIModeFix {
			loadingLabel = " Suggesting corrective command..."
		} else if m.Mode == AIModeCommand {
			loadingLabel = " Generating shell command..."
		}
		spinnerLine := fmt.Sprintf("%s%s", m.Spinner.View(), lipgloss.NewStyle().Foreground(ColorAccentCyan).Bold(true).Render(loadingLabel))
		b.WriteString(lipgloss.NewStyle().Padding(0, 1).Render(spinnerLine))
		b.WriteString("\n")
	}

	m.Viewport.SetContent(b.String())
}

func (m AIModel) renderUserCard(content string) string {
	badge := lipgloss.NewStyle().Foreground(ColorAccentBlue).Bold(true).Render(" You")
	body := lipgloss.NewStyle().Foreground(ColorFgDefault).Render(content)
	card := lipgloss.JoinVertical(lipgloss.Left, badge, body)
	return lipgloss.NewStyle().
		Border(lipgloss.Border{Left: "▎"}, false, false, false, true).
		BorderForeground(ColorAccentBlue).
		Padding(0, 1).
		MarginBottom(1).
		Render(card)
}

func (m AIModel) renderRefinementCard(content string) string {
	badge := lipgloss.NewStyle().Foreground(ColorAccentYellow).Bold(true).Render("󰑕 Refinement")
	body := lipgloss.NewStyle().Foreground(ColorFgDefault).Render(content)
	card := lipgloss.JoinVertical(lipgloss.Left, badge, body)
	return lipgloss.NewStyle().
		Border(lipgloss.Border{Left: "▎"}, false, false, false, true).
		BorderForeground(ColorAccentYellow).
		Padding(0, 1).
		MarginBottom(1).
		Render(card)
}

func (m AIModel) renderAssistantCard(content string) string {
	if m.isCompactMode() {
		badge := lipgloss.NewStyle().Foreground(ColorAccentGreen).Bold(true).Render("󰘳 Candidate Command")
		cmdBox := lipgloss.NewStyle().
			Foreground(ColorAccentGreen).
			Background(ColorBgCard).
			Bold(true).
			Padding(0, 1).
			Render(content)
		card := lipgloss.JoinVertical(lipgloss.Left, badge, cmdBox)
		return lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorAccentGreen).
			Padding(0, 1).
			MarginBottom(1).
			Render(card)
	}

	badge := lipgloss.NewStyle().Foreground(ColorAccentPurple).Bold(true).Render("󰧑 Assistant")
	body := lipgloss.NewStyle().Foreground(ColorFgDefault).Render(content)
	card := lipgloss.JoinVertical(lipgloss.Left, badge, body)
	return lipgloss.NewStyle().
		Border(lipgloss.Border{Left: "▎"}, false, false, false, true).
		BorderForeground(ColorAccentPurple).
		Padding(0, 1).
		MarginBottom(1).
		Render(card)
}

func (m AIModel) renderErrorCard(content string) string {
	badge := lipgloss.NewStyle().Foreground(ColorStatusError).Bold(true).Render("✖ Error")
	body := lipgloss.NewStyle().Foreground(ColorStatusError).Render(content)
	card := lipgloss.JoinVertical(lipgloss.Left, badge, body)
	return lipgloss.NewStyle().
		Border(lipgloss.Border{Left: "▎"}, false, false, false, true).
		BorderForeground(ColorStatusError).
		Padding(0, 1).
		MarginBottom(1).
		Render(card)
}

func (m AIModel) getLatestCopyableText() string {
	if m.CandidateCommand != "" {
		return m.CandidateCommand
	}
	for i := len(m.Messages) - 1; i >= 0; i-- {
		if m.Messages[i].Role == "assistant" {
			return m.Messages[i].Content
		}
	}
	return ""
}

func (m *AIModel) insertCandidateCommand(cmdText string) {
	if cmdText == "" {
		return
	}
	aiAssistPath := findAIAssistScript()
	cmd := exec.Command(aiAssistPath, "insert-command", m.PaneID, cmdText)
	cmd.Env = append(os.Environ(), "TMUX_AI_ASSIST_NO_PAUSE=1")
	_ = cmd.Run()
}

func findAIAssistScript() string {
	homeDir, _ := os.UserHomeDir()
	candidates := []string{
		filepath.Join(homeDir, ".config", "tmux", "scripts", "ai-assist.sh"),
		filepath.Join(".", "scripts", "ai-assist.sh"),
		filepath.Join("..", "scripts", "ai-assist.sh"),
	}
	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return "ai-assist.sh"
}

func (m *AIModel) runQueryCmd(promptText string) tea.Cmd {
	m.IsBusy = true

	if promptText != "" {
		if m.isCompactMode() && m.CandidateCommand != "" {
			m.Messages = append(m.Messages, ChatMessage{
				Role:      "refinement",
				Content:   promptText,
				Timestamp: time.Now(),
			})
		} else {
			m.Messages = append(m.Messages, ChatMessage{
				Role:      "user",
				Content:   promptText,
				Timestamp: time.Now(),
			})
		}
	}
	m.updateViewportContent()

	mode := m.Mode
	paneID := m.PaneID
	sessionID := m.SessionID
	depth := scrollbackDepths[m.DepthIndex]
	turnCount := m.TurnCount
	candidateCmd := m.CandidateCommand
	origPrompt := m.OriginalPrompt

	return func() tea.Msg {
		aiAssistPath := findAIAssistScript()

		args := []string{mode, paneID}
		env := os.Environ()
		env = append(env, "TMUX_AI_ASSIST_NO_PAUSE=1")
		env = append(env, fmt.Sprintf("TMUX_AI_ASSIST_SCROLLBACK=%s", depth))

		if mode == AIModeCommand || mode == AIModeFix {
			env = append(env, "TMUX_AI_ASSIST_PRINT_ONLY=1")
			if candidateCmd != "" && promptText != "" {
				env = append(env, "TMUX_AI_ASSIST_REFINE=1")
				env = append(env, fmt.Sprintf("TMUX_AI_ASSIST_ORIGINAL_PROMPT=%s", origPrompt))
				env = append(env, fmt.Sprintf("TMUX_AI_ASSIST_PREVIOUS_COMMAND=%s", candidateCmd))
				args = append(args, promptText)
			} else if promptText != "" {
				args = append(args, promptText)
			}
		} else {
			env = append(env, fmt.Sprintf("TMUX_AI_ASSIST_SESSION=%s", sessionID))
			if promptText == "/refresh" {
				env = append(env, "TMUX_AI_ASSIST_REFRESH=1")
				args = append(args, promptText)
			} else if turnCount > 0 && promptText != "" {
				env = append(env, "TMUX_AI_ASSIST_FOLLOW_UP=1")
				args = append(args, promptText)
			} else if promptText != "" {
				args = append(args, promptText)
			}
		}

		ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
		defer cancel()

		cmd := exec.CommandContext(ctx, aiAssistPath, args...)
		cmd.Env = env

		out, err := cmd.CombinedOutput()
		if err != nil {
			return aiResultMsg{err: fmt.Errorf("aichat failed: %s (%w)", strings.TrimSpace(string(out)), err)}
		}

		return aiResultMsg{content: string(out)}
	}
}

func (m AIModel) toastCmd() tea.Cmd {
	return tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
		return aiTickMsg(t)
	})
}

func (m AIModel) View() string {
	if m.Width == 0 {
		return "Initializing..."
	}

	// 1. Header
	titleStr := aiModeTitles[m.Mode]
	if titleStr == "" {
		titleStr = "AI Assistant"
	}
	titleBadge := lipgloss.NewStyle().Foreground(ColorAccentBlue).Bold(true).Render(titleStr)

	depthBadge := ""
	if m.Mode == AIModeSummarize {
		depthBadge = lipgloss.NewStyle().Foreground(ColorAccentCyan).Bold(true).Render(fmt.Sprintf(" [%s lines]", scrollbackDepths[m.DepthIndex]))
	}

	brandBadge := lipgloss.NewStyle().Foreground(ColorAccentPurple).Bold(true).Render("󰧑 aichat")
	topBar := lipgloss.JoinHorizontal(lipgloss.Center, titleBadge, depthBadge, "  ", brandBadge)

	pathStr := m.PanePath
	if len(pathStr) > m.Width-10 {
		pathStr = "..." + pathStr[len(pathStr)-(m.Width-13):]
	}
	subHeader := lipgloss.NewStyle().Foreground(ColorFgMuted).Render(fmt.Sprintf(" %s   %s", pathStr, m.PaneCommand))

	headerBox := lipgloss.NewStyle().
		Background(ColorBgSurface).
		Padding(0, 1).
		Width(m.Width).
		Render(lipgloss.JoinVertical(lipgloss.Left, topBar, subHeader))

	// 2. Viewport
	vpView := m.Viewport.View()

	// 3. Input Box
	var inputBox string
	if m.isCompactMode() {
		inputBox = lipgloss.NewStyle().
			Padding(0, 1).
			Width(m.Width).
			Render(m.CmdInput.View())
	} else {
		inputBox = lipgloss.NewStyle().
			Padding(0, 1).
			Width(m.Width).
			Render(m.Input.View())
	}

	// 4. Footer
	var hints string
	if m.isCompactMode() {
		if m.CandidateCommand != "" {
			hints = "󰌌 <Enter/s> Send   󰅍 <y/c> Copy   <Esc> Cancel   <?> Help"
		} else {
			hints = "󰌌 <Enter> Generate   <Esc> Cancel   <?> Help"
		}
	} else if m.Mode == AIModeSummarize {
		hints = "󰌌 <1-5/d> Depth   <Enter> Send   <Tab> Focus   󰅍 <y/c> Copy   <?> Help   <Esc> Quit"
	} else {
		hints = "󰌌 <Enter> Send   <Tab> Focus   </refresh> Context   󰅍 <y/c> Copy   <?> Help   <Esc> Quit"
	}

	if m.ToastMsg != "" {
		hints = lipgloss.NewStyle().Foreground(ColorAccentGreen).Bold(true).Render(m.ToastMsg)
	}

	footerLine := lipgloss.NewStyle().
		Foreground(ColorFgMuted).
		Background(ColorBgSurface).
		Padding(0, 1).
		Width(m.Width).
		Render(hints)

	content := lipgloss.JoinVertical(lipgloss.Left,
		headerBox,
		vpView,
		inputBox,
		footerLine,
	)

	// 5. Help Modal Overlay if active
	if m.ShowHelp {
		return m.renderHelpModal(content)
	}

	return content
}

func (m AIModel) renderHelpModal(background string) string {
	helpText := `
  AI Assistant Shortcuts:

  <Enter>       Submit prompt / Send candidate command
  <Shift+Enter> Insert newline in multiline input
  <Tab>         Toggle focus between input and transcript
  <j> / <k>     Scroll transcript up / down (when transcript focused)
  <Ctrl+d/u>    Half-page scroll down / up
  <g> / <G>     Jump to top / bottom
  <y> / <c>     Copy latest assistant response or candidate command
  <1> - <5>     Switch scrollback depth (100, 200, 500, 1000, all)
  <d>           Cycle scrollback depth
  <r>           Reload query with current pane context
  </refresh>    Type /refresh to update pane context in session
  <?>           Toggle this help dialog
  <Esc> / <q>   Quit AI assistant
`
	dialogBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorAccentBlue).
		Background(ColorBgCard).
		Foreground(ColorFgDefault).
		Padding(1, 2).
		Render(helpText)

	return lipgloss.Place(m.Width, m.Height, lipgloss.Center, lipgloss.Center, dialogBox)
}
