package main

import (
	"bufio"
	"fmt"
	"os/exec"
	"regexp"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sahilm/fuzzy"
)

// InspectItem represents a generic item in any inspector view.
type InspectItem struct {
	Col1        string // e.g. Key combo, Command name, Option name, Env var, Client name
	Col2        string // e.g. Category, Alias, Scope, Value, TTY
	Col3        string // e.g. Description, Syntax/Usage, Current value, PID
	Col4        string // e.g. Raw command / additional detail
	SearchText  string
	RawCopy     string
	ActionCmd   string
	IsToggle    bool
	ToggleScope string // "global" or "window"
}

// InspectModel is a unified, searchable table model for all tmux list inspectors.
type InspectModel struct {
	Title        string
	Icon         string
	Col1Header   string
	Col2Header   string
	Col3Header   string
	Col4Header   string
	Items        []InspectItem
	Filtered     []InspectItem
	CursorIndex  int
	ScrollTop    int
	TextInput    textinput.Model
	Width        int
	Height       int
	Col1Width    int
	Col2Width    int
	Col3Width    int
	StatusMsg    string
	PendingExec  string
	NeedsRefresh bool
}

func NewInspectModel(title, icon, col1H, col2H, col3H, col4H string, items []InspectItem, col1W, col2W, col3W int) InspectModel {
	ti := textinput.New()
	ti.Placeholder = fmt.Sprintf("Search %s...", strings.ToLower(title))
	ti.Focus()
	ti.Prompt = "   "
	ti.PromptStyle = PromptIconStyle
	ti.TextStyle = lipgloss.NewStyle().Foreground(ColorFgEmphasis)
	ti.PlaceholderStyle = lipgloss.NewStyle().Foreground(ColorFgMuted)
	ti.Cursor.Style = lipgloss.NewStyle().Foreground(ColorAccentBlue)

	m := InspectModel{
		Title:       title,
		Icon:        icon,
		Col1Header:  col1H,
		Col2Header:  col2H,
		Col3Header:  col3H,
		Col4Header:  col4H,
		Items:       items,
		Filtered:    items,
		TextInput:   ti,
		Col1Width:   col1W,
		Col2Width:   col2W,
		Col3Width:   col3W,
		CursorIndex: 0,
		ScrollTop:   0,
	}
	return m
}

func (m *InspectModel) Filter() {
	query := strings.TrimSpace(m.TextInput.Value())
	if query == "" {
		m.Filtered = m.Items
		m.CursorIndex = 0
		m.ScrollTop = 0
		return
	}

	lower := strings.ToLower(query)
	words := strings.Fields(lower)

	var candidates []InspectItem
	var texts []string
	for _, item := range m.Items {
		matchAll := true
		for _, w := range words {
			if !strings.Contains(item.SearchText, w) {
				matchAll = false
				break
			}
		}
		if matchAll {
			candidates = append(candidates, item)
			texts = append(texts, item.SearchText)
		}
	}

	if len(candidates) > 1 && query != "" {
		matches := fuzzy.Find(lower, texts)
		if len(matches) > 0 {
			var ranked []InspectItem
			seen := make(map[int]bool)
			for _, match := range matches {
				ranked = append(ranked, candidates[match.Index])
				seen[match.Index] = true
			}
			for idx, item := range candidates {
				if !seen[idx] {
					ranked = append(ranked, item)
				}
			}
			m.Filtered = ranked
		} else {
			m.Filtered = candidates
		}
	} else {
		m.Filtered = candidates
	}

	if m.CursorIndex >= len(m.Filtered) {
		m.CursorIndex = 0
	}
	m.ScrollTop = 0
}

func (m *InspectModel) CursorDown(maxVisible int) {
	if len(m.Filtered) == 0 {
		return
	}
	m.CursorIndex++
	if m.CursorIndex >= len(m.Filtered) {
		m.CursorIndex = 0
		m.ScrollTop = 0
		return
	}
	if maxVisible > 0 && m.CursorIndex >= m.ScrollTop+maxVisible {
		m.ScrollTop = m.CursorIndex - maxVisible + 1
	}
}

