package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// NavFrame represents one level in the Which-Key navigation stack.
type NavFrame struct {
	Title string
	Icon  string
	Items []MenuItem
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
				Title: rootTitle,
				Icon:  "󰍡",
				Items: rootItems,
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
		FooterKeyStyle.Render("[<esc>]"), " ", FooterDescStyle.Render("back"), "   ",
		FooterKeyStyle.Render("[q]"), " ", FooterDescStyle.Render("quit"),
	)

	brand := FooterBrandStyle.Render("󰌌 Tokyo Night")

	availLeft := width - lipgloss.Width(brand) - 4
	if availLeft < 10 {
		availLeft = 10
	}
	leftBlock := lipgloss.NewStyle().Width(availLeft).Render(hints)

	bar := lipgloss.JoinHorizontal(lipgloss.Center, leftBlock, brand)
	footerLine := lipgloss.NewStyle().Padding(0, 1).Render(bar)

	return lipgloss.JoinVertical(lipgloss.Left, divider, footerLine)
}

// RenderItemLine renders a single which-key item in column alignment.
func (m WhichKeyModel) RenderItemLine(item MenuItem, colWidth int) string {
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
	maxTitleWidth := colWidth - 7
	if maxTitleWidth < 5 {
		maxTitleWidth = 5
	}

	if isGroup {
		groupLabel := fmt.Sprintf("+%s", title)
		if lipgloss.Width(groupLabel) > maxTitleWidth {
			groupLabel = truncateWithEllipsis(groupLabel, maxTitleWidth)
		}
		titleStr = WkGroupTitleStyle.Render(groupLabel)
	} else {
		if lipgloss.Width(title) > maxTitleWidth {
			title = truncateWithEllipsis(title, maxTitleWidth)
		}
		titleStr = WkLeafTitleStyle.Render(title)
	}

	content := lipgloss.JoinHorizontal(lipgloss.Left, keyStr, " ", iconStr, " ", titleStr)
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
	availWidth := width - 4 - totalGutter
	if availWidth < cols*15 {
		availWidth = cols * 15
	}
	colWidth := availWidth / cols

	// Column-major filling for neat alphabetical / visual scanning
	numItems := len(items)
	rowsPerCol := (numItems + cols - 1) / cols
	if rowsPerCol < 1 {
		rowsPerCol = 1
	}

	var rowLines []string
	for r := 0; r < rowsPerCol; r++ {
		var rowCols []string
		for c := 0; c < cols; c++ {
			idx := c*rowsPerCol + r
			if idx < numItems {
				rowCols = append(rowCols, m.RenderItemLine(items[idx], colWidth))
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
	contentHeight := m.Height - headerHeight - footerHeight
	if contentHeight < 1 {
		contentHeight = 1
	}

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
