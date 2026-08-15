package main

import (
	"github.com/charmbracelet/lipgloss"
)

// Tokyo Night Color Tokens
var (
	ColorBgBase       = lipgloss.Color("#1a1b26")
	ColorBgSurface    = lipgloss.Color("#1f2335")
	ColorBgCard       = lipgloss.Color("#24283b")
	ColorBgHighlight  = lipgloss.Color("#283457")
	ColorBgOverlay    = lipgloss.Color("#292e42")
	ColorBorder       = lipgloss.Color("#3b4261")
	ColorBorderMuted  = lipgloss.Color("#292e42")
	ColorFgDefault    = lipgloss.Color("#c0caf5")
	ColorFgEmphasis   = lipgloss.Color("#ffffff")
	ColorFgMuted      = lipgloss.Color("#565f89")
	ColorFgSubtle     = lipgloss.Color("#737aa2")
	ColorAccentBlue   = lipgloss.Color("#7aa2f7")
	ColorAccentPurple = lipgloss.Color("#bb9af7")
	ColorAccentCyan   = lipgloss.Color("#7dcfff")
	ColorAccentGreen  = lipgloss.Color("#9ece6a")
	ColorAccentOrange = lipgloss.Color("#ff9e64")
	ColorAccentYellow = lipgloss.Color("#e0af68")
	ColorStatusError  = lipgloss.Color("#f7768e")
)

// LazyVim / NvChad inspired UI styles
var (
	// Outer window frame
	WindowStyle = lipgloss.NewStyle().
			Background(ColorBgBase).
			Foreground(ColorFgDefault)

	// Top title / breadcrumbs bar
	HeaderStyle = lipgloss.NewStyle().
			Foreground(ColorAccentBlue).
			Bold(true).
			Padding(0, 1)

	HeaderLeaderIcon = lipgloss.NewStyle().
				Foreground(ColorAccentPurple).
				Bold(true)

	HeaderSeparator = lipgloss.NewStyle().
			Foreground(ColorFgMuted)

	HeaderCounterStyle = lipgloss.NewStyle().
				Foreground(ColorAccentCyan).
				Bold(true)

	// Which-Key column entry styles
	WkKeyStyle = lipgloss.NewStyle().
			Foreground(ColorAccentOrange).
			Bold(true)

	WkIconStyle = lipgloss.NewStyle().
			Foreground(ColorAccentCyan)

	WkLeafTitleStyle = lipgloss.NewStyle().
				Foreground(ColorFgDefault)

	WkGroupTitleStyle = lipgloss.NewStyle().
				Foreground(ColorAccentPurple).
				Bold(true)

	WkGroupPlusStyle = lipgloss.NewStyle().
				Foreground(ColorAccentPurple).
				Bold(true)

	WkDescStyle = lipgloss.NewStyle().
			Foreground(ColorFgMuted)

	// Palette / Search styles
	PromptIconStyle = lipgloss.NewStyle().
			Foreground(ColorAccentCyan).
			Bold(true)

	PaletteRowNormalStyle = lipgloss.NewStyle().
				Background(ColorBgBase).
				Padding(0, 1)

	PaletteRowSelectedStyle = lipgloss.NewStyle().
				Background(ColorBgHighlight).
				Padding(0, 1)

	PaletteIndicatorSelected = lipgloss.NewStyle().
					Foreground(ColorAccentBlue).
					Bold(true)

	PaletteIndicatorNormal = lipgloss.NewStyle().
				Foreground(lipgloss.Color("transparent"))

	PaletteTitleNormal = lipgloss.NewStyle().
				Foreground(ColorFgDefault)

	PaletteTitleSelected = lipgloss.NewStyle().
				Foreground(ColorFgEmphasis).
				Bold(true)

	PaletteCategoryNormal = lipgloss.NewStyle().
				Foreground(ColorAccentPurple)

	PaletteCategorySelected = lipgloss.NewStyle().
				Foreground(ColorAccentPurple).
				Bold(true)

	PaletteKeyBadgeNormal = lipgloss.NewStyle().
				Foreground(ColorAccentOrange).
				Background(ColorBgSurface).
				Bold(true).
				Padding(0, 1)

	PaletteKeyBadgeSelected = lipgloss.NewStyle().
				Foreground(ColorBgBase).
				Background(ColorAccentBlue).
				Bold(true).
				Padding(0, 1)

	PaletteDescNormal = lipgloss.NewStyle().
				Foreground(ColorFgMuted)

	PaletteDescSelected = lipgloss.NewStyle().
				Foreground(ColorFgSubtle)

	// Footer bar
	FooterStyle = lipgloss.NewStyle().
			Foreground(ColorFgMuted).
			Padding(0, 1)

	FooterKeyStyle = lipgloss.NewStyle().
			Foreground(ColorAccentBlue).
			Bold(true)

	FooterDescStyle = lipgloss.NewStyle().
			Foreground(ColorFgMuted)

	FooterBrandStyle = lipgloss.NewStyle().
				Foreground(ColorAccentGreen).
				Bold(true)

	// Divider line
	DividerStyle = lipgloss.NewStyle().
			Foreground(ColorBorder)
)
