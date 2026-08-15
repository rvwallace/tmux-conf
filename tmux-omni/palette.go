package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
	"github.com/sahilm/fuzzy"
)

// PaletteModel represents the fuzzy Command Palette state.
type PaletteModel struct {
	TextInput   textinput.Model
	AllCommands []FlatCommand
	Filtered    []FlatCommand
	CursorIndex int
	ScrollTop   int
	Width       int
	Height      int
}

// NewPaletteModel creates an initialized PaletteModel.
func NewPaletteModel(commands []FlatCommand) PaletteModel {
	ti := textinput.New()
	ti.Placeholder = "Type to search commands, categories, keys..."
	ti.Focus()
	ti.Prompt = "   "
	ti.PromptStyle = PromptIconStyle
	ti.TextStyle = lipgloss.NewStyle().Foreground(ColorFgEmphasis)
	ti.PlaceholderStyle = lipgloss.NewStyle().Foreground(ColorFgMuted)
	ti.Cursor.Style = lipgloss.NewStyle().Foreground(ColorAccentBlue)

	return PaletteModel{
		TextInput:   ti,
		AllCommands: commands,
		Filtered:    commands,
		CursorIndex: 0,
		ScrollTop:   0,
	}
}

// FilterCommands filters commands matching the current search input.
func (m *PaletteModel) FilterCommands() {
	query := strings.TrimSpace(m.TextInput.Value())
	if query == "" {
		m.Filtered = m.AllCommands
		m.CursorIndex = 0
		m.ScrollTop = 0
		return
	}

	lowerQuery := strings.ToLower(query)
	words := strings.Fields(lowerQuery)

	// Step 1: Filter all items containing all query tokens
	var candidates []FlatCommand
	var candidateTexts []string
	for _, cmd := range m.AllCommands {
		matchAll := true
		for _, w := range words {
			if !strings.Contains(cmd.SearchableText, w) {
				matchAll = false
				break
			}
		}
		if matchAll {
			candidates = append(candidates, cmd)
			candidateTexts = append(candidateTexts, cmd.SearchableText)
		}
	}

	// Step 2: Rank candidates with fuzzy search
	if len(candidates) > 1 && query != "" {
		matches := fuzzy.Find(lowerQuery, candidateTexts)
		if len(matches) > 0 {
			var ranked []FlatCommand
			seen := make(map[int]bool)
			for _, match := range matches {
				ranked = append(ranked, candidates[match.Index])
				seen[match.Index] = true
			}
			// Append any token matches not explicitly in fuzzy result
			for idx, cmd := range candidates {
				if !seen[idx] {
					ranked = append(ranked, cmd)
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

// CursorDown moves the selection down with scroll adjustment.
func (m *PaletteModel) CursorDown(maxVisible int) {
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

// CursorUp moves the selection up with scroll adjustment.
func (m *PaletteModel) CursorUp(maxVisible int) {
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

// SelectedCommand returns the currently highlighted command or nil if empty.
func (m *PaletteModel) SelectedCommand() *FlatCommand {
	if len(m.Filtered) == 0 || m.CursorIndex < 0 || m.CursorIndex >= len(m.Filtered) {
		return nil
	}
	cmd := m.Filtered[m.CursorIndex]
	return &cmd
}

// RenderSearchBar renders the top search bar and divider.
func (m PaletteModel) RenderSearchBar(width int) string {
	counterStr := fmt.Sprintf("%d/%d", len(m.Filtered), len(m.AllCommands))
	counter := HeaderCounterStyle.Render(counterStr)

	availForInput := max(width-lipgloss.Width(counter)-4, 10)
	m.TextInput.Width = availForInput

	inputView := m.TextInput.View()
	leftBlock := lipgloss.NewStyle().Width(width - lipgloss.Width(counter) - 2).Render(inputView)

	searchLine := lipgloss.JoinHorizontal(lipgloss.Center, leftBlock, counter)
	topBar := lipgloss.NewStyle().Padding(0, 1).Render(searchLine)

	divider := DividerStyle.Render(strings.Repeat("─", width))
	return lipgloss.JoinVertical(lipgloss.Left, topBar, divider)
}

// RenderRow renders a single item with strict, aligned column widths.
func (m PaletteModel) RenderRow(
	cmd FlatCommand,
	isSelected bool,
	rowWidth int,
	titleWidth int,
	catWidth int,
	keyWidth int,
	descWidth int,
) string {
	var indicator string
	if isSelected {
		indicator = PaletteIndicatorSelected.Render("▎ ")
	} else {
		indicator = "  "
	}

	// 1. Icon (fixed 3 cols: icon + space)
	icon := cmd.Icon
	if icon == "" {
		icon = "󰘳"
	}
	iconBlock := lipgloss.NewStyle().Width(3).Render(WkIconStyle.Render(icon))

	// 2. Title Column (strictly aligned)
	titleText := cmd.Title
	if lipgloss.Width(titleText) > titleWidth {
		titleText = truncateWithEllipsis(titleText, titleWidth)
	}
	var titleBlock string
	if isSelected {
		titleBlock = PaletteTitleSelected.Width(titleWidth).Align(lipgloss.Left).Render(titleText)
	} else {
		titleBlock = PaletteTitleNormal.Width(titleWidth).Align(lipgloss.Left).Render(titleText)
	}

	// 3. Category Column (strictly aligned with subtle breadcrumb arrow)
	catText := cmd.Category
	if catText == "Root" {
		catText = ""
	} else if catText != "" {
		catText = strings.ReplaceAll(catText, " > ", "  ")
	}
	if lipgloss.Width(catText) > catWidth {
		catText = truncateWithEllipsis(catText, catWidth)
	}
	var catBlock string
	if isSelected {
		catBlock = PaletteCategorySelected.Width(catWidth).Align(lipgloss.Left).Render(catText)
	} else {
		catBlock = PaletteCategoryNormal.Width(catWidth).Align(lipgloss.Left).Render(catText)
	}

	// 4. Key Sequence Column (strictly aligned)
	var keyBlock string
	if cmd.KeySeq != "" {
		var badge string
		if isSelected {
			badge = PaletteKeyBadgeSelected.Render(cmd.KeySeq)
		} else {
			badge = PaletteKeyBadgeNormal.Render(cmd.KeySeq)
		}
		keyBlock = lipgloss.NewStyle().Width(keyWidth).Align(lipgloss.Left).Render(badge)
	} else {
		keyBlock = lipgloss.NewStyle().Width(keyWidth).Render("")
	}

	// 5. Description Column (takes remaining space)
	descText := cmd.Description
	if lipgloss.Width(descText) > descWidth {
		descText = truncateWithEllipsis(descText, descWidth)
	}
	var descBlock string
	if isSelected {
		descBlock = PaletteDescSelected.Width(descWidth).Align(lipgloss.Left).Render(descText)
	} else {
		descBlock = PaletteDescNormal.Width(descWidth).Align(lipgloss.Left).Render(descText)
	}

	rowContent := lipgloss.JoinHorizontal(
		lipgloss.Center,
		indicator,
		iconBlock,
		titleBlock,
		" ",
		catBlock,
		" ",
		keyBlock,
		" ",
		descBlock,
	)

	if isSelected {
		return PaletteRowSelectedStyle.Width(rowWidth).Render(rowContent)
	}
	return PaletteRowNormalStyle.Width(rowWidth).Render(rowContent)
}

// RenderList renders the scrollable list of matching commands with tabular columns.
func (m PaletteModel) RenderList(width, height int) string {
	if len(m.Filtered) == 0 {
		return lipgloss.NewStyle().
			Width(width).
			Height(height).
			Align(lipgloss.Center, lipgloss.Center).
			Foreground(ColorFgMuted).
			Render("No matching commands found.")
	}

	rowHeight := 1
	maxVisible := max(height/rowHeight, 1)

	start := m.ScrollTop
	if start >= len(m.Filtered) {
		start = 0
	}
	end := min(start+maxVisible, len(m.Filtered))

	// Calculate deterministic column widths based on available width
	// Fixed elements: Indicator (2) + Icon (3) + 3 single-space gutters (3) + margins (2) = 10
	fixedOverhead := 10
	avail := max(width-fixedOverhead, 30)

	// Title column: 30% of available (min 22, max 35)
	titleWidth := max((avail*30)/100, 22)
	if titleWidth > 35 {
		titleWidth = 35
	}

	// Category column: 22% of available (min 16, max 28)
	catWidth := max((avail*22)/100, 16)
	if catWidth > 28 {
		catWidth = 28
	}

	// Key sequence badge column: fixed 8 chars
	keyWidth := 8

	// Description takes the remaining width
	descWidth := max(avail-titleWidth-catWidth-keyWidth, 10)

	var rows []string
	for i := start; i < end; i++ {
		isSelected := (i == m.CursorIndex)
		rows = append(rows, m.RenderRow(
			m.Filtered[i],
			isSelected,
			width,
			titleWidth,
			catWidth,
			keyWidth,
			descWidth,
		))
	}

	listContent := lipgloss.JoinVertical(lipgloss.Left, rows...)
	return lipgloss.NewStyle().
		Height(height).
		Width(width).
		Render(listContent)
}

// RenderFooter renders the bottom modifier bar.
func (m PaletteModel) RenderFooter(width int) string {
	divider := DividerStyle.Render(strings.Repeat("─", width))

	modifiers := lipgloss.JoinHorizontal(
		lipgloss.Left,
		FooterKeyStyle.Render("<CR>"), " ", FooterDescStyle.Render("Run"), "   ",
		FooterKeyStyle.Render("<C-v>"), " ", FooterDescStyle.Render("V-Split"), "   ",
		FooterKeyStyle.Render("<C-s>"), " ", FooterDescStyle.Render("H-Split"), "   ",
		FooterKeyStyle.Render("<C-w>"), " ", FooterDescStyle.Render("Win"), "   ",
		FooterKeyStyle.Render("<C-i>"), " ", FooterDescStyle.Render("Send"), "   ",
		FooterKeyStyle.Render("<Esc>"), " ", FooterDescStyle.Render("Leader"),
	)

	brand := FooterBrandStyle.Render("󰍉 Command Palette")

	availLeft := max(width-lipgloss.Width(brand)-4, 10)
	leftBlock := lipgloss.NewStyle().Width(availLeft).Render(modifiers)

	bar := lipgloss.JoinHorizontal(lipgloss.Center, leftBlock, brand)
	footerLine := lipgloss.NewStyle().Padding(0, 1).Render(bar)

	return lipgloss.JoinVertical(lipgloss.Left, divider, footerLine)
}

// View renders the complete Palette interface.
func (m PaletteModel) View() string {
	if m.Width == 0 {
		m.Width = 80
	}
	if m.Height == 0 {
		m.Height = 24
	}

	searchBar := m.RenderSearchBar(m.Width)
	footer := m.RenderFooter(m.Width)

	searchHeight := lipgloss.Height(searchBar)
	footerHeight := lipgloss.Height(footer)
	listHeight := max(m.Height-searchHeight-footerHeight, 1)

	list := m.RenderList(m.Width, listHeight)

	return lipgloss.JoinVertical(lipgloss.Left, searchBar, list, footer)
}
