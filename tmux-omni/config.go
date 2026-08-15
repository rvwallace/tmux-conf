package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// MenuItem represents an item or sub-group in the menu configuration.
type MenuItem struct {
	Key          string     `json:"key"`
	Title        string     `json:"title"`
	Icon         string     `json:"icon"`
	Description  string     `json:"description"`
	Action       string     `json:"action"`
	Target       string     `json:"target"`
	PopupWidth   string     `json:"popup_width"`
	PopupHeight  string     `json:"popup_height"`
	PersistShell bool       `json:"persist_shell"`
	Shell        bool       `json:"shell"` // alias for persist_shell
	Items        []MenuItem `json:"items"`
}

// Config is the root configuration structure.
type Config struct {
	Title string     `json:"title"`
	Items []MenuItem `json:"items"`
}

// FlatCommand represents a flattened leaf command for the command palette.
type FlatCommand struct {
	Title          string
	Icon           string
	Category       string
	Description    string
	KeySeq         string
	Action         string
	Target         string
	PopupWidth     string
	PopupHeight    string
	PersistShell   bool
	SearchableText string
}

// FindConfigFile resolves the config.json file path following discovery hierarchy.
func FindConfigFile(customPath string) (string, error) {
	if customPath != "" {
		if _, err := os.Stat(customPath); err == nil {
			return customPath, nil
		}
		return "", fmt.Errorf("specified config file not found: %s", customPath)
	}

	homeDir, err := os.UserHomeDir()
	var candidates []string

	if err == nil {
		candidates = append(candidates,
			filepath.Join(homeDir, ".config", "tmux-omni", "config.json"),
			filepath.Join(homeDir, ".config", "tmux-menu", "config.json"),
			filepath.Join(homeDir, ".config", "tmux", "menu.json"),
		)
	}

	// Also check relative to the binary / working dir for repo development
	if exePath, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exePath)
		candidates = append(candidates,
			filepath.Join(exeDir, "..", "tmux-menu", "config.json"),
			filepath.Join(exeDir, "config.json"),
		)
	}

	// Working dir fallback
	candidates = append(candidates,
		filepath.Join(".", "tmux-menu", "config.json"),
		filepath.Join("..", "tmux-menu", "config.json"),
	)

	for _, path := range candidates {
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}

	return "", fmt.Errorf("no config.json found in standard paths")
}

// LoadConfig reads and parses the JSON configuration.
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", path, err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse JSON from %s: %w", path, err)
	}

	if cfg.Title == "" {
		cfg.Title = "Tmux Omni"
	}

	return &cfg, nil
}

// FlattenCommands recursively flattens menu items into a list of searchable leaf commands.
func FlattenCommands(items []MenuItem, breadcrumbs []string, keyPrefix string) []FlatCommand {
	var results []FlatCommand

	for _, item := range items {
		currentKey := item.Key
		var currentSeq string
		if keyPrefix != "" && currentKey != "" {
			currentSeq = keyPrefix + " " + currentKey
		} else if currentKey != "" {
			currentSeq = currentKey
		} else {
			currentSeq = keyPrefix
		}

		icon := item.Icon
		if icon == "" {
			icon = "󰘳"
		}

		if len(item.Items) > 0 {
			nextCrumbs := append(append([]string(nil), breadcrumbs...), item.Title)
			results = append(results, FlattenCommands(item.Items, nextCrumbs, currentSeq)...)
		} else if item.Action != "" {
			category := "Root"
			if len(breadcrumbs) > 0 {
				category = strings.Join(breadcrumbs, " > ")
			}

			target := item.Target
			if target == "" {
				target = "tmux"
			}
			popupW := item.PopupWidth
			if popupW == "" {
				popupW = "80%"
			}
			popupH := item.PopupHeight
			if popupH == "" {
				popupH = "80%"
			}

			searchable := strings.ToLower(fmt.Sprintf("%s %s %s %s", item.Title, category, item.Description, currentSeq))

			results = append(results, FlatCommand{
				Title:          item.Title,
				Icon:           icon,
				Category:       category,
				Description:    item.Description,
				KeySeq:         currentSeq,
				Action:         item.Action,
				Target:         target,
				PopupWidth:     popupW,
				PopupHeight:    popupH,
				PersistShell:   item.PersistShell || item.Shell,
				SearchableText: searchable,
			})
		}
	}

	return results
}