func (m *InspectModel) CursorUp(maxVisible int) {
	if len(m.Filtered) == 0 {
		return
	}
	m.CursorIndex--
	if m.CursorIndex < 0 {
		m.CursorIndex = len(m.Filtered) - 1
		if maxVisible > 0 && len(m.Filtered) > maxVisible {
			m.ScrollTop = len(m.Filtered) - maxVisible
		} else {
			m.ScrollTop = 0
		}
		return
	}
	if m.CursorIndex < m.ScrollTop {
		m.ScrollTop = m.CursorIndex
	}
}

func (m *InspectModel) SelectedItem() *InspectItem {
	if len(m.Filtered) == 0 || m.CursorIndex < 0 || m.CursorIndex >= len(m.Filtered) {
		return nil
	}
	item := m.Filtered[m.CursorIndex]
	return &item
}

func (m InspectModel) RenderRow(item InspectItem, isSelected bool, width int) string {
	var indicator string
	if isSelected {
		indicator = PaletteIndicatorSelected.Render("▎ ")
	} else {
		indicator = "  "
	}

	fixedWidth := 2 + m.Col1Width + 1 + m.Col2Width + 1
	if m.Col3Width > 0 {
		fixedWidth += m.Col3Width + 1
	}
	availForLast := max(width-fixedWidth-4, 10)

	// Col 1
	c1Text := item.Col1
	if lipgloss.Width(c1Text) > m.Col1Width {
		c1Text = truncateWithEllipsis(c1Text, m.Col1Width)
	}
	var c1Block string
	if isSelected {
		c1Block = lipgloss.NewStyle().
			Foreground(ColorAccentOrange).
			Bold(true).
			Width(m.Col1Width).
			Align(lipgloss.Left).
			Render(c1Text)
	} else {
		c1Block = lipgloss.NewStyle().
			Foreground(ColorAccentOrange).
			Width(m.Col1Width).
			Align(lipgloss.Left).
			Render(c1Text)
	}

	// Col 2
	c2Text := item.Col2
	if lipgloss.Width(c2Text) > m.Col2Width {
		c2Text = truncateWithEllipsis(c2Text, m.Col2Width)
	}
	var c2Block string
	if isSelected {
		c2Block = lipgloss.NewStyle().
			Foreground(ColorAccentPurple).
			Bold(true).
			Width(m.Col2Width).
			Align(lipgloss.Left).
			Render(c2Text)
	} else {
		c2Block = lipgloss.NewStyle().
			Foreground(ColorAccentPurple).
			Width(m.Col2Width).
			Align(lipgloss.Left).
			Render(c2Text)
	}

	var rowContent string

	if m.Col3Width > 0 {
		// Col 3
		c3Text := item.Col3
		if lipgloss.Width(c3Text) > m.Col3Width {
			c3Text = truncateWithEllipsis(c3Text, m.Col3Width)
		}
		var c3Block string
		if isSelected {
			c3Block = lipgloss.NewStyle().
				Foreground(ColorFgEmphasis).
				Bold(true).
				Width(m.Col3Width).
				Align(lipgloss.Left).
				Render(c3Text)
		} else {
			c3Block = lipgloss.NewStyle().
				Foreground(ColorFgDefault).
				Width(m.Col3Width).
				Align(lipgloss.Left).
				Render(c3Text)
		}

		// Col 4 (Remaining)
		c4Text := item.Col4
		if lipgloss.Width(c4Text) > availForLast {
			c4Text = truncateWithEllipsis(c4Text, availForLast)
		}
		var c4Block string
		if isSelected {
			c4Block = lipgloss.NewStyle().
				Foreground(ColorFgSubtle).
				Width(availForLast).
				Align(lipgloss.Left).
				Render(c4Text)
		} else {
			c4Block = lipgloss.NewStyle().
				Foreground(ColorFgMuted).
				Width(availForLast).
				Align(lipgloss.Left).
				Render(c4Text)
		}

		rowContent = lipgloss.JoinHorizontal(
			lipgloss.Center,
			indicator,
			c1Block, " ",
			c2Block, " ",
			c3Block, " ",
			c4Block,
		)
	} else {
		// 3-column format: Col 1, Col 2, and Col 3 takes all remaining space
		c3Text := item.Col3
		if lipgloss.Width(c3Text) > availForLast {
			c3Text = truncateWithEllipsis(c3Text, availForLast)
		}
		var c3Block string
		if isSelected {
			c3Block = lipgloss.NewStyle().
				Foreground(ColorFgEmphasis).
				Width(availForLast).
				Align(lipgloss.Left).
				Render(c3Text)
		} else {
			c3Block = lipgloss.NewStyle().
				Foreground(ColorFgDefault).
				Width(availForLast).
				Align(lipgloss.Left).
				Render(c3Text)
		}

		rowContent = lipgloss.JoinHorizontal(
			lipgloss.Center,
			indicator,
			c1Block, " ",
			c2Block, " ",
			c3Block,
		)
	}

	if isSelected {
		return PaletteRowSelectedStyle.Width(width).Render(rowContent)
	}
	return PaletteRowNormalStyle.Width(width).Render(rowContent)
}

