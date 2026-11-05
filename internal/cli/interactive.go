package cli

import (
	"errors"
	"fmt"

	"github.com/charmbracelet/bubbletea"
	"github.com/tiagokriok/cdp/internal/config"
	"github.com/tiagokriok/cdp/internal/ui"
)

type model struct {
	profiles       []config.Profile
	currentProfile string
	cursor         int
	selected       bool
}

func initialModel(profiles []config.Profile, current string) model {
	return model{
		profiles:       profiles,
		currentProfile: current,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.cursor < len(m.profiles)-1 {
				m.cursor++
			}

		case "enter":
			m.selected = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m model) View() string {
	s := ""
	s += ui.HeaderStyle.Render("Select a profile to activate:") + "\n\n"

	for i, profile := range m.profiles {
		cursor := " "
		if m.cursor == i {
			cursor = ui.CurrentSymbol + " "
		}

		name := profile.Name
		if profile.Name == m.currentProfile {
			name = ui.CurrentProfileStyle.Render(name + " (current)")
		}

		row := fmt.Sprintf("%s%s", cursor, name)
		s += row + "\n"
	}

	s += ui.DimStyle.Render("\n(j/k or up/down to navigate, enter to select, q to quit)\n")
	return s
}

// RunInteractiveMenu starts the Bubble Tea TUI to select a profile
func RunInteractiveMenu(noRun bool) error {
	cfg, err := loadConfig()
	if err != nil {
		if errors.Is(err, ErrNotInitialized) {
			return HandleInit()
		}
		return err
	}

	pm := config.NewProfileManager(cfg)
	profiles, err := pm.ListProfiles()
	if err != nil {
		return fmt.Errorf("failed to list profiles: %w", err)
	}

	if len(profiles) == 0 {
		ui.Info("No profiles found. Let's create one.")
		fmt.Println("Run `cdp create <name> [description]` to get started.")
		return nil
	}

	initial := initialModel(profiles, cfg.GetCurrentProfile())
	p := tea.NewProgram(initial)

	m, err := p.Run()
	if err != nil {
		return fmt.Errorf("error running interactive menu: %w", err)
	}

	if finalModel, ok := m.(model); ok && finalModel.selected {
		selectedProfile := finalModel.profiles[finalModel.cursor]
		fmt.Println()
		return HandleSwitch(selectedProfile.Name, []string{}, noRun)
	}

	return nil
}
