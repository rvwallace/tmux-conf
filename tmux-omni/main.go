package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type AppMode int

const (
	ModeWhichKey AppMode = iota
	ModePalette
	ModeInspector
)

// PendingExecution records the action to execute after TUI cleanly exits.
type PendingExecution struct {
	Action         string
	Target         string
	Title          string
	PersistShell   bool
	OriginalTarget string
}

// AppModel is the top-level Bubble Tea application model.
type AppModel struct {
	Mode                  AppMode
	PreviousMode          AppMode
	WhichKey              WhichKeyModel
	Palette               PaletteModel
	Inspector             InspectModel
	IsStandaloneInspector bool
	Config                *Config
	FlatCommands          []FlatCommand
	StartSearch           bool
	PaneID                string
	Width                 int
	Height                int
	PendingExec           *PendingExecution
}

func InitialModel(cfg *Config, flat []FlatCommand, startSearch bool, paneID string, inspectorType string) AppModel {
	var wk WhichKeyModel
	if cfg != nil {
		wk = NewWhichKeyModel(cfg.Title, cfg.Items)
	}
	pal := NewPaletteModel(flat)

	initialMode := ModeWhichKey
	var insp InspectModel
	isStandalone := false

	if inspectorType != "" {
		initialMode = ModeInspector
		insp = CreateInspector(inspectorType)
		isStandalone = true
	} else if startSearch {
		initialMode = ModePalette
	}

	return AppModel{
		Mode:                  initialMode,
		PreviousMode:          ModeWhichKey,
		WhichKey:              wk,
		Palette:               pal,
		Inspector:             insp,
		IsStandaloneInspector: isStandalone,
		Config:                cfg,
		FlatCommands:          flat,
		StartSearch:           startSearch,
		PaneID:                paneID,
	}
}

func (m AppModel) Init() tea.Cmd {
	if m.Mode == ModePalette || m.Mode == ModeInspector {
		return textinput.Blink
	}
	return nil
}