func (m InspectModel) View() string {
	width := m.Width
	height := m.Height
	if width == 0 {
		width = 80
	}
	if height == 0 {
		height = 24
	}

	// Top Bar: Title + Counter
	titleLabel := HeaderStyle.Render(fmt.Sprintf("%s %s", m.Icon, m.Title))
	counterStr := fmt.Sprintf("%d/%d", len(m.Filtered), len(m.Items))
	counter := HeaderCounterStyle.Render(counterStr)

	availForInput := max(width-lipgloss.Width(titleLabel)-lipgloss.Width(counter)-8, 10)
	m.TextInput.Width = availForInput

	inputView := m.TextInput.View()
	topLine := lipgloss.JoinHorizontal(lipgloss.Center, titleLabel, "  ", inputView, "  ", counter)
	topBar := lipgloss.NewStyle().Padding(0, 1).Render(topLine)
	divider := DividerStyle.Render(strings.Repeat("─", width))

	// Footer
	var legendText string
	if m.StatusMsg != "" {
		legendText = lipgloss.NewStyle().Foreground(ColorAccentGreen).Bold(true).Render(m.StatusMsg)
	} else if m.Title == "Paste Buffers" {
		legendText = lipgloss.JoinHorizontal(
			lipgloss.Left,
			FooterKeyStyle.Render("<CR>"), " ", FooterDescStyle.Render("Paste"), "   ",
			FooterKeyStyle.Render("<y/c>"), " ", FooterDescStyle.Render("Copy"), "   ",
			FooterKeyStyle.Render("<d/x>"), " ", FooterDescStyle.Render("Delete"), "   ",
			FooterKeyStyle.Render("<Tab/↑/↓>"), " ", FooterDescStyle.Render("Move"), "   ",
			FooterKeyStyle.Render("<Esc/q>"), " ", FooterDescStyle.Render("Close"),
		)
	} else if m.Title == "Keybindings" {
		legendText = lipgloss.JoinHorizontal(
			lipgloss.Left,
			FooterKeyStyle.Render("<CR>"), " ", FooterDescStyle.Render("Execute"), "   ",
			FooterKeyStyle.Render("<y/c>"), " ", FooterDescStyle.Render("Copy Command"), "   ",
			FooterKeyStyle.Render("<Tab/↑/↓>"), " ", FooterDescStyle.Render("Move"), "   ",
			FooterKeyStyle.Render("<Esc/q>"), " ", FooterDescStyle.Render("Close"),
		)
	} else if m.Title == "Tmux Options" {
		legendText = lipgloss.JoinHorizontal(
			lipgloss.Left,
			FooterKeyStyle.Render("<CR>"), " ", FooterDescStyle.Render("Toggle/Edit"), "   ",
			FooterKeyStyle.Render("<y/c>"), " ", FooterDescStyle.Render("Copy"), "   ",
			FooterKeyStyle.Render("<Tab/↑/↓>"), " ", FooterDescStyle.Render("Move"), "   ",
			FooterKeyStyle.Render("<Esc/q>"), " ", FooterDescStyle.Render("Close"),
		)
	} else {
		legendText = lipgloss.JoinHorizontal(
			lipgloss.Left,
			FooterKeyStyle.Render("<CR>"), " ", FooterDescStyle.Render("Apply/Run"), "   ",
			FooterKeyStyle.Render("<y/c>"), " ", FooterDescStyle.Render("Copy"), "   ",
			FooterKeyStyle.Render("<Tab/↑/↓>"), " ", FooterDescStyle.Render("Navigate"), "   ",
			FooterKeyStyle.Render("<Esc/q>"), " ", FooterDescStyle.Render("Close"),
		)
	}

	brand := FooterBrandStyle.Render("󰌌 Tokyo Night")
	availLeft := max(width-lipgloss.Width(brand)-4, 10)
	leftBlock := lipgloss.NewStyle().Width(availLeft).Render(legendText)
	footerBar := lipgloss.JoinHorizontal(lipgloss.Center, leftBlock, brand)
	footerLine := lipgloss.NewStyle().Padding(0, 1).Render(footerBar)
	bottomArea := lipgloss.JoinVertical(lipgloss.Left, divider, footerLine)

	headerHeight := lipgloss.Height(topBar) + 1
	footerHeight := lipgloss.Height(bottomArea)
	listHeight := max(height-headerHeight-footerHeight, 1)

	// List items
	var rows []string
	if len(m.Filtered) == 0 {
		rows = append(rows, lipgloss.NewStyle().
			Width(width).
			Height(listHeight).
			Align(lipgloss.Center, lipgloss.Center).
			Foreground(ColorFgMuted).
			Render("No matching items found."))
	} else {
		start := m.ScrollTop
		if start >= len(m.Filtered) {
			start = 0
		}
		end := min(start+listHeight, len(m.Filtered))
		for i := start; i < end; i++ {
			isSelected := (i == m.CursorIndex)
			rows = append(rows, m.RenderRow(m.Filtered[i], isSelected, width))
		}
	}

	listContent := lipgloss.NewStyle().
		Height(listHeight).
		Width(width).
		Render(lipgloss.JoinVertical(lipgloss.Left, rows...))

	return WindowStyle.Render(lipgloss.JoinVertical(lipgloss.Left, topBar, divider, listContent, bottomArea))
}

