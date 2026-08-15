package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type AppMode int

const (
	ModeWhichKey AppMode = iota
	ModePalette
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
	Mode         AppMode
	WhichKey     WhichKeyModel
	Palette      PaletteModel
	Config       *Config
	FlatCommands []FlatCommand
	StartSearch  bool
	PaneID       string
	Width        int
	Height       int
	PendingExec  *PendingExecution
}

func InitialModel(cfg *Config, flat []FlatCommand, startSearch bool, paneID string) AppModel {
	wk := NewWhichKeyModel(cfg.Title, cfg.Items)
	pal := NewPaletteModel(flat)

	initialMode := ModeWhichKey
	if startSearch {
		initialMode = ModePalette
	}

	return AppModel{
		Mode:         initialMode,
		WhichKey:     wk,
		Palette:      pal,
		Config:       cfg,
		FlatCommands: flat,
		StartSearch:  startSearch,
		PaneID:       paneID,
	}
}

func (m AppModel) Init() tea.Cmd {
	if m.Mode == ModePalette {
		return textinput.Blink
	}
	return nil
}

func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		m.WhichKey.Width = msg.Width
		m.WhichKey.Height = msg.Height
		m.Palette.Width = msg.Width
		m.Palette.Height = msg.Height
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
		}
	}

	// Update textinput ticks if in palette mode
	if m.Mode == ModePalette {
		var cmd tea.Cmd
		m.Palette.TextInput, cmd = m.Palette.TextInput.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m AppModel) updateWhichKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	keyStr := msg.String()

	switch keyStr {
	case "/", "ctrl+p":
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
					Title: item.Title,
					Icon:  icon,
					Items: item.Items,
				})
				return m, nil
			} else if item.Action != "" {
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
	case "alt+v", "ctrl+v", "√":
		return m.executeSelectedPaletteItem("split_h")
	case "alt+s", "ctrl+s", "ß":
		return m.executeSelectedPaletteItem("split_v")
	case "alt+w", "ctrl+w", "ctrl+t", "∑":
		return m.executeSelectedPaletteItem("window")
	case "alt+i", "ctrl+i", "ctrl+y", "ˆ", "^":
		return m.executeSelectedPaletteItem("send_keys")
	}

	switch keyStr {
	case "ctrl+space", "ctrl+l", "alt+m":
		m.Mode = ModeWhichKey
		return m, nil

	case "down", "ctrl+n", "ctrl+j", "tab":
		m.Palette.CursorDown(maxVisible)
		return m, nil

	case "up", "ctrl+p", "ctrl+k", "shift+tab":
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

	target := targetOverride
	if target == "" {
		target = cmd.Target
	}

	m.PendingExec = &PendingExecution{
		Action:         cmd.Action,
		Target:         target,
		Title:          cmd.Title,
		PersistShell:   cmd.PersistShell,
		OriginalTarget: cmd.Target,
	}

	return m, tea.Quit
}

func (m AppModel) View() string {
	if m.Mode == ModeWhichKey {
		return WindowStyle.Render(m.WhichKey.View())
	}
	return WindowStyle.Render(m.Palette.View())
}

func runInspector(inspModel InspectModel, paneID string) {
	p := tea.NewProgram(InspectorAppModel{Model: inspModel}, tea.WithAltScreen())
	finalModel, err := p.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if app, ok := finalModel.(InspectorAppModel); ok && app.Model.PendingExec != "" {
		execErr := RunTmuxTarget(
			app.Model.PendingExec,
			"tmux",
			paneID,
			app.Model.Title,
			false,
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
	clientsFlag := flag.Bool("clients", false, "Open Connected Clients Inspector")
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
		case "clients":
			*clientsFlag = true
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

	// Mode 1: Keybindings Inspector
	if *keysFlag || *keysLong {
		items := LoadKeybindingsData()
		insp := NewInspectModel("Keybindings", "󰋗", "Key", "Category", "Description", "Command", items, 15, 12, 32)
		runInspector(insp, paneID)
		return
	}

	// Mode 2: Commands Inspector
	if *cmdsFlag || *cmdsLong {
		items := LoadCommandsData()
		insp := NewInspectModel("Tmux Commands", "󰘳", "Command", "Alias", "Syntax / Usage", "", items, 24, 10, 0)
		runInspector(insp, paneID)
		return
	}

	// Mode 3: Options Inspector
	if *optsFlag || *optsLong {
		items := LoadOptionsData()
		insp := NewInspectModel("Tmux Options", "󰘳", "Option Name", "Scope", "Current Value", "", items, 32, 10, 0)
		runInspector(insp, paneID)
		return
	}

	// Mode 4: Environment Inspector
	if *envFlag || *envLong {
		items := LoadEnvironmentData()
		insp := NewInspectModel("Environment", "󰈞", "Variable", "Scope", "Value", "", items, 28, 10, 0)
		runInspector(insp, paneID)
		return
	}

	// Mode 5: Paste Buffers Inspector
	if *buffersFlag || *buffersLong {
		items := LoadBuffersData()
		insp := NewInspectModel("Paste Buffers", "󰅍", "Buffer Name", "Size", "Sample Content", "", items, 16, 12, 0)
		runInspector(insp, paneID)
		return
	}

	// Mode 6: Clients Inspector
	if *clientsFlag {
		items := LoadClientsData()
		insp := NewInspectModel("Connected Clients", "󰒍", "Client", "Session", "Dimensions (TTY)", "PID", items, 18, 14, 22)
		runInspector(insp, paneID)
		return
	}

	// Mode 6: Which-Key Menu & Command Palette (Default)
	startSearch := *searchFlag || *searchLong

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

	flatCommands := FlattenCommands(cfg.Items, nil, "")

	model := InitialModel(cfg, flatCommands, startSearch, paneID)
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