func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case ClearInspectStatusMsg:
		if msg.ID == m.Inspector.StatusMsgID {
			m.Inspector.StatusMsg = ""
		}
		return m, nil

	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		m.WhichKey.Width = msg.Width
		m.WhichKey.Height = msg.Height
		m.Palette.Width = msg.Width
		m.Palette.Height = msg.Height
		m.Inspector.Width = msg.Width
		m.Inspector.Height = msg.Height
		return m, nil

	case tea.KeyMsg:
		keyStr := msg.String()

		// Global quit
		if keyStr == "ctrl+c" {
			return m, tea.Quit
		}

		if m.Mode == ModeWhichKey {
			return m.updateWhichKey(msg)
		} else if m.Mode == ModePalette {
			return m.updatePalette(msg)
		} else if m.Mode == ModeInspector {
			return m.updateInspector(msg)
		}
	}

	// Update textinput ticks if in palette or inspector mode
	if m.Mode == ModePalette {
		var cmd tea.Cmd
		m.Palette.TextInput, cmd = m.Palette.TextInput.Update(msg)
		return m, cmd
	} else if m.Mode == ModeInspector {
		var cmd tea.Cmd
		m.Inspector.TextInput, cmd = m.Inspector.TextInput.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m AppModel) updateWhichKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	keyStr := msg.String()

	switch keyStr {
	case "/", "ctrl+f":
		m.Mode = ModePalette
		m.Palette.TextInput.Focus()
		m.Palette.FilterCommands()
		return m, textinput.Blink

	case "esc", "backspace":
		if len(m.WhichKey.NavStack) > 1 {
			m.WhichKey.NavStack = m.WhichKey.NavStack[:len(m.WhichKey.NavStack)-1]
			return m, nil
		}
		return m, tea.Quit

	case "q":
		return m, tea.Quit

	case "down", "ctrl+n", "ctrl+j":
		m.WhichKey.CursorDown()
		return m, nil

	case "up", "ctrl+p", "ctrl+k":
		m.WhichKey.CursorUp()
		return m, nil

	case "enter":
		item := m.WhichKey.SelectedItem()
		if item != nil {
			if len(item.Items) > 0 {
				icon := item.Icon
				if icon == "" {
					icon = "󰘳"
				}
				m.WhichKey.NavStack = append(m.WhichKey.NavStack, NavFrame{
					Title:       item.Title,
					Icon:        icon,
					Items:       item.Items,
					CursorIndex: 0,
				})
				return m, nil
			} else if item.Action != "" {
				if inspType, ok := IsInspectorCommand(item.Action); ok {
					m.PreviousMode = ModeWhichKey
					m.Inspector = CreateInspector(inspType)
					m.Inspector.Width = m.Width
					m.Inspector.Height = m.Height
					m.Mode = ModeInspector
					m.Inspector.TextInput.Focus()
					return m, tea.Batch(textinput.Blink, m.Inspector.TextInput.Focus())
				}
				target := item.Target
				if target == "" {
					target = "tmux"
				}
				m.PendingExec = &PendingExecution{
					Action:         item.Action,
					Target:         target,
					Title:          item.Title,
					PersistShell:   item.PersistShell || item.Shell,
					OriginalTarget: target,
				}
				return m, tea.Quit
			}
		}
		// Fallback: switch to palette search
		m.Mode = ModePalette
		m.Palette.TextInput.Focus()
		m.Palette.FilterCommands()
		return m, textinput.Blink

	case "alt+v", "ctrl+v", "alt+h", "ctrl+h", "√":
		item := m.WhichKey.SelectedItem()
		if item != nil && item.Action != "" {
			m.PendingExec = &PendingExecution{
				Action:         item.Action,
				Target:         "split_h",
				Title:          item.Title,
				PersistShell:   true,
				OriginalTarget: item.Target,
			}
			return m, tea.Quit
		}

	case "alt+s", "ctrl+s", "alt+x", "ctrl+x", "ß":
		item := m.WhichKey.SelectedItem()
		if item != nil && item.Action != "" {
			m.PendingExec = &PendingExecution{
				Action:         item.Action,
				Target:         "split_v",
				Title:          item.Title,
				PersistShell:   true,
				OriginalTarget: item.Target,
			}
			return m, tea.Quit
		}

	case "alt+w", "ctrl+w", "ctrl+t", "alt+t", "∑":
		item := m.WhichKey.SelectedItem()
		if item != nil && item.Action != "" {
			m.PendingExec = &PendingExecution{
				Action:         item.Action,
				Target:         "window",
				Title:          item.Title,
				PersistShell:   true,
				OriginalTarget: item.Target,
			}
			return m, tea.Quit
		}

	case "alt+i", "ctrl+i", "tab", "shift+tab", "ˆ", "^":
		item := m.WhichKey.SelectedItem()
		if item != nil && item.Action != "" {
			m.PendingExec = &PendingExecution{
				Action:         item.Action,
				Target:         "send_keys",
				Title:          item.Title,
				PersistShell:   false,
				OriginalTarget: item.Target,
			}
			return m, tea.Quit
		}

	case "y":
		item := m.WhichKey.SelectedItem()
		if item != nil && item.Action != "" {
			CopyToClipboard(item.Action)
			return m, tea.Quit
		}
	}

	// Match key against current items
	frame := m.WhichKey.CurrentFrame()
	for _, item := range frame.Items {
		if item.Key == keyStr || (len(keyStr) == 1 && item.Key == keyStr) {
			if len(item.Items) > 0 {
				// Sub-group navigation
				icon := item.Icon
				if icon == "" {
					icon = "󰘳"
				}
				m.WhichKey.NavStack = append(m.WhichKey.NavStack, NavFrame{
					Title:       item.Title,
					Icon:        icon,
					Items:       item.Items,
					CursorIndex: 0,
				})
				return m, nil
			} else if item.Action != "" {
				if inspType, ok := IsInspectorCommand(item.Action); ok {
					m.PreviousMode = ModeWhichKey
					m.Inspector = CreateInspector(inspType)
					m.Inspector.Width = m.Width
					m.Inspector.Height = m.Height
					m.Mode = ModeInspector
					m.Inspector.TextInput.Focus()
					return m, tea.Batch(textinput.Blink, m.Inspector.TextInput.Focus())
				}
				// Leaf action execution
				target := item.Target
				if target == "" {
					target = "tmux"
				}
				m.PendingExec = &PendingExecution{
					Action:         item.Action,
					Target:         target,
					Title:          item.Title,
					PersistShell:   item.PersistShell || item.Shell,
					OriginalTarget: target,
				}
				return m, tea.Quit
			}
		}
	}

	return m, nil
}