// Data Loaders for Tmux Introspection

// LoadKeybindingsData queries active tmux keybindings and incorporates notes from tmux list-keys -N.
func LoadKeybindingsData() []InspectItem {
	prefixKey := "C-a"
	if out, err := exec.Command("tmux", "show-option", "-gqv", "prefix").Output(); err == nil {
		p := strings.TrimSpace(string(out))
		if p != "" {
			prefixKey = p
		}
	}

	notes := make(map[string]string)
	noteRegex := regexp.MustCompile(`^(\S+(?:\s+\S+)?)\s{2,}(.+)$`)
	if out, err := exec.Command("tmux", "list-keys", "-N").Output(); err == nil {
		scanner := bufio.NewScanner(strings.NewReader(string(out)))
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				continue
			}
			m := noteRegex.FindStringSubmatch(line)
			if len(m) >= 3 {
				comboStr := strings.TrimSpace(m[1])
				note := strings.TrimSpace(m[2])
				notes[comboStr] = note
				fields := strings.Fields(comboStr)
				if len(fields) > 0 {
					notes[fields[len(fields)-1]] = note
				}
			}
		}
	}

	var items []InspectItem
	seen := make(map[string]bool)

	// 1. Prefix keybindings
	if out, err := exec.Command("tmux", "list-keys", "-T", "prefix").Output(); err == nil {
		scanner := bufio.NewScanner(strings.NewReader(string(out)))
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				continue
			}
			tokens := strings.Fields(line)
			prefixIdx := -1
			for i, t := range tokens {
				if t == "prefix" {
					prefixIdx = i
					break
				}
			}
			if prefixIdx == -1 || prefixIdx+1 >= len(tokens) {
				continue
			}
			key := tokens[prefixIdx+1]
			if key == "-N" {
				if prefixIdx+3 >= len(tokens) {
					continue
				}
				key = tokens[prefixIdx+3]
			}
			cmd := strings.Join(tokens[prefixIdx+2:], " ")
			if seen["prefix:"+key] {
				continue
			}
			seen["prefix:"+key] = true

			combo := fmt.Sprintf("<%s %s>", prefixKey, key)
			desc := notes[key]
			if desc == "" {
				desc = notes[fmt.Sprintf("%s %s", prefixKey, key)]
			}
			if desc == "" {
				desc = cmd
			}
			cat := categorizeKey(key, desc, cmd)

			searchable := strings.ToLower(fmt.Sprintf("%s %s %s %s %s", key, combo, cat, desc, cmd))

			items = append(items, InspectItem{
				Col1:       combo,
				Col2:       cat,
				Col3:       desc,
				Col4:       cmd,
				SearchText: searchable,
				RawCopy:    cmd, // Copy exact command
				ActionCmd:  cmd,
			})
		}
	}

	// 2. Root keybindings with descriptions (like C-p)
	if out, err := exec.Command("tmux", "list-keys", "-T", "root").Output(); err == nil {
		scanner := bufio.NewScanner(strings.NewReader(string(out)))
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				continue
			}
			tokens := strings.Fields(line)
			rootIdx := -1
			for i, t := range tokens {
				if t == "root" {
					rootIdx = i
					break
				}
			}
			if rootIdx == -1 || rootIdx+1 >= len(tokens) {
				continue
			}
			key := tokens[rootIdx+1]
			if key == "-N" {
				if rootIdx+3 >= len(tokens) {
					continue
				}
				key = tokens[rootIdx+3]
			}
			cmd := strings.Join(tokens[rootIdx+2:], " ")
			// Only include root bindings that are not mouse events or default character sends
			if strings.HasPrefix(key, "Mouse") || strings.HasPrefix(cmd, "send-keys") {
				continue
			}
			if seen["root:"+key] {
				continue
			}
			seen["root:"+key] = true

			combo := fmt.Sprintf("<%s>", key)
			desc := notes[key]
			if desc == "" {
				desc = cmd
			}
			cat := categorizeKey(key, desc, cmd)

			searchable := strings.ToLower(fmt.Sprintf("%s %s %s %s %s", key, combo, cat, desc, cmd))

			items = append(items, InspectItem{
				Col1:       combo,
				Col2:       cat,
				Col3:       desc,
				Col4:       cmd,
				SearchText: searchable,
				RawCopy:    cmd, // Copy exact command
				ActionCmd:  cmd,
			})
		}
	}

	// Sort by category then key
	catOrder := map[string]int{"Leader": 0, "Panes": 1, "Windows": 2, "Sessions": 3, "Buffers": 4, "Tools": 5, "System": 6}
	sort.Slice(items, func(i, j int) bool {
		c1 := catOrder[items[i].Col2]
		c2 := catOrder[items[j].Col2]
		if c1 != c2 {
			return c1 < c2
		}
		return items[i].Col1 < items[j].Col1
	})

	return items
}

