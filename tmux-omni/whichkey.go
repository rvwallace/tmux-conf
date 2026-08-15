package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// NavFrame represents one level in the Which-Key navigation stack.
type NavFrame struct {
	Title       string
	Icon        string
	Items       []MenuItem
	CursorIndex int
}

// WhichKeyModel represents the state of the Which-Key menu.
type WhichKeyModel struct {
	NavStack []NavFrame
	Width    int
	Height   int
}

// NewWhichKeyModel creates an initialized WhichKeyModel.
func NewWhichKeyModel(rootTitle string, rootItems []MenuItem) WhichKeyModel {
	if rootTitle == "" {
		rootTitle = "Leader"
	}
	return WhichKeyModel{
		NavStack: []NavFrame{
			{
				Title:       rootTitle,
				Icon:        "󰍡",
				Items:       rootItems,
				CursorIndex: 0,
			},
		},
	}
}

// CurrentFrame returns the top of the navigation stack.
func (m *WhichKeyModel) CurrentFrame() NavFrame {
	if len(m.NavStack) == 0 {
		return NavFrame{}
	}
	return m.NavStack[len(m.NavStack)-1]
}

// SelectedItem returns the currently highlighted item in the active frame.
func (m *WhichKeyModel) SelectedItem() *MenuItem {
	frame := m.CurrentFrame()
	if frame.CursorIndex >= 0 && frame.CursorIndex < len(frame.Items) {
		return &frame.Items[frame.CursorIndex]
	}
	return nil
}

// CursorDown moves selection to the next item.
func (m *WhichKeyModel) CursorDown() {
	frame := m.CurrentFrame()
	if len(frame.Items) == 0 {
		return
	}
	idx := frame.CursorIndex + 1
	if idx >= len(frame.Items) {
		idx = 0
	}
	m.NavStack[len(m.NavStack)-1].CursorIndex = idx
}

// CursorUp moves selection to the previous item.
func (m *WhichKeyModel) CursorUp() {
	frame := m.CurrentFrame()
	if len(frame.Items) == 0 {
		return
	}
	idx := frame.CursorIndex - 1
	if idx < 0 {
		idx = len(frame.Items) - 1
	}
	m.NavStack[len(m.NavStack)-1].CursorIndex = idx
}

// RenderBreadcrumbs renders the hierarchical breadcrumb header trail.
func (m WhichKeyModel) RenderBreadcrumbs() string {
	var parts []string
	for i, frame := range m.NavStack {
		icon := frame.Icon
		if icon == "" {
			icon = "󰍡"
		}
		if i == 0 {
			parts = append(parts, fmt.Sprintf("%s %s", HeaderLeaderIcon.Render(icon), HeaderStyle.Render(frame.Title)))
		} else {
			parts = append(parts, fmt.Sprintf("%s %s", WkIconStyle.Render(icon), HeaderStyle.Render(frame.Title)))
		}
	}
	sep := HeaderSeparator.Render("  ")
	return strings.Join(parts, sep)
}

// RenderHeader renders the top title bar and divider line.
func (m WhichKeyModel) RenderHeader(width int) string {
	trail := m.RenderBreadcrumbs()
	headerLine := lipgloss.NewStyle().Padding(0, 1).Render(trail)

	divider := DividerStyle.Render(strings.Repeat("─", width))
	return lipgloss.JoinVertical(lipgloss.Left, headerLine, divider)
}

// RenderFooter renders the bottom shortcut hint and status bar.
func (m WhichKeyModel) RenderFooter(width int) string {
	divider := DividerStyle.Render(strings.Repeat("─", width))

	hints := lipgloss.JoinHorizontal(
		lipgloss.Left,
		FooterKeyStyle.Render("[/]"), " ", FooterDescStyle.Render("search"), "   ",
		FooterKeyStyle.Render("[<cr>]"), " ", FooterDescStyle.Render("run"), "   ",
		FooterKeyStyle.Render("[<c-i>]"), " ", FooterDescStyle.Render("send"), "   ",
		FooterKeyStyle.Render("[<c-w>]"), " ", FooterDescStyle.Render("win"), "   ",
		FooterKeyStyle.Render("[y]"), " ", FooterDescStyle.Render("copy"), "   ",
		FooterKeyStyle.Render("[<esc>]"), " ", FooterDescStyle.Render("back"),
	)

	brand := FooterBrandStyle.Render("󰌌 Tokyo Night")

	availLeft := max(width-lipgloss.Width(brand)-4, 10)
	leftBlock := lipgloss.NewStyle().Width(availLeft).Render(hints)

	bar := lipgloss.JoinHorizontal(lipgloss.Center, leftBlock, brand)
	footerLine := lipgloss.NewStyle().Padding(0, 1).Render(bar)

	return lipgloss.JoinVertical(lipgloss.Left, divider, footerLine)
}