func (m AppModel) updatePalette(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	keyStr := msg.String()

	maxVisible := max(m.Palette.Height-4, 1)

	// Target modifiers
	switch keyStr {
	case "alt+v", "ctrl+v", "alt+h", "ctrl+h", "√":
		return m.executeSelectedPaletteItem("split_h")
	case "alt+s", "ctrl+s", "alt+x", "ctrl+x", "ß":
		return m.executeSelectedPaletteItem("split_v")
	case "alt+w", "ctrl+w", "ctrl+t", "alt+t", "∑":
		return m.executeSelectedPaletteItem("window")
	case "alt+i", "ctrl+i", "tab", "shift+tab", "ˆ", "^":
		return m.executeSelectedPaletteItem("send_keys")
	case "ctrl+y", "alt+y":
		cmd := m.Palette.SelectedCommand()
		if cmd != nil && cmd.Action != "" {
			CopyToClipboard(cmd.Action)
			return m, tea.Quit
		}
	}

	switch keyStr {
	case "ctrl+space", "ctrl+l", "alt+m":
		m.Mode = ModeWhichKey
		return m, nil

	case "down", "ctrl+n", "ctrl+j":
		m.Palette.CursorDown(maxVisible)
		return m, nil

	case "up", "ctrl+p", "ctrl+k":
		m.Palette.CursorUp(maxVisible)
		return m, nil

	case "enter":
		return m.executeSelectedPaletteItem("")

	case "esc":
		if m.Palette.TextInput.Value() != "" {
			m.Palette.TextInput.SetValue("")
			m.Palette.FilterCommands()
			return m, nil
		}
		m.Mode = ModeWhichKey
		return m, nil

	case "backspace":
		if m.Palette.TextInput.Value() == "" {
			m.Mode = ModeWhichKey
			return m, nil
		}
	}

	// Handle typing into textinput
	oldVal := m.Palette.TextInput.Value()
	var cmd tea.Cmd
	m.Palette.TextInput, cmd = m.Palette.TextInput.Update(msg)
	if m.Palette.TextInput.Value() != oldVal {
		m.Palette.FilterCommands()
	}

	return m, cmd
}

func (m AppModel) executeSelectedPaletteItem(targetOverride string) (tea.Model, tea.Cmd) {
	cmd := m.Palette.SelectedCommand()
	if cmd == nil {
		return m, nil
	}

	if targetOverride == "" {
		if inspType, ok := IsInspectorCommand(cmd.Action); ok {
			m.PreviousMode = ModePalette
			m.Inspector = CreateInspector(inspType)
			m.Inspector.Width = m.Width
			m.Inspector.Height = m.Height
			m.Mode = ModeInspector
			m.Inspector.TextInput.Focus()
			return m, tea.Batch(textinput.Blink, m.Inspector.TextInput.Focus())
		}
	}

	target := targetOverride
	if target == "" {
		target = cmd.Target
	}

	persist := cmd.PersistShell
	if target == "window" || target == "split_h" || target == "split_v" {
		persist = true
	}

	m.PendingExec = &PendingExecution{
		Action:         cmd.Action,
		Target:         target,
		Title:          cmd.Title,
		PersistShell:   persist,
		OriginalTarget: cmd.Target,
	}

	return m, tea.Quit
}