func categorizeKey(key, desc, cmd string) string {
	d := strings.ToLower(desc)
	c := strings.ToLower(cmd)
	if strings.Contains(d, "leader") || strings.Contains(d, "palette") || strings.Contains(d, "menu") || key == "Space" || key == "P" || key == "?" {
		return "Leader"
	}
	if strings.Contains(d, "pane") || strings.Contains(d, "split") || strings.Contains(d, "layout") || strings.Contains(d, "zoom") || strings.Contains("vszx-|\"%o;q", key) {
		return "Panes"
	}
	if strings.Contains(d, "window") || strings.Contains("cnpw&,.lf", key) {
		return "Windows"
	}
	if strings.Contains(d, "session") || strings.Contains(d, "client") || strings.Contains(d, "detach") || strings.Contains("ds$DL()", key) {
		return "Sessions"
	}
	if strings.Contains(d, "buffer") || strings.Contains(d, "copy") || strings.Contains(d, "paste") || strings.Contains("[]#=", key) {
		return "Buffers"
	}
	if strings.Contains(d, "plugin") || strings.Contains(c, "extrakto") || strings.Contains(c, "snaglord") || strings.Contains(c, "cowboy") || strings.Contains(c, "fzf") || strings.Contains(c, "yazi") || strings.Contains(c, "picker") {
		return "Tools"
	}
	return "System"
}

// LoadCommandsData queries all tmux commands and usage syntax.
func LoadCommandsData() []InspectItem {
	var items []InspectItem
	if out, err := exec.Command("tmux", "list-commands").Output(); err == nil {
		scanner := bufio.NewScanner(strings.NewReader(string(out)))
		cmdRegex := regexp.MustCompile(`^(\S+)(?:\s+\((\S+)\))?\s*(.*)$`)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				continue
			}
			m := cmdRegex.FindStringSubmatch(line)
			if len(m) >= 4 {
				name := m[1]
				alias := m[2]
				syntax := m[3]

				aliasDisplay := ""
				if alias != "" {
					aliasDisplay = fmt.Sprintf("(%s)", alias)
				}

				searchable := strings.ToLower(fmt.Sprintf("%s %s %s", name, alias, syntax))
				rawCopy := fmt.Sprintf("%s %s", name, syntax)

				items = append(items, InspectItem{
					Col1:       name,
					Col2:       aliasDisplay,
					Col3:       syntax,
					SearchText: searchable,
					RawCopy:    rawCopy,
					ActionCmd:  fmt.Sprintf("command-prompt -I '%s '", name),
				})
			}
		}
	}
	return items
}