// RenderItemLine renders a single which-key item in column alignment.
func (m WhichKeyModel) RenderItemLine(item MenuItem, isSelected bool, colWidth int) string {
	key := item.Key
	icon := item.Icon
	if icon == "" {
		icon = "󰘳"
	}
	title := item.Title
	isGroup := len(item.Items) > 0

	// Format key: fixed width (e.g. 3 chars wide)
	keyStr := WkKeyStyle.Render(fmt.Sprintf("%-2s", key))
	iconStr := WkIconStyle.Render(icon)

	// Subgroup vs leaf action styling
	var titleStr string
	maxTitleWidth := max(colWidth-7, 5)

	if isGroup {
		groupLabel := fmt.Sprintf("+%s", title)
		if lipgloss.Width(groupLabel) > maxTitleWidth {
			groupLabel = truncateWithEllipsis(groupLabel, maxTitleWidth)
		}
		if isSelected {
			titleStr = lipgloss.NewStyle().Foreground(ColorAccentPurple).Bold(true).Underline(true).Render(groupLabel)
		} else {
			titleStr = WkGroupTitleStyle.Render(groupLabel)
		}
	} else {
		if lipgloss.Width(title) > maxTitleWidth {
			title = truncateWithEllipsis(title, maxTitleWidth)
		}
		if isSelected {
			titleStr = lipgloss.NewStyle().Foreground(ColorFgEmphasis).Bold(true).Underline(true).Render(title)
		} else {
			titleStr = WkLeafTitleStyle.Render(title)
		}
	}

	content := lipgloss.JoinHorizontal(lipgloss.Left, keyStr, " ", iconStr, " ", titleStr)
	if isSelected {
		return lipgloss.NewStyle().Background(ColorBgHighlight).Width(colWidth).Render(content)
	}
	return lipgloss.NewStyle().Width(colWidth).Render(content)
}

// RenderColumns arranges items into balanced multi-column layout (LazyVim style).
func (m WhichKeyModel) RenderColumns(width, height int) string {
	frame := m.CurrentFrame()
	items := frame.Items
	if len(items) == 0 {
		return lipgloss.NewStyle().
			Width(width).
			Height(height).
			Align(lipgloss.Center, lipgloss.Center).
			Foreground(ColorFgMuted).
			Render("No actions configured in this group.")
	}

	cols := 3
	if width >= 110 {
		cols = 4
	} else if width < 70 {
		cols = 2
	}
	if width < 45 {
		cols = 1
	}

	colGutter := 3
	totalGutter := (cols - 1) * colGutter
	availWidth := max(width-4-totalGutter, cols*15)
	colWidth := availWidth / cols

	// Column-major filling for neat alphabetical / visual scanning
	numItems := len(items)
	rowsPerCol := max((numItems+cols-1)/cols, 1)

	var rowLines []string
	for r := 0; r < rowsPerCol; r++ {
		var rowCols []string
		for c := 0; c < cols; c++ {
			idx := c*rowsPerCol + r
			if idx < numItems {
				isSelected := (idx == frame.CursorIndex)
				rowCols = append(rowCols, m.RenderItemLine(items[idx], isSelected, colWidth))
			} else {
				rowCols = append(rowCols, lipgloss.NewStyle().Width(colWidth).Render(""))
			}
			if c < cols-1 {
				rowCols = append(rowCols, strings.Repeat(" ", colGutter))
			}
		}
		rowLines = append(rowLines, lipgloss.JoinHorizontal(lipgloss.Left, rowCols...))
	}

	gridContent := lipgloss.JoinVertical(lipgloss.Left, rowLines...)
	return lipgloss.NewStyle().
		Padding(1, 2).
		Width(width).
		Render(gridContent)
}

// View renders the complete Which-Key interface.
func (m WhichKeyModel) View() string {
	if m.Width == 0 {
		m.Width = 80
	}
	if m.Height == 0 {
		m.Height = 24
	}

	header := m.RenderHeader(m.Width)
	footer := m.RenderFooter(m.Width)

	headerHeight := lipgloss.Height(header)
	footerHeight := lipgloss.Height(footer)
	contentHeight := max(m.Height-headerHeight-footerHeight, 1)

	body := m.RenderColumns(m.Width, contentHeight)

	contentArea := lipgloss.NewStyle().
		Height(contentHeight).
		Width(m.Width).
		Render(body)

	return lipgloss.JoinVertical(lipgloss.Left, header, contentArea, footer)
}

func truncateWithEllipsis(s string, maxLen int) string {
	if maxLen <= 3 {
		return s
	}
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen-1]) + "…"
}