func (m AppModel) updateInspector(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	keyStr := msg.String()
	maxVisible := max(m.Inspector.Height-4, 1)

	// Handle Action Picker modal if open
	if m.Inspector.ShowActionPicker {
		switch keyStr {
		case "ctrl+c":
			return m, tea.Quit
		case "esc", "q", "backspace":
			m.Inspector.CloseActionPicker()
			return m, nil
		case "down", "ctrl+n", "ctrl+j", "j":
			if len(m.Inspector.ActionOptions) > 0 {
				m.Inspector.ActionPickerCursor = (m.Inspector.ActionPickerCursor + 1) % len(m.Inspector.ActionOptions)
			}
			return m, nil
		case "up", "ctrl+p", "ctrl+k", "k":
			if len(m.Inspector.ActionOptions) > 0 {
				m.Inspector.ActionPickerCursor = (m.Inspector.ActionPickerCursor - 1 + len(m.Inspector.ActionOptions)) % len(m.Inspector.ActionOptions)
			}
			return m, nil
		case "enter":
			if len(m.Inspector.ActionOptions) > 0 {
				opt := m.Inspector.ActionOptions[m.Inspector.ActionPickerCursor]
				return m.executeActionOption(opt)
			}
			return m, nil
		}

		// Direct hotkeys inside modal
		for _, opt := range m.Inspector.ActionOptions {
			if keyStr == opt.Key || strings.ToLower(keyStr) == opt.Key {
				return m.executeActionOption(opt)
			}
		}
		return m, nil
	}

	// Action Picker modal trigger
	switch keyStr {
	case "ctrl+a", "alt+a", "ctrl+o", "alt+o", "å", "ø":
		if m.Inspector.OpenActionPicker() {
			return m, nil
		}
	}

	switch keyStr {
	case "ctrl+space", "ctrl+l", "alt+m":
		m.Mode = ModeWhichKey
		m.Inspector.TextInput.Blur()
		return m, nil

	case "ctrl+f":
		m.Mode = ModePalette
		m.Inspector.TextInput.Blur()
		m.Palette.TextInput.Focus()
		m.Palette.FilterCommands()
		return m, textinput.Blink

	case "backspace":
		if m.Inspector.TextInput.Value() == "" {
			if m.IsStandaloneInspector {
				return m, tea.Quit
			}
			m.Mode = m.PreviousMode
			m.Inspector.TextInput.Blur()
			if m.Mode == ModePalette {
				m.Palette.TextInput.Focus()
				m.Palette.FilterCommands()
				return m, textinput.Blink
			}
			return m, nil
		}

	case "esc":
		if m.Inspector.TextInput.Value() != "" {
			m.Inspector.TextInput.SetValue("")
			m.Inspector.Filter()
			return m, nil
		}
		if m.IsStandaloneInspector {
			return m, tea.Quit
		}
		m.Mode = m.PreviousMode
		m.Inspector.TextInput.Blur()
		if m.Mode == ModePalette {
			m.Palette.TextInput.Focus()
			m.Palette.FilterCommands()
			return m, textinput.Blink
		}
		return m, nil

	case "down", "ctrl+n", "ctrl+j":
		m.Inspector.CursorDown(maxVisible)
		return m, nil

	case "up", "ctrl+p", "ctrl+k":
		m.Inspector.CursorUp(maxVisible)
		return m, nil

	case "enter":
		item := m.Inspector.SelectedItem()
		if item == nil {
			return m, nil
		}
		if m.Inspector.AllowExecute && item.ActionCmd != "" {
			m.PendingExec = &PendingExecution{
				Action:         item.ActionCmd,
				Target:         "tmux",
				Title:          item.Col1,
				PersistShell:   false,
				OriginalTarget: "tmux",
			}
			return m, tea.Quit
		} else if m.Inspector.AllowCopy && item.RawCopy != "" {
			CopyToClipboard(item.RawCopy)
			disp := item.RawCopy
			if len(disp) > 40 {
				disp = disp[:37] + "..."
			}
			cmd := m.Inspector.SetStatus(fmt.Sprintf("Copied to clipboard: %s", disp))
			return m, cmd
		}
		return m, nil

	case "alt+i", "ctrl+i", "tab", "shift+tab", "ˆ", "^":
		if !m.Inspector.AllowSendKeys {
			return m, nil
		}
		item := m.Inspector.SelectedItem()
		if item != nil && item.ActionCmd != "" {
			m.PendingExec = &PendingExecution{
				Action:         item.ActionCmd,
				Target:         "send_keys",
				Title:          item.Col1,
				PersistShell:   false,
				OriginalTarget: "tmux",
			}
			return m, tea.Quit
		}
		return m, nil

	case "alt+w", "ctrl+w", "ctrl+t", "alt+t", "∑":
		if !m.Inspector.AllowWindow {
			return m, nil
		}
		item := m.Inspector.SelectedItem()
		if item != nil && item.ActionCmd != "" {
			m.PendingExec = &PendingExecution{
				Action:         item.ActionCmd,
				Target:         "window",
				Title:          item.Col1,
				PersistShell:   true,
				OriginalTarget: "tmux",
			}
			return m, tea.Quit
		}
		return m, nil

	case "alt+v", "ctrl+v", "alt+h", "ctrl+h", "√":
		if !m.Inspector.AllowSplit {
			return m, nil
		}
		item := m.Inspector.SelectedItem()
		if item != nil && item.ActionCmd != "" {
			m.PendingExec = &PendingExecution{
				Action:         item.ActionCmd,
				Target:         "split_h",
				Title:          item.Col1,
				PersistShell:   true,
				OriginalTarget: "tmux",
			}
			return m, tea.Quit
		}
		return m, nil

	case "alt+s", "ctrl+s", "alt+x", "ctrl+x", "ß":
		if !m.Inspector.AllowSplit {
			return m, nil
		}
		item := m.Inspector.SelectedItem()
		if item != nil && item.ActionCmd != "" {
			m.PendingExec = &PendingExecution{
				Action:         item.ActionCmd,
				Target:         "split_v",
				Title:          item.Col1,
				PersistShell:   true,
				OriginalTarget: "tmux",
			}
			return m, tea.Quit
		}
		return m, nil

	case "ctrl+d", "alt+d", "∂":
		if m.Inspector.AllowDelete && m.Inspector.Title == "Paste Buffers" {
			item := m.Inspector.SelectedItem()
			if item != nil && item.Col1 != "" {
				_ = exec.Command("tmux", "delete-buffer", "-b", item.Col1).Run()
				m.Inspector.Items = LoadBuffersData()
				m.Inspector.Filter()
				cmd := m.Inspector.SetStatus(fmt.Sprintf("Deleted buffer: %s", item.Col1))
				return m, cmd
			}
		}

	case "ctrl+y", "alt+y", "¥":
		if !m.Inspector.AllowCopy {
			return m, nil
		}
		item := m.Inspector.SelectedItem()
		if item != nil {
			copyText := item.RawCopy
			if copyText == "" && item.Col3 != "" {
				copyText = item.Col3
			} else if copyText == "" && item.Col1 != "" {
				copyText = item.Col1
			}
			if copyText != "" {
				CopyToClipboard(copyText)
				disp := copyText
				if len(disp) > 40 {
					disp = disp[:37] + "..."
				}
				cmd := m.Inspector.SetStatus(fmt.Sprintf("Copied to clipboard: %s", disp))
				return m, cmd
			}
		}
	}

	// Update text input
	oldVal := m.Inspector.TextInput.Value()
	var cmd tea.Cmd
	m.Inspector.TextInput, cmd = m.Inspector.TextInput.Update(msg)
	if m.Inspector.TextInput.Value() != oldVal {
		m.Inspector.Filter()
		m.Inspector.StatusMsg = ""
	}

	return m, cmd
}