// LoadOptionsData queries global and window options.
func LoadOptionsData() []InspectItem {
	var items []InspectItem

	// Global options
	if out, err := exec.Command("tmux", "show-options", "-g").Output(); err == nil {
		items = append(items, parseOptionLines(string(out), "Global")...)
	}
	// Window options
	if out, err := exec.Command("tmux", "show-window-options", "-g").Output(); err == nil {
		items = append(items, parseOptionLines(string(out), "Window")...)
	}

	sort.Slice(items, func(i, j int) bool {
		if items[i].Col2 != items[j].Col2 {
			return items[i].Col2 < items[j].Col2
		}
		return items[i].Col1 < items[j].Col1
	})

	return items
}

func parseOptionLines(raw, scope string) []InspectItem {
	var items []InspectItem
	scanner := bufio.NewScanner(strings.NewReader(raw))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, " ", 2)
		name := parts[0]
		val := ""
		if len(parts) == 2 {
			val = parts[1]
		}

		isToggle := (val == "on" || val == "off")
		var toggleScope string
		if scope == "Global" {
			toggleScope = "global"
		} else {
			toggleScope = "window"
		}

		searchable := strings.ToLower(fmt.Sprintf("%s %s %s", name, scope, val))
		scopeFlag := "-g"
		if scope == "Window" {
			scopeFlag = "-w"
		}
		rawCopy := fmt.Sprintf("set %s %s %s", scopeFlag, name, val)

		var actionCmd string
		if isToggle {
			newVal := "off"
			if val == "off" {
				newVal = "on"
			}
			actionCmd = fmt.Sprintf("set %s %s %s ; display-message '%s set to %s'", scopeFlag, name, newVal, name, newVal)
		} else {
			actionCmd = fmt.Sprintf("command-prompt -I 'set %s %s %s'", scopeFlag, name, strings.ReplaceAll(val, "'", "\\'"))
		}

		items = append(items, InspectItem{
			Col1:        name,
			Col2:        scope,
			Col3:        val,
			SearchText:  searchable,
			RawCopy:     rawCopy,
			ActionCmd:   actionCmd,
			IsToggle:    isToggle,
			ToggleScope: toggleScope,
		})
	}
	return items
}

// LoadEnvironmentData queries tmux environment variables.
func LoadEnvironmentData() []InspectItem {
	var items []InspectItem
	if out, err := exec.Command("tmux", "show-environment", "-g").Output(); err == nil {
		scanner := bufio.NewScanner(strings.NewReader(string(out)))
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				continue
			}
			if after, ok := strings.CutPrefix(line, "-"); ok {
				varName := after
				items = append(items, InspectItem{
					Col1:       varName,
					Col2:       "Unset",
					Col3:       "(not set in global env)",
					SearchText: strings.ToLower(varName),
					RawCopy:    fmt.Sprintf("unset %s", varName),
				})
			} else {
				parts := strings.SplitN(line, "=", 2)
				varName := parts[0]
				varVal := ""
				if len(parts) == 2 {
					varVal = parts[1]
				}
				items = append(items, InspectItem{
					Col1:       varName,
					Col2:       "Global",
					Col3:       varVal,
					SearchText: strings.ToLower(fmt.Sprintf("%s %s", varName, varVal)),
					RawCopy:    fmt.Sprintf("export %s=\"%s\"", varName, varVal),
				})
			}
		}
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].Col1 < items[j].Col1
	})
	return items
}