func (m AppModel) executeActionOption(opt ActionOption) (tea.Model, tea.Cmd) {
	m.Inspector.CloseActionPicker()

	switch opt.ActionType {
	case "copy":
		CopyToClipboard(opt.Payload)
		disp := opt.Payload
		if len(disp) > 40 {
			disp = disp[:37] + "..."
		}
		cmd := m.Inspector.SetStatus(fmt.Sprintf("Copied %s: %s", strings.ToLower(opt.Title), disp))
		return m, cmd

	case "exec":
		m.PendingExec = &PendingExecution{
			Action:         opt.Payload,
			Target:         "tmux",
			Title:          opt.Title,
			PersistShell:   false,
			OriginalTarget: "tmux",
		}
		return m, tea.Quit

	case "send":
		m.PendingExec = &PendingExecution{
			Action:         opt.Payload,
			Target:         "send_keys",
			Title:          opt.Title,
			PersistShell:   false,
			OriginalTarget: "tmux",
		}
		return m, tea.Quit
	}

	return m, nil
}

func (m AppModel) View() string {
	if m.Mode == ModeWhichKey {
		return WindowStyle.Render(m.WhichKey.View())
	} else if m.Mode == ModePalette {
		return WindowStyle.Render(m.Palette.View())
	}
	return m.Inspector.View()
}

func runInspector(inspModel InspectModel, paneID string) {
	p := tea.NewProgram(InspectorAppModel{Model: inspModel}, tea.WithAltScreen())
	finalModel, err := p.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if app, ok := finalModel.(InspectorAppModel); ok && app.Model.PendingExec != "" {
		target := app.Model.PendingExecTarget
		if target == "" {
			target = "tmux"
		}
		title := app.Model.PendingExecTitle
		if title == "" {
			title = app.Model.Title
		}
		persist := (target == "window" || target == "split_h" || target == "split_v")
		execErr := RunTmuxTarget(
			app.Model.PendingExec,
			target,
			paneID,
			title,
			persist,
			"tmux",
		)
		if execErr != nil {
			fmt.Fprintf(os.Stderr, "Execution error: %v\n", execErr)
			os.Exit(1)
		}
	}
}

func runAI(mode, paneID string) {
	model := NewAIModel(mode, paneID)
	p := tea.NewProgram(model, tea.WithAltScreen(), tea.WithMouseCellMotion())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running AI assistant: %v\n", err)
		os.Exit(1)
	}
}

func main() {
	searchFlag := flag.Bool("s", false, "Start in Command Palette search mode")
	searchLong := flag.Bool("search", false, "Start in Command Palette search mode")
	keysFlag := flag.Bool("k", false, "Open Keybindings Inspector")
	keysLong := flag.Bool("keys", false, "Open Keybindings Inspector")
	cmdsFlag := flag.Bool("C", false, "Open Tmux Commands Inspector")
	cmdsLong := flag.Bool("commands", false, "Open Tmux Commands Inspector")
	optsFlag := flag.Bool("O", false, "Open Tmux Options Inspector")
	optsLong := flag.Bool("options", false, "Open Tmux Options Inspector")
	envFlag := flag.Bool("E", false, "Open Tmux Environment Inspector")
	envLong := flag.Bool("env", false, "Open Tmux Environment Inspector")
	buffersFlag := flag.Bool("B", false, "Open Paste Buffers Inspector")
	buffersLong := flag.Bool("buffers", false, "Open Paste Buffers Inspector")
	messagesFlag := flag.Bool("M", false, "Open Tmux Messages Inspector")
	messagesLong := flag.Bool("messages", false, "Open Tmux Messages Inspector")
	logsFlag := flag.Bool("logs", false, "Open Tmux Messages Inspector")
	clientsFlag := flag.Bool("clients", false, "Open Connected Clients Inspector")
	statesFlag := flag.Bool("states", false, "Open Saved Tmux States Inspector")
	aiFlag := flag.String("ai", "", "Run AI Assistant mode (ask, error, fix, summarize, command, explain, explain-copy)")
	validateFlag := flag.Bool("validate", false, "Validate config.json and print diagnostics")
	validateLong := flag.Bool("check-config", false, "Validate config.json and print diagnostics")
	configFlag := flag.String("config", "", "Path to config.json")
	flag.Parse()

	// Check positional subcommand or pane ID
	subcmd := ""
	aiModeArg := ""
	paneIDArg := ""
	args := flag.Args()
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "search", "palette":
			*searchFlag = true
		case "keys", "keybindings", "help":
			*keysFlag = true
		case "commands", "cmds":
			*cmdsFlag = true
		case "options", "opts":
			*optsFlag = true
		case "env", "environment":
			*envFlag = true
		case "buffers", "buffer":
			*buffersFlag = true
		case "messages", "msg", "logs", "log":
			*messagesFlag = true
		case "clients":
			*clientsFlag = true
		case "states", "saves", "snapshots", "autosaves":
			*statesFlag = true
		case "validate", "check-config", "lint":
			*validateFlag = true
		case "ai", "aichat":
			if i+1 < len(args) {
				aiModeArg = args[i+1]
				i++
			}
		case "ask", "error", "fix", "summarize", "command", "explain", "explain-copy":
			aiModeArg = arg
		default:
			if paneIDArg == "" {
				paneIDArg = arg
			}
		}
	}
	_ = subcmd

	if *aiFlag != "" {
		aiModeArg = *aiFlag
	}
	// Mode 0: Config Validation / Linter
	if *validateFlag || *validateLong {
		configPath, err := FindConfigFile(*configFlag)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		cfg, err := LoadConfig(configPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		errs := ValidateConfig(cfg)
		if len(errs) > 0 {
			fmt.Fprintf(os.Stderr, "Config validation FAILED for %s (%d errors):\n", configPath, len(errs))
			for _, e := range errs {
				fmt.Fprintf(os.Stderr, "  ✖ %s\n", e)
			}
			os.Exit(1)
		}
		flat := FlattenCommands(cfg.Items, nil, "")
		fmt.Printf("Config validation PASSED: %s (%d top-level items, %d total commands)\n", configPath, len(cfg.Items), len(flat))
		return
	}

	paneID := GetCurrentPaneID(paneIDArg)

	// Mode 0.5: AI Assistant
	if aiModeArg != "" {
		runAI(aiModeArg, paneID)
		return
	}

	inspectorType := ""
	if *keysFlag || *keysLong {
		inspectorType = "keys"
	} else if *cmdsFlag || *cmdsLong {
		inspectorType = "commands"
	} else if *optsFlag || *optsLong {
		inspectorType = "options"
	} else if *envFlag || *envLong {
		inspectorType = "env"
	} else if *buffersFlag || *buffersLong {
		inspectorType = "buffers"
	} else if *messagesFlag || *messagesLong || *logsFlag {
		inspectorType = "messages"
	} else if *clientsFlag {
		inspectorType = "clients"
	} else if *statesFlag {
		inspectorType = "states"
	}

	startSearch := *searchFlag || *searchLong

	var cfg *Config
	var flatCommands []FlatCommand

	configPath, err := FindConfigFile(*configFlag)
	if err == nil {
		cfg, err = LoadConfig(configPath)
		if err == nil {
			flatCommands = FlattenCommands(cfg.Items, nil, "")
		}
	}

	model := InitialModel(cfg, flatCommands, startSearch, paneID, inspectorType)
	p := tea.NewProgram(model, tea.WithAltScreen())

	finalModel, err := p.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error running tmux-omni: %v\n", err)
		os.Exit(1)
	}

	// Execute chosen command cleanly after terminal restore
	if app, ok := finalModel.(AppModel); ok && app.PendingExec != nil {
		execErr := RunTmuxTarget(
			app.PendingExec.Action,
			app.PendingExec.Target,
			paneID,
			app.PendingExec.Title,
			app.PendingExec.PersistShell,
			app.PendingExec.OriginalTarget,
		)
		if execErr != nil {
			fmt.Fprintf(os.Stderr, "Execution error: %v\n", execErr)
			os.Exit(1)
		}
	}
}