// LoadClientsData queries connected tmux clients.
func LoadClientsData() []InspectItem {
	var items []InspectItem
	format := "#{client_name}\t#{client_tty}\t#{client_width}x#{client_height}\t#{client_session}\t#{client_pid}"
	if out, err := exec.Command("tmux", "list-clients", "-F", format).Output(); err == nil {
		scanner := bufio.NewScanner(strings.NewReader(string(out)))
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				continue
			}
			parts := strings.Split(line, "\t")
			if len(parts) >= 5 {
				name := parts[0]
				tty := parts[1]
				size := parts[2]
				session := parts[3]
				pid := parts[4]

				searchable := strings.ToLower(fmt.Sprintf("%s %s %s %s %s", name, tty, size, session, pid))
				rawCopy := fmt.Sprintf("%s (%s) %s on session %s [PID %s]", name, tty, size, session, pid)

				items = append(items, InspectItem{
					Col1:       name,
					Col2:       session,
					Col3:       fmt.Sprintf("%s (%s)", size, tty),
					Col4:       fmt.Sprintf("PID %s", pid),
					SearchText: searchable,
					RawCopy:    rawCopy,
					ActionCmd:  fmt.Sprintf("switch-client -t '%s'", session),
				})
			}
		}
	}
	return items
}

// LoadBuffersData queries tmux paste buffers.
func LoadBuffersData() []InspectItem {
	var items []InspectItem
	format := "#{buffer_name}\t#{buffer_size}\t#{buffer_sample}"
	if out, err := exec.Command("tmux", "list-buffers", "-F", format).Output(); err == nil {
		scanner := bufio.NewScanner(strings.NewReader(string(out)))
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				continue
			}
			parts := strings.Split(line, "\t")
			if len(parts) >= 3 {
				bufName := parts[0]
				bufSize := fmt.Sprintf("%s bytes", parts[1])
				sample := parts[2]

				// Get full buffer text for copy action
				fullText := sample
				if bufOut, err := exec.Command("tmux", "show-buffer", "-b", bufName).Output(); err == nil {
					fullText = string(bufOut)
				}

				searchable := strings.ToLower(fmt.Sprintf("%s %s %s", bufName, bufSize, sample))

				items = append(items, InspectItem{
					Col1:       bufName,
					Col2:       bufSize,
					Col3:       sample,
					SearchText: searchable,
					RawCopy:    fullText,
					ActionCmd:  fmt.Sprintf("paste-buffer -b '%s'", bufName),
				})
			}
		}
	}
	return items
}

// InspectorAppModel wraps InspectModel into a Bubble Tea interactive app.
type InspectorAppModel struct {
	Model InspectModel
}

func (m InspectorAppModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m InspectorAppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Model.Width = msg.Width
		m.Model.Height = msg.Height
		return m, nil

	case tea.KeyMsg:
		keyStr := msg.String()
		maxVisible := max(m.Model.Height-4, 1)

		switch keyStr {
		case "ctrl+c":
			return m, tea.Quit

		case "esc", "q":
			if m.Model.TextInput.Value() != "" {
				m.Model.TextInput.SetValue("")
				m.Model.Filter()
				return m, nil
			}
			return m, tea.Quit

		case "down", "ctrl+n", "ctrl+j", "tab":
			m.Model.CursorDown(maxVisible)
			return m, nil

		case "up", "ctrl+p", "ctrl+k", "shift+tab":
			m.Model.CursorUp(maxVisible)
			return m, nil

		case "enter":
			item := m.Model.SelectedItem()
			if item != nil && item.ActionCmd != "" {
				m.Model.PendingExec = item.ActionCmd
				return m, tea.Quit
			}
			return m, nil

		case "d", "x":
			if m.Model.Title == "Paste Buffers" {
				item := m.Model.SelectedItem()
				if item != nil && item.Col1 != "" {
					_ = exec.Command("tmux", "delete-buffer", "-b", item.Col1).Run()
					m.Model.Items = LoadBuffersData()
					m.Model.Filter()
					m.Model.StatusMsg = fmt.Sprintf("Deleted buffer: %s", item.Col1)
					return m, nil
				}
			}

		case "y", "c":
			item := m.Model.SelectedItem()
			if item != nil && item.RawCopy != "" {
				CopyToClipboard(item.RawCopy)
				disp := item.RawCopy
				if len(disp) > 40 {
					disp = disp[:37] + "..."
				}
				m.Model.StatusMsg = fmt.Sprintf("Copied to clipboard: %s", disp)
				return m, nil
			}
		}
	}

	// Update text input
	oldVal := m.Model.TextInput.Value()
	var cmd tea.Cmd
	m.Model.TextInput, cmd = m.Model.TextInput.Update(msg)
	if m.Model.TextInput.Value() != oldVal {
		m.Model.Filter()
		m.Model.StatusMsg = ""
	}

	return m, cmd
}

func (m InspectorAppModel) View() string {
	return m.Model.View()
}
